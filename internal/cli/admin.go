package cli

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/admin"
	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	adminapi "google.golang.org/api/admin/directory/v1"
)

type AdminCmd struct {
	Users  AdminUsersCmd  `cmd:"" help:"User management operations"`
	Groups AdminGroupsCmd `cmd:"" help:"Group management operations"`
}

type AdminUsersCmd struct {
	List      AdminUsersListCmd      `cmd:"" help:"List users"`
	Get       AdminUsersGetCmd       `cmd:"" help:"Get user details"`
	Create    AdminUsersCreateCmd    `cmd:"" help:"Create a new user"`
	Delete    AdminUsersDeleteCmd    `cmd:"" help:"Delete a user"`
	Update    AdminUsersUpdateCmd    `cmd:"" help:"Update a user"`
	Suspend   AdminUsersSuspendCmd   `cmd:"" help:"Suspend a user"`
	Unsuspend AdminUsersUnsuspendCmd `cmd:"" help:"Unsuspend a user"`
}

type AdminGroupsCmd struct {
	List    AdminGroupsListCmd    `cmd:"" help:"List groups"`
	Get     AdminGroupsGetCmd     `cmd:"" help:"Get group details"`
	Create  AdminGroupsCreateCmd  `cmd:"" help:"Create a new group"`
	Delete  AdminGroupsDeleteCmd  `cmd:"" help:"Delete a group"`
	Update  AdminGroupsUpdateCmd  `cmd:"" help:"Update a group"`
	Members AdminGroupsMembersCmd `cmd:"" help:"Group membership management"`
}

type AdminGroupsMembersCmd struct {
	List   AdminGroupsMembersListCmd   `cmd:"" help:"List group members"`
	Add    AdminGroupsMembersAddCmd    `cmd:"" help:"Add member to group"`
	Remove AdminGroupsMembersRemoveCmd `cmd:"" help:"Remove member from group"`
}

type AdminUsersListCmd struct {
	Domain    string `help:"Domain to list users from" name:"domain"`
	Customer  string `help:"Customer ID" name:"customer"`
	Query     string `help:"Search query" name:"query"`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
	OrderBy   string `help:"Sort order" name:"order-by"`
}

type AdminUsersGetCmd struct {
	UserKey string `arg:"" name:"user-key" help:"User email or ID"`
	Fields  string `help:"Fields to return" name:"fields"`
}

type AdminUsersCreateCmd struct {
	Email      string `arg:"" name:"email" help:"User email address"`
	GivenName  string `help:"First name" name:"given-name" required:""`
	FamilyName string `help:"Last name" name:"family-name" required:""`
	Password   string `help:"Password" name:"password" required:""`
}

type AdminUsersDeleteCmd struct {
	UserKey string `arg:"" name:"user-key" help:"User email or ID"`
}

type AdminUsersUpdateCmd struct {
	UserKey     string `arg:"" name:"user-key" help:"User email or ID"`
	GivenName   string `help:"Update first name" name:"given-name"`
	FamilyName  string `help:"Update last name" name:"family-name"`
	Suspended   string `help:"Set suspension status (true/false)" name:"suspended"`
	OrgUnitPath string `help:"Update organizational unit path" name:"org-unit-path"`
}

type AdminUsersSuspendCmd struct {
	UserKey string `arg:"" name:"user-key" help:"User email or ID"`
}

type AdminUsersUnsuspendCmd struct {
	UserKey string `arg:"" name:"user-key" help:"User email or ID"`
}

// --- Group command structs ---

type AdminGroupsListCmd struct {
	Domain    string `help:"Domain to list groups from" name:"domain"`
	Customer  string `help:"Customer ID" name:"customer"`
	Query     string `help:"Search query" name:"query"`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
	OrderBy   string `help:"Sort order" name:"order-by"`
}

type AdminGroupsGetCmd struct {
	GroupKey string `arg:"" name:"group-key" help:"Group email, alias, or ID"`
	Fields   string `help:"Fields to return" name:"fields"`
}

type AdminGroupsCreateCmd struct {
	Email       string `arg:"" name:"email" help:"Group email address"`
	Name        string `arg:"" name:"name" help:"Group name"`
	Description string `help:"Group description" name:"description"`
}

type AdminGroupsDeleteCmd struct {
	GroupKey string `arg:"" name:"group-key" help:"Group email, alias, or ID"`
}

type AdminGroupsUpdateCmd struct {
	GroupKey    string `arg:"" name:"group-key" help:"Group email, alias, or ID"`
	Name        string `help:"Update group name" name:"name"`
	Description string `help:"Update group description" name:"description"`
}

