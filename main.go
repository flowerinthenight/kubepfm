package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()

	targets []string
	cs      map[string]*exec.Cmd

	rootCmd = &cobra.Command{
		Use:   "kubepfm",
		Short: "simple port-forward wrapper tool for multiple pods",
		Long:  "Simple port-forward wrapper tool for multiple pods.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(targets) == 0 {
				failx("need at least one target")
			}

			cs = make(map[string]*exec.Cmd)
			podsList, err := exec.Command("kubectl", "get", "pod", "--all-namespaces").CombinedOutput()
			if err != nil {
				failx(err)
			}

			info("Your pods:")
			fmt.Printf(string(podsList))

			for _, c := range targets {
				var args []string
				var name, portPair string
				t := strings.Split(c, ":")
				switch len(t) {
				case 3:
					name = t[0]
					portPair = t[1] + ":" + t[2]
					args = []string{
						"get",
						"pod",
						"--no-headers=true",
						"--namespace=default",
						"-o",
						"custom-columns=:metadata.name,:metadata.namespace",
					}
				case 4:
					name = t[1]
					portPair = t[2] + ":" + t[3]
					args = []string{
						"get",
						"pod",
						"--no-headers=true",
						fmt.Sprintf("--namespace=%s", t[0]),
						"-o",
						"custom-columns=:metadata.name,:metadata.namespace",
					}
				default:
					fail("invalid target", c)
				}

				pods, _ := exec.Command("kubectl", args...).CombinedOutput()
				rows := strings.Split(string(pods), "\n")
				for _, row := range rows {
					parts := strings.Fields(row)
					if len(parts) != 2 {
						continue
					}

					re := regexp.MustCompile(name + ".*")
					targetList := re.FindAllString(parts[0], -1)
					if len(targetList) > 0 {
						addcmd := exec.Command("kubectl", "port-forward", "-n", parts[1], targetList[0], portPair)
						if _, ok := cs[name]; !ok {
							cs[name] = addcmd
						}
					}
				}
			}

			done := make(chan error)

			// Start all cmds.
			for _, c := range cs {
				go func(kcmd *exec.Cmd) {
					outpipe, err := kcmd.StdoutPipe()
					if err != nil {
						failx(err)
					}

					errpipe, err := kcmd.StderrPipe()
					if err != nil {
						failx(err)
					}

					err = kcmd.Start()
					if err != nil {
						failx(err)
					}

					go func() {
						outscan := bufio.NewScanner(outpipe)
						for {
							chk := outscan.Scan()
							if !chk {
								break
							}

							stxt := outscan.Text()
							log.Printf("%v|stdout: %v", green(kcmd.Args), stxt)
						}
					}()

					go func() {
						errscan := bufio.NewScanner(errpipe)
						for {
							chk := errscan.Scan()
							if !chk {
								break
							}

							stxt := errscan.Text()
							log.Printf("%v|stderr: %v", green(kcmd.Args), stxt)
						}
					}()

					kcmd.Wait()
				}(c)
			}

			<-done
		},
	}
)

func info(v ...interface{}) {
	m := fmt.Sprintln(v...)
	log.Printf("%s %s", green("[info]"), m)
}

func fail(v ...interface{}) {
	m := fmt.Sprintln(v...)
	log.Printf("%s %s", red("[error]"), m)
}

func failx(v ...interface{}) {
	fail(v...)
	os.Exit(1)
}

func main() {
	go func() {
		s := make(chan os.Signal)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
		sig := errors.Errorf("%s", <-s)
		_ = sig

		for _, c := range cs {
			err := c.Process.Signal(syscall.SIGTERM)
			if err != nil {
				info("failed to terminate process, force kill...")
				_ = c.Process.Signal(syscall.SIGKILL)
			}
		}

		os.Exit(0)
	}()

	rootCmd.Flags().StringSliceVar(&targets, "target", targets, "fmt: pod-name-pattern:local-port:pod-port")
	rootCmd.Execute()
}
