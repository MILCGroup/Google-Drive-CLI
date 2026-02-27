package permissions

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
)

// AuditPublic finds all files with public access (type="anyone")
func (m *Manager) AuditPublic(ctx context.Context, reqCtx *types.RequestContext, opts types.AuditOptions) (*types.AuditResult, error) {
	query := "visibility = 'anyoneCanFind' or visibility = 'anyoneWithLink'"
	return m.auditByQuery(ctx, reqCtx, query, opts, func(perms []*types.Permission) bool {
		for _, p := range perms {
			if p.Type == "anyone" {
				return true
			}
		}
		return false
	})
}

// AuditExternal finds all files shared with external domains
func (m *Manager) AuditExternal(ctx context.Context, reqCtx *types.RequestContext, opts types.AuditOptions) (*types.AuditResult, error) {
	if opts.InternalDomain == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"InternalDomain is required for external audit").Build())
	}

	return m.auditByQuery(ctx, reqCtx, "", opts, func(perms []*types.Permission) bool {
		for _, p := range perms {
			switch p.Type {
			case "user", "group":
				if p.EmailAddress != "" && !isInternalEmail(p.EmailAddress, opts.InternalDomain) {
					return true
				}
			case "domain":
				if p.Domain != "" && p.Domain != opts.InternalDomain {
					return true
				}
			}
		}
		return false
	})
}

// AuditAnyoneWithLink finds all files with "anyone with link" access
func (m *Manager) AuditAnyoneWithLink(ctx context.Context, reqCtx *types.RequestContext, opts types.AuditOptions) (*types.AuditResult, error) {
	query := "visibility = 'anyoneWithLink'"
	return m.auditByQuery(ctx, reqCtx, query, opts, func(perms []*types.Permission) bool {
		for _, p := range perms {
			if p.Type == "anyone" {
				return true
			}
		}
		return false
	})
}

