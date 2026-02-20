package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/labels"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

type LabelsCmd struct {
	List    LabelsListCmd    `cmd:"" help:"List available labels"`
	Get     LabelsGetCmd     `cmd:"" help:"Get label schema"`
	Create  LabelsCreateCmd  `cmd:"" help:"Create a label (admin only)"`
	Publish LabelsPublishCmd `cmd:"" help:"Publish a label (admin only)"`
	Disable LabelsDisableCmd `cmd:"" help:"Disable a label (admin only)"`
	File    LabelsFileCmd    `cmd:"" help:"File label operations"`
}

type LabelsFileCmd struct {
	List   LabelsFileListCmd   `cmd:"" help:"List labels applied to a file"`
	Apply  LabelsFileApplyCmd  `cmd:"" help:"Apply a label to a file"`
	Update LabelsFileUpdateCmd `cmd:"" help:"Update label fields on a file"`
	Remove LabelsFileRemoveCmd `cmd:"" help:"Remove a label from a file"`
}

type LabelsListCmd struct {
	Customer      string `help:"Customer ID (for admin operations)" name:"customer"`
	View          string `help:"View mode (LABEL_VIEW_BASIC, LABEL_VIEW_FULL)" name:"view"`
	MinimumRole   string `help:"Minimum role (READER, APPLIER, ORGANIZER, EDITOR)" name:"minimum-role"`
	PublishedOnly bool   `help:"Only return published labels" name:"published-only"`
	Limit         int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken     string `help:"Pagination token" name:"page-token"`
	Fields        string `help:"Fields to return" name:"fields"`
}

type LabelsGetCmd struct {
	LabelID        string `arg:"" name:"label-id" help:"Label ID"`
	View           string `help:"View mode (LABEL_VIEW_BASIC, LABEL_VIEW_FULL)" name:"view"`
	UseAdminAccess bool   `help:"Use admin access" name:"use-admin-access"`
	Fields         string `help:"Fields to return" name:"fields"`
}

type LabelsCreateCmd struct {
	Name           string `arg:"" name:"name" help:"Label name"`
	UseAdminAccess bool   `help:"Use admin access" name:"use-admin-access"`
	LanguageCode   string `help:"Language code (e.g., en)" name:"language-code"`
}

type LabelsPublishCmd struct {
	LabelID        string `arg:"" name:"label-id" help:"Label ID"`
	UseAdminAccess bool   `help:"Use admin access" name:"use-admin-access"`
}

type LabelsDisableCmd struct {
	LabelID        string `arg:"" name:"label-id" help:"Label ID"`
	UseAdminAccess bool   `help:"Use admin access" name:"use-admin-access"`
	HideInSearch   bool   `help:"Hide in search" name:"hide-in-search"`
	ShowInApply    bool   `help:"Show in apply" name:"show-in-apply"`
}

type LabelsFileListCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID"`
	View   string `help:"View mode (LABEL_VIEW_BASIC, LABEL_VIEW_FULL)" name:"view"`
	Fields string `help:"Fields to return" name:"fields"`
}

type LabelsFileApplyCmd struct {
	FileID  string `arg:"" name:"file-id" help:"File ID"`
	LabelID string `arg:"" name:"label-id" help:"Label ID"`
	Fields  string `help:"JSON object of field values" name:"fields"`
}

type LabelsFileUpdateCmd struct {
	FileID  string `arg:"" name:"file-id" help:"File ID"`
	LabelID string `arg:"" name:"label-id" help:"Label ID"`
	Fields  string `help:"JSON object of field values" name:"fields"`
}

type LabelsFileRemoveCmd struct {
	FileID  string `arg:"" name:"file-id" help:"File ID"`
	LabelID string `arg:"" name:"label-id" help:"Label ID"`
}

func getLabelsManager(ctx context.Context, flags types.GlobalFlags) (*labels.Manager, *api.Client, *types.RequestContext, *OutputWriter, error) {
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
	mgr := labels.NewManager(client)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)

	return mgr, client, reqCtx, out, nil
}

