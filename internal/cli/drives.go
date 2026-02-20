package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/drives"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type DrivesCmd struct {
	List DrivesListCmd `cmd:"" help:"List all Shared Drives"`
	Get  DrivesGetCmd  `cmd:"" help:"Get Shared Drive details"`
}

type DrivesListCmd struct {
	PageSize  int    `help:"Maximum number of drives to return per page" default:"100" name:"page-size"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type DrivesGetCmd struct {
	DriveID string `arg:"" name:"drive-id" help:"Shared Drive ID"`
}

func (cmd *DrivesListCmd) Run(globals *Globals) error {
	ctx := context.Background()
	flags := globals.ToGlobalFlags()

	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Get authenticated client
	client, err := getAPIClient(ctx, flags.Profile)
	if err != nil {
		return handleCLIError(writer, "drives list", err)
	}

	// Create drives manager
	manager := drives.NewManager(client)

	// Create request context
	reqCtx := api.NewRequestContext(flags.Profile, "", types.RequestTypeListOrSearch)

	// If --paginate flag is set, fetch all pages
	if cmd.Paginate {
		var allDrives []*drives.SharedDrive
		pageToken := cmd.PageToken
		for {
			result, err := manager.List(ctx, reqCtx, cmd.PageSize, pageToken)
			if err != nil {
				return handleCLIError(writer, "drives list", err)
			}
			allDrives = append(allDrives, result.Drives...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		return writer.WriteSuccess("drives list", map[string]interface{}{
			"drives": allDrives,
		})
	}

	// List drives
	result, err := manager.List(ctx, reqCtx, cmd.PageSize, cmd.PageToken)
	if err != nil {
		return handleCLIError(writer, "drives list", err)
	}

	// Output result
	return writer.WriteSuccess("drives list", result)
}

func (cmd *DrivesGetCmd) Run(globals *Globals) error {
	ctx := context.Background()
	flags := globals.ToGlobalFlags()

	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Get authenticated client
	client, err := getAPIClient(ctx, flags.Profile)
	if err != nil {
		return handleCLIError(writer, "drives get", err)
	}

	// Create drives manager
	manager := drives.NewManager(client)

	// Create request context
	reqCtx := api.NewRequestContext(flags.Profile, cmd.DriveID, types.RequestTypeGetByID)

	// Get drive
	result, err := manager.Get(ctx, reqCtx, cmd.DriveID, "")
	if err != nil {
		return handleCLIError(writer, "drives get", err)
	}

	// Output result
	return writer.WriteSuccess("drives get", result)
}

// getAPIClient creates an API client for the given profile
func getAPIClient(ctx context.Context, profile string) (*api.Client, error) {
	// Get config directory
	configDir := getConfigDir()

	// Get auth manager
	authMgr := auth.NewManager(configDir)

	// Get valid credentials for profile
	creds, err := authMgr.GetValidCredentials(ctx, profile)
	if err != nil {
		return nil, err
	}

	// Create Drive service
	driveService, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeAuthRequired,
			"Failed to create Drive service: "+err.Error()).Build())
	}

	// Create API client
	return api.NewClient(driveService, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger()), nil
}
