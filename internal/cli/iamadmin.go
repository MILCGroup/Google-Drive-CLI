package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	iamadminmgr "github.com/dl-alexandre/gdrv/internal/iamadmin"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/option"
)

type IAMAdminCmd struct {
	ServiceAccounts IAMAdminServiceAccountsCmd `cmd:"" help:"Manage service accounts"`
	Roles           IAMAdminRolesCmd           `cmd:"" help:"Manage IAM roles"`
}

type IAMAdminServiceAccountsCmd struct {
	List   IAMAdminServiceAccountsListCmd   `cmd:"" help:"List service accounts"`
	Get    IAMAdminServiceAccountsGetCmd    `cmd:"" help:"Get a service account"`
	Create IAMAdminServiceAccountsCreateCmd `cmd:"" help:"Create a service account"`
	Delete IAMAdminServiceAccountsDeleteCmd `cmd:"" help:"Delete a service account"`
}

type IAMAdminRolesCmd struct {
	List IAMAdminRolesListCmd `cmd:"" help:"List IAM roles"`
	Get  IAMAdminRolesGetCmd  `cmd:"" help:"Get an IAM role"`
}

type IAMAdminServiceAccountsListCmd struct {
	ProjectID string `arg:"" name:"project-id" help:"Project ID"`
	PageSize  int32  `help:"Page size" default:"100" name:"page-size"`
}

type IAMAdminServiceAccountsGetCmd struct {
	Name string `arg:"" name:"name" help:"Service account name"`
}

type IAMAdminServiceAccountsCreateCmd struct {
	ProjectID   string `arg:"" name:"project-id" help:"Project ID"`
	AccountID   string `arg:"" name:"account-id" help:"Service account ID"`
	DisplayName string `help:"Display name" name:"display-name"`
	Description string `help:"Description" name:"description"`
}

type IAMAdminServiceAccountsDeleteCmd struct {
	Name string `arg:"" name:"name" help:"Service account name"`
}

type IAMAdminRolesListCmd struct {
	Parent   string `help:"Parent resource" name:"parent"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type IAMAdminRolesGetCmd struct {
	Name string `arg:"" name:"name" help:"Role name"`
}

func (cmd *IAMAdminServiceAccountsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getIAMAdminManager(ctx, flags)
	if err != nil {
		return out.WriteError("iamadmin.service-accounts.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer mgr.Close()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListServiceAccounts(ctx, reqCtx, cmd.ProjectID, "", cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "iamadmin.service-accounts.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("iamadmin.service-accounts.list", result.Accounts)
	}
	return out.WriteSuccess("iamadmin.service-accounts.list", result)
}

func (cmd *IAMAdminServiceAccountsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getIAMAdminManager(ctx, flags)
	if err != nil {
		return out.WriteError("iamadmin.service-accounts.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer mgr.Close()

	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateServiceAccount(ctx, reqCtx, cmd.ProjectID, cmd.AccountID, cmd.DisplayName, cmd.Description)
	if err != nil {
		return handleCLIError(out, "iamadmin.service-accounts.create", err)
	}

	return out.WriteSuccess("iamadmin.service-accounts.create", result)
}

func (cmd *IAMAdminRolesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getIAMAdminManager(ctx, flags)
	if err != nil {
		return out.WriteError("iamadmin.roles.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer mgr.Close()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListRoles(ctx, reqCtx, cmd.Parent, "", cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "iamadmin.roles.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("iamadmin.roles.list", result.Roles)
	}
	return out.WriteSuccess("iamadmin.roles.list", result)
}

func getIAMAdminManager(ctx context.Context, flags types.GlobalFlags) (*iamadminmgr.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	tokenSource := authMgr.GetTokenSource(ctx, creds)
	opt := option.WithTokenSource(tokenSource)

	mgr, err := iamadminmgr.NewManager(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, nil
}
