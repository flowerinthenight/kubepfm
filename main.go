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
		Use:          "kubepfm",
		Short:        "A simple port-forward wrapper tool for multiple pods/deployments/services",
		Long:         "A simple port-forward wrapper tool for multiple pods/deployments/services.",
		RunE:         Run,
		SilenceUsage: true,
	}
)

// Returns the kubectl args, kubectl context name, resource name, and the port pair (i.e. 8080:1222) from the input.
func parse(in string) ([]string, string, string, string) {
	var args []string
	var ctx, name, ports string
	rctype := "pod"
	ns := "default"
	t := strings.Split(in, ":")

	// Type of supported input formats:
	//   --target [namespace:]name:8080:1222
	//   --target ctx=minikube:ns=default:deployment/testdeployment:8080:1222
	switch {
	case len(t) == 3:
		// Simplest form: name:port:port
		name = t[0]
		if nn := strings.Split(name, "/"); len(nn) > 1 {
			rctype = nn[0]
		}

		ports = t[len(t)-2] + ":" + t[len(t)-1]
		args = []string{
			"get",
			rctype,
			"--no-headers=true",
			fmt.Sprintf("--namespace=%s", ns),
			"-o",
			"custom-columns=:metadata.name,:metadata.namespace",
		}
	case (len(t) == 4 && !strings.HasPrefix(in, "ctx=")) || (len(t) == 4 && strings.HasPrefix(in, "ns=")):
		// First check: old optional namespace (namespace:name:port:port)
		// Second check: old optional namespace with optional prefix ns= (ns=namespace:name:port:port)
		ns = t[0]
		if strings.HasPrefix(ns, "ns=") {
			ns = strings.Split(ns, "=")[1]
		}

		name = t[1]
		if nn := strings.Split(name, "/"); len(nn) > 1 {
			rctype = nn[0]
		}

		ports = t[len(t)-2] + ":" + t[len(t)-1]
		args = []string{
			"get",
			rctype,
			"--no-headers=true",
			fmt.Sprintf("--namespace=%s", ns),
			"-o",
			"custom-columns=:metadata.name,:metadata.namespace",
		}
	case len(t) > 4 && strings.HasPrefix(in, "ctx="):
		// With context and optional namespace: [ctx=context:ns=namespace:]name:port:port
		ctx = strings.Split(t[0], "=")[1]
		ns = t[1]
		if strings.HasPrefix(ns, "ns=") {
			ns = strings.Split(ns, "=")[1]
		}

		name = t[2]
		if nn := strings.Split(name, "/"); len(nn) > 1 {
			rctype = nn[0]
		}

		ports = t[len(t)-2] + ":" + t[len(t)-1]
		args = []string{
			"get",
			rctype,
			"--no-headers=true",
			fmt.Sprintf("--context=%s", ctx),
			fmt.Sprintf("--namespace=%s", ns),
			"-o",
			"custom-columns=:metadata.name,:metadata.namespace",
		}
	default:
		// unknown combination
		fail("skip unknown target: " + in)
	}

	return args, ctx, name, ports
}

func Run(cmd *cobra.Command, args []string) error {
	if len(targets) == 0 {
		info("no target inputs, read from stdin")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			targets = append(targets, scanner.Text())
		}

		info(fmt.Sprintf("read %d targets from stdin", len(targets)))
	}

	cs = make(map[string]*exec.Cmd)

	// Range through our input targets.
	for _, c := range targets {
		v, ctx, name, portpair := parse(c)
		if v == nil {
			return fmt.Errorf("invalid target: %v", c)
		}

		rctype := v[1]
		if rctype == "pod" {
			v = append(v, "--field-selector=status.phase=Running")
		}

		rcs, err := exec.Command("kubectl", v...).CombinedOutput()
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
				if ctx == "" {
					addcmd = exec.Command("kubectl", "port-forward", "-n", parts[1], rctype+"/"+targetList[0], portpair)
				} else {
					addcmd = exec.Command("kubectl", "--context", ctx, "port-forward", "-n", parts[1], rctype+"/"+targetList[0], portpair)
				}

				key := fmt.Sprintf("%s:%s:%s:%s", ctx, parts[1], name, portpair)
				if _, ok := cs[key]; !ok {
					cs[key] = addcmd
				}
			}
		}
	}

	if len(cs) == 0 {
		return nil
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
	return nil
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

	rootCmd.Flags().StringSliceVar(&targets, "target", targets, "fmt: [[ctx=context:ns=[namespace:]]pod-name-pattern:local-port:pod-port")
	rootCmd.Execute()
}
