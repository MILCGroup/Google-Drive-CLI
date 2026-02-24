package cli

import (
	"context"

	"github.com/milcgroup/gdrv/internal/activity"
	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
)

type ActivityCmd struct {
	Query ActivityQueryCmd `cmd:"" help:"Query Drive activity"`
}

type ActivityQueryCmd struct {
	FileID       string `help:"Filter by specific file ID" name:"file-id"`
	FolderID     string `help:"Filter by folder ID (includes descendants)" name:"folder-id"`
	AncestorName string `help:"Filter by ancestor folder (e.g., folders/123)" name:"ancestor-name"`
	StartTime    string `help:"Start of time range (RFC3339 format)" name:"start-time"`
	EndTime      string `help:"End of time range (RFC3339 format)" name:"end-time"`
	ActionTypes  string `help:"Comma-separated action types (edit,comment,share,permission_change,move,delete,restore,create,rename)" name:"action-types"`
	User         string `help:"Filter by user email" name:"user"`
	Limit        int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken    string `help:"Pagination token" name:"page-token"`
	Fields       string `help:"Fields to return" name:"fields"`
}

func getActivityManager(ctx context.Context, flags types.GlobalFlags) (*activity.Manager, *api.Client, *types.RequestContext, *OutputWriter, error) {
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
	mgr := activity.NewManager(client)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	return mgr, client, reqCtx, out, nil
}

func (cmd *ActivityQueryCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getActivityManager(ctx, flags)
	if err != nil {
		return out.WriteError("activity.query", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.QueryOptions{
		FileID:       cmd.FileID,
		FolderID:     cmd.FolderID,
		AncestorName: cmd.AncestorName,
		StartTime:    cmd.StartTime,
		EndTime:      cmd.EndTime,
		ActionTypes:  cmd.ActionTypes,
		User:         cmd.User,
		Limit:        cmd.Limit,
		PageToken:    cmd.PageToken,
		Fields:       cmd.Fields,
	}

	activities, err := mgr.Query(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "activity.query", err)
	}

	result := &ActivityQueryResult{Activities: activities}
	return out.WriteSuccess("activity.query", result)
}

type ActivityQueryResult struct {
	Activities []types.Activity
}

func (r *ActivityQueryResult) Headers() []string {
	return []string{"Timestamp", "Action", "Actor", "Target"}
}

func (r *ActivityQueryResult) Rows() [][]string {
	rows := make([][]string, len(r.Activities))
	for i, activity := range r.Activities {
		timestamp := activity.Timestamp.Format("2006-01-02 15:04:05")
		action := activity.PrimaryActionDetail.Type

		actor := "unknown"
		if len(activity.Actors) > 0 {
			if activity.Actors[0].User != nil {
				actor = activity.Actors[0].User.Email
			} else {
				actor = activity.Actors[0].Type
			}
		}

		target := "unknown"
		if len(activity.Targets) > 0 {
			if activity.Targets[0].DriveItem != nil {
				target = activity.Targets[0].DriveItem.Title
				if target == "" {
					target = activity.Targets[0].DriveItem.Name
				}
			} else {
				target = activity.Targets[0].Type
			}
		}

		rows[i] = []string{timestamp, action, actor, target}
	}
	return rows
}

func (r *ActivityQueryResult) EmptyMessage() string {
	return "No activity found"
}
