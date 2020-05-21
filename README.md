[![CircleCI](https://circleci.com/gh/flowerinthenight/kubepfm/tree/master.svg?style=svg)](https://circleci.com/gh/flowerinthenight/kubepfm/tree/master)

## Overview

`kubepfm` is a simple wrapper to [`kubectl port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) command for multiple pods/deployments/services. It can start multiple `kubectl port-forward` processes based on the number of input targets. Terminating the tool (Ctrl-C) will also terminate all running `kubectl` sub-processes.

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
$ kubepfm --target [ctx=context:[ns=]namespace:]name-or-pattern:local-port:remote-port [--target ...]
```
If the `[ctx=context]` part is not specified, the current context is used. If the `[ns=namespace:]` part is not specified, the `default` namespace is used.

This tool uses [`regexp.FindAllString`](https://golang.org/pkg/regexp/#Regexp.FindAllString) to resolve the input pattern. Since this tool uses the `:` character as its input separator, it cannot process input patterns that contain `:`, such as `[[:alpha:]]`).

```bash
# Simple pattern, namespace not needed.
$ kubepfm --target mypod:8080:1222

# With namespace input.
$ kubepfm --target default:mypod:8080:1222
$ kubepfm --target ns=default:mypod:8080:1222

# With context. Useful if you need to port-forward to different clusters in one go.
$ kubepfm --target ctx=devcluster:ns=default:mypod:8080:1222 \
          --target ctx=prodcluster:ns=default:mypod:8081:1222
```

You can also pipe input from STDIN as well.
```bash
$ cat input.txt
mypod:8080:1222
otherpod:8081:1222
$ cat input.txt | kubepfm

# Or you can do it this way:
$ kubepfm
mypod:8080:1222
otherpod:8081:1222
Ctrl+D
```

By default, this tool will port-forward to pods. If you want to forward to deployments or services, you can prefix the name/pattern with the resource type.

```bash
# Using deployment
$ kubepfm --target deployment/dep1:8080:8080

# Using services
$ kubepfm --target service/svc1:8080:8080 --target service/svc2:8081:80
```

Finally, the `.*` string is appended to the input name/pattern before it is resolved.

## Examples

View our running pods:
```bash
$ kubectl get pod
NAME                                 READY     STATUS      RESTARTS   AGE
mypod-7c497c9d94-8xls2               1/1       Running     0          7d
otherpod-5987f84db4-9mhxf            1/1       Running     0          4d
hispod-7d8c4cbd9-dqjc6               1/1       Running     0          21d
herpod-7d48964997-d6pgs              1/1       Running     0          3d
...
```

Do a port-forward to two pods using port 1222 to our local 8080 and 8081 ports:
```bash
$ kubepfm --target mypod:8080:1222 --target otherpod:8081:1222
[kubectl port-forward -n default pod/mypod-xxx 8080:1222] Forwarding from 127.0.0.1:8080 -> 1222
[kubectl port-forward -n default pod/mypod-xxx 8080:1222] Forwarding from [::1]:8080 -> 1222
[kubectl port-forward -n default pod/otherpod-xxx 8081:1222] Forwarding from 127.0.0.1:8081 -> 1222
[kubectl port-forward -n default pod/otherpod-xxx 8081:1222] Forwarding from [::1]:8081 -> 1222
```

Both pods are now accessible from localhost:
```bash
localhost:8080 --> mypod
localhost:8081 --> otherpod
```
