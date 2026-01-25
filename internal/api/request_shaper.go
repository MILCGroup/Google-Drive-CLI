package api

import (
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
)

// RequestShaper applies proper parameters based on request context
type RequestShaper struct {
	client *Client
}

// NewRequestShaper creates a new request shaper
func NewRequestShaper(client *Client) *RequestShaper {
	return &RequestShaper{client: client}
}

// ShapeFilesGet applies parameters to files.get request
func (s *RequestShaper) ShapeFilesGet(call *drive.FilesGetCall, ctx *types.RequestContext) *drive.FilesGetCall {
	call = call.SupportsAllDrives(true)

	// Add resource keys header if available
	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeFilesList applies parameters to files.list request
func (s *RequestShaper) ShapeFilesList(call *drive.FilesListCall, ctx *types.RequestContext) *drive.FilesListCall {
	call = call.SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true)

	// Set corpora based on context
	if ctx.DriveID != "" {
		call = call.Corpora("drive").DriveId(ctx.DriveID)
	} else {
		call = call.Corpora("user")
	}

	return call
}

// ShapeFilesCreate applies parameters to files.create request
func (s *RequestShaper) ShapeFilesCreate(call *drive.FilesCreateCall, ctx *types.RequestContext) *drive.FilesCreateCall {
	call = call.SupportsAllDrives(true)

	// Add resource keys for parent folders
	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedParentIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeFilesUpdate applies parameters to files.update request
func (s *RequestShaper) ShapeFilesUpdate(call *drive.FilesUpdateCall, ctx *types.RequestContext) *drive.FilesUpdateCall {
	call = call.SupportsAllDrives(true)

	// Include both file and parent resource keys
	allIDs := append(ctx.InvolvedFileIDs, ctx.InvolvedParentIDs...)
	header := s.client.ResourceKeys().BuildHeader(allIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeFilesDelete applies parameters to files.delete request
func (s *RequestShaper) ShapeFilesDelete(call *drive.FilesDeleteCall, ctx *types.RequestContext) *drive.FilesDeleteCall {
	call = call.SupportsAllDrives(true)

	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeFilesCopy applies parameters to files.copy request
func (s *RequestShaper) ShapeFilesCopy(call *drive.FilesCopyCall, ctx *types.RequestContext) *drive.FilesCopyCall {
	call = call.SupportsAllDrives(true)

	allIDs := append(ctx.InvolvedFileIDs, ctx.InvolvedParentIDs...)
	header := s.client.ResourceKeys().BuildHeader(allIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapePermissionsList applies parameters to permissions.list request
func (s *RequestShaper) ShapePermissionsList(call *drive.PermissionsListCall, ctx *types.RequestContext) *drive.PermissionsListCall {
	call = call.SupportsAllDrives(true)

	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapePermissionsCreate applies parameters to permissions.create request
func (s *RequestShaper) ShapePermissionsCreate(call *drive.PermissionsCreateCall, ctx *types.RequestContext) *drive.PermissionsCreateCall {
	call = call.SupportsAllDrives(true)

	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapePermissionsUpdate applies parameters to permissions.update request
func (s *RequestShaper) ShapePermissionsUpdate(call *drive.PermissionsUpdateCall, ctx *types.RequestContext) *drive.PermissionsUpdateCall {
	call = call.SupportsAllDrives(true)

	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapePermissionsDelete applies parameters to permissions.delete request
func (s *RequestShaper) ShapePermissionsDelete(call *drive.PermissionsDeleteCall, ctx *types.RequestContext) *drive.PermissionsDeleteCall {
	call = call.SupportsAllDrives(true)

	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeRevisionsList applies parameters to revisions.list request
func (s *RequestShaper) ShapeRevisionsList(call *drive.RevisionsListCall, ctx *types.RequestContext) *drive.RevisionsListCall {
	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeRevisionsGet applies parameters to revisions.get request
func (s *RequestShaper) ShapeRevisionsGet(call *drive.RevisionsGetCall, ctx *types.RequestContext) *drive.RevisionsGetCall {
	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeRevisionsUpdate applies parameters to revisions.update request
func (s *RequestShaper) ShapeRevisionsUpdate(call *drive.RevisionsUpdateCall, ctx *types.RequestContext) *drive.RevisionsUpdateCall {
	header := s.client.ResourceKeys().BuildHeader(ctx.InvolvedFileIDs)
	if header != "" {
		call.Header().Set("X-Goog-Drive-Resource-Keys", header)
	}

	return call
}

// ShapeDrivesList applies parameters to drives.list request
func (s *RequestShaper) ShapeDrivesList(call *drive.DrivesListCall, ctx *types.RequestContext) *drive.DrivesListCall {
	// No special shaping needed for drives.list
	return call
}
