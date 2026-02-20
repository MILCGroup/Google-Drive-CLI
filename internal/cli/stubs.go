package cli

import "fmt"

// TODO: remove this file as each domain command migrates to kong.

type CompletionCmd struct{}

func (c *CompletionCmd) Run(globals *Globals) error {
	return fmt.Errorf("completion command not yet migrated to kong")
}
