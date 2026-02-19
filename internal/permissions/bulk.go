package permissions

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
)

// BulkRemovePublic removes public access from all files in a folder
func (m *Manager) BulkRemovePublic(ctx context.Context, reqCtx *types.RequestContext, opts types.BulkOptions) (*types.BulkOperationResult, error) {
	if opts.FolderID == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"FolderID is required for bulk operations").Build())
	}

	result := &types.BulkOperationResult{
		DryRun: opts.DryRun,
	}

	files, err := m.findFilesInFolder(ctx, reqCtx, opts)
	if err != nil {
		return nil, err
	}

	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		files = files[:opts.MaxFiles]
	}

	result.TotalFiles = len(files)

	for _, file := range files {
		perms, err := m.List(ctx, reqCtx, file.Id, ListOptions{})
		if err != nil {
			result.FailureCount++
			result.FailedFiles = append(result.FailedFiles, &types.BulkOperationItem{
				FileID:       file.Id,
				FileName:     file.Name,
				Operation:    "remove_public",
				Status:       "failure",
				ErrorMessage: err.Error(),
			})
			if !opts.ContinueOnError {
				return result, err
			}
			continue
		}

		hasPublic := false
		for _, p := range perms {
			if p.Type == "anyone" {
				hasPublic = true
				if opts.DryRun {
					result.SuccessCount++
					result.SuccessfulFiles = append(result.SuccessfulFiles, &types.BulkOperationItem{
						FileID:    file.Id,
						FileName:  file.Name,
						Operation: "remove_public",
						Status:    "success",
					})
				} else {
					err := m.Delete(ctx, reqCtx, file.Id, p.ID, DeleteOptions{})
					if err != nil {
						result.FailureCount++
						result.FailedFiles = append(result.FailedFiles, &types.BulkOperationItem{
							FileID:       file.Id,
							FileName:     file.Name,
							Operation:    "remove_public",
							Status:       "failure",
							ErrorMessage: err.Error(),
						})
						if !opts.ContinueOnError {
							return result, err
						}
					} else {
						result.SuccessCount++
						result.SuccessfulFiles = append(result.SuccessfulFiles, &types.BulkOperationItem{
							FileID:    file.Id,
							FileName:  file.Name,
							Operation: "remove_public",
							Status:    "success",
						})
					}
				}
				break
			}
		}

		if !hasPublic {
			result.SkippedCount++
			result.SkippedFiles = append(result.SkippedFiles, &types.BulkOperationItem{
				FileID:    file.Id,
				FileName:  file.Name,
				Operation: "remove_public",
				Status:    "skipped",
			})
		}
	}

	return result, nil
}
// BulkUpdateRole updates permissions from one role to another in a folder
func (m *Manager) BulkUpdateRole(ctx context.Context, reqCtx *types.RequestContext, fromRole, toRole string, opts types.BulkOptions) (*types.BulkOperationResult, error) {
	if opts.FolderID == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"FolderID is required for bulk operations").Build())
	}

	if fromRole == "" || toRole == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Both fromRole and toRole are required").Build())
	}

	result := &types.BulkOperationResult{
		DryRun: opts.DryRun,
	}

	files, err := m.findFilesInFolder(ctx, reqCtx, opts)
	if err != nil {
		return nil, err
	}

	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		files = files[:opts.MaxFiles]
	}

	result.TotalFiles = len(files)

	for _, file := range files {
		perms, err := m.List(ctx, reqCtx, file.Id, ListOptions{})
		if err != nil {
			result.FailureCount++
			result.FailedFiles = append(result.FailedFiles, &types.BulkOperationItem{
				FileID:       file.Id,
				FileName:     file.Name,
				Operation:    "update_role",
				Status:       "failure",
				ErrorMessage: err.Error(),
			})
			if !opts.ContinueOnError {
				return result, err
			}
			continue
		}

		updated := false
		for _, p := range perms {
			if p.Role == fromRole {
				if opts.DryRun {
					result.SuccessCount++
					result.SuccessfulFiles = append(result.SuccessfulFiles, &types.BulkOperationItem{
						FileID:    file.Id,
						FileName:  file.Name,
						Operation: "update_role",
						Status:    "success",
					})
					updated = true
				} else {
					_, err := m.Update(ctx, reqCtx, file.Id, p.ID, UpdateOptions{Role: toRole})
					if err != nil {
						result.FailureCount++
						result.FailedFiles = append(result.FailedFiles, &types.BulkOperationItem{
							FileID:       file.Id,
							FileName:     file.Name,
							Operation:    "update_role",
							Status:       "failure",
							ErrorMessage: err.Error(),
						})
						if !opts.ContinueOnError {
							return result, err
						}
					} else {
						result.SuccessCount++
						result.SuccessfulFiles = append(result.SuccessfulFiles, &types.BulkOperationItem{
							FileID:    file.Id,
							FileName:  file.Name,
							Operation: "update_role",
							Status:    "success",
						})
						updated = true
					}
				}
			}
		}

		if !updated {
			result.SkippedCount++
			result.SkippedFiles = append(result.SkippedFiles, &types.BulkOperationItem{
				FileID:    file.Id,
				FileName:  file.Name,
				Operation: "update_role",
				Status:    "skipped",
			})
		}
	}

	return result, nil
}
// SearchByEmail finds all files accessible by a specific email address
func (m *Manager) SearchByEmail(ctx context.Context, reqCtx *types.RequestContext, opts types.SearchOptions) (*types.AuditResult, error) {
	if opts.Email == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Email is required for search").Build())
	}

	auditOpts := types.AuditOptions{
		FolderID:           opts.FolderID,
		Recursive:          opts.Recursive,
		IncludeTrashed:     opts.IncludeTrashed,
		MimeType:           opts.MimeType,
		Query:              opts.Query,
		PageSize:           opts.PageSize,
		PageToken:          opts.PageToken,
		IncludePermissions: opts.IncludePermissions,
	}

	return m.AuditUser(ctx, reqCtx, opts.Email, auditOpts)
}
// SearchByRole finds all files with a specific permission role
func (m *Manager) SearchByRole(ctx context.Context, reqCtx *types.RequestContext, opts types.SearchOptions) (*types.AuditResult, error) {
	if opts.Role == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Role is required for search").Build())
	}

	auditOpts := types.AuditOptions{
		FolderID:           opts.FolderID,
		Recursive:          opts.Recursive,
		IncludeTrashed:     opts.IncludeTrashed,
		MimeType:           opts.MimeType,
		Query:              opts.Query,
		PageSize:           opts.PageSize,
		PageToken:          opts.PageToken,
		IncludePermissions: opts.IncludePermissions,
	}

	return m.auditByQuery(ctx, reqCtx, "", auditOpts, func(perms []*types.Permission) bool {
		for _, p := range perms {
			if p.Role == opts.Role {
				return true
			}
		}
		return false
	})
}
func (m *Manager) findFilesInFolder(ctx context.Context, reqCtx *types.RequestContext, opts types.BulkOptions) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents", opts.FolderID)
	if !opts.IncludeTrashed {
		query += " and trashed = false"
	}
	if opts.MimeType != "" {
		query += fmt.Sprintf(" and mimeType = '%s'", opts.MimeType)
	}
	if opts.Query != "" {
		query += " and " + opts.Query
	}

	listCall := m.client.Service().Files.List().Q(query).Fields("files(id,name,mimeType)")
	listCall = m.shaper.ShapeFilesList(listCall, reqCtx)

	fileList, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.FileList, error) {
		return listCall.Do()
	})
	if err != nil {
		return nil, err
	}

	files := fileList.Files

	if opts.Recursive {
		for _, file := range fileList.Files {
			if file.MimeType == "application/vnd.google-apps.folder" {
				subOpts := opts
				subOpts.FolderID = file.Id
				subFiles, err := m.findFilesInFolder(ctx, reqCtx, subOpts)
				if err == nil {
					files = append(files, subFiles...)
				}
			}
		}
	}

	return files, nil
}
