package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/export"
	"github.com/milcgroup/gdrv/internal/files"
	"github.com/milcgroup/gdrv/internal/revisions"
	"github.com/milcgroup/gdrv/internal/safety"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
)

type FilesCmd struct {
	List          FilesListCmd          `cmd:"" help:"List files"`
	Get           FilesGetCmd           `cmd:"" help:"Get file metadata"`
	Upload        FilesUploadCmd        `cmd:"" help:"Upload a file"`
	Download      FilesDownloadCmd      `cmd:"" help:"Download a file"`
	Delete        FilesDeleteCmd        `cmd:"" help:"Delete a file"`
	Copy          FilesCopyCmd          `cmd:"" help:"Copy a file"`
	Move          FilesMoveCmd          `cmd:"" help:"Move a file"`
	Trash         FilesTrashCmd         `cmd:"" help:"Move file to trash"`
	Restore       FilesRestoreCmd       `cmd:"" help:"Restore file from trash"`
	Revisions     FilesRevisionsCmd     `cmd:"" help:"List file revisions"`
	ListTrashed   FilesListTrashedCmd   `cmd:"list-trashed" help:"List trashed files"`
	ExportFormats FilesExportFormatsCmd `cmd:"export-formats" help:"Show available export formats for a file"`
	BatchUpload   FilesBatchUploadCmd   `cmd:"batch-upload" help:"Upload multiple files"`
	BatchDownload FilesBatchDownloadCmd `cmd:"batch-download" help:"Download multiple files"`
	BatchDelete   FilesBatchDeleteCmd   `cmd:"batch-delete" help:"Delete multiple files"`
}

type FilesListCmd struct {
	Parent         string `help:"Parent folder ID" name:"parent"`
	Query          string `help:"Search query" name:"query"`
	Limit          int    `help:"Maximum files to return per page" default:"100" name:"limit"`
	PageToken      string `help:"Page token for pagination" name:"page-token"`
	OrderBy        string `help:"Sort order" name:"order-by"`
	IncludeTrashed bool   `help:"Include trashed files" name:"include-trashed"`
	Fields         string `help:"Fields to return" name:"fields"`
	Paginate       bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type FilesGetCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
	Fields string `help:"Fields to return" name:"fields"`
}

type FilesUploadCmd struct {
	LocalPath string `arg:"" name:"local-path" help:"Local file path"`
	Parent    string `help:"Parent folder ID" name:"parent"`
	Name      string `help:"File name" name:"name"`
	MimeType  string `help:"MIME type" name:"mime-type"`
}

type FilesDownloadCmd struct {
	FileID     string `arg:"" name:"file-id" help:"File ID or path"`
	OutputPath string `help:"Output path" name:"file-output"`
	MimeType   string `help:"Export MIME type" name:"mime-type"`
	Doc        bool   `help:"Export Google Docs as plain text" name:"doc" aliases:"doc-text"`
}

type FilesDeleteCmd struct {
	FileID           string `arg:"" name:"file-id" help:"File ID or path"`
	Permanent        bool   `help:"Permanently delete" name:"permanent"`
	SkipConfirmation bool   `help:"Skip confirmation" name:"skip-confirmation"`
}

type FilesCopyCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
	Name   string `help:"New file name" name:"name"`
	Parent string `help:"Destination folder ID" name:"parent"`
}

type FilesMoveCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
	Parent string `help:"New parent folder ID" name:"parent" required:""`
}

type FilesTrashCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
}

type FilesRestoreCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
}

type FilesRevisionsCmd struct {
	List     FilesRevisionsListCmd     `cmd:"" default:"withargs" help:"List file revisions"`
	Download FilesRevisionsDownloadCmd `cmd:"" help:"Download a specific revision"`
	Restore  FilesRevisionsRestoreCmd  `cmd:"" help:"Restore file to a specific revision"`
}

type FilesRevisionsListCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
}

type FilesRevisionsDownloadCmd struct {
	FileID     string `arg:"" name:"file-id" help:"File ID or path"`
	RevisionID string `arg:"" name:"revision-id" help:"Revision ID"`
	OutputPath string `help:"Output path for revision download" name:"revision-output" required:""`
}

