package main

import (
	"go-svc-gophkeeper/internal/transport/client/CLI/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cobra.RunCLI(version, commit, date)
}
