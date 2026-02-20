package cli

import (
	"context"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/changes"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type ChangesCmd struct {
	StartPageToken ChangesStartPageTokenCmd `cmd:"start-page-token" help:"Get the starting page token"`
	List           ChangesListCmd           `cmd:"" help:"List changes since a page token"`
	Watch          ChangesWatchCmd          `cmd:"" help:"Watch for changes (webhook setup)"`
	Stop           ChangesStopCmd           `cmd:"" help:"Stop watching for changes"`
}

type ChangesStartPageTokenCmd struct {
}

type ChangesListCmd struct {
	PageToken                 string `help:"Page token to list changes from (required)" name:"page-token" required:""`
	IncludeCorpusRemovals     bool   `help:"Include changes outside target corpus" name:"include-corpus-removals"`
	IncludeItemsFromAllDrives bool   `help:"Include items from all drives" name:"include-items-from-all-drives"`
	IncludePermissionsForView string `help:"Include permissions with published view" name:"include-permissions-for-view"`
	IncludeRemoved            bool   `help:"Include removed items" name:"include-removed"`
	RestrictToMyDrive         bool   `help:"Restrict to My Drive only" name:"restrict-to-my-drive"`
	SupportsAllDrives         bool   `help:"Support all drives" default:"true" name:"supports-all-drives"`
	Limit                     int    `help:"Maximum results per page" default:"100" name:"limit"`
	Fields                    string `help:"Fields to return" name:"fields"`
	Spaces                    string `help:"Comma-separated list of spaces (drive, appDataFolder, photos)" name:"spaces"`
	Paginate                  bool   `help:"Auto-paginate through all changes" name:"paginate"`
}

type ChangesWatchCmd struct {
	PageToken                 string `help:"Page token to watch from (required)" name:"page-token" required:""`
	WebhookURL                string `help:"Webhook URL for notifications (required)" name:"webhook-url" required:""`
	IncludeCorpusRemovals     bool   `help:"Include changes outside target corpus" name:"include-corpus-removals"`
	IncludeItemsFromAllDrives bool   `help:"Include items from all drives" name:"include-items-from-all-drives"`
	IncludePermissionsForView string `help:"Include permissions with published view" name:"include-permissions-for-view"`
	IncludeRemoved            bool   `help:"Include removed items" name:"include-removed"`
	RestrictToMyDrive         bool   `help:"Restrict to My Drive only" name:"restrict-to-my-drive"`
	SupportsAllDrives         bool   `help:"Support all drives" default:"true" name:"supports-all-drives"`
	Spaces                    string `help:"Comma-separated list of spaces (drive, appDataFolder, photos)" name:"spaces"`
	Expiration                int64  `help:"Webhook expiration time (Unix timestamp in milliseconds)" name:"expiration"`
	Token                     string `help:"Arbitrary token for webhook verification" name:"token"`
}

type ChangesStopCmd struct {
	ChannelID  string `arg:"" name:"channel-id" help:"Channel ID from watch command"`
	ResourceID string `arg:"" name:"resource-id" help:"Resource ID from watch command"`
}

func getChangesManager(ctx context.Context, flags types.GlobalFlags) (*changes.Manager, *api.Client, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, out, err
	}

	service, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, out, err
	}

	client := api.NewClient(service, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	mgr := changes.NewManager(client)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	return mgr, client, reqCtx, out, nil
}

func (cmd *ChangesStartPageTokenCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	mgr, _, reqCtx, out, err := getChangesManager(ctx, flags)
	if err != nil {
		return err
	}

	token, err := mgr.GetStartPageToken(ctx, reqCtx, flags.DriveID)
	if err != nil {
		return err
	}

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("changes start-page-token", map[string]string{"startPageToken": token})
	}

	fmt.Println(token)
	return nil
}

func (cmd *ChangesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	mgr, _, reqCtx, out, err := getChangesManager(ctx, flags)
	if err != nil {
		return err
	}

	opts := types.ListOptions{
		PageToken:                 cmd.PageToken,
		DriveID:                   flags.DriveID,
		IncludeCorpusRemovals:     cmd.IncludeCorpusRemovals,
		IncludeItemsFromAllDrives: cmd.IncludeItemsFromAllDrives,
		IncludePermissionsForView: cmd.IncludePermissionsForView,
		IncludeRemoved:            cmd.IncludeRemoved,
		RestrictToMyDrive:         cmd.RestrictToMyDrive,
		SupportsAllDrives:         cmd.SupportsAllDrives,
		Limit:                     cmd.Limit,
		Fields:                    cmd.Fields,
		Spaces:                    cmd.Spaces,
	}

	if cmd.Paginate {
		return paginateChanges(ctx, mgr, reqCtx, out, opts)
	}

	result, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return err
	}

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("changes list", result)
	}

	if len(result.Changes) == 0 {
		fmt.Println("No changes found")
		return nil
	}

	fmt.Printf("Found %d change(s)\n", len(result.Changes))
	for _, change := range result.Changes {
		if change.Removed {
			fmt.Printf("  [REMOVED] %s (File ID: %s) at %s\n", change.ChangeType, change.FileID, change.Time.Format("2006-01-02 15:04:05"))
		} else if change.File != nil {
			fmt.Printf("  [%s] %s (ID: %s) at %s\n", change.ChangeType, change.File.Name, change.FileID, change.Time.Format("2006-01-02 15:04:05"))
		} else if change.Drive != nil {
			fmt.Printf("  [%s] %s (ID: %s) at %s\n", change.ChangeType, change.Drive.Name, change.DriveID, change.Time.Format("2006-01-02 15:04:05"))
		}
	}

	if result.NextPageToken != "" {
		fmt.Printf("\nNext page token: %s\n", result.NextPageToken)
	}

	if result.NewStartPageToken != "" {
		fmt.Printf("New start page token: %s\n", result.NewStartPageToken)
	}

	return nil
}