type FilesRevisionsRestoreCmd struct {
	FileID     string `arg:"" name:"file-id" help:"File ID or path"`
	RevisionID string `arg:"" name:"revision-id" help:"Revision ID"`
}

type FilesListTrashedCmd struct {
	Query     string `help:"Search query" name:"query"`
	Limit     int    `help:"Maximum files to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	OrderBy   string `help:"Sort order" name:"order-by"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type FilesExportFormatsCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
}

type FilesBatchUploadCmd struct {
	Paths         []string `arg:"" name:"paths" help:"File paths to upload"`
	FilesFrom     string   `help:"Read file list from JSON or text file" name:"files-from"`
	Parent        string   `help:"Parent folder ID" name:"parent"`
	Workers       int      `help:"Number of concurrent workers (1-10)" default:"3" name:"workers"`
	ContinueOnErr bool     `help:"Continue on error" name:"continue-on-error"`
}

type FilesBatchDownloadCmd struct {
	FileIDs       []string `arg:"" name:"file-ids" help:"File IDs to download"`
	IDsFrom       string   `help:"Read file IDs from JSON or text file" name:"ids-from"`
	OutputDir     string   `help:"Output directory" name:"output-dir"`
	Workers       int      `help:"Number of concurrent workers (1-10)" default:"3" name:"workers"`
	ContinueOnErr bool     `help:"Continue on error" name:"continue-on-error"`
	MimeType      string   `help:"Export MIME type" name:"mime-type"`
}

type FilesBatchDeleteCmd struct {
	FileIDs       []string `arg:"" name:"file-ids" help:"File IDs to delete"`
	IDsFrom       string   `help:"Read file IDs from JSON or text file" name:"ids-from"`
	Workers       int      `help:"Number of concurrent workers (1-10)" default:"3" name:"workers"`
	ContinueOnErr bool     `help:"Continue on error" name:"continue-on-error"`
	Permanent     bool     `help:"Permanently delete (skip trash)" name:"permanent"`
}

func getFileManager(ctx context.Context, flags types.GlobalFlags) (*files.Manager, *api.Client, *types.RequestContext, *OutputWriter, error) {
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
	mgr := files.NewManager(client)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	return mgr, client, reqCtx, out, nil
}

