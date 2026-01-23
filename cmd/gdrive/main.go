package main

import (
	"os"

	"github.com/dl-alexandre/gdrive/internal/cli"
	"github.com/dl-alexandre/gdrive/internal/utils"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(utils.ExitUnknown)
	}
}