func (cmd *LabelsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.LabelListOptions{
		Customer:      cmd.Customer,
		View:          cmd.View,
		MinimumRole:   cmd.MinimumRole,
		PublishedOnly: cmd.PublishedOnly,
		Limit:         cmd.Limit,
		PageToken:     cmd.PageToken,
		Fields:        cmd.Fields,
	}

	labelsList, nextPageToken, err := mgr.List(ctx, reqCtx, opts)
	if err != nil {
		return handleCLIError(out, "labels.list", err)
	}

	result := &LabelsListResult{
		Labels:        labelsList,
		NextPageToken: nextPageToken,
	}
	return out.WriteSuccess("labels.list", result)
}

func (cmd *LabelsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.LabelGetOptions{
		View:           cmd.View,
		UseAdminAccess: cmd.UseAdminAccess,
		Fields:         cmd.Fields,
	}

	label, err := mgr.Get(ctx, reqCtx, cmd.LabelID, opts)
	if err != nil {
		return handleCLIError(out, "labels.get", err)
	}

	result := &LabelResult{Label: label}
	return out.WriteSuccess("labels.get", result)
}

func (cmd *LabelsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	label := &types.Label{
		LabelType: types.LabelTypeShared,
		Properties: &types.LabelProperties{
			Title: cmd.Name,
		},
	}

	opts := types.LabelCreateOptions{
		UseAdminAccess: cmd.UseAdminAccess,
		LanguageCode:   cmd.LanguageCode,
	}

	createdLabel, err := mgr.CreateLabel(ctx, reqCtx, label, opts)
	if err != nil {
		return handleCLIError(out, "labels.create", err)
	}

	result := &LabelResult{Label: createdLabel}
	return out.WriteSuccess("labels.create", result)
}

func (cmd *LabelsPublishCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.publish", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.LabelPublishOptions{
		UseAdminAccess: cmd.UseAdminAccess,
	}

	publishedLabel, err := mgr.PublishLabel(ctx, reqCtx, cmd.LabelID, opts)
	if err != nil {
		return handleCLIError(out, "labels.publish", err)
	}

	result := &LabelResult{Label: publishedLabel}
	return out.WriteSuccess("labels.publish", result)
}

func (cmd *LabelsDisableCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.disable", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.LabelDisableOptions{
		UseAdminAccess: cmd.UseAdminAccess,
	}

	if cmd.HideInSearch || cmd.ShowInApply {
		opts.DisabledPolicy = &types.LabelDisabledPolicy{
			HideInSearch: cmd.HideInSearch,
			ShowInApply:  cmd.ShowInApply,
		}
	}

	disabledLabel, err := mgr.DisableLabel(ctx, reqCtx, cmd.LabelID, opts)
	if err != nil {
		return handleCLIError(out, "labels.disable", err)
	}

	result := &LabelResult{Label: disabledLabel}
	return out.WriteSuccess("labels.disable", result)
}

func (cmd *LabelsFileListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.file.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.FileLabelListOptions{
		View:   cmd.View,
		Fields: cmd.Fields,
	}

	fileLabels, err := mgr.ListFileLabels(ctx, reqCtx, cmd.FileID, opts)
	if err != nil {
		return handleCLIError(out, "labels.file.list", err)
	}

	result := &FileLabelsListResult{FileLabels: fileLabels}
	return out.WriteSuccess("labels.file.list", result)
}

func (cmd *LabelsFileApplyCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.file.apply", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.FileLabelApplyOptions{
		Fields: make(map[string]*types.LabelFieldValue),
	}

	if cmd.Fields != "" {
		if err := json.Unmarshal([]byte(cmd.Fields), &opts.Fields); err != nil {
			return out.WriteError("labels.file.apply", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				fmt.Sprintf("Invalid field values JSON: %s", err)).Build())
		}
	}

	fileLabel, err := mgr.ApplyLabel(ctx, reqCtx, cmd.FileID, cmd.LabelID, opts)
	if err != nil {
		return handleCLIError(out, "labels.file.apply", err)
	}

	result := &FileLabelResult{FileLabel: fileLabel}
	return out.WriteSuccess("labels.file.apply", result)
}

