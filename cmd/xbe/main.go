package main

import (
	"os"

	"github.com/xbe-inc/xbe-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