func paginateChanges(ctx context.Context, mgr *changes.Manager, reqCtx *types.RequestContext, out *OutputWriter, opts types.ListOptions) error {
	allChanges := []types.Change{}
	pageToken := opts.PageToken
	var newStartPageToken string

	for {
		opts.PageToken = pageToken
		result, err := mgr.List(ctx, reqCtx, opts)
		if err != nil {
			return err
		}

		allChanges = append(allChanges, result.Changes...)

		if result.NewStartPageToken != "" {
			newStartPageToken = result.NewStartPageToken
		}

		if result.NextPageToken == "" {
			break
		}

		pageToken = result.NextPageToken
	}

	flags := globalFlags
	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("changes list", map[string]interface{}{
			"changes":           allChanges,
			"newStartPageToken": newStartPageToken,
			"totalChanges":      len(allChanges),
		})
	}

	fmt.Printf("Found %d total change(s)\n", len(allChanges))
	for _, change := range allChanges {
		if change.Removed {
			fmt.Printf("  [REMOVED] %s (File ID: %s) at %s\n", change.ChangeType, change.FileID, change.Time.Format("2006-01-02 15:04:05"))
		} else if change.File != nil {
			fmt.Printf("  [%s] %s (ID: %s) at %s\n", change.ChangeType, change.File.Name, change.FileID, change.Time.Format("2006-01-02 15:04:05"))
		} else if change.Drive != nil {
			fmt.Printf("  [%s] %s (ID: %s) at %s\n", change.ChangeType, change.Drive.Name, change.DriveID, change.Time.Format("2006-01-02 15:04:05"))
		}
	}

	if newStartPageToken != "" {
		fmt.Printf("\nNew start page token: %s\n", newStartPageToken)
	}

	return nil
}

func (cmd *ChangesWatchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	mgr, _, reqCtx, out, err := getChangesManager(ctx, flags)
	if err != nil {
		return err
	}

	opts := types.WatchOptions{
		PageToken:                 cmd.PageToken,
		DriveID:                   flags.DriveID,
		IncludeCorpusRemovals:     cmd.IncludeCorpusRemovals,
		IncludeItemsFromAllDrives: cmd.IncludeItemsFromAllDrives,
		IncludePermissionsForView: cmd.IncludePermissionsForView,
		IncludeRemoved:            cmd.IncludeRemoved,
		RestrictToMyDrive:         cmd.RestrictToMyDrive,
		SupportsAllDrives:         cmd.SupportsAllDrives,
		Spaces:                    cmd.Spaces,
		WebhookURL:                cmd.WebhookURL,
		Expiration:                cmd.Expiration,
		Token:                     cmd.Token,
	}

	channel, err := mgr.Watch(ctx, reqCtx, cmd.PageToken, cmd.WebhookURL, opts)
	if err != nil {
		return err
	}

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("changes watch", channel)
	}

	fmt.Printf("Watching for changes\n")
	fmt.Printf("Channel ID: %s\n", channel.ID)
	fmt.Printf("Resource ID: %s\n", channel.ResourceID)
	if channel.Expiration > 0 {
		fmt.Printf("Expiration: %d\n", channel.Expiration)
	}
	fmt.Printf("\nTo stop watching, run:\n")
	fmt.Printf("  gdrv changes stop %s %s\n", channel.ID, channel.ResourceID)

	return nil
}

func (cmd *ChangesStopCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	mgr, _, reqCtx, out, err := getChangesManager(ctx, flags)
	if err != nil {
		return err
	}

	err = mgr.Stop(ctx, reqCtx, cmd.ChannelID, cmd.ResourceID)
	if err != nil {
		return err
	}

	if flags.OutputFormat == types.OutputFormatJSON {
		return out.WriteSuccess("changes stop", map[string]string{
			"status":     "stopped",
			"channelId":  cmd.ChannelID,
			"resourceId": cmd.ResourceID,
		})
	}

	fmt.Printf("Stopped watching channel %s\n", cmd.ChannelID)
	return nil
}
