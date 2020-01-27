[![CircleCI](https://circleci.com/gh/flowerinthenight/kubepfm/tree/master.svg?style=svg)](https://circleci.com/gh/flowerinthenight/kubepfm/tree/master)

## Overview

`kubepfm` is a simple wrapper to [`kubectl port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) command. It can start multiple `kubectl port-forward` processes based on the number of input targets. Terminating the tool (Ctrl-C) will also terminate all running `kubectl` sub-processes.

## Installation

Using [Homebrew](https://brew.sh/):
```bash
$ brew tap flowerinthenight/tap
$ brew install kubepfm
```

If you have a Go environment:
```bash
$ go get -u -v github.com/flowerinthenight/kubepfm
```

## Usage

```bash
$ kubepfm --target [namespace:]name-or-pattern:local-port:remote-port [--target ...]
```
If the `[namespace:]` part is not specified, the `default` namespace is used.

This tool uses [`regexp.FindAllString`](https://golang.org/pkg/regexp/#Regexp.FindAllString) to resolve the input pattern. If your pattern includes `:` in it (i.e. `[[:alpha:]]`), then you need to include the `namespace` part, as this tool uses the `:` character as its input separator.

```bash
# Simple pattern, namespace not needed
$ kubepfm --target mypod:8080:1222

# Pattern with a `:` in it, namespace is required
$ kubepfm --target "default:mypo[[:alpha:]]:8080:1222"
```

By default, this tool will port-forward to pods. If you want to use deployments or service, you can prefix the name/pattern with the resource type.

```bash
# Using deployment
$ kubepfm --target deployment/dep1:8080:8080

# Using service
$ kubepfm --target service/svc1:8080:8080
```

Finally, the `.*` string is appended to the input `pod-name-or-pattern` before it is resolved.

## Examples

```bash
# Example pods:
$ kubectl get pod
NAME                                 READY     STATUS      RESTARTS   AGE
mypod-7c497c9d94-8xls2               1/1       Running     0          7d
otherpod-5987f84db4-9mhxf            1/1       Running     0          4d
hispod-7d8c4cbd9-dqjc6               1/1       Running     0          21d
herpod-7d48964997-d6pgs              1/1       Running     0          3d

# Do a port-forward to two pods using port 1222 to our local 8080 and 8081 ports:
$ kubepfm --target mypod:8080:1222 --target otherpod:8081:1222
2019/02/19 18:40:05 [info] Your pods:
NAMESPACE     NAME                                 READY     STATUS      RESTARTS   AGE
default       mypod-7c497c9d94-8xls2               1/1       Running     0          7d
default       otherpod-5987f84db4-9mhxf            1/1       Running     0          4d
default       hispod-7d8c4cbd9-dqjc6               1/1       Running     0          21d
default       herpod-7d48964997-d6pgs              1/1       Running     0          3d
...
kube-system   heapster-v1.5.3-6b684ff798-98tn2     3/3       Running     0          27d
kube-system   kube-dns-788979dc8f-287w8            4/4       Running     0          45d
...

# Both pods are now accessible from localhost:
# localhost:8080 -> mypod, localhost:8081 -> otherpod
```