func (cmd *LabelsFileUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.file.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	opts := types.FileLabelUpdateOptions{
		Fields: make(map[string]*types.LabelFieldValue),
	}

	if cmd.Fields != "" {
		if err := json.Unmarshal([]byte(cmd.Fields), &opts.Fields); err != nil {
			return out.WriteError("labels.file.update", utils.NewCLIError(utils.ErrCodeInvalidArgument,
				fmt.Sprintf("Invalid field values JSON: %s", err)).Build())
		}
	}

	fileLabel, err := mgr.UpdateLabel(ctx, reqCtx, cmd.FileID, cmd.LabelID, opts)
	if err != nil {
		return handleCLIError(out, "labels.file.update", err)
	}

	result := &FileLabelResult{FileLabel: fileLabel}
	return out.WriteSuccess("labels.file.update", result)
}

func (cmd *LabelsFileRemoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, _, reqCtx, out, err := getLabelsManager(ctx, flags)
	if err != nil {
		return out.WriteError("labels.file.remove", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	err = mgr.RemoveLabel(ctx, reqCtx, cmd.FileID, cmd.LabelID)
	if err != nil {
		return handleCLIError(out, "labels.file.remove", err)
	}

	result := &SuccessResult{Message: fmt.Sprintf("Label %s removed from file %s", cmd.LabelID, cmd.FileID)}
	return out.WriteSuccess("labels.file.remove", result)
}

type LabelsListResult struct {
	Labels        []*types.Label
	NextPageToken string
}

func (r *LabelsListResult) Headers() []string {
	return []string{"ID", "Name", "Type", "State"}
}

func (r *LabelsListResult) Rows() [][]string {
	rows := make([][]string, len(r.Labels))
	for i, label := range r.Labels {
		state := ""
		if label.Lifecycle != nil {
			state = label.Lifecycle.State
		}
		rows[i] = []string{label.ID, label.Name, label.LabelType, state}
	}
	return rows
}

func (r *LabelsListResult) EmptyMessage() string {
	return "No labels found"
}

type LabelResult struct {
	Label *types.Label
}

func (r *LabelResult) Headers() []string {
	return []string{"ID", "Name", "Type", "State", "Fields"}
}

func (r *LabelResult) Rows() [][]string {
	state := ""
	if r.Label.Lifecycle != nil {
		state = r.Label.Lifecycle.State
	}
	fieldCount := fmt.Sprintf("%d", len(r.Label.Fields))
	return [][]string{{r.Label.ID, r.Label.Name, r.Label.LabelType, state, fieldCount}}
}

func (r *LabelResult) EmptyMessage() string {
	return "No label found"
}

type FileLabelsListResult struct {
	FileLabels []*types.FileLabel
}

func (r *FileLabelsListResult) Headers() []string {
	return []string{"Label ID", "Revision ID", "Fields"}
}

func (r *FileLabelsListResult) Rows() [][]string {
	rows := make([][]string, len(r.FileLabels))
	for i, fileLabel := range r.FileLabels {
		fieldCount := fmt.Sprintf("%d", len(fileLabel.Fields))
		rows[i] = []string{fileLabel.ID, fileLabel.RevisionID, fieldCount}
	}
	return rows
}

func (r *FileLabelsListResult) EmptyMessage() string {
	return "No labels found on file"
}

type FileLabelResult struct {
	FileLabel *types.FileLabel
}

func (r *FileLabelResult) Headers() []string {
	return []string{"Label ID", "Revision ID", "Fields"}
}

func (r *FileLabelResult) Rows() [][]string {
	fieldCount := fmt.Sprintf("%d", len(r.FileLabel.Fields))
	return [][]string{{r.FileLabel.ID, r.FileLabel.RevisionID, fieldCount}}
}

func (r *FileLabelResult) EmptyMessage() string {
	return "No label found"
}

type SuccessResult struct {
	Message string
}

func (r *SuccessResult) Headers() []string {
	return []string{"Status"}
}

func (r *SuccessResult) Rows() [][]string {
	return [][]string{{r.Message}}
}

func (r *SuccessResult) EmptyMessage() string {
	return ""
}
