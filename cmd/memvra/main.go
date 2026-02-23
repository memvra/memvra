package main

import "github.com/memvra/memvra/internal/cli"

// version, commit, date are injected by the linker via -ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}
