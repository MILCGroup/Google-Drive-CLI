package cli

import "fmt"

// TODO: remove this file as each domain command migrates to kong.

type FoldersCmd struct{}

func (c *FoldersCmd) Run(globals *Globals) error {
	return fmt.Errorf("folders command not yet migrated to kong")
}

type PermissionsCmd struct{}

func (c *PermissionsCmd) Run(globals *Globals) error {
	return fmt.Errorf("permissions command not yet migrated to kong")
}

type DrivesCmd struct{}

func (c *DrivesCmd) Run(globals *Globals) error {
	return fmt.Errorf("drives command not yet migrated to kong")
}

type SheetsCmd struct{}

func (c *SheetsCmd) Run(globals *Globals) error {
	return fmt.Errorf("sheets command not yet migrated to kong")
}

type DocsCmd struct{}

func (c *DocsCmd) Run(globals *Globals) error {
	return fmt.Errorf("docs command not yet migrated to kong")
}

type SlidesCmd struct{}

func (c *SlidesCmd) Run(globals *Globals) error {
	return fmt.Errorf("slides command not yet migrated to kong")
}

type AdminCmd struct{}

func (c *AdminCmd) Run(globals *Globals) error {
	return fmt.Errorf("admin command not yet migrated to kong")
}

type ChangesCmd struct{}

func (c *ChangesCmd) Run(globals *Globals) error {
	return fmt.Errorf("changes command not yet migrated to kong")
}

type LabelsCmd struct{}

func (c *LabelsCmd) Run(globals *Globals) error {
	return fmt.Errorf("labels command not yet migrated to kong")
}

type ActivityCmd struct{}

func (c *ActivityCmd) Run(globals *Globals) error {
	return fmt.Errorf("activity command not yet migrated to kong")
}

type ChatCmd struct{}

func (c *ChatCmd) Run(globals *Globals) error {
	return fmt.Errorf("chat command not yet migrated to kong")
}

type SyncCmd struct{}

func (c *SyncCmd) Run(globals *Globals) error {
	return fmt.Errorf("sync command not yet migrated to kong")
}

type ConfigCmd struct{}

func (c *ConfigCmd) Run(globals *Globals) error {
	return fmt.Errorf("config command not yet migrated to kong")
}

type CompletionCmd struct{}

func (c *CompletionCmd) Run(globals *Globals) error {
	return fmt.Errorf("completion command not yet migrated to kong")
}
