package main

import (
	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/gdrv/internal/cli"
	"github.com/dl-alexandre/gdrv/pkg/version"
)

func main() {
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
	ctx.FatalIfErrorf(err)
}
