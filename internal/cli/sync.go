package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/logging"
	syncengine "github.com/dl-alexandre/gdrv/internal/sync"
	"github.com/dl-alexandre/gdrv/internal/sync/conflict"
	"github.com/dl-alexandre/gdrv/internal/sync/diff"
	"github.com/dl-alexandre/gdrv/internal/sync/index"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"github.com/google/uuid"
)

// SyncCmd is the parent for all sync subcommands.
type SyncCmd struct {
	Init   SyncInitCmd   `cmd:"" help:"Initialize a sync configuration"`
	Push   SyncPushCmd   `cmd:"" help:"Push local changes to Drive"`
	Pull   SyncPullCmd   `cmd:"" help:"Pull remote changes to local"`
	Status SyncStatusCmd `cmd:"" help:"Show pending sync changes"`
	List   SyncListCmd   `cmd:"" help:"List sync configurations"`
	Remove SyncRemoveCmd `cmd:"" help:"Remove a sync configuration"`
}

type SyncInitCmd struct {
	LocalPath    string `arg:"" name:"local-path" help:"Local path to sync"`
	RemoteFolder string `arg:"" name:"remote-folder" help:"Remote folder ID or path"`
	Exclude      string `help:"Comma-separated exclude patterns" name:"exclude"`
	Conflict     string `help:"Conflict policy (local-wins, remote-wins, rename-both)" default:"rename-both" name:"conflict"`
	Direction    string `help:"Sync direction (push, pull, bidirectional)" default:"bidirectional" name:"direction"`
	ID           string `help:"Optional sync configuration ID" name:"id"`
}

type SyncPushCmd struct {
	ConfigID    string `arg:"" name:"config-id" help:"Sync configuration ID"`
	Delete      bool   `help:"Propagate deletions" name:"delete"`
	Conflict    string `help:"Override conflict policy" name:"conflict"`
	Concurrency int    `help:"Concurrent transfers" default:"5" name:"concurrency"`
	UseChanges  bool   `help:"Use Drive Changes API when available" default:"true" name:"use-changes"`
}

type SyncPullCmd struct {
	ConfigID    string `arg:"" name:"config-id" help:"Sync configuration ID"`
	Delete      bool   `help:"Propagate deletions" name:"delete"`
	Conflict    string `help:"Override conflict policy" name:"conflict"`
	Concurrency int    `help:"Concurrent transfers" default:"5" name:"concurrency"`
	UseChanges  bool   `help:"Use Drive Changes API when available" default:"true" name:"use-changes"`
}

type SyncStatusCmd struct {
	ConfigID   string `arg:"" name:"config-id" help:"Sync configuration ID"`
	Delete     bool   `help:"Include deletions in status" name:"delete"`
	Conflict   string `help:"Override conflict policy" name:"conflict"`
	UseChanges bool   `help:"Use Drive Changes API when available" default:"true" name:"use-changes"`
}

type SyncListCmd struct{}

type SyncRemoveCmd struct {
	ConfigID string `arg:"" name:"config-id" help:"Sync configuration ID"`
}

