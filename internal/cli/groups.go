package cli

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	groupsmgr "github.com/milcgroup/gdrv/internal/groups"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
)

// GroupsCmd provides Cloud Identity Groups operations.
// These are distinct from 'admin groups' which uses the Admin SDK Directory API.
type GroupsCmd struct {
	List    GroupsListCmd    `cmd:"" help:"List Cloud Identity groups (uses Cloud Identity API, not Admin SDK)"`
	Members GroupsMembersCmd `cmd:"" help:"List memberships of a Cloud Identity group (uses Cloud Identity API, not Admin SDK)"`
}

// GroupsListCmd lists Cloud Identity groups under a customer.
type GroupsListCmd struct {
	Customer  string `help:"Customer ID (e.g. C01234abc) — required" name:"customer" required:""`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

// GroupsMembersCmd lists memberships of a Cloud Identity group.
type GroupsMembersCmd struct {
	GroupName string `arg:"" name:"group-name" help:"Group resource name (e.g. groups/abc123)"`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

// Run executes the Cloud Identity groups list command.
func (cmd *GroupsListCmd) Run(globals *Globals) error {
	ctx := context.Background()
	flags := globals.ToGlobalFlags()

	mgr, reqCtx, out, err := getGroupsManager(ctx, flags)
	if err != nil {
		writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
		return writer.WriteError("groups.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch
	parent := fmt.Sprintf("customers/%s", cmd.Customer)

	if cmd.Paginate {
		var allGroups []types.CloudIdentityGroup
		pageToken := cmd.PageToken
		for {
			result, nextToken, err := mgr.ListGroups(ctx, reqCtx, parent, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "groups.list", err)
			}
			allGroups = append(allGroups, result.Groups...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("groups.list", &types.CloudIdentityGroupList{Groups: allGroups})
	}

	result, nextToken, err := mgr.ListGroups(ctx, reqCtx, parent, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "groups.list", err)
	}

	response := map[string]interface{}{
		"groups": result.Groups,
	}
	if nextToken != "" {
		response["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("groups.list", response)
}

// Run executes the Cloud Identity group members list command.
func (cmd *GroupsMembersCmd) Run(globals *Globals) error {
	ctx := context.Background()
	flags := globals.ToGlobalFlags()

	mgr, reqCtx, out, err := getGroupsManager(ctx, flags)
	if err != nil {
		writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
		return writer.WriteError("groups.members", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allMembers []types.CloudIdentityMember
		pageToken := cmd.PageToken
		for {
			result, nextToken, err := mgr.ListMembers(ctx, reqCtx, cmd.GroupName, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "groups.members", err)
			}
			allMembers = append(allMembers, result.Members...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("groups.members", &types.CloudIdentityMemberList{Members: allMembers})
	}

	result, nextToken, err := mgr.ListMembers(ctx, reqCtx, cmd.GroupName, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "groups.members", err)
	}

	response := map[string]interface{}{
		"members": result.Members,
	}
	if nextToken != "" {
		response["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("groups.members", response)
}

// getGroupsManager creates a Cloud Identity Groups manager with authenticated services.
func getGroupsManager(ctx context.Context, flags types.GlobalFlags) (*groupsmgr.Manager, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, out, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceCloudIdentity); err != nil {
		return nil, nil, out, err
	}

	svc, err := authMgr.GetCloudIdentityService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	mgr := groupsmgr.NewManager(client, svc)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, out, nil
}