func (cmd *FilesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	// Resolve parent path if provided
	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("files.list", appErr.CLIError)
			}
			return out.WriteError("files.list", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	opts := files.ListOptions{
		ParentID:       parentID,
		Query:          cmd.Query,
		PageSize:       cmd.Limit,
		PageToken:      cmd.PageToken,
		OrderBy:        cmd.OrderBy,
		IncludeTrashed: cmd.IncludeTrashed,
		Fields:         cmd.Fields,
	}

	// If --paginate flag is set, fetch all pages
	if cmd.Paginate {
		allFiles, err := mgr.ListAll(ctx, reqCtx, opts)
		if err != nil {
			return handleCLIError(out, "files.list", err)
		}
		// Return result without nextPageToken (all pages fetched)
		return out.WriteSuccess("files.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "files.list", err)
	}

	return out.WriteSuccess("files.list", result)
}

func (cmd *FilesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	// Resolve file ID from path if needed
	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.get", appErr.CLIError)
		}
		return out.WriteError("files.get", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	file, err := mgr.Get(ctx, reqCtx, fileID, cmd.Fields)
	if err != nil {
		return handleCLIError(out, "files.get", err)
	}

	return out.WriteSuccess("files.get", file)
}

func (cmd *FilesUploadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.upload", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	// Resolve parent path if provided
	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("files.upload", appErr.CLIError)
			}
			return out.WriteError("files.upload", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation
	file, err := mgr.Upload(ctx, reqCtx, cmd.LocalPath, files.UploadOptions{
		ParentID: parentID,
		Name:     cmd.Name,
		MimeType: cmd.MimeType,
	})
	if err != nil {
		return handleCLIError(out, "files.upload", err)
	}

	out.Log("Uploaded: %s", file.Name)
	return out.WriteSuccess("files.upload", file)
}

func (cmd *FilesDownloadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.download", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	// Resolve file ID from path if needed
	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.download", appErr.CLIError)
		}
		return out.WriteError("files.download", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeDownloadOrExport
	mimeType := cmd.MimeType
	if cmd.Doc && mimeType == "" {
		mimeType = "text/plain"
	}

	err = mgr.Download(ctx, reqCtx, fileID, files.DownloadOptions{
		OutputPath: cmd.OutputPath,
		MimeType:   mimeType,
	})
	if err != nil {
		return handleCLIError(out, "files.download", err)
	}

	out.Log("Downloaded to: %s", cmd.OutputPath)
	return out.WriteSuccess("files.download", map[string]string{"path": cmd.OutputPath})
}

func (cmd *FilesDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	// Resolve file ID from path if needed
	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.delete", appErr.CLIError)
		}
		return out.WriteError("files.delete", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	err = mgr.Delete(ctx, reqCtx, fileID, cmd.Permanent)
	if err != nil {
		return handleCLIError(out, "files.delete", err)
	}

	action := "trashed"
	if cmd.Permanent {
		action = "deleted"
	}
	out.Log("File %s: %s", action, fileID)
	return out.WriteSuccess("files.delete", map[string]string{"id": fileID, "status": action})
}

func (cmd *FilesCopyCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.copy", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.copy", appErr.CLIError)
		}
		return out.WriteError("files.copy", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("files.copy", appErr.CLIError)
			}
			return out.WriteError("files.copy", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation
	file, err := mgr.Copy(ctx, reqCtx, fileID, cmd.Name, parentID)
	if err != nil {
		return handleCLIError(out, "files.copy", err)
	}

	out.Log("Copied to: %s", file.Name)
	return out.WriteSuccess("files.copy", file)
}

func (cmd *FilesMoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.move", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.move", appErr.CLIError)
		}
		return out.WriteError("files.move", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	parentID, err := ResolveFileID(ctx, client, flags, cmd.Parent)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.move", appErr.CLIError)
		}
		return out.WriteError("files.move", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	file, err := mgr.Move(ctx, reqCtx, fileID, parentID)
	if err != nil {
		return handleCLIError(out, "files.move", err)
	}

	out.Log("Moved: %s", file.Name)
	return out.WriteSuccess("files.move", file)
}

func (cmd *FilesTrashCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.trash", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.trash", appErr.CLIError)
		}
		return out.WriteError("files.trash", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	file, err := mgr.Trash(ctx, reqCtx, fileID)
	if err != nil {
		return handleCLIError(out, "files.trash", err)
	}

	out.Log("Trashed: %s", file.Name)
	return out.WriteSuccess("files.trash", file)
}

func (cmd *FilesRestoreCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.restore", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.restore", appErr.CLIError)
		}
		return out.WriteError("files.restore", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	file, err := mgr.Restore(ctx, reqCtx, fileID)
	if err != nil {
		return handleCLIError(out, "files.restore", err)
	}

	out.Log("Restored: %s", file.Name)
	return out.WriteSuccess("files.restore", file)
}

func (cmd *FilesRevisionsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	_, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.revisions", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.revisions", appErr.CLIError)
		}
		return out.WriteError("files.revisions", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	revMgr := revisions.NewManager(client)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := revMgr.List(ctx, reqCtx, fileID, revisions.ListOptions{})
	if err != nil {
		return handleCLIError(out, "files.revisions", err)
	}

	return out.WriteSuccess("files.revisions", result)
}

func (cmd *FilesRevisionsDownloadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	_, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.revisions.download", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.revisions.download", appErr.CLIError)
		}
		return out.WriteError("files.revisions.download", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	revMgr := revisions.NewManager(client)
	reqCtx.RequestType = types.RequestTypeDownloadOrExport

	err = revMgr.Download(ctx, reqCtx, fileID, cmd.RevisionID, revisions.DownloadOptions{
		OutputPath: cmd.OutputPath,
	})
	if err != nil {
		return handleCLIError(out, "files.revisions.download", err)
	}

	out.Log("Downloaded revision %s to: %s", cmd.RevisionID, cmd.OutputPath)
	return out.WriteSuccess("files.revisions.download", map[string]string{"revisionId": cmd.RevisionID, "path": cmd.OutputPath})
}

func (cmd *FilesRevisionsRestoreCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	_, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.revisions.restore", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.revisions.restore", appErr.CLIError)
		}
		return out.WriteError("files.revisions.restore", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	revMgr := revisions.NewManager(client)
	reqCtx.RequestType = types.RequestTypeMutation

	file, err := revMgr.Restore(ctx, reqCtx, fileID, cmd.RevisionID)
	if err != nil {
		return handleCLIError(out, "files.revisions.restore", err)
	}

	out.Log("Restored file to revision: %s", cmd.RevisionID)
	return out.WriteSuccess("files.revisions.restore", file)
}

func (cmd *FilesListTrashedCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.list-trashed", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := files.ListOptions{
		Query:     cmd.Query,
		PageSize:  cmd.Limit,
		PageToken: cmd.PageToken,
		OrderBy:   cmd.OrderBy,
		Fields:    cmd.Fields,
	}

	if cmd.Paginate {
		opts.IncludeTrashed = true
		if opts.Query != "" {
			opts.Query = "trashed = true and (" + opts.Query + ")"
		} else {
			opts.Query = "trashed = true"
		}
		allFiles, err := mgr.ListAll(ctx, reqCtx, opts)
		if err != nil {
			return handleCLIError(out, "files.list-trashed", err)
		}
		return out.WriteSuccess("files.list-trashed", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.ListTrashed(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "files.list-trashed", err)
	}

	return out.WriteSuccess("files.list-trashed", result)
}

func (cmd *FilesExportFormatsCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.export-formats", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	fileID, err := ResolveFileID(ctx, client, flags, cmd.FileID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("files.export-formats", appErr.CLIError)
		}
		return out.WriteError("files.export-formats", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	file, err := mgr.Get(ctx, reqCtx, fileID, "mimeType,name")
	if err != nil {
		return handleCLIError(out, "files.export-formats", err)
	}

	formats, err := export.GetAvailableFormats(file.MimeType)
	if err != nil {
		return out.WriteError("files.export-formats", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	result := map[string]interface{}{
		"file": map[string]string{
			"id":       fileID,
			"name":     file.Name,
			"mimeType": file.MimeType,
		},
		"availableFormats": formats,
	}

	return out.WriteSuccess("files.export-formats", result)
}

func (cmd *FilesBatchUploadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.batch-upload", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	var paths []string
	paths = append(paths, cmd.Paths...)

	if cmd.FilesFrom != "" {
		filePaths, err := loadFileList(cmd.FilesFrom)
		if err != nil {
			return out.WriteError("files.batch-upload", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
		}
		paths = append(paths, filePaths...)
	}

	if len(paths) == 0 {
		return out.WriteError("files.batch-upload", utils.NewCLIError(utils.ErrCodeInvalidArgument, "No files to upload. Provide file paths as arguments or use --files-from").Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			return out.WriteError("files.batch-upload", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation

	progressFunc := defaultProgressFunc("upload")
	result, err := mgr.BatchUpload(ctx, reqCtx, paths, files.BatchUploadOptions{
		ParentID:      parentID,
		Workers:       cmd.Workers,
		ContinueOnErr: cmd.ContinueOnErr,
		ProgressFunc: func(index, total int, path string, success bool, err error) {
			progressFunc(index, total, path, "", success, err)
		},
	})

	if err != nil && !cmd.ContinueOnErr {
		return handleCLIError(out, "files.batch-upload", err)
	}

	fmt.Printf("\n=== Batch Upload Summary ===\n")
	fmt.Printf("Total: %d, Success: %d, Failed: %d\n", result.TotalCount, result.SuccessCount, result.FailedCount)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.Path, e.Error)
		}
	}

	return out.WriteSuccess("files.batch-upload", map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   result.TotalCount,
			"success": result.SuccessCount,
			"failed":  result.FailedCount,
		},
		"files":  result.Files,
		"errors": result.Errors,
	})
}

func (cmd *FilesBatchDownloadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.batch-download", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	var fileIDs []string
	fileIDs = append(fileIDs, cmd.FileIDs...)

	if cmd.IDsFrom != "" {
		ids, err := loadIDList(cmd.IDsFrom)
		if err != nil {
			return out.WriteError("files.batch-download", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
		}
		fileIDs = append(fileIDs, ids...)
	}

	if len(fileIDs) == 0 {
		return out.WriteError("files.batch-download", utils.NewCLIError(utils.ErrCodeInvalidArgument, "No file IDs to download. Provide IDs as arguments or use --ids-from").Build())
	}

	reqCtx.RequestType = types.RequestTypeDownloadOrExport

	progressFunc := defaultProgressFunc("download")
	result, err := mgr.BatchDownload(ctx, reqCtx, fileIDs, files.BatchDownloadOptions{
		OutputDir:     cmd.OutputDir,
		Workers:       cmd.Workers,
		ContinueOnErr: cmd.ContinueOnErr,
		MimeType:      cmd.MimeType,
		ProgressFunc: func(index, total int, id, name string, success bool, err error) {
			progressFunc(index, total, id, name, success, err)
		},
	})

	if err != nil && !cmd.ContinueOnErr {
		return handleCLIError(out, "files.batch-download", err)
	}

	fmt.Printf("\n=== Batch Download Summary ===\n")
	fmt.Printf("Total: %d, Success: %d, Failed: %d\n", result.TotalCount, result.SuccessCount, result.FailedCount)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.ID, e.Error)
		}
	}

	return out.WriteSuccess("files.batch-download", map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   result.TotalCount,
			"success": result.SuccessCount,
			"failed":  result.FailedCount,
		},
		"files":  result.Files,
		"errors": result.Errors,
	})
}

