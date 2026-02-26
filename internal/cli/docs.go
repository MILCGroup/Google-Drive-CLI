package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	docsmgr "github.com/dl-alexandre/gdrv/internal/docs"
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	docsapi "google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

type DocsCmd struct {
	List   DocsListCmd   `cmd:"" help:"List documents"`
	Get    DocsGetCmd    `cmd:"" help:"Get document metadata"`
	Read   DocsReadCmd   `cmd:"" help:"Extract plain text from document"`
	Create DocsCreateCmd `cmd:"" help:"Create a document"`
	Update DocsUpdateCmd `cmd:"" help:"Batch update document"`
}

type DocsListCmd struct {
	Parent    string `help:"Parent folder ID" name:"parent"`
	Query     string `help:"Additional search query" name:"query"`
	Limit     int    `help:"Maximum files to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	OrderBy   string `help:"Sort order" name:"order-by"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type DocsGetCmd struct {
	DocumentID string `arg:"" name:"document-id" help:"Document ID or path"`
}

type DocsReadCmd struct {
	DocumentID string `arg:"" name:"document-id" help:"Document ID or path"`
}

type DocsCreateCmd struct {
	Name   string `arg:"" name:"name" help:"Document name"`
	Parent string `help:"Parent folder ID" name:"parent"`
}

type DocsUpdateCmd struct {
	DocumentID   string `arg:"" name:"document-id" help:"Document ID or path"`
	Requests     string `help:"Batch update requests JSON" name:"requests"`
	RequestsFile string `help:"Path to JSON file with batch update requests" name:"requests-file"`
}

func (cmd *DocsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getDocsService(ctx, flags)
	if err != nil {
		return out.WriteError("docs.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("docs.list", appErr.CLIError)
			}
			return out.WriteError("docs.list", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	query := fmt.Sprintf("mimeType = '%s'", utils.MimeTypeDocument)
	if parentID != "" {
		query = fmt.Sprintf("'%s' in parents and %s", parentID, query)
	}
	if cmd.Query != "" {
		query = fmt.Sprintf("%s and (%s)", query, cmd.Query)
	}

	opts := files.ListOptions{
		ParentID:       parentID,
		Query:          query,
		PageSize:       cmd.Limit,
		PageToken:      cmd.PageToken,
		OrderBy:        cmd.OrderBy,
		IncludeTrashed: false,
		Fields:         cmd.Fields,
	}

	mgr := files.NewManager(client)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		allFiles, err := mgr.ListAll(ctx, reqCtx, opts)
		if err != nil {
			return handleCLIError(out, "docs.list", err)
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("docs.list", allFiles)
		}
		return out.WriteSuccess("docs.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "docs.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("docs.list", result.Files)
	}
	return out.WriteSuccess("docs.list", result)
}

func (cmd *DocsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getDocsService(ctx, flags)
	if err != nil {
		return out.WriteError("docs.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	documentID, err := ResolveFileID(ctx, client, flags, cmd.DocumentID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("docs.get", appErr.CLIError)
		}
		return out.WriteError("docs.get", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := docsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetDocument(ctx, reqCtx, documentID)
	if err != nil {
		return handleCLIError(out, "docs.get", err)
	}

	return out.WriteSuccess("docs.get", result)
}

func (cmd *DocsReadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getDocsService(ctx, flags)
	if err != nil {
		return out.WriteError("docs.read", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	documentID, err := ResolveFileID(ctx, client, flags, cmd.DocumentID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("docs.read", appErr.CLIError)
		}
		return out.WriteError("docs.read", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := docsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.ReadDocument(ctx, reqCtx, documentID)
	if err != nil {
		return handleCLIError(out, "docs.read", err)
	}

	return out.WriteSuccess("docs.read", result)
}

func (cmd *DocsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getDocsService(ctx, flags)
	if err != nil {
		return out.WriteError("docs.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("docs.create", appErr.CLIError)
			}
			return out.WriteError("docs.create", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation
	metadata := &drive.File{
		Name:     cmd.Name,
		MimeType: utils.MimeTypeDocument,
	}
	if parentID != "" {
		metadata.Parents = []string{parentID}
		reqCtx.InvolvedParentIDs = append(reqCtx.InvolvedParentIDs, parentID)
	}

	call := client.Service().Files.Create(metadata)
	call = api.NewRequestShaper(client).ShapeFilesCreate(call, reqCtx)
	call = call.Fields("id,name,mimeType,size,createdTime,modifiedTime,parents,resourceKey,trashed,capabilities")

	result, err := api.ExecuteWithRetry(ctx, client, reqCtx, func() (*drive.File, error) {
		return call.Do()
	})
	if err != nil {
		return handleCLIError(out, "docs.create", err)
	}

	if result.ResourceKey != "" {
		client.ResourceKeys().UpdateFromAPIResponse(result.Id, result.ResourceKey)
	}

	return out.WriteSuccess("docs.create", convertDriveFile(result))
}

func (cmd *DocsUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getDocsService(ctx, flags)
	if err != nil {
		return out.WriteError("docs.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	requests, err := parseDocsRequests(cmd.Requests, cmd.RequestsFile)
	if err != nil {
		return out.WriteError("docs.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	documentID, err := ResolveFileID(ctx, client, flags, cmd.DocumentID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("docs.update", appErr.CLIError)
		}
		return out.WriteError("docs.update", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := docsmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdateDocument(ctx, reqCtx, documentID, requests)
	if err != nil {
		return handleCLIError(out, "docs.update", err)
	}

	return out.WriteSuccess("docs.update", result)
}

func getDocsService(ctx context.Context, flags types.GlobalFlags) (*docsapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceDocs); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetDocsService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return svc, client, reqCtx, nil
}

func parseDocsRequests(requestsJSON, requestsFile string) ([]*docsapi.Request, error) {
	if requestsJSON == "" && requestsFile == "" {
		return nil, fmt.Errorf("requests required via --requests or --requests-file")
	}

	var raw []byte
	if requestsFile != "" {
		data, err := os.ReadFile(requestsFile)
		if err != nil {
			return nil, err
		}
		raw = data
	} else {
		raw = []byte(requestsJSON)
	}

	var requests []*docsapi.Request
	if err := json.Unmarshal(raw, &requests); err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one request is required")
	}
	return requests, nil
}
