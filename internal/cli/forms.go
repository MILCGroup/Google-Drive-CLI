package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	formsmgr "github.com/dl-alexandre/gdrv/internal/forms"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type FormsCmd struct {
	Get       FormsGetCmd       `cmd:"" help:"Get form metadata and questions"`
	Responses FormsResponsesCmd `cmd:"" help:"List form responses"`
	Create    FormsCreateCmd    `cmd:"" help:"Create a new form"`
}

type FormsGetCmd struct {
	FormID string `arg:"" name:"form-id" help:"Form ID"`
}

type FormsResponsesCmd struct {
	FormID    string `arg:"" name:"form-id" help:"Form ID"`
	Limit     int    `help:"Maximum responses to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type FormsCreateCmd struct {
	Title string `help:"Form title" required:"" name:"title"`
}

func (cmd *FormsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, _, err := getFormsManager(ctx, flags)
	if err != nil {
		return out.WriteError("forms.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetForm(ctx, reqCtx, cmd.FormID)
	if err != nil {
		return handleCLIError(out, "forms.get", err)
	}

	return out.WriteSuccess("forms.get", result)
}

func (cmd *FormsResponsesCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, _, err := getFormsManager(ctx, flags)
	if err != nil {
		return out.WriteError("forms.responses", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allResponses []types.FormResponse
		pageToken := cmd.PageToken
		for {
			result, nextToken, err := mgr.ListResponses(ctx, reqCtx, cmd.FormID, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "forms.responses", err)
			}
			allResponses = append(allResponses, result.Responses...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("forms.responses", &types.FormResponseList{Responses: allResponses})
	}

	result, nextToken, err := mgr.ListResponses(ctx, reqCtx, cmd.FormID, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "forms.responses", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("forms.responses", result)
	}
	return out.WriteSuccess("forms.responses", map[string]interface{}{
		"responses":     result.Responses,
		"nextPageToken": nextToken,
	})
}

func (cmd *FormsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, _, err := getFormsManager(ctx, flags)
	if err != nil {
		return out.WriteError("forms.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateForm(ctx, reqCtx, cmd.Title)
	if err != nil {
		return handleCLIError(out, "forms.create", err)
	}

	return out.WriteSuccess("forms.create", result)
}

func getFormsManager(ctx context.Context, flags types.GlobalFlags) (*formsmgr.Manager, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, out, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceForms); err != nil {
		return nil, nil, out, err
	}

	svc, err := authMgr.GetFormsService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	mgr := formsmgr.NewManager(client, svc)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, out, nil
}