func (cmd *FilesBatchDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, client, reqCtx, out, err := getFileManager(ctx, flags)
	if err != nil {
		return out.WriteError("files.batch-delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	var fileIDs []string
	fileIDs = append(fileIDs, cmd.FileIDs...)

	if cmd.IDsFrom != "" {
		ids, err := loadIDList(cmd.IDsFrom)
		if err != nil {
			return out.WriteError("files.batch-delete", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
		}
		fileIDs = append(fileIDs, ids...)
	}

	if len(fileIDs) == 0 {
		return out.WriteError("files.batch-delete", utils.NewCLIError(utils.ErrCodeInvalidArgument, "No file IDs to delete. Provide IDs as arguments or use --ids-from").Build())
	}

	safetyOpts := safety.Default()
	safetyOpts.Yes = globals.Yes
	if globals.DryRun {
		safetyOpts.DryRun = true
	}

	if !globals.Yes && !globals.DryRun {
		names := make([]string, 0, len(fileIDs))
		for _, id := range fileIDs {
			file, err := mgr.Get(ctx, reqCtx, id, "id,name")
			if err != nil {
				resolvedID, resolveErr := ResolveFileID(ctx, client, flags, id)
				if resolveErr != nil {
					names = append(names, fmt.Sprintf("%s (unable to resolve)", id))
					continue
				}
				file, err = mgr.Get(ctx, reqCtx, resolvedID, "id,name")
				if err != nil {
					names = append(names, fmt.Sprintf("%s (error: %s)", id, err.Error()))
					continue
				}
			}
			names = append(names, file.Name)
		}

		operation := "trash"
		if cmd.Permanent {
			operation = "permanently delete"
		}
		confirmed, err := safety.ConfirmDestructive(names, operation, safetyOpts)
		if err != nil {
			return out.WriteError("files.batch-delete", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
		}
		if !confirmed {
			return out.WriteSuccess("files.batch-delete", map[string]string{"status": "cancelled"})
		}
	}

	reqCtx.RequestType = types.RequestTypeMutation

	progressFunc := defaultProgressFunc("delete")
	result, err := mgr.BatchDelete(ctx, reqCtx, fileIDs, files.BatchDeleteOptions{
		Permanent:     cmd.Permanent,
		Workers:       cmd.Workers,
		ContinueOnErr: cmd.ContinueOnErr,
		SafetyOpts:    safetyOpts,
		DryRun:        globals.DryRun,
		ProgressFunc: func(index, total int, id string, success bool, err error) {
			progressFunc(index, total, id, "", success, err)
		},
	})

	if globals.DryRun {
		fmt.Printf("\n=== Dry Run Mode ===\n")
		fmt.Printf("Would %s %d file(s):\n", map[bool]string{true: "permanently delete", false: "trash"}[cmd.Permanent], len(fileIDs))
		for _, id := range fileIDs {
			file, _ := mgr.Get(ctx, reqCtx, id, "id,name")
			if file != nil {
				fmt.Printf("  - %s (%s)\n", file.Name, id)
			} else {
				fmt.Printf("  - %s\n", id)
			}
		}
		fmt.Printf("\nNo files were actually deleted.\n")
		return out.WriteSuccess("files.batch-delete", map[string]string{"status": "dry-run", "count": fmt.Sprintf("%d", len(fileIDs))})
	}

	if err != nil && !cmd.ContinueOnErr {
		return handleCLIError(out, "files.batch-delete", err)
	}

	action := "trashed"
	if cmd.Permanent {
		action = "deleted"
	}
	fmt.Printf("\n=== Batch Delete Summary ===\n")
	fmt.Printf("Total: %d, Success: %d, Failed: %d\n", result.TotalCount, result.SuccessCount, result.FailedCount)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors:\n")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.ID, e.Error)
		}
	}

	return out.WriteSuccess("files.batch-delete", map[string]interface{}{
		"summary": map[string]interface{}{
			"total":   result.TotalCount,
			"success": result.SuccessCount,
			"failed":  result.FailedCount,
			"action":  action,
		},
		"deletedIDs": result.DeletedIDs,
		"errors":     result.Errors,
	})
}

// loadFileList loads file paths from a JSON or text file
func loadFileList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file list: %w", err)
	}

	// Try JSON first
	var jsonPaths []string
	if err := json.Unmarshal(data, &jsonPaths); err == nil {
		return jsonPaths, nil
	}

	// Try JSON object with "files" field
	var jsonObj struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(data, &jsonObj); err == nil && len(jsonObj.Files) > 0 {
		return jsonObj.Files, nil
	}

	// Fall back to line-by-line text file
	var paths []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse file list: %w", err)
	}

	return paths, nil
}