// --- Members command structs ---

type AdminGroupsMembersListCmd struct {
	GroupKey  string `arg:"" name:"group-key" help:"Group email or ID"`
	Limit     int    `help:"Maximum results per page" default:"200" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Roles     string `help:"Filter by role (OWNER, MANAGER, MEMBER)" name:"roles"`
	Fields    string `help:"Fields to return" name:"fields"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type AdminGroupsMembersAddCmd struct {
	GroupKey    string `arg:"" name:"group-key" help:"Group email or ID"`
	MemberEmail string `arg:"" name:"member-email" help:"Member email"`
	Role        string `help:"Member role (OWNER, MANAGER, MEMBER)" default:"MEMBER" name:"role"`
}

type AdminGroupsMembersRemoveCmd struct {
	GroupKey  string `arg:"" name:"group-key" help:"Group email or ID"`
	MemberKey string `arg:"" name:"member-key" help:"Member email or ID"`
}

// --- Run methods ---

func (cmd *AdminUsersListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	if cmd.Domain == "" && cmd.Customer == "" {
		return out.WriteError("admin.users.list", utils.NewCLIError(utils.ErrCodeInvalidArgument, "domain or customer is required").Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListUsers(ctx, reqCtx, &admin.ListUsersOptions{
		Domain:     cmd.Domain,
		Customer:   cmd.Customer,
		Query:      cmd.Query,
		MaxResults: int64(cmd.Limit),
		PageToken:  cmd.PageToken,
		OrderBy:    cmd.OrderBy,
		Fields:     cmd.Fields,
		Paginate:   cmd.Paginate,
	})
	if err != nil {
		return handleCLIError(out, "admin.users.list", err)
	}

	return out.WriteSuccess("admin.users.list", result)
}

func (cmd *AdminUsersGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetUser(ctx, reqCtx, cmd.UserKey, cmd.Fields)
	if err != nil {
		return handleCLIError(out, "admin.users.get", err)
	}

	return out.WriteSuccess("admin.users.get", result)
}

func (cmd *AdminUsersCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateUser(ctx, reqCtx, &types.CreateUserRequest{
		Email:      cmd.Email,
		GivenName:  cmd.GivenName,
		FamilyName: cmd.FamilyName,
		Password:   cmd.Password,
	})
	if err != nil {
		return handleCLIError(out, "admin.users.create", err)
	}

	return out.WriteSuccess("admin.users.create", result)
}

func (cmd *AdminUsersDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	if err := mgr.DeleteUser(ctx, reqCtx, cmd.UserKey); err != nil {
		return handleCLIError(out, "admin.users.delete", err)
	}

	return out.WriteSuccess("admin.users.delete", map[string]string{
		"message": fmt.Sprintf("User %s deleted successfully", cmd.UserKey),
	})
}

func (cmd *AdminUsersUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	req := &types.UpdateUserRequest{}
	if cmd.GivenName != "" {
		req.GivenName = &cmd.GivenName
	}
	if cmd.FamilyName != "" {
		req.FamilyName = &cmd.FamilyName
	}
	if cmd.Suspended != "" {
		value, err := parseSuspendedFlag(cmd.Suspended)
		if err != nil {
			return out.WriteError("admin.users.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, err.Error()).Build())
		}
		req.Suspended = value
	}
	if cmd.OrgUnitPath != "" {
		req.OrgUnitPath = &cmd.OrgUnitPath
	}
	if req.GivenName == nil && req.FamilyName == nil && req.Suspended == nil && req.OrgUnitPath == nil {
		return out.WriteError("admin.users.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one field must be provided").Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdateUser(ctx, reqCtx, cmd.UserKey, req)
	if err != nil {
		return handleCLIError(out, "admin.users.update", err)
	}

	return out.WriteSuccess("admin.users.update", result)
}

func (cmd *AdminUsersSuspendCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.suspend", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.SuspendUser(ctx, reqCtx, cmd.UserKey)
	if err != nil {
		return handleCLIError(out, "admin.users.suspend", err)
	}

	return out.WriteSuccess("admin.users.suspend", result)
}

func (cmd *AdminUsersUnsuspendCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.users.unsuspend", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UnsuspendUser(ctx, reqCtx, cmd.UserKey)
	if err != nil {
		return handleCLIError(out, "admin.users.unsuspend", err)
	}

	return out.WriteSuccess("admin.users.unsuspend", result)
}

func (cmd *AdminGroupsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	if cmd.Domain == "" && cmd.Customer == "" {
		return out.WriteError("admin.groups.list", utils.NewCLIError(utils.ErrCodeInvalidArgument, "domain or customer is required").Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListGroups(ctx, reqCtx, &admin.ListGroupsOptions{
		Domain:     cmd.Domain,
		Customer:   cmd.Customer,
		Query:      cmd.Query,
		MaxResults: int64(cmd.Limit),
		PageToken:  cmd.PageToken,
		OrderBy:    cmd.OrderBy,
		Fields:     cmd.Fields,
		Paginate:   cmd.Paginate,
	})
	if err != nil {
		return handleCLIError(out, "admin.groups.list", err)
	}

	return out.WriteSuccess("admin.groups.list", result)
}

func (cmd *AdminGroupsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetGroup(ctx, reqCtx, cmd.GroupKey, &admin.GetGroupOptions{Fields: cmd.Fields})
	if err != nil {
		return handleCLIError(out, "admin.groups.get", err)
	}

	return out.WriteSuccess("admin.groups.get", result)
}

func (cmd *AdminGroupsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateGroup(ctx, reqCtx, &types.CreateGroupRequest{
		Email:       cmd.Email,
		Name:        cmd.Name,
		Description: cmd.Description,
	})
	if err != nil {
		return handleCLIError(out, "admin.groups.create", err)
	}

	return out.WriteSuccess("admin.groups.create", result)
}

func (cmd *AdminGroupsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	if err := mgr.DeleteGroup(ctx, reqCtx, cmd.GroupKey); err != nil {
		return handleCLIError(out, "admin.groups.delete", err)
	}

	return out.WriteSuccess("admin.groups.delete", map[string]string{
		"message": fmt.Sprintf("Group %s deleted successfully", cmd.GroupKey),
	})
}

func (cmd *AdminGroupsUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	req := &types.UpdateGroupRequest{}
	if cmd.Name != "" {
		req.Name = &cmd.Name
	}
	if cmd.Description != "" {
		req.Description = &cmd.Description
	}
	if req.Name == nil && req.Description == nil {
		return out.WriteError("admin.groups.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one field must be provided").Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdateGroup(ctx, reqCtx, cmd.GroupKey, req)
	if err != nil {
		return handleCLIError(out, "admin.groups.update", err)
	}

	return out.WriteSuccess("admin.groups.update", result)
}

func (cmd *AdminGroupsMembersListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.members.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListMembers(ctx, reqCtx, cmd.GroupKey, &admin.ListMembersOptions{
		MaxResults: int64(cmd.Limit),
		PageToken:  cmd.PageToken,
		Roles:      cmd.Roles,
		Fields:     cmd.Fields,
		Paginate:   cmd.Paginate,
	})
	if err != nil {
		return handleCLIError(out, "admin.groups.members.list", err)
	}

	return out.WriteSuccess("admin.groups.members.list", result)
}

func (cmd *AdminGroupsMembersAddCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.members.add", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	role := cmd.Role
	if role != "OWNER" && role != "MANAGER" && role != "MEMBER" {
		return out.WriteError("admin.groups.members.add", utils.NewCLIError(utils.ErrCodeInvalidArgument, "role must be OWNER, MANAGER, or MEMBER").Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.AddMember(ctx, reqCtx, cmd.GroupKey, &types.AddMemberRequest{
		Email: cmd.MemberEmail,
		Role:  role,
	})
	if err != nil {
		return handleCLIError(out, "admin.groups.members.add", err)
	}

	return out.WriteSuccess("admin.groups.members.add", result)
}

func (cmd *AdminGroupsMembersRemoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getAdminService(ctx, flags)
	if err != nil {
		return out.WriteError("admin.groups.members.remove", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := admin.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation
	if err := mgr.RemoveMember(ctx, reqCtx, cmd.GroupKey, cmd.MemberKey); err != nil {
		return handleCLIError(out, "admin.groups.members.remove", err)
	}

	return out.WriteSuccess("admin.groups.members.remove", map[string]string{
		"message": fmt.Sprintf("Member %s removed from group %s", cmd.MemberKey, cmd.GroupKey),
	})
}

func parseSuspendedFlag(value string) (*bool, error) {
	switch value {
	case "true":
		result := true
		return &result, nil
	case "false":
		result := false
		return &result, nil
	default:
		return nil, fmt.Errorf("suspended must be true or false")
	}
}

func getAdminService(ctx context.Context, flags types.GlobalFlags) (*adminapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}
	if creds.Type != types.AuthTypeServiceAccount && creds.Type != types.AuthTypeImpersonated {
		return nil, nil, nil, fmt.Errorf("admin operations require service account authentication")
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceAdminDir); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetAdminService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, "", types.RequestTypeListOrSearch)
	return svc, client, reqCtx, nil
}
