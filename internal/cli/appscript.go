package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/api"
	appscriptmgr "github.com/dl-alexandre/gdrv/internal/appscript"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type AppScriptCmd struct {
	Get     AppScriptGetCmd     `cmd:"" help:"Get Apps Script project metadata"`
	Content AppScriptContentCmd `cmd:"" help:"Get Apps Script project file content"`
	Create  AppScriptCreateCmd  `cmd:"" help:"Create a new Apps Script project"`
	Run     AppScriptRunCmd     `cmd:"" help:"EXPERIMENTAL: Run a function in an Apps Script project. Requires the script to be deployed as an API executable, the caller and script must share the same GCP project, and the caller must have all scopes the script uses. 6-minute execution timeout."`
}

type AppScriptGetCmd struct {
	ScriptID string `arg:"" name:"script-id" help:"Apps Script project ID"`
}

type AppScriptContentCmd struct {
	ScriptID string `arg:"" name:"script-id" help:"Apps Script project ID"`
}

type AppScriptCreateCmd struct {
	Title  string `help:"Project title" name:"title" required:""`
	Parent string `help:"Parent Drive folder ID to bind the script to" name:"parent"`
}

type AppScriptRunCmd struct {
	ScriptID   string `arg:"" name:"script-id" help:"Apps Script project ID"`
	Function   string `help:"Function name to execute" name:"function" required:""`
	Parameters string `help:"JSON array of parameters to pass to the function" name:"parameters"`
}

func getAppScriptManager(ctx context.Context, flags types.GlobalFlags) (*appscriptmgr.Manager, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, out, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceAppScript); err != nil {
		return nil, nil, out, err
	}

	svc, err := authMgr.GetAppScriptService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	mgr := appscriptmgr.NewManager(client, svc)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	return mgr, reqCtx, out, nil
}

func (cmd *AppScriptGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getAppScriptManager(ctx, flags)
	if err != nil {
		return out.WriteError("appscript.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	project, err := mgr.GetProject(ctx, reqCtx, cmd.ScriptID)
	if err != nil {
		return handleCLIError(out, "appscript.get", err)
	}

	return out.WriteSuccess("appscript.get", project)
}

func (cmd *AppScriptContentCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getAppScriptManager(ctx, flags)
	if err != nil {
		return out.WriteError("appscript.content", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	content, err := mgr.GetContent(ctx, reqCtx, cmd.ScriptID)
	if err != nil {
		return handleCLIError(out, "appscript.content", err)
	}

	return out.WriteSuccess("appscript.content", content)
}

func (cmd *AppScriptCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getAppScriptManager(ctx, flags)
	if err != nil {
		return out.WriteError("appscript.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateProject(ctx, reqCtx, cmd.Title, cmd.Parent)
	if err != nil {
		return handleCLIError(out, "appscript.create", err)
	}

	out.Log("Created Apps Script project: %s", result.Title)
	return out.WriteSuccess("appscript.create", result)
}

func (cmd *AppScriptRunCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getAppScriptManager(ctx, flags)
	if err != nil {
		return out.WriteError("appscript.run", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	var params []interface{}
	if cmd.Parameters != "" {
		if err := json.Unmarshal([]byte(cmd.Parameters), &params); err != nil {
			return out.WriteError("appscript.run", utils.NewCLIError(utils.ErrCodeInvalidArgument, fmt.Sprintf("invalid --parameters JSON: %v", err)).Build())
		}
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.RunFunction(ctx, reqCtx, cmd.ScriptID, cmd.Function, params)
	if err != nil {
		return handleCLIError(out, "appscript.run", err)
	}

	return out.WriteSuccess("appscript.run", result)
}
