package permissions

import (
	"context"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
)

// AnalyzeFolder analyzes permissions for a folder and optionally its descendants
func (m *Manager) AnalyzeFolder(ctx context.Context, reqCtx *types.RequestContext, folderID string, opts types.AnalyzeOptions) (*types.PermissionAnalysis, error) {
	reqCtx.InvolvedParentIDs = append(reqCtx.InvolvedParentIDs, folderID)

	filesManager := m.client.Service().Files

	folderCall := filesManager.Get(folderID).Fields("id,name,mimeType")
	folderCall = folderCall.SupportsAllDrives(true)
	header := m.client.ResourceKeys().BuildHeader(reqCtx.InvolvedParentIDs)
	if header != "" {
		folderCall.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	folder, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.File, error) {
		return folderCall.Do()
	})
	if err != nil {
		return nil, err
	}

	analysis := &types.PermissionAnalysis{
		FolderID:         folderID,
		FolderName:       folder.Name,
		Recursive:        opts.Recursive,
		RiskDistribution: make(map[string]int),
		PermissionTypes:  make(map[string]int),
		RoleDistribution: make(map[string]int),
	}

	query := fmt.Sprintf("'%s' in parents", folderID)
	if !opts.IncludeTrashed {
		query += " and trashed = false"
	}

	listCall := filesManager.List().Q(query).Fields("files(id,name,mimeType,webViewLink,createdTime,modifiedTime)")
	listCall = m.shaper.ShapeFilesList(listCall, reqCtx)

	fileList, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.FileList, error) {
		return listCall.Do()
	})
	if err != nil {
		return nil, err
	}

	for _, file := range fileList.Files {
		if file.MimeType == "application/vnd.google-apps.folder" {
			analysis.TotalFolders++
		} else {
			analysis.TotalFiles++
		}

		perms, err := m.List(ctx, reqCtx, file.Id, ListOptions{})
		if err != nil {
			continue
		}

		fileInfo := analyzeFilePermissions(file, perms, opts.InternalDomain)

		for _, p := range perms {
			analysis.PermissionTypes[p.Type]++
			analysis.RoleDistribution[p.Role]++
		}

		analysis.RiskDistribution[fileInfo.RiskLevel]++

		if fileInfo.RiskLevel == types.RiskLevelHigh || fileInfo.RiskLevel == types.RiskLevelCritical {
			if file.MimeType == "application/vnd.google-apps.folder" {
				analysis.FoldersWithRisks++
			} else {
				analysis.FilesWithRisks++
			}
		}

		if opts.IncludeDetails {
			if fileInfo.HasPublicAccess {
				analysis.PublicFiles = append(analysis.PublicFiles, fileInfo)
			}
			if fileInfo.HasExternalAccess {
				analysis.ExternalShares = append(analysis.ExternalShares, fileInfo)
			}
			if fileInfo.HasAnyoneWithLink {
				analysis.AnyoneWithLink = append(analysis.AnyoneWithLink, fileInfo)
			}
			if fileInfo.RiskLevel == types.RiskLevelHigh || fileInfo.RiskLevel == types.RiskLevelCritical {
				analysis.HighRiskFiles = append(analysis.HighRiskFiles, fileInfo)
			}
		}

		if opts.MaxFiles > 0 && (analysis.TotalFiles+analysis.TotalFolders) >= opts.MaxFiles {
			break
		}
	}

	if opts.Recursive {
		for _, file := range fileList.Files {
			if file.MimeType == "application/vnd.google-apps.folder" {
				if opts.MaxDepth > 0 {
					opts.MaxDepth--
					if opts.MaxDepth == 0 {
						break
					}
				}
				subAnalysis, err := m.AnalyzeFolder(ctx, reqCtx, file.Id, opts)
				if err == nil {
					analysis.Subfolders = append(analysis.Subfolders, subAnalysis)
				}
			}
		}
	}

	return analysis, nil
}
// GenerateReport generates a detailed permission report for a file or folder
func (m *Manager) GenerateReport(ctx context.Context, reqCtx *types.RequestContext, fileID string, internalDomain string) (*types.PermissionReport, error) {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	fileCall := m.client.Service().Files.Get(fileID).Fields("id,name,mimeType,webViewLink,createdTime,modifiedTime,owners")
	fileCall = m.shaper.ShapeFilesGet(fileCall, reqCtx)

	file, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.File, error) {
		return fileCall.Do()
	})
	if err != nil {
		return nil, err
	}

	perms, err := m.List(ctx, reqCtx, fileID, ListOptions{})
	if err != nil {
		return nil, err
	}

	report := &types.PermissionReport{
		ResourceID:      fileID,
		ResourceName:    file.Name,
		ResourceType:    "file",
		MimeType:        file.MimeType,
		WebViewLink:     file.WebViewLink,
		CreatedTime:     file.CreatedTime,
		ModifiedTime:    file.ModifiedTime,
		InternalDomain:  internalDomain,
		PermissionCount: len(perms),
		Permissions:     make([]*types.PermissionDetail, 0, len(perms)),
	}

	if file.MimeType == "application/vnd.google-apps.folder" {
		report.ResourceType = "folder"
	}

	if len(file.Owners) > 0 {
		report.Owner = file.Owners[0].EmailAddress
	}

	externalDomains := make(map[string]bool)
	riskScore := 0

	for _, p := range perms {
		detail := &types.PermissionDetail{
			ID:           p.ID,
			Type:         p.Type,
			Role:         p.Role,
			EmailAddress: p.EmailAddress,
			Domain:       p.Domain,
			DisplayName:  p.DisplayName,
		}

		if p.Type == "anyone" {
			detail.IsPublic = true
			report.HasPublicAccess = true
			report.HasAnyoneWithLink = true
			detail.RiskLevel = types.RiskLevelCritical
			riskScore += 40
		} else if p.Type == "domain" && p.Domain != internalDomain {
			detail.IsExternal = true
			report.HasExternalAccess = true
			externalDomains[p.Domain] = true
			detail.RiskLevel = types.RiskLevelHigh
			riskScore += 20
		} else if (p.Type == "user" || p.Type == "group") && p.EmailAddress != "" {
			if !isInternalEmail(p.EmailAddress, internalDomain) {
				detail.IsExternal = true
				report.HasExternalAccess = true
				domain := extractDomain(p.EmailAddress)
				if domain != "" {
					externalDomains[domain] = true
				}
				detail.RiskLevel = types.RiskLevelMedium
				riskScore += 10
			} else {
				detail.RiskLevel = types.RiskLevelLow
			}
		}

		switch p.Role {
		case "writer", "organizer":
			if detail.IsPublic || detail.IsExternal {
				riskScore += 10
			}
		case "owner":
			if detail.IsExternal {
				riskScore += 20
			}
		}

		report.Permissions = append(report.Permissions, detail)
	}

	for domain := range externalDomains {
		report.ExternalDomains = append(report.ExternalDomains, domain)
	}

	report.RiskScore = riskScore
	if riskScore >= 60 {
		report.RiskLevel = types.RiskLevelCritical
		report.RiskReasons = append(report.RiskReasons, "Multiple high-risk permissions detected")
	} else if riskScore >= 40 {
		report.RiskLevel = types.RiskLevelHigh
		report.RiskReasons = append(report.RiskReasons, "Public or external access with elevated permissions")
	} else if riskScore >= 20 {
		report.RiskLevel = types.RiskLevelMedium
		report.RiskReasons = append(report.RiskReasons, "External access detected")
	} else {
		report.RiskLevel = types.RiskLevelLow
	}

	if report.HasPublicAccess {
		report.Recommendations = append(report.Recommendations, "Consider removing public access and sharing with specific users")
	}
	if report.HasExternalAccess {
		report.Recommendations = append(report.Recommendations, "Review external shares and ensure they are necessary")
	}
	if len(report.Permissions) > 10 {
		report.Recommendations = append(report.Recommendations, "Consider using groups to simplify permission management")
	}

	return report, nil
}