func (cmd *SyncInitCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	localPath := cmd.LocalPath
	remotePath := cmd.RemoteFolder

	stat, err := os.Stat(localPath)
	if err != nil || !stat.IsDir() {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeInvalidArgument, "Local path must be a directory").Build())
	}

	absLocal, err := filepath.Abs(localPath)
	if err != nil {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	_, client, _, _, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	remoteID, err := ResolveFileID(ctx, client, flags, remotePath)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("sync.init", appErr.CLIError)
		}
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	db, err := openSyncDB()
	if err != nil {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			GetLogger().Warn("failed to close sync database", logging.F("error", closeErr))
		}
	}()

	configID := cmd.ID
	if configID == "" {
		configID = uuid.New().String()
	}

	excludes := []string{}
	if cmd.Exclude != "" {
		for _, part := range strings.Split(cmd.Exclude, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				excludes = append(excludes, part)
			}
		}
	}

	cfg := index.SyncConfig{
		ID:              configID,
		LocalRoot:       absLocal,
		RemoteRootID:    remoteID,
		ExcludePatterns: excludes,
		ConflictPolicy:  cmd.Conflict,
		Direction:       cmd.Direction,
	}
	if err := syncengine.EnsureConfig(&cfg); err != nil {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	if err := db.UpsertConfig(ctx, cfg); err != nil {
		return out.WriteError("sync.init", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	return out.WriteSuccess("sync.init", cfg)
}

func (cmd *SyncPushCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	return runSyncWithMode(flags, cmd.ConfigID, diff.ModePush, "sync.push", false, cmd.Delete, cmd.Conflict, cmd.Concurrency, cmd.UseChanges)
}

func (cmd *SyncPullCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	return runSyncWithMode(flags, cmd.ConfigID, diff.ModePull, "sync.pull", false, cmd.Delete, cmd.Conflict, cmd.Concurrency, cmd.UseChanges)
}

func (cmd *SyncStatusCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	return runSyncWithMode(flags, cmd.ConfigID, diff.ModeBidirectional, "sync.status", true, cmd.Delete, cmd.Conflict, 0, cmd.UseChanges)
}

func (cmd *SyncListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	db, err := openSyncDB()
	if err != nil {
		return out.WriteError("sync.list", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			GetLogger().Warn("failed to close sync database", logging.F("error", closeErr))
		}
	}()

	configs, err := db.ListConfigs(context.Background())
	if err != nil {
		return out.WriteError("sync.list", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	return out.WriteSuccess("sync.list", map[string]interface{}{
		"configs": configs,
	})
}

func (cmd *SyncRemoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	db, err := openSyncDB()
	if err != nil {
		return out.WriteError("sync.remove", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			GetLogger().Warn("failed to close sync database", logging.F("error", closeErr))
		}
	}()

	if err := db.DeleteEntries(context.Background(), cmd.ConfigID); err != nil {
		return out.WriteError("sync.remove", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}
	if err := db.DeleteConfig(context.Background(), cmd.ConfigID); err != nil {
		return out.WriteError("sync.remove", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	return out.WriteSuccess("sync.remove", map[string]interface{}{
		"configId": cmd.ConfigID,
	})
}

func runSyncWithMode(flags types.GlobalFlags, configID string, mode diff.Mode, command string, planOnly bool, delete bool, conflictStr string, concurrency int, useChanges bool) error {
	ctx := context.Background()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	engine, reqCtx, cfg, err := loadSyncEngine(ctx, flags, configID)
	if err != nil {
		return out.WriteError(command, utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}
	defer func() {
		if closeErr := engine.Close(); closeErr != nil {
			GetLogger().Warn("failed to close sync engine", logging.F("error", closeErr))
		}
	}()

	opts := syncengine.Options{
		Mode:        mode,
		Delete:      delete,
		DryRun:      flags.DryRun || planOnly,
		Force:       flags.Force,
		Yes:         flags.Yes,
		Concurrency: concurrency,
		UseChanges:  useChanges,
	}
	if conflictStr != "" {
		opts.ConflictPolicy = conflict.Policy(conflictStr)
	}

	plan, err := engine.Plan(ctx, cfg, opts, reqCtx)
	if err != nil {
		return out.WriteError(command, utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	if planOnly {
		return out.WriteSuccess(command, map[string]interface{}{
			"configId":  cfg.ID,
			"actions":   plan.Actions,
			"conflicts": plan.Conflicts,
		})
	}

	if len(plan.Conflicts) > 0 {
		return out.WriteError(command, utils.NewCLIError(utils.ErrCodeUnknown, "Conflicts detected").Build())
	}

	applyCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	result, err := engine.Apply(ctx, cfg, plan, opts, applyCtx)
	if err != nil {
		return out.WriteError(command, utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	return out.WriteSuccess(command, map[string]interface{}{
		"configId": cfg.ID,
		"summary":  result.Summary,
	})
}

func openSyncDB() (*index.DB, error) {
	configDir := getConfigDir()
	dbPath := filepath.Join(configDir, "sync", "index.db")
	return index.Open(dbPath)
}

func loadSyncEngine(ctx context.Context, flags types.GlobalFlags, configID string) (*syncengine.Engine, *types.RequestContext, index.SyncConfig, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, index.SyncConfig{}, err
	}

	service, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, index.SyncConfig{}, err
	}

	client := api.NewClient(service, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	db, err := openSyncDB()
	if err != nil {
		return nil, nil, index.SyncConfig{}, err
	}

	cfg, err := db.GetConfig(ctx, configID)
	if err != nil {
		_ = db.Close()
		return nil, nil, index.SyncConfig{}, err
	}

	engine := syncengine.NewEngine(client, db)
	return engine, reqCtx, *cfg, nil
}
