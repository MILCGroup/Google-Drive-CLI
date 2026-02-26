package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/files"
	slidesmgr "github.com/dl-alexandre/gdrv/internal/slides"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
	slidesapi "google.golang.org/api/slides/v1"
)

type SlidesCmd struct {
	List    SlidesListCmd    `cmd:"" help:"List presentations"`
	Get     SlidesGetCmd     `cmd:"" help:"Get presentation metadata"`
	Read    SlidesReadCmd    `cmd:"" help:"Extract text from all slides"`
	Create  SlidesCreateCmd  `cmd:"" help:"Create a presentation"`
	Update  SlidesUpdateCmd  `cmd:"" help:"Batch update presentation"`
	Replace SlidesReplaceCmd `cmd:"" help:"Replace text placeholders"`
}

type SlidesListCmd struct {
	Parent    string `help:"Parent folder ID" name:"parent"`
	Query     string `help:"Additional search query" name:"query"`
	Limit     int    `help:"Maximum files to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	OrderBy   string `help:"Sort order" name:"order-by"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type SlidesGetCmd struct {
	PresentationID string `arg:"" name:"presentation-id" help:"Presentation ID or path"`
}

type SlidesReadCmd struct {
	PresentationID string `arg:"" name:"presentation-id" help:"Presentation ID or path"`
}

type SlidesCreateCmd struct {
	Name   string `arg:"" name:"name" help:"Presentation name"`
	Parent string `help:"Parent folder ID" name:"parent"`
}

type SlidesUpdateCmd struct {
	PresentationID string `arg:"" name:"presentation-id" help:"Presentation ID or path"`
	Requests       string `help:"Batch update requests JSON" name:"requests"`
	RequestsFile   string `help:"Path to JSON file with batch update requests" name:"requests-file"`
}

type SlidesReplaceCmd struct {
	PresentationID string `arg:"" name:"presentation-id" help:"Presentation ID or path"`
	Data           string `help:"JSON string with replacements map" name:"data"`
	File           string `help:"Path to JSON file with replacements map" name:"file"`
}

func (cmd *SlidesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("slides.list", appErr.CLIError)
			}
			return out.WriteError("slides.list", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	query := fmt.Sprintf("mimeType = '%s'", utils.MimeTypePresentation)
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
			return handleCLIError(out, "slides.list", err)
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("slides.list", allFiles)
		}
		return out.WriteSuccess("slides.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "slides.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("slides.list", result.Files)
	}
	return out.WriteSuccess("slides.list", result)
}

func (cmd *SlidesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	presentationID, err := ResolveFileID(ctx, client, flags, cmd.PresentationID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("slides.get", appErr.CLIError)
		}
		return out.WriteError("slides.get", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := slidesmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetPresentation(ctx, reqCtx, presentationID)
	if err != nil {
		return handleCLIError(out, "slides.get", err)
	}

	return out.WriteSuccess("slides.get", result)
}

func (cmd *SlidesReadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.read", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	presentationID, err := ResolveFileID(ctx, client, flags, cmd.PresentationID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("slides.read", appErr.CLIError)
		}
		return out.WriteError("slides.read", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := slidesmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.ReadPresentation(ctx, reqCtx, presentationID)
	if err != nil {
		return handleCLIError(out, "slides.read", err)
	}

	return out.WriteSuccess("slides.read", result)
}

func (cmd *SlidesCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	_, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	parentID := cmd.Parent
	if parentID != "" {
		resolvedID, err := ResolveFileID(ctx, client, flags, parentID)
		if err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				return out.WriteError("slides.create", appErr.CLIError)
			}
			return out.WriteError("slides.create", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
		}
		parentID = resolvedID
	}

	reqCtx.RequestType = types.RequestTypeMutation
	metadata := &drive.File{
		Name:     cmd.Name,
		MimeType: utils.MimeTypePresentation,
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
		return handleCLIError(out, "slides.create", err)
	}

	if result.ResourceKey != "" {
		client.ResourceKeys().UpdateFromAPIResponse(result.Id, result.ResourceKey)
	}

	return out.WriteSuccess("slides.create", convertDriveFile(result))
}

func (cmd *SlidesUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	requests, err := parseSlidesRequests(cmd.Requests, cmd.RequestsFile)
	if err != nil {
		return out.WriteError("slides.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	presentationID, err := ResolveFileID(ctx, client, flags, cmd.PresentationID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("slides.update", appErr.CLIError)
		}
		return out.WriteError("slides.update", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := slidesmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdatePresentation(ctx, reqCtx, presentationID, requests)
	if err != nil {
		return handleCLIError(out, "slides.update", err)
	}

	return out.WriteSuccess("slides.update", result)
}

func (cmd *SlidesReplaceCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getSlidesService(ctx, flags)
	if err != nil {
		return out.WriteError("slides.replace", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	replacements, err := parseSlidesReplacements(cmd.Data, cmd.File)
	if err != nil {
		return out.WriteError("slides.replace", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
	}

	presentationID, err := ResolveFileID(ctx, client, flags, cmd.PresentationID)
	if err != nil {
		var appErr *utils.AppError
		if errors.As(err, &appErr) {
			return out.WriteError("slides.replace", appErr.CLIError)
		}
		return out.WriteError("slides.replace", utils.NewCLIError(utils.ErrCodeInvalidPath, err.Error()).Build())
	}

	mgr := slidesmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.ReplaceAllText(ctx, reqCtx, presentationID, replacements)
	if err != nil {
		return handleCLIError(out, "slides.replace", err)
	}

	return out.WriteSuccess("slides.replace", result)
}

func getSlidesService(ctx context.Context, flags types.GlobalFlags) (*slidesapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceSlides); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetSlidesService(ctx, creds)
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

func parseSlidesRequests(requestsJSON, requestsFile string) ([]*slidesapi.Request, error) {
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

	var requests []*slidesapi.Request
	if err := json.Unmarshal(raw, &requests); err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one request is required")
	}
	return requests, nil
}

func parseSlidesReplacements(data, file string) (map[string]string, error) {
	if data == "" && file == "" {
		return nil, fmt.Errorf("replacements required via --data or --file")
	}

	var raw []byte
	if file != "" {
		fileData, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		raw = fileData
	} else {
		raw = []byte(data)
	}

	var replacements map[string]string
	if err := json.Unmarshal(raw, &replacements); err != nil {
		return nil, err
	}
	if len(replacements) == 0 {
		return nil, fmt.Errorf("at least one replacement is required")
	}
	return replacements, nil
}
