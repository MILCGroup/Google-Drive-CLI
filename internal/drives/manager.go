package drives

import (
	"context"

	"github.com/dl-alexandre/gdrive/internal/api"
	"github.com/dl-alexandre/gdrive/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// Manager handles Shared Drive operations
type Manager struct {
	client *api.Client
	shaper *api.RequestShaper
}

// NewManager creates a new drives manager
func NewManager(client *api.Client) *Manager {
	return &Manager{
		client: client,
		shaper: api.NewRequestShaper(client),
	}
}

// SharedDrive represents a Shared Drive with metadata
type SharedDrive struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Kind                  string            `json:"kind,omitempty"`
	ColorRgb              string            `json:"colorRgb,omitempty"`
	BackgroundImageLink   string            `json:"backgroundImageLink,omitempty"`
	ThemeID               string            `json:"themeId,omitempty"`
	CreatedTime           string            `json:"createdTime,omitempty"`
	Hidden                bool              `json:"hidden,omitempty"`
	Capabilities          *DriveCapabilities `json:"capabilities,omitempty"`
	Restrictions          *DriveRestrictions `json:"restrictions,omitempty"`
	OrgUnitID             string            `json:"orgUnitId,omitempty"`
}

// DriveCapabilities represents capabilities for a Shared Drive
type DriveCapabilities struct {
	CanAddChildren                       bool `json:"canAddChildren,omitempty"`
	CanChangeCopyRequiresWriterPermissionRestriction bool `json:"canChangeCopyRequiresWriterPermissionRestriction,omitempty"`
	CanChangeDomainUsersOnlyRestriction  bool `json:"canChangeDomainUsersOnlyRestriction,omitempty"`
	CanChangeDriveBackground             bool `json:"canChangeDriveBackground,omitempty"`
	CanChangeDriveMembersOnlyRestriction bool `json:"canChangeDriveMembersOnlyRestriction,omitempty"`
	CanComment                           bool `json:"canComment,omitempty"`
	CanCopy                              bool `json:"canCopy,omitempty"`
	CanDeleteDrive                       bool `json:"canDeleteDrive,omitempty"`
	CanDownload                          bool `json:"canDownload,omitempty"`
	CanEdit                              bool `json:"canEdit,omitempty"`
	CanListChildren                      bool `json:"canListChildren,omitempty"`
	CanManageMembers                     bool `json:"canManageMembers,omitempty"`
	CanReadRevisions                     bool `json:"canReadRevisions,omitempty"`
	CanRename                            bool `json:"canRename,omitempty"`
	CanRenameDrive                       bool `json:"canRenameDrive,omitempty"`
	CanResetDriveRestrictions            bool `json:"canResetDriveRestrictions,omitempty"`
	CanShare                             bool `json:"canShare,omitempty"`
	CanTrashChildren                     bool `json:"canTrashChildren,omitempty"`
}

// DriveRestrictions represents restrictions for a Shared Drive
type DriveRestrictions struct {
	AdminManagedRestrictions             bool `json:"adminManagedRestrictions,omitempty"`
	CopyRequiresWriterPermission         bool `json:"copyRequiresWriterPermission,omitempty"`
	DomainUsersOnly                      bool `json:"domainUsersOnly,omitempty"`
	DriveMembersOnly                     bool `json:"driveMembersOnly,omitempty"`
}

// ListResult contains shared drive listing results
type ListResult struct {
	Drives        []*SharedDrive `json:"drives"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

// List lists all Shared Drives
func (m *Manager) List(ctx context.Context, reqCtx *types.RequestContext, pageSize int, pageToken string) (*ListResult, error) {
	call := m.client.Service().Drives.List()
	call = m.shaper.ShapeDrivesList(call, reqCtx)

	if pageSize > 0 {
		call = call.PageSize(int64(pageSize))
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.DriveList, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	drives := make([]*SharedDrive, len(result.Drives))
	for i, d := range result.Drives {
		drives[i] = mapDriveToSharedDrive(d)
	}

	return &ListResult{
		Drives:        drives,
		NextPageToken: result.NextPageToken,
	}, nil
}

// Get retrieves a Shared Drive by ID with full metadata
func (m *Manager) Get(ctx context.Context, reqCtx *types.RequestContext, driveID string, fields string) (*SharedDrive, error) {
	call := m.client.Service().Drives.Get(driveID)
	
	// Apply fields mask if specified
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.Drive, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return mapDriveToSharedDrive(result), nil
}

// GetDriveRootID returns the root ID for a Shared Drive, which is the drive ID itself
// This is used as an anchor point for path resolution within the Shared Drive
func (m *Manager) GetDriveRootID(driveID string) string {
	// In Google Drive, the Shared Drive ID itself serves as the root folder ID
	return driveID
}

// mapDriveToSharedDrive converts Google Drive API Drive to internal SharedDrive
func mapDriveToSharedDrive(d *drive.Drive) *SharedDrive {
	sd := &SharedDrive{
		ID:                  d.Id,
		Name:                d.Name,
		Kind:                d.Kind,
		ColorRgb:            d.ColorRgb,
		BackgroundImageLink: d.BackgroundImageLink,
		ThemeID:             d.ThemeId,
		CreatedTime:         d.CreatedTime,
		Hidden:              d.Hidden,
		OrgUnitID:           d.OrgUnitId,
	}

	// Map capabilities
	if d.Capabilities != nil {
		sd.Capabilities = &DriveCapabilities{
			CanAddChildren:                       d.Capabilities.CanAddChildren,
			CanChangeCopyRequiresWriterPermissionRestriction: d.Capabilities.CanChangeCopyRequiresWriterPermissionRestriction,
			CanChangeDomainUsersOnlyRestriction:  d.Capabilities.CanChangeDomainUsersOnlyRestriction,
			CanChangeDriveBackground:             d.Capabilities.CanChangeDriveBackground,
			CanChangeDriveMembersOnlyRestriction: d.Capabilities.CanChangeDriveMembersOnlyRestriction,
			CanComment:                           d.Capabilities.CanComment,
			CanCopy:                              d.Capabilities.CanCopy,
			CanDeleteDrive:                       d.Capabilities.CanDeleteDrive,
			CanDownload:                          d.Capabilities.CanDownload,
			CanEdit:                              d.Capabilities.CanEdit,
			CanListChildren:                      d.Capabilities.CanListChildren,
			CanManageMembers:                     d.Capabilities.CanManageMembers,
			CanReadRevisions:                     d.Capabilities.CanReadRevisions,
			CanRename:                            d.Capabilities.CanRename,
			CanRenameDrive:                       d.Capabilities.CanRenameDrive,
			CanResetDriveRestrictions:            d.Capabilities.CanResetDriveRestrictions,
			CanShare:                             d.Capabilities.CanShare,
			CanTrashChildren:                     d.Capabilities.CanTrashChildren,
		}
	}

	// Map restrictions
	if d.Restrictions != nil {
		sd.Restrictions = &DriveRestrictions{
			AdminManagedRestrictions:     d.Restrictions.AdminManagedRestrictions,
			CopyRequiresWriterPermission: d.Restrictions.CopyRequiresWriterPermission,
			DomainUsersOnly:              d.Restrictions.DomainUsersOnly,
			DriveMembersOnly:             d.Restrictions.DriveMembersOnly,
		}
	}

	return sd
}
