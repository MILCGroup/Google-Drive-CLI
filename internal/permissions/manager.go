// Package permissions provides Drive API permission management functionality.
// It handles creating, listing, updating, and deleting permissions on Drive files and folders.
// The package supports all permission types (user, group, domain, anyone) and roles
// (reader, commenter, writer, organizer, owner), including Shared Drive-specific behaviors.
package permissions

import (
	"context"
	"fmt"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/safety"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
)

// Manager handles permission operations for Google Drive files and folders.
// It provides methods for creating, listing, updating, and deleting permissions
// with support for:
//   - All permission types: user, group, domain, anyone
//   - All permission roles: reader, commenter, writer, organizer, owner
//   - Notification emails and custom messages
//   - Ownership transfer
//   - Public link creation with discoverability control
//   - Shared Drive-specific permission behaviors
//   - Domain admin access for Workspace environments
//   - Resource key handling for link-shared files
type Manager struct {
	client *api.Client
	shaper *api.RequestShaper
}

// NewManager creates a new permission manager
func NewManager(client *api.Client) *Manager {
	return &Manager{
		client: client,
		shaper: api.NewRequestShaper(client),
	}
}

// CreateOptions configures permission creation.
//
// Type specifies the permission type:
//   - "user": Permission for a specific user (requires EmailAddress)
//   - "group": Permission for a group (requires EmailAddress)
//   - "domain": Permission for an entire domain (requires Domain)
//   - "anyone": Public permission (link-sharing)
//
// Role specifies the permission level:
//   - "reader": Can view and download
//   - "commenter": Can view, download, and comment
//   - "writer": Can view, download, comment, and edit
//   - "organizer": Can organize files in Shared Drives (Shared Drive only)
//   - "owner": Full ownership (requires TransferOwnership=true)
//
// Requirements:
//   - Requirement 4.1: Support user/group/domain/anyone permission types
//   - Requirement 4.2: Support reader/commenter/writer/organizer roles
//   - Requirement 4.3: Support sendNotificationEmail and emailMessage parameters
//   - Requirement 4.4: Support transferOwnership parameter
//   - Requirement 4.5: Support allowFileDiscovery for public permissions
//   - Requirement 4.6: Support useDomainAdminAccess for Workspace environments
type CreateOptions struct {
	Type                  string // user, group, domain, anyone
	Role                  string // reader, commenter, writer, organizer, owner
	EmailAddress          string // Required for user and group types
	Domain                string // Required for domain type
	SendNotificationEmail bool   // Send email notification to recipients
	EmailMessage          string // Custom message to include in notification email
	TransferOwnership     bool   // Transfer ownership (only valid when Role="owner")
	AllowFileDiscovery    bool   // Allow file to be discovered via search (anyone type only)
	UseDomainAdminAccess  bool   // Use domain administrator access for Workspace environments
}

// UpdateOptions configures permission updates.
//
// Requirements:
//   - Requirement 4.10: Modify existing permission levels
//   - Requirement 4.6: Support useDomainAdminAccess for Workspace environments
type UpdateOptions struct {
	Role                 string // New role: reader, commenter, writer, organizer
	UseDomainAdminAccess bool   // Use domain administrator access
}

// DeleteOptions configures permission deletion.
//
// Requirements:
//   - Requirement 4.9: Revoke access for specified user or group
//   - Requirement 4.6: Support useDomainAdminAccess for Workspace environments
type DeleteOptions struct {
	UseDomainAdminAccess bool // Use domain administrator access
}

// ListOptions configures permission listing.
//
// Requirements:
//   - Requirement 4.7: Return all current permissions
//   - Requirement 4.8: Support pagination with pageToken and nextPageToken
//   - Requirement 4.6: Support useDomainAdminAccess for Workspace environments
type ListOptions struct {
	UseDomainAdminAccess bool // Use domain administrator access
	PageSize             int  // Number of permissions per page (0 = API default)
}

// List lists all permissions for a file or folder.
// It automatically handles pagination to retrieve all permissions.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - opts: Options for listing (domain admin access, page size)
//
// Returns all permissions or an error. Supports Shared Drive files when
// reqCtx.DriveID is set.
//
// Requirements:
//   - Requirement 4.7: Return all current permissions
//   - Requirement 4.8: Iterate through all pages using pageToken
//   - Requirement 4.14: Include supportsAllDrives=true parameter
func (m *Manager) List(ctx context.Context, reqCtx *types.RequestContext, fileID string, opts ListOptions) ([]*types.Permission, error) {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	call := m.client.Service().Permissions.List(fileID)
	call = m.shaper.ShapePermissionsList(call, reqCtx)
	call = call.Fields("permissions(id,type,role,emailAddress,domain,displayName),nextPageToken")

	if opts.UseDomainAdminAccess {
		call = call.UseDomainAdminAccess(true)
	}
	if opts.PageSize > 0 {
		call = call.PageSize(int64(opts.PageSize))
	}
	call = call.Context(ctx)

	var allPerms []*types.Permission
	pageToken := ""

	for {
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.PermissionList, error) {
			return call.Do()
		})
		if err != nil {
			return nil, err
		}

		for _, p := range result.Permissions {
			allPerms = append(allPerms, convertPermission(p))
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allPerms, nil
}

