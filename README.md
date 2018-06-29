## Overview

`kubepfm` is a simple wrapper to `kubectl port-forward` command. It can start multiple kubectl port-forward processes based on the number of input target pods. At the moment, if you have multiple pods per deployment, it will choose the first one listed from the `kubectl get pod` command.

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
# port-forward two pods using port 1222 to my local 8080 and 8081 ports:
$ kubepfm --target mypod:8080:1222 --target anotherpod:8081:1222
```
