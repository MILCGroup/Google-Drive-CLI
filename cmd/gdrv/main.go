package main

import (
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/gdrv/internal/cache"
	"github.com/dl-alexandre/gdrv/internal/cli"
	"github.com/dl-alexandre/gdrv/pkg/version"
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

	// Perform automatic update check in background (non-blocking)
	// This will only run once per day and only notify if an update is available
	updateCache := cache.NewMemoryCache(cache.MemoryCacheOptions{
		DefaultTTL: 24 * time.Hour,
	})
	cli.AutoUpdateCheck(updateCache)

	err := ctx.Run(&c.Globals)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	return 0
}
