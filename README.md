## Overview

`kubepfm` is a simple wrapper to `kubectl port-forward` command. It can start multiple kubectl port-forward processes based on the number of input target pods. At the moment, if you have multiple pods per deployment, it will choose the first one listed from the `kubectl get pod` command.

## Installation

```bash
$ go get -u -v github.com/flowerinthenight/kubepfm
```
