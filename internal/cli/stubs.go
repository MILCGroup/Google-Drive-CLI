package cli

import "fmt"

// TODO: remove this file as each domain command migrates to kong.

type FoldersCmd struct{}

func (c *FoldersCmd) Run(globals *Globals) error {
	return fmt.Errorf("folders command not yet migrated to kong")
}

type DrivesCmd struct{}

func (c *DrivesCmd) Run(globals *Globals) error {
	return fmt.Errorf("drives command not yet migrated to kong")
}

type DocsCmd struct{}

func (c *DocsCmd) Run(globals *Globals) error {
	return fmt.Errorf("docs command not yet migrated to kong")
}

type SlidesCmd struct{}

func (c *SlidesCmd) Run(globals *Globals) error {
	return fmt.Errorf("slides command not yet migrated to kong")
}

type CompletionCmd struct{}

func (c *CompletionCmd) Run(globals *Globals) error {
	return fmt.Errorf("completion command not yet migrated to kong")
}
