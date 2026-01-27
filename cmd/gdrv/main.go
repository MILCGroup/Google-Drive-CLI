package main

import (
	"os"

	"github.com/dl-alexandre/gdrv/internal/cli"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := cli.Execute(); err != nil {
		return utils.ExitUnknown
	}
	return 0
}
