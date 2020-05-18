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
		Run:   Run,
	}
)

func Run(cmd *cobra.Command, args []string) {
	if len(targets) == 0 {
		failx("need at least one target")
	}

	cs = make(map[string]*exec.Cmd)

	// Range through our input targets.
	for _, c := range targets {
		var args []string
		var name, portPair string
		rctype := "pod"
		t := strings.Split(c, ":")
		portPair = t[len(t)-2] + ":" + t[len(t)-1]

		var context string

		switch len(t) {
		case 3:
			// name:port:port combination
			name = t[0]
			if nn := strings.Split(name, "/"); len(nn) > 1 {
				rctype = nn[0]
			}

			args = []string{
				"get",
				rctype,
				"--no-headers=true",
				"--namespace=default",
				"-o",
				"custom-columns=:metadata.name,:metadata.namespace",
			}
		case 4:
			// namespace:name:port:port combination

			// Rejoin the names excluding namespace and port pair.
			name = strings.Join(t[1:len(t)-2], ":")
			if nn := strings.Split(name, "/"); len(nn) > 1 {
				rctype = nn[0]
			}

			args = []string{
				"get",
				rctype,
				"--no-headers=true",
				fmt.Sprintf("--namespace=%s", t[0]),
				"-o",
				"custom-columns=:metadata.name,:metadata.namespace",
			}
		case 5:
			// context:namespace:name:port:port combination
			context = t[0]

			// Rejoin the names excluding namespace and port pair.
			name = strings.Join(t[2:len(t)-2], ":")
			if nn := strings.Split(name, "/"); len(nn) > 1 {
				rctype = nn[0]
			}

			args = []string{
				"get",
				rctype,
				"--no-headers=true",
				fmt.Sprintf("--context=%s", context),
				fmt.Sprintf("--namespace=%s", t[1]),
				"-o",
				"custom-columns=:metadata.name,:metadata.namespace",
			}
		default:
			// unknown combination
			info("Ignoring unrecognized target definition " + c)
		}

		if rctype == "pod" {
			args = append(args, "--field-selector=status.phase=Running")
		}

		rcs, err := exec.Command("kubectl", args...).CombinedOutput()
		if err != nil {
			fail(err, string(rcs))
			continue
		}

		rows := strings.Split(string(rcs), "\n")
		for _, row := range rows {
			parts := strings.Fields(row)
			if len(parts) != 2 {
				continue
			}

			search := name
			if nn := strings.Split(search, "/"); len(nn) > 1 {
				search = nn[1]
			}

			re := regexp.MustCompile(search + ".*")
			targetList := re.FindAllString(parts[0], -1)
			if len(targetList) > 0 {
				var addcmd *exec.Cmd
				if context == "" {
					addcmd = exec.Command("kubectl", "port-forward", "-n", parts[1], rctype+"/"+targetList[0], portPair)
				} else {
					addcmd = exec.Command("kubectl", "--context", context, "port-forward", "-n", parts[1], rctype+"/"+targetList[0], portPair)
				}
				if _, ok := cs[context+":"+parts[1]+":"+name]; !ok {
					cs[context+":"+parts[1]+":"+name] = addcmd
				}
			}
		}
	}

	if len(cs) > 0 {
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
	}
}

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

	rootCmd.Flags().StringSliceVar(&targets, "target", targets, "fmt: [namespace:]pod-name-pattern:local-port:pod-port")
	rootCmd.Execute()
}