// loadIDList loads file IDs from a JSON or text file
func loadIDList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID list: %w", err)
	}

	// Try JSON first
	var jsonIDs []string
	if err := json.Unmarshal(data, &jsonIDs); err == nil {
		return jsonIDs, nil
	}

	// Try JSON object with "ids" field
	var jsonObj struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(data, &jsonObj); err == nil && len(jsonObj.IDs) > 0 {
		return jsonObj.IDs, nil
	}

	// Fall back to line-by-line text file
	var ids []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ids = append(ids, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse ID list: %w", err)
	}

	return ids, nil
}

// defaultProgressFunc creates a default progress function that prints to stdout
func defaultProgressFunc(operation string) func(index, total int, idOrPath, name string, success bool, err error) {
	startTime := time.Now()
	var mu sync.Mutex
	var lastProgress int

	return func(index, total int, idOrPath, name string, success bool, err error) {
		mu.Lock()
		defer mu.Unlock()

		// Only print progress at certain intervals to avoid spam
		progress := (index * 100) / total
		if progress != lastProgress || index == total || index == 1 {
			lastProgress = progress

			elapsed := time.Since(startTime)
			var eta time.Duration
			if index > 0 {
				rate := elapsed / time.Duration(index)
				remaining := total - index
				eta = rate * time.Duration(remaining)
			}

			status := "✓"
			if !success {
				status = "✗"
			}

			display := idOrPath
			if name != "" {
				display = name
			}

			fmt.Printf("\r[%3d%%] %s %s (%d/%d) - ETA: %s",
				progress, status, display, index, total, formatDuration(eta))

			if index == total {
				fmt.Println() // New line at end
			}
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