// AuditUser finds all files accessible by a specific user email
func (m *Manager) AuditUser(ctx context.Context, reqCtx *types.RequestContext, email string, opts types.AuditOptions) (*types.AuditResult, error) {
	if email == "" {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Email is required for user audit").Build())
	}

	return m.auditByQuery(ctx, reqCtx, "", opts, func(perms []*types.Permission) bool {
		for _, p := range perms {
			if p.EmailAddress == email {
				return true
			}
		}
		return false
	})
}
func (m *Manager) auditByQuery(ctx context.Context, reqCtx *types.RequestContext, baseQuery string, opts types.AuditOptions, filter func([]*types.Permission) bool) (*types.AuditResult, error) {
	query := baseQuery
	if opts.FolderID != "" {
		if query != "" {
			query += " and "
		}
		query += fmt.Sprintf("'%s' in parents", opts.FolderID)
		reqCtx.InvolvedParentIDs = append(reqCtx.InvolvedParentIDs, opts.FolderID)
	}
	if !opts.IncludeTrashed {
		if query != "" {
			query += " and "
		}
		query += "trashed = false"
	}
	if opts.MimeType != "" {
		if query != "" {
			query += " and "
		}
		query += fmt.Sprintf("mimeType = '%s'", opts.MimeType)
	}
	if opts.Query != "" {
		if query != "" {
			query += " and "
		}
		query += opts.Query
	}

	listCall := m.client.Service().Files.List()
	listCall = m.shaper.ShapeFilesList(listCall, reqCtx)
	if query != "" {
		listCall = listCall.Q(query)
	}
	listCall = listCall.Fields("files(id,name,mimeType,webViewLink,createdTime,modifiedTime)")
	if opts.PageSize > 0 {
		listCall = listCall.PageSize(int64(opts.PageSize))
	}
	if opts.PageToken != "" {
		listCall = listCall.PageToken(opts.PageToken)
	}

	fileList, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.FileList, error) {
		return listCall.Do()
	})
	if err != nil {
		return nil, err
	}

	result := &types.AuditResult{
		Files:   make([]*types.FilePermissionInfo, 0),
		Summary: make(map[string]int),
	}

	for _, file := range fileList.Files {
		perms, err := m.List(ctx, reqCtx, file.Id, ListOptions{})
		if err != nil {
			continue
		}

		if filter(perms) {
			fileInfo := &types.FilePermissionInfo{
				FileID:          file.Id,
				FileName:        file.Name,
				MimeType:        file.MimeType,
				WebViewLink:     file.WebViewLink,
				CreatedTime:     file.CreatedTime,
				ModifiedTime:    file.ModifiedTime,
				PermissionCount: len(perms),
			}

			if opts.IncludePermissions {
				fileInfo.Permissions = perms
			}

			fileInfo = analyzeFilePermissions(file, perms, opts.InternalDomain)
			result.Files = append(result.Files, fileInfo)
			result.Summary[fileInfo.RiskLevel]++
		}
	}

	result.TotalCount = len(result.Files)

	if result.TotalCount == 0 {
		result.RiskLevel = types.RiskLevelLow
	} else {
		criticalCount := result.Summary[types.RiskLevelCritical]
		highCount := result.Summary[types.RiskLevelHigh]
		if criticalCount > 0 {
			result.RiskLevel = types.RiskLevelCritical
		} else if highCount > 0 {
			result.RiskLevel = types.RiskLevelHigh
		} else if result.Summary[types.RiskLevelMedium] > 0 {
			result.RiskLevel = types.RiskLevelMedium
		} else {
			result.RiskLevel = types.RiskLevelLow
		}
	}

	return result, nil
}
func analyzeFilePermissions(file *drive.File, perms []*types.Permission, internalDomain string) *types.FilePermissionInfo {
	info := &types.FilePermissionInfo{
		FileID:          file.Id,
		FileName:        file.Name,
		MimeType:        file.MimeType,
		WebViewLink:     file.WebViewLink,
		CreatedTime:     file.CreatedTime,
		ModifiedTime:    file.ModifiedTime,
		Permissions:     perms,
		PermissionCount: len(perms),
		RiskReasons:     make([]string, 0),
		ExternalDomains: make([]string, 0),
	}

	externalDomains := make(map[string]bool)
	riskScore := 0

	for _, p := range perms {
		if p.Type == "anyone" {
			info.HasPublicAccess = true
			info.HasAnyoneWithLink = true
			info.RiskReasons = append(info.RiskReasons, "Public access enabled")
			riskScore += 40
		} else if p.Type == "domain" && p.Domain != internalDomain {
			info.HasExternalAccess = true
			externalDomains[p.Domain] = true
			info.RiskReasons = append(info.RiskReasons, fmt.Sprintf("Shared with external domain: %s", p.Domain))
			riskScore += 20
		} else if (p.Type == "user" || p.Type == "group") && p.EmailAddress != "" {
			if !isInternalEmail(p.EmailAddress, internalDomain) {
				info.HasExternalAccess = true
				domain := extractDomain(p.EmailAddress)
				if domain != "" {
					externalDomains[domain] = true
				}
				info.RiskReasons = append(info.RiskReasons, fmt.Sprintf("Shared with external user: %s", p.EmailAddress))
				riskScore += 10
			}
		}

		if p.Role == "writer" || p.Role == "organizer" {
			if info.HasPublicAccess || info.HasExternalAccess {
				riskScore += 10
			}
		}
	}

	for domain := range externalDomains {
		info.ExternalDomains = append(info.ExternalDomains, domain)
	}

	if riskScore >= 60 {
		info.RiskLevel = types.RiskLevelCritical
	} else if riskScore >= 40 {
		info.RiskLevel = types.RiskLevelHigh
	} else if riskScore >= 20 {
		info.RiskLevel = types.RiskLevelMedium
	} else {
		info.RiskLevel = types.RiskLevelLow
	}

	return info
}