// Create creates a new permission on a file or folder.
//
// Supports all permission types (user, group, domain, anyone) and roles
// (reader, commenter, writer, organizer, owner). Can send notification emails,
// transfer ownership, and control discoverability.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - opts: Permission creation options (type, role, email, etc.)
//
// Returns the created permission or an error. Handles Shared Drive-specific
// behaviors when reqCtx.DriveID is set.
//
// Error Handling:
//   - Returns ErrCodePolicyViolation for domain policy restrictions
//   - Returns ErrCodeSharingRestricted for invalid sharing requests
//   - Returns structured errors with capability indicators
//
// Requirements:
//   - Requirement 4.1: Support user/group/domain/anyone permission types
//   - Requirement 4.2: Support reader/commenter/writer/organizer roles
//   - Requirement 4.3: Support sendNotificationEmail and emailMessage
//   - Requirement 4.4: Support transferOwnership parameter
//   - Requirement 4.5: Support allowFileDiscovery parameter
//   - Requirement 4.6: Support useDomainAdminAccess parameter
//   - Requirement 4.12: Return policy violation errors with capability indicators
//   - Requirement 4.13: Validate ownership transfer restrictions
//   - Requirement 4.14: Include supportsAllDrives=true parameter
func (m *Manager) Create(ctx context.Context, reqCtx *types.RequestContext, fileID string, opts CreateOptions) (*types.Permission, error) {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	perm := &drive.Permission{
		Type: opts.Type,
		Role: opts.Role,
	}
	if opts.EmailAddress != "" {
		perm.EmailAddress = opts.EmailAddress
	}
	if opts.Domain != "" {
		perm.Domain = opts.Domain
	}
	if opts.Type == "anyone" {
		perm.AllowFileDiscovery = opts.AllowFileDiscovery
	}

	call := m.client.Service().Permissions.Create(fileID, perm)
	call = m.shaper.ShapePermissionsCreate(call, reqCtx)
	call = call.SendNotificationEmail(opts.SendNotificationEmail)
	call = call.Fields("id,type,role,emailAddress,domain,displayName")

	if opts.EmailMessage != "" {
		call = call.EmailMessage(opts.EmailMessage)
	}
	if opts.TransferOwnership {
		call = call.TransferOwnership(true)
	}
	if opts.UseDomainAdminAccess {
		call = call.UseDomainAdminAccess(true)
	}
	call = call.Context(ctx)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.Permission, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertPermission(result), nil
}

// Update updates an existing permission's role.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - permissionID: The ID of the permission to update
//   - opts: Update options (new role, domain admin access)
//
// Returns the updated permission or an error.
//
// Requirements:
//   - Requirement 4.10: Modify existing permission levels
//   - Requirement 4.14: Include supportsAllDrives=true parameter
func (m *Manager) Update(ctx context.Context, reqCtx *types.RequestContext, fileID, permissionID string, opts UpdateOptions) (*types.Permission, error) {
	return m.UpdateWithSafety(ctx, reqCtx, fileID, permissionID, opts, safety.Default(), nil)
}

// UpdateWithSafety updates an existing permission's role with safety controls.
// Supports dry-run mode and confirmation.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
func (m *Manager) UpdateWithSafety(ctx context.Context, reqCtx *types.RequestContext, fileID, permissionID string, opts UpdateOptions, safetyOpts safety.SafetyOptions, recorder safety.DryRunRecorder) (*types.Permission, error) {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	// Get current permission for dry-run display
	if safetyOpts.DryRun && recorder != nil {
		safety.RecordPermissionUpdate(recorder, fileID, fileID, permissionID, opts.Role)
		// Return a placeholder permission
		return &types.Permission{
			ID:   permissionID,
			Role: opts.Role,
		}, nil
	}

	perm := &drive.Permission{
		Role: opts.Role,
	}

	call := m.client.Service().Permissions.Update(fileID, permissionID, perm)
	call = call.SupportsAllDrives(true)
	call = call.Fields("id,type,role,emailAddress,domain,displayName")

	if opts.UseDomainAdminAccess {
		call = call.UseDomainAdminAccess(true)
	}

	header := m.client.ResourceKeys().BuildHeader(reqCtx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}
	call = call.Context(ctx)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.Permission, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertPermission(result), nil
}

