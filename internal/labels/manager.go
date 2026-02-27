package labels

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/drivelabels/v2"
	"google.golang.org/api/googleapi"
)

type Manager struct {
	client *api.Client
}

func NewManager(client *api.Client) *Manager {
	return &Manager{
		client: client,
	}
}

func (m *Manager) List(ctx context.Context, reqCtx *types.RequestContext, opts types.LabelListOptions) ([]*types.Label, string, error) {
	service, err := drivelabels.NewService(ctx)
	if err != nil {
		return nil, "", utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to create Drive Labels service: %s", err)).Build())
	}

	call := service.Labels.List()

	if opts.Customer != "" {
		call = call.Customer(opts.Customer)
	}

	if opts.View != "" {
		call = call.View(opts.View)
	}

	if opts.MinimumRole != "" {
		call = call.MinimumRole(opts.MinimumRole)
	}

	if opts.PublishedOnly {
		call = call.PublishedOnly(opts.PublishedOnly)
	}

	if opts.Limit > 0 {
		call = call.PageSize(int64(opts.Limit))
	}

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	if opts.Fields != "" {
		call = call.Fields(googleapi.Field(opts.Fields))
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drivelabels.GoogleAppsDriveLabelsV2ListLabelsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	labels := make([]*types.Label, 0, len(result.Labels))
	for _, label := range result.Labels {
		labels = append(labels, convertLabel(label))
	}

	return labels, result.NextPageToken, nil
}

func (m *Manager) Get(ctx context.Context, reqCtx *types.RequestContext, labelID string, opts types.LabelGetOptions) (*types.Label, error) {
	service, err := drivelabels.NewService(ctx)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to create Drive Labels service: %s", err)).Build())
	}

	call := service.Labels.Get(labelID)

	if opts.View != "" {
		call = call.View(opts.View)
	}

	if opts.UseAdminAccess {
		call = call.UseAdminAccess(opts.UseAdminAccess)
	}

	if opts.Fields != "" {
		call = call.Fields(googleapi.Field(opts.Fields))
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drivelabels.GoogleAppsDriveLabelsV2Label, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertLabel(result), nil
}

func (m *Manager) ListFileLabels(ctx context.Context, reqCtx *types.RequestContext, fileID string, opts types.FileLabelListOptions) ([]*types.FileLabel, error) {
	driveService := m.client.Service()

	call := driveService.Files.Get(fileID).SupportsAllDrives(true)

	fields := "labelInfo"
	if opts.Fields != "" {
		fields = opts.Fields
	}
	call = call.Fields(googleapi.Field(fields))

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.File, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	if result.LabelInfo == nil || len(result.LabelInfo.Labels) == 0 {
		return []*types.FileLabel{}, nil
	}

	fileLabels := make([]*types.FileLabel, 0, len(result.LabelInfo.Labels))
	for _, label := range result.LabelInfo.Labels {
		fileLabels = append(fileLabels, convertDriveLabel(label))
	}

	return fileLabels, nil
}

func (m *Manager) ApplyLabel(ctx context.Context, reqCtx *types.RequestContext, fileID string, labelID string, opts types.FileLabelApplyOptions) (*types.FileLabel, error) {
	driveService := m.client.Service()

	modifyRequest := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:            labelID,
				FieldModifications: convertFieldModificationsToDrive(opts.Fields),
			},
		},
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.ModifyLabelsResponse, error) {
		return driveService.Files.ModifyLabels(fileID, modifyRequest).Do()
	})
	if err != nil {
		return nil, err
	}

	if len(result.ModifiedLabels) == 0 {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown, "No labels were applied").Build())
	}

	return convertDriveLabel(result.ModifiedLabels[0]), nil
}

func (m *Manager) UpdateLabel(ctx context.Context, reqCtx *types.RequestContext, fileID string, labelID string, opts types.FileLabelUpdateOptions) (*types.FileLabel, error) {
	driveService := m.client.Service()

	modifyRequest := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:            labelID,
				FieldModifications: convertFieldModificationsToDrive(opts.Fields),
			},
		},
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.ModifyLabelsResponse, error) {
		return driveService.Files.ModifyLabels(fileID, modifyRequest).Do()
	})
	if err != nil {
		return nil, err
	}

	if len(result.ModifiedLabels) == 0 {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown, "No labels were updated").Build())
	}

	return convertDriveLabel(result.ModifiedLabels[0]), nil
}

func (m *Manager) RemoveLabel(ctx context.Context, reqCtx *types.RequestContext, fileID string, labelID string) error {
	driveService := m.client.Service()

	modifyRequest := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:     labelID,
				RemoveLabel: true,
			},
		},
	}

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.ModifyLabelsResponse, error) {
		return driveService.Files.ModifyLabels(fileID, modifyRequest).Do()
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) CreateLabel(ctx context.Context, reqCtx *types.RequestContext, label *types.Label, opts types.LabelCreateOptions) (*types.Label, error) {
	service, err := drivelabels.NewService(ctx)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to create Drive Labels service: %s", err)).Build())
	}

	apiLabel := convertToAPILabel(label)

	call := service.Labels.Create(apiLabel)

	if opts.UseAdminAccess {
		call = call.UseAdminAccess(opts.UseAdminAccess)
	}

	if opts.LanguageCode != "" {
		call = call.LanguageCode(opts.LanguageCode)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drivelabels.GoogleAppsDriveLabelsV2Label, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertLabel(result), nil
}

func (m *Manager) PublishLabel(ctx context.Context, reqCtx *types.RequestContext, labelID string, opts types.LabelPublishOptions) (*types.Label, error) {
	service, err := drivelabels.NewService(ctx)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to create Drive Labels service: %s", err)).Build())
	}

	publishRequest := &drivelabels.GoogleAppsDriveLabelsV2PublishLabelRequest{
		UseAdminAccess: opts.UseAdminAccess,
	}

	if opts.WriteControl != nil {
		publishRequest.WriteControl = &drivelabels.GoogleAppsDriveLabelsV2WriteControl{
			RequiredRevisionId: opts.WriteControl.RequiredRevisionID,
		}
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drivelabels.GoogleAppsDriveLabelsV2Label, error) {
		return service.Labels.Publish(labelID, publishRequest).Do()
	})
	if err != nil {
		return nil, err
	}

	return convertLabel(result), nil
}

func (m *Manager) DisableLabel(ctx context.Context, reqCtx *types.RequestContext, labelID string, opts types.LabelDisableOptions) (*types.Label, error) {
	service, err := drivelabels.NewService(ctx)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to create Drive Labels service: %s", err)).Build())
	}

	disableRequest := &drivelabels.GoogleAppsDriveLabelsV2DisableLabelRequest{
		UseAdminAccess: opts.UseAdminAccess,
	}

	if opts.WriteControl != nil {
		disableRequest.WriteControl = &drivelabels.GoogleAppsDriveLabelsV2WriteControl{
			RequiredRevisionId: opts.WriteControl.RequiredRevisionID,
		}
	}

	if opts.DisabledPolicy != nil {
		disableRequest.DisabledPolicy = &drivelabels.GoogleAppsDriveLabelsV2LifecycleDisabledPolicy{
			HideInSearch: opts.DisabledPolicy.HideInSearch,
			ShowInApply:  opts.DisabledPolicy.ShowInApply,
		}
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drivelabels.GoogleAppsDriveLabelsV2Label, error) {
		return service.Labels.Disable(labelID, disableRequest).Do()
	})
	if err != nil {
		return nil, err
	}

	return convertLabel(result), nil
}
