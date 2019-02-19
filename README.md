## Overview

`kubepfm` is a simple wrapper to [`kubectl port-forward`](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) command. It can start multiple kubectl port-forward processes based on the number of input target pods. At the moment, if you have multiple pods per deployment, it will choose the first one listed from the `kubectl get pod` command. Terminating the tool (Ctrl-C) will also terminate all running kubectl sub-processes.

## Installation

```bash
$ go get -u -v github.com/flowerinthenight/kubepfm
```

## Usage

```bash
$ kubepfm --target pod-name-or-pattern:local-port:pod-port --target ...
```

## Examples

```bash
# Example pods:
$ kubectl get pod
NAME                                 READY     STATUS      RESTARTS   AGE
mypod-7c497c9d94-8xls2               1/1       Running     0          7d
otherpod-5987f84db4-9mhxf            2/2       Running     0          4d
hispod-7d8c4cbd9-dqjc6               2/2       Running     0          21d
herpod-7d48964997-d6pgs              1/1       Running     0          3d

# Do a port-forward to two pods using port 1222 to our local 8080 and 8081 ports:
$ kubepfm --target mypod:8080:1222 --target otherpod:8081:1222

# localhost:8080 -> mypod, localhost:8081 -> otherpod
```