// Delete removes a permission from a file or folder.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - permissionID: The ID of the permission to delete
//   - opts: Delete options (domain admin access)
//
// Returns an error if the deletion fails.
//
// Requirements:
//   - Requirement 4.9: Revoke access for specified user or group
//   - Requirement 4.14: Include supportsAllDrives=true parameter
func (m *Manager) Delete(ctx context.Context, reqCtx *types.RequestContext, fileID, permissionID string, opts DeleteOptions) error {
	return m.DeleteWithSafety(ctx, reqCtx, fileID, permissionID, opts, safety.Default(), nil)
}

// DeleteWithSafety removes a permission from a file or folder with safety controls.
// Supports dry-run mode, confirmation, and idempotency.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
//   - Requirement 13.2: Support --force flag to skip confirmations
func (m *Manager) DeleteWithSafety(ctx context.Context, reqCtx *types.RequestContext, fileID, permissionID string, opts DeleteOptions, safetyOpts safety.SafetyOptions, recorder safety.DryRunRecorder) error {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	// Get permission details for confirmation
	perm, err := m.Get(ctx, reqCtx, fileID, permissionID)
	if err != nil && !safetyOpts.DryRun {
		return err
	}

	// Dry-run mode: record operation without executing
	if safetyOpts.DryRun && recorder != nil {
		safety.RecordPermissionDelete(recorder, fileID, fileID, permissionID)
		return nil
	}

	// Confirmation for destructive operations
	if safetyOpts.ShouldConfirm() {
		displayName := permissionID
		if perm != nil && perm.EmailAddress != "" {
			displayName = perm.EmailAddress
		} else if perm != nil && perm.DisplayName != "" {
			displayName = perm.DisplayName
		}

		confirmed, err := safety.Confirm(
			fmt.Sprintf("About to revoke permission for '%s'. Continue?", displayName),
			safetyOpts,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			return utils.NewAppError(utils.NewCLIError(utils.ErrCodeCancelled, "Operation cancelled by user").Build())
		}
	}

	call := m.client.Service().Permissions.Delete(fileID, permissionID)
	call = call.SupportsAllDrives(true)

	if opts.UseDomainAdminAccess {
		call = call.UseDomainAdminAccess(true)
	}

	header := m.client.ResourceKeys().BuildHeader(reqCtx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}
	call = call.Context(ctx)

	_, err = api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (interface{}, error) {
		return nil, call.Do()
	})
	return err
}

// CreatePublicLink creates a public "anyone with link" permission.
//
// This is a convenience method for creating public sharing links.
// For domain-only sharing, use Create with Type="domain".
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - role: Permission role (reader, commenter, writer)
//   - allowDiscovery: If true, file can be discovered via search
//
// Returns the created permission or an error.
//
// Requirements:
//   - Requirement 4.11: Support "anyone with link" sharing
//   - Requirement 4.5: Support allowFileDiscovery for discoverability control
func (m *Manager) CreatePublicLink(ctx context.Context, reqCtx *types.RequestContext, fileID string, role string, allowDiscovery bool) (*types.Permission, error) {
	return m.Create(ctx, reqCtx, fileID, CreateOptions{
		Type:               "anyone",
		Role:               role,
		AllowFileDiscovery: allowDiscovery,
	})
}

// Get retrieves a specific permission by ID.
//
// Parameters:
//   - ctx: Context for request cancellation
//   - reqCtx: Request context with profile, drive context, and trace ID
//   - fileID: The ID of the file or folder
//   - permissionID: The ID of the permission to retrieve
//
// Returns the permission or an error if not found.
//
// Requirements:
//   - Requirement 4.14: Include supportsAllDrives=true parameter
func (m *Manager) Get(ctx context.Context, reqCtx *types.RequestContext, fileID, permissionID string) (*types.Permission, error) {
	reqCtx.InvolvedFileIDs = append(reqCtx.InvolvedFileIDs, fileID)

	call := m.client.Service().Permissions.Get(fileID, permissionID)
	call = call.SupportsAllDrives(true)
	call = call.Fields("id,type,role,emailAddress,domain,displayName")

	header := m.client.ResourceKeys().BuildHeader(reqCtx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}
	call = call.Context(ctx)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*drive.Permission, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertPermission(result), nil
}

func convertPermission(p *drive.Permission) *types.Permission {
	return &types.Permission{
		ID:           p.Id,
		Type:         p.Type,
		Role:         p.Role,
		EmailAddress: p.EmailAddress,
		Domain:       p.Domain,
		DisplayName:  p.DisplayName,
	}
}

func isInternalEmail(email, internalDomain string) bool {
	if internalDomain == "" {
		return true
	}
	domain := extractDomain(email)
	return domain == internalDomain
}

func extractDomain(email string) string {
	parts := []rune(email)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '@' {
			return string(parts[i+1:])
		}
	}
	return ""
}
