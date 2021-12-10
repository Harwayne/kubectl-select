# kubectl-select
Simple kubectl extension to select kubectl contexts.

### kubectl plugin

Install as a `kubectl`
[plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) by
placing anywhere on your path.

### Development

#### Building

Build with `go build -ldflags="-s -w" kubectl-select.go`. The
[linker flags](https://pkg.go.dev/cmd/link) remove some debug information from
the binary, making it smaller.
