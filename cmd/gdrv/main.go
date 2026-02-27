package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/milcgroup/gdrv/internal/cli"
	"github.com/milcgroup/gdrv/pkg/version"
)

func main() {
	os.Exit(run())
}

func run() int {
	var c cli.CLI
	ctx := kong.Parse(
		&c,
		kong.Name("gdrv"),
		kong.Description(`gdrv is a command-line tool for interacting with Google Drive.
It supports file operations, folder management, permissions, and more.

All commands support JSON output for automation and scripting.`),
		kong.Vars{"version": version.Version},
		kong.UsageOnError(),
	)

	err := ctx.Run(&c.Globals)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	return 0
}
