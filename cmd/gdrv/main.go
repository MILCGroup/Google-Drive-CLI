package main

import (
	"os"

	"github.com/milcgroup/gdrv/internal/cli"
	"github.com/milcgroup/gdrv/internal/utils"
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
