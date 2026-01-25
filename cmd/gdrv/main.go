package main

import (
	"os"

	"github.com/dl-alexandre/gdrv/internal/cli"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(utils.ExitUnknown)
	}
}
