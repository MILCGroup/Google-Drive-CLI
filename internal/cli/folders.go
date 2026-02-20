package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/folders"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type FoldersCmd struct {
	Create FoldersCreateCmd `cmd:"" help:"Create a folder"`
	List   FoldersListCmd   `cmd:"" help:"List folder contents"`
	Delete FoldersDeleteCmd `cmd:"" help:"Delete a folder"`
	Move   FoldersMoveCmd   `cmd:"" help:"Move a folder"`
	Get    FoldersGetCmd    `cmd:"" help:"Get folder metadata"`
}

type FoldersCreateCmd struct {
	Name   string `arg:"" name:"name" help:"Folder name"`
	Parent string `help:"Parent folder ID" name:"parent"`
}

type FoldersListCmd struct {
	FolderID  string `arg:"" name:"folder-id" help:"Folder ID"`
	PageSize  int    `help:"Number of items per page" default:"100" name:"page-size"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type FoldersDeleteCmd struct {
	FolderID  string `arg:"" name:"folder-id" help:"Folder ID"`
	Recursive bool   `help:"Delete folder contents recursively" name:"recursive"`
}

type FoldersMoveCmd struct {
	FolderID    string `arg:"" name:"folder-id" help:"Folder ID"`
	NewParentID string `arg:"" name:"new-parent-id" help:"New parent folder ID"`
}

type FoldersGetCmd struct {
	FolderID string `arg:"" name:"folder-id" help:"Folder ID"`
	Fields   string `help:"Fields to retrieve (comma-separated)" name:"fields"`
}

func getFolderManager(ctx context.Context, flags types.GlobalFlags) (*folders.Manager, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)
	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, err
	}

	service, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, err
	}

	client := api.NewClient(service, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	return folders.NewManager(client), nil
}

func (cmd *FoldersCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager(ctx, flags)
	if err != nil {
		return writer.WriteError("folder.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	name := cmd.Name

	result, err := mgr.Create(ctx, reqCtx, name, cmd.Parent)
	if err != nil {
		return handleCLIError(writer, "folder.create", err)
	}

	return writer.WriteSuccess("folder.create", result)
}

func (cmd *FoldersListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager(ctx, flags)
	if err != nil {
		return writer.WriteError("folder.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	folderID := cmd.FolderID

	// If --paginate flag is set, fetch all pages
	if cmd.Paginate {
		var allFiles []*types.DriveFile
		pageToken := cmd.PageToken
		for {
			result, err := mgr.List(ctx, reqCtx, folderID, cmd.PageSize, pageToken)
			if err != nil {
				return handleCLIError(writer, "folder.list", err)
			}
			allFiles = append(allFiles, result.Files...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		return writer.WriteSuccess("folder.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(ctx, reqCtx, folderID, cmd.PageSize, cmd.PageToken)
	if err != nil {
		return handleCLIError(writer, "folder.list", err)
	}

	return writer.WriteSuccess("folder.list", result)
}

func (cmd *FoldersDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager(ctx, flags)
	if err != nil {
		return writer.WriteError("folder.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	folderID := cmd.FolderID

	err = mgr.Delete(ctx, reqCtx, folderID, cmd.Recursive)
	if err != nil {
		return handleCLIError(writer, "folder.delete", err)
	}

	return writer.WriteSuccess("folder.delete", map[string]interface{}{
		"deleted":   true,
		"folderId":  folderID,
		"recursive": cmd.Recursive,
	})
}

func (cmd *FoldersMoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager(ctx, flags)
	if err != nil {
		return writer.WriteError("folder.move", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	folderID := cmd.FolderID
	newParentID := cmd.NewParentID

	result, err := mgr.Move(ctx, reqCtx, folderID, newParentID)
	if err != nil {
		return handleCLIError(writer, "folder.move", err)
	}

	return writer.WriteSuccess("folder.move", result)
}

func (cmd *FoldersGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager(ctx, flags)
	if err != nil {
		return writer.WriteError("folder.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeGetByID)
	folderID := cmd.FolderID

	result, err := mgr.Get(ctx, reqCtx, folderID, cmd.Fields)
	if err != nil {
		return handleCLIError(writer, "folder.get", err)
	}

	return writer.WriteSuccess("folder.get", result)
}
