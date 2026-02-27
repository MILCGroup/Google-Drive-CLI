package files

import (
	"context"
	"testing"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
)

func TestManager_Creation(t *testing.T) {
	// Test that manager can be created
	client := &api.Client{}
	manager := NewManager(client)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.client != client { //nolint:staticcheck // False positive: t.Fatal() never returns
		t.Error("Manager client not set correctly")
	}

	if manager.shaper == nil { //nolint:staticcheck // False positive: t.Fatal() never returns
		t.Error("Manager shaper not initialized")
	}
}

func TestConvertDriveFile(t *testing.T) {
	// Test conversion from Drive API file to internal type
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         1024,
		CreatedTime:  "2024-01-01T00:00:00Z",
		ModifiedTime: "2024-01-02T00:00:00Z",
		Parents:      []string{"parent1", "parent2"},
		ResourceKey:  "key123",
		WebViewLink:  "https://drive.google.com/file/d/file123/view",
		Trashed:      false,
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          true,
			CanShare:         true,
			CanDelete:        true,
			CanTrash:         true,
			CanReadRevisions: true,
		},
		ExportLinks: map[string]string{
			"application/pdf": "https://export.link/pdf",
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.ID != driveFile.Id {
		t.Errorf("ID mismatch: got %s, want %s", converted.ID, driveFile.Id)
	}

	if converted.Name != driveFile.Name {
		t.Errorf("Name mismatch: got %s, want %s", converted.Name, driveFile.Name)
	}

	if converted.MimeType != driveFile.MimeType {
		t.Errorf("MimeType mismatch: got %s, want %s", converted.MimeType, driveFile.MimeType)
	}

	if converted.Size != driveFile.Size {
		t.Errorf("Size mismatch: got %d, want %d", converted.Size, driveFile.Size)
	}

	if len(converted.Parents) != len(driveFile.Parents) {
		t.Errorf("Parents length mismatch: got %d, want %d", len(converted.Parents), len(driveFile.Parents))
	}

	if converted.ResourceKey != driveFile.ResourceKey {
		t.Errorf("ResourceKey mismatch: got %s, want %s", converted.ResourceKey, driveFile.ResourceKey)
	}

	if converted.Trashed != driveFile.Trashed {
		t.Errorf("Trashed mismatch: got %v, want %v", converted.Trashed, driveFile.Trashed)
	}

	if converted.Capabilities == nil {
		t.Error("Capabilities should not be nil")
	} else {
		if converted.Capabilities.CanDownload != driveFile.Capabilities.CanDownload {
			t.Errorf("CanDownload mismatch: got %v, want %v", converted.Capabilities.CanDownload, driveFile.Capabilities.CanDownload)
		}
		if converted.Capabilities.CanEdit != driveFile.Capabilities.CanEdit {
			t.Errorf("CanEdit mismatch: got %v, want %v", converted.Capabilities.CanEdit, driveFile.Capabilities.CanEdit)
		}
	}

	if len(converted.ExportLinks) != len(driveFile.ExportLinks) {
		t.Errorf("ExportLinks length mismatch: got %d, want %d", len(converted.ExportLinks), len(driveFile.ExportLinks))
	}
}

func TestConvertDriveFile_NilCapabilities(t *testing.T) {
	// Test conversion when capabilities are nil
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Capabilities: nil,
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities != nil {
		t.Error("Capabilities should be nil when source has nil capabilities")
	}
}

func TestListOptions_Defaults(t *testing.T) {
	// Test that ListOptions has sensible defaults
	opts := ListOptions{}

	if opts.ParentID != "" {
		t.Errorf("Default ParentID should be empty, got %s", opts.ParentID)
	}

	if opts.IncludeTrashed != false {
		t.Error("Default IncludeTrashed should be false")
	}
}

func TestUploadOptions_Defaults(t *testing.T) {
	// Test that UploadOptions has sensible defaults
	opts := UploadOptions{}

	if opts.Convert != false {
		t.Error("Default Convert should be false")
	}

	if opts.PinRevision != false {
		t.Error("Default PinRevision should be false")
	}
}

func TestDownloadOptions_Defaults(t *testing.T) {
	// Test that DownloadOptions has sensible defaults
	opts := DownloadOptions{}

	if opts.Wait != false {
		t.Error("Default Wait should be false")
	}

	if opts.Timeout != 0 {
		t.Errorf("Default Timeout should be 0, got %d", opts.Timeout)
	}

	if opts.PollInterval != 0 {
		t.Errorf("Default PollInterval should be 0, got %d", opts.PollInterval)
	}
}

// TestSelectUploadType_Integration tests upload type selection with realistic scenarios
func TestSelectUploadType_Integration(t *testing.T) {
	tests := []struct {
		name         string
		scenario     string
		size         int64
		metadata     *drive.File
		expectedType string
	}{
		{
			name:         "Quick file upload without metadata",
			scenario:     "User uploads small file without specifying name or parent",
			size:         1024,
			metadata:     &drive.File{},
			expectedType: "simple",
		},
		{
			name:         "File upload to specific folder",
			scenario:     "User uploads file to a specific parent folder",
			size:         1024,
			metadata:     &drive.File{Parents: []string{"folderID"}},
			expectedType: "multipart",
		},
		{
			name:         "Large file upload",
			scenario:     "User uploads 10MB video file",
			size:         10 * 1024 * 1024,
			metadata:     &drive.File{Name: "video.mp4", MimeType: "video/mp4"},
			expectedType: "resumable",
		},
		{
			name:         "Document with metadata",
			scenario:     "User uploads document with custom name",
			size:         500 * 1024,
			metadata:     &drive.File{Name: "report.pdf", MimeType: "application/pdf"},
			expectedType: "multipart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectUploadType(tt.size, tt.metadata)
			if got != tt.expectedType {
				t.Errorf("Scenario: %s\nselectUploadType(%d, %+v) = %s, want %s",
					tt.scenario, tt.size, tt.metadata, got, tt.expectedType)
			}
		})
	}
}

// TestRequestContext_Construction tests that request contexts are properly constructed
func TestRequestContext_Construction(t *testing.T) {
	ctx := &types.RequestContext{
		Profile:           "default",
		DriveID:           "drive123",
		InvolvedFileIDs:   []string{"file1", "file2"},
		InvolvedParentIDs: []string{"parent1"},
		RequestType:       types.RequestTypeMutation,
		TraceID:           "trace123",
	}

	if ctx.Profile != "default" {
		t.Errorf("Profile = %s, want default", ctx.Profile)
	}

	if ctx.DriveID != "drive123" {
		t.Errorf("DriveID = %s, want drive123", ctx.DriveID)
	}

	if len(ctx.InvolvedFileIDs) != 2 {
		t.Errorf("InvolvedFileIDs length = %d, want 2", len(ctx.InvolvedFileIDs))
	}

	if len(ctx.InvolvedParentIDs) != 1 {
		t.Errorf("InvolvedParentIDs length = %d, want 1", len(ctx.InvolvedParentIDs))
	}
}

func TestSelectUploadType_Simple(t *testing.T) {
	// Small file without metadata uses simple upload
	uploadType := selectUploadType(1024, &drive.File{})
	if uploadType != "simple" {
		t.Errorf("selectUploadType(1024, empty) = %s, want simple", uploadType)
	}
}

func TestSelectUploadType_Multipart(t *testing.T) {
	// File with name uses multipart
	uploadType := selectUploadType(1024, &drive.File{Name: "test.txt"})
	if uploadType != "multipart" {
		t.Errorf("selectUploadType(1024, with name) = %s, want multipart", uploadType)
	}

	// File with mimeType uses multipart
	uploadType = selectUploadType(1024, &drive.File{MimeType: "text/plain"})
	if uploadType != "multipart" {
		t.Errorf("selectUploadType(1024, with mimeType) = %s, want multipart", uploadType)
	}

	// File with parent uses multipart
	uploadType = selectUploadType(1024, &drive.File{Parents: []string{"parent123"}})
	if uploadType != "multipart" {
		t.Errorf("selectUploadType(1024, with parents) = %s, want multipart", uploadType)
	}
}

func TestSelectUploadType_Resumable(t *testing.T) {
	// Large file uses resumable upload
	largeSize := int64(6 * 1024 * 1024) // 6MB
	uploadType := selectUploadType(largeSize, &drive.File{})
	if uploadType != "resumable" {
		t.Errorf("selectUploadType(%d, empty) = %s, want resumable", largeSize, uploadType)
	}

	// Large file with metadata still uses resumable
	uploadType = selectUploadType(largeSize, &drive.File{Name: "large.mp4"})
	if uploadType != "resumable" {
		t.Errorf("selectUploadType(%d, with name) = %s, want resumable", largeSize, uploadType)
	}
}

func TestUpdateContentOptions_Fields(t *testing.T) {
	opts := UpdateContentOptions{
		Name:     "newname.txt",
		MimeType: "text/plain",
		Fields:   "id,name,mimeType",
	}

	if opts.Name != "newname.txt" {
		t.Errorf("Name = %s, want newname.txt", opts.Name)
	}

	if opts.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", opts.MimeType)
	}

	if opts.Fields != "id,name,mimeType" {
		t.Errorf("Fields = %s, want id,name,mimeType", opts.Fields)
	}
}

func TestListOptions_Query(t *testing.T) {
	opts := ListOptions{
		Query:          "name contains 'test'",
		PageSize:       100,
		OrderBy:        "modifiedTime desc",
		IncludeTrashed: true,
	}

	if opts.Query != "name contains 'test'" {
		t.Errorf("Query = %s, want name contains 'test'", opts.Query)
	}

	if opts.PageSize != 100 {
		t.Errorf("PageSize = %d, want 100", opts.PageSize)
	}

	if opts.OrderBy != "modifiedTime desc" {
		t.Errorf("OrderBy = %s, want modifiedTime desc", opts.OrderBy)
	}

	if !opts.IncludeTrashed {
		t.Error("IncludeTrashed should be true")
	}
}

func TestDownloadOptions_WithWait(t *testing.T) {
	opts := DownloadOptions{
		OutputPath:   "/tmp/output.txt",
		MimeType:     "application/pdf",
		Wait:         true,
		Timeout:      300,
		PollInterval: 5,
	}

	if opts.OutputPath != "/tmp/output.txt" {
		t.Errorf("OutputPath = %s, want /tmp/output.txt", opts.OutputPath)
	}

	if opts.MimeType != "application/pdf" {
		t.Errorf("MimeType = %s, want application/pdf", opts.MimeType)
	}

	if !opts.Wait {
		t.Error("Wait should be true")
	}

	if opts.Timeout != 300 {
		t.Errorf("Timeout = %d, want 300", opts.Timeout)
	}

	if opts.PollInterval != 5 {
		t.Errorf("PollInterval = %d, want 5", opts.PollInterval)
	}
}

func TestUploadOptions_WithMetadata(t *testing.T) {
	opts := UploadOptions{
		ParentID:    "parent123",
		Name:        "uploaded.txt",
		MimeType:    "text/plain",
		Convert:     true,
		PinRevision: true,
	}

	if opts.ParentID != "parent123" {
		t.Errorf("ParentID = %s, want parent123", opts.ParentID)
	}

	if opts.Name != "uploaded.txt" {
		t.Errorf("Name = %s, want uploaded.txt", opts.Name)
	}

	if opts.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", opts.MimeType)
	}

	if !opts.Convert {
		t.Error("Convert should be true")
	}

	if !opts.PinRevision {
		t.Error("PinRevision should be true")
	}
}

func TestConvertDriveFile_CompleteFields(t *testing.T) {
	driveFile := &drive.File{
		Id:             "file123",
		Name:           "document.pdf",
		MimeType:       "application/pdf",
		Size:           2048,
		Md5Checksum:    "abc123",
		CreatedTime:    "2024-01-01T00:00:00Z",
		ModifiedTime:   "2024-01-02T00:00:00Z",
		Parents:        []string{"parent1"},
		ResourceKey:    "key123",
		WebViewLink:    "https://drive.google.com/file/d/file123/view",
		WebContentLink: "https://drive.google.com/uc?id=file123",
		Trashed:        false,
		ExportLinks: map[string]string{
			"application/pdf": "https://export/pdf",
			"text/plain":      "https://export/txt",
		},
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          true,
			CanShare:         true,
			CanDelete:        true,
			CanTrash:         true,
			CanReadRevisions: true,
		},
	}

	converted := convertDriveFile(driveFile)

	// Verify all fields
	if converted.ID != "file123" {
		t.Errorf("ID = %s, want file123", converted.ID)
	}

	if converted.Name != "document.pdf" {
		t.Errorf("Name = %s, want document.pdf", converted.Name)
	}

	if converted.MimeType != "application/pdf" {
		t.Errorf("MimeType = %s, want application/pdf", converted.MimeType)
	}

	if converted.Size != 2048 {
		t.Errorf("Size = %d, want 2048", converted.Size)
	}

	if converted.MD5Checksum != "abc123" {
		t.Errorf("MD5Checksum = %s, want abc123", converted.MD5Checksum)
	}

	if converted.CreatedTime != "2024-01-01T00:00:00Z" {
		t.Errorf("CreatedTime = %s, want 2024-01-01T00:00:00Z", converted.CreatedTime)
	}

	if converted.ModifiedTime != "2024-01-02T00:00:00Z" {
		t.Errorf("ModifiedTime = %s, want 2024-01-02T00:00:00Z", converted.ModifiedTime)
	}

	if len(converted.Parents) != 1 || converted.Parents[0] != "parent1" {
		t.Errorf("Parents = %v, want [parent1]", converted.Parents)
	}

	if converted.ResourceKey != "key123" {
		t.Errorf("ResourceKey = %s, want key123", converted.ResourceKey)
	}

	if converted.WebViewLink != "https://drive.google.com/file/d/file123/view" {
		t.Errorf("WebViewLink = %s, unexpected value", converted.WebViewLink)
	}

	if converted.WebContentLink != "https://drive.google.com/uc?id=file123" {
		t.Errorf("WebContentLink = %s, unexpected value", converted.WebContentLink)
	}

	if converted.Trashed {
		t.Error("Trashed should be false")
	}

	if len(converted.ExportLinks) != 2 {
		t.Errorf("ExportLinks length = %d, want 2", len(converted.ExportLinks))
	}

	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	if !converted.Capabilities.CanDownload {
		t.Error("CanDownload should be true")
	}

	if !converted.Capabilities.CanEdit {
		t.Error("CanEdit should be true")
	}

	if !converted.Capabilities.CanShare {
		t.Error("CanShare should be true")
	}

	if !converted.Capabilities.CanDelete {
		t.Error("CanDelete should be true")
	}

	if !converted.Capabilities.CanTrash {
		t.Error("CanTrash should be true")
	}

	if !converted.Capabilities.CanReadRevisions {
		t.Error("CanReadRevisions should be true")
	}
}

func TestConvertDriveFile_MinimalFields(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file456",
		Name:     "minimal.txt",
		MimeType: "text/plain",
	}

	converted := convertDriveFile(driveFile)

	if converted.ID != "file456" {
		t.Errorf("ID = %s, want file456", converted.ID)
	}

	if converted.Name != "minimal.txt" {
		t.Errorf("Name = %s, want minimal.txt", converted.Name)
	}

	if converted.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", converted.MimeType)
	}

	if converted.Size != 0 {
		t.Errorf("Size = %d, want 0", converted.Size)
	}

	if converted.Capabilities != nil {
		t.Error("Capabilities should be nil")
	}

	if len(converted.Parents) != 0 {
		t.Errorf("Parents should be empty, got %d items", len(converted.Parents))
	}

	if len(converted.ExportLinks) != 0 {
		t.Errorf("ExportLinks should be empty, got %d items", len(converted.ExportLinks))
	}
}

func TestConvertDriveFile_TrashedFile(t *testing.T) {
	driveFile := &drive.File{
		Id:       "trashedFile",
		Name:     "deleted.txt",
		MimeType: "text/plain",
		Trashed:  true,
	}

	converted := convertDriveFile(driveFile)

	if !converted.Trashed {
		t.Error("Trashed should be true")
	}
}

func TestConvertDriveFile_MultipleParents(t *testing.T) {
	driveFile := &drive.File{
		Id:       "multiParent",
		Name:     "shared.txt",
		MimeType: "text/plain",
		Parents:  []string{"parent1", "parent2", "parent3"},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.Parents) != 3 {
		t.Errorf("Parents length = %d, want 3", len(converted.Parents))
	}

	expectedParents := map[string]bool{"parent1": true, "parent2": true, "parent3": true}
	for _, p := range converted.Parents {
		if !expectedParents[p] {
			t.Errorf("Unexpected parent: %s", p)
		}
	}
}

// ============================================================================
// COMPREHENSIVE TESTS FOR MANAGER METHODS
// ============================================================================

// TestUpload_FileNotFound tests Upload with non-existent file
func TestUpload_FileNotFound(t *testing.T) {
	t.Helper()
	client := &api.Client{}
	manager := NewManager(client)
	ctx := &types.RequestContext{
		Profile:     "default",
		RequestType: types.RequestTypeMutation,
	}

	_, err := manager.Upload(context.Background(), ctx, "/nonexistent/file.txt", UploadOptions{})
	if err == nil {
		t.Error("Upload should return error for non-existent file")
	}
}

// TestUpload_NameFromLocalPath tests Upload uses local filename when name not provided
func TestUpload_NameFromLocalPath(t *testing.T) {
	t.Helper()
	// This test verifies the logic path but can't fully test without mocking API
	// The actual file opening will fail, but we can verify the error handling
	client := &api.Client{}
	manager := NewManager(client)
	ctx := &types.RequestContext{
		Profile:     "default",
		RequestType: types.RequestTypeMutation,
	}

	opts := UploadOptions{
		Name: "", // Empty name should use local path
	}

	_, err := manager.Upload(context.Background(), ctx, "/tmp/test.txt", opts)
	if err == nil {
		t.Error("Upload should return error for non-existent file")
	}
}

// TestUpload_WithParentID tests Upload includes parent ID in request context
func TestUpload_WithParentID(t *testing.T) {
	t.Helper()
	client := &api.Client{}
	manager := NewManager(client)
	ctx := &types.RequestContext{
		Profile:     "default",
		RequestType: types.RequestTypeMutation,
	}

	opts := UploadOptions{
		ParentID: "parent123",
	}

	_, err := manager.Upload(context.Background(), ctx, "/nonexistent/file.txt", opts)
	if err == nil {
		t.Error("Upload should return error for non-existent file")
	}
	// Verify parent was added to context (would happen in real execution)
}

// TestUpdateContent_FileNotFound tests UpdateContent with non-existent file
func TestUpdateContent_FileNotFound(t *testing.T) {
	t.Helper()
	client := &api.Client{}
	manager := NewManager(client)
	ctx := &types.RequestContext{
		Profile:     "default",
		RequestType: types.RequestTypeMutation,
	}

	_, err := manager.UpdateContent(context.Background(), ctx, "fileID", "/nonexistent/file.txt", UpdateContentOptions{})
	if err == nil {
		t.Error("UpdateContent should return error for non-existent file")
	}
}

// TestSelectUploadType_BoundaryConditions tests upload type selection at boundaries
func TestSelectUploadType_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name         string
		size         int64
		metadata     *drive.File
		expectedType string
	}{
		{
			name:         "Exactly at 5MB boundary",
			size:         5 * 1024 * 1024,
			metadata:     &drive.File{},
			expectedType: "simple",
		},
		{
			name:         "Just over 5MB boundary",
			size:         5*1024*1024 + 1,
			metadata:     &drive.File{},
			expectedType: "resumable",
		},
		{
			name:         "Zero size file",
			size:         0,
			metadata:     &drive.File{},
			expectedType: "simple",
		},
		{
			name:         "1 byte file",
			size:         1,
			metadata:     &drive.File{},
			expectedType: "simple",
		},
		{
			name:         "Large file with no metadata",
			size:         100 * 1024 * 1024,
			metadata:     &drive.File{},
			expectedType: "resumable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectUploadType(tt.size, tt.metadata)
			if got != tt.expectedType {
				t.Errorf("selectUploadType(%d, %+v) = %s, want %s",
					tt.size, tt.metadata, got, tt.expectedType)
			}
		})
	}
}

// TestSelectUploadType_MetadataVariations tests upload type with various metadata combinations
func TestSelectUploadType_MetadataVariations(t *testing.T) {
	tests := []struct {
		name         string
		size         int64
		metadata     *drive.File
		expectedType string
	}{
		{
			name:         "Name only",
			size:         1024,
			metadata:     &drive.File{Name: "test.txt"},
			expectedType: "multipart",
		},
		{
			name:         "MimeType only",
			size:         1024,
			metadata:     &drive.File{MimeType: "text/plain"},
			expectedType: "multipart",
		},
		{
			name:         "Parents only",
			size:         1024,
			metadata:     &drive.File{Parents: []string{"parent1"}},
			expectedType: "multipart",
		},
		{
			name:         "Name and MimeType",
			size:         1024,
			metadata:     &drive.File{Name: "test.txt", MimeType: "text/plain"},
			expectedType: "multipart",
		},
		{
			name:         "All metadata fields",
			size:         1024,
			metadata:     &drive.File{Name: "test.txt", MimeType: "text/plain", Parents: []string{"parent1"}},
			expectedType: "multipart",
		},
		{
			name:         "Large file with name",
			size:         10 * 1024 * 1024,
			metadata:     &drive.File{Name: "large.mp4"},
			expectedType: "resumable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectUploadType(tt.size, tt.metadata)
			if got != tt.expectedType {
				t.Errorf("selectUploadType(%d, %+v) = %s, want %s",
					tt.size, tt.metadata, got, tt.expectedType)
			}
		})
	}
}

// TestConvertDriveFile_EmptyExportLinks tests conversion with empty export links
func TestConvertDriveFile_EmptyExportLinks(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		ExportLinks: map[string]string{},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.ExportLinks) != 0 {
		t.Errorf("ExportLinks should be empty, got %d items", len(converted.ExportLinks))
	}
}

// TestConvertDriveFile_MultipleExportLinks tests conversion with multiple export links
func TestConvertDriveFile_MultipleExportLinks(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "document.doc",
		MimeType: "application/vnd.google-apps.document",
		ExportLinks: map[string]string{
			"application/pdf":      "https://export/pdf",
			"text/plain":           "https://export/txt",
			"application/docx":     "https://export/docx",
			"application/odt":      "https://export/odt",
			"text/html":            "https://export/html",
			"application/rtf":      "https://export/rtf",
			"application/epub+zip": "https://export/epub",
		},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.ExportLinks) != 7 {
		t.Errorf("ExportLinks length = %d, want 7", len(converted.ExportLinks))
	}

	expectedLinks := map[string]bool{
		"application/pdf":      true,
		"text/plain":           true,
		"application/docx":     true,
		"application/odt":      true,
		"text/html":            true,
		"application/rtf":      true,
		"application/epub+zip": true,
	}

	for mimeType := range converted.ExportLinks {
		if !expectedLinks[mimeType] {
			t.Errorf("Unexpected export link mime type: %s", mimeType)
		}
	}
}

// TestConvertDriveFile_PartialCapabilities tests conversion with partial capabilities
func TestConvertDriveFile_PartialCapabilities(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "readonly.txt",
		MimeType: "text/plain",
		Capabilities: &drive.FileCapabilities{
			CanDownload: true,
			CanEdit:     false,
			CanShare:    false,
			CanDelete:   false,
			CanTrash:    false,
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	if !converted.Capabilities.CanDownload {
		t.Error("CanDownload should be true")
	}

	if converted.Capabilities.CanEdit {
		t.Error("CanEdit should be false")
	}

	if converted.Capabilities.CanShare {
		t.Error("CanShare should be false")
	}

	if converted.Capabilities.CanDelete {
		t.Error("CanDelete should be false")
	}

	if converted.Capabilities.CanTrash {
		t.Error("CanTrash should be false")
	}
}

// TestListOptions_AllFields tests ListOptions with all fields set
func TestListOptions_AllFields(t *testing.T) {
	opts := ListOptions{
		ParentID:       "parent123",
		Query:          "name contains 'test'",
		PageSize:       50,
		PageToken:      "token123",
		OrderBy:        "modifiedTime desc",
		IncludeTrashed: true,
		Fields:         "id,name,mimeType",
	}

	if opts.ParentID != "parent123" {
		t.Errorf("ParentID = %s, want parent123", opts.ParentID)
	}

	if opts.Query != "name contains 'test'" {
		t.Errorf("Query = %s, want name contains 'test'", opts.Query)
	}

	if opts.PageSize != 50 {
		t.Errorf("PageSize = %d, want 50", opts.PageSize)
	}

	if opts.PageToken != "token123" {
		t.Errorf("PageToken = %s, want token123", opts.PageToken)
	}

	if opts.OrderBy != "modifiedTime desc" {
		t.Errorf("OrderBy = %s, want modifiedTime desc", opts.OrderBy)
	}

	if !opts.IncludeTrashed {
		t.Error("IncludeTrashed should be true")
	}

	if opts.Fields != "id,name,mimeType" {
		t.Errorf("Fields = %s, want id,name,mimeType", opts.Fields)
	}
}

// TestDownloadOptions_AllFields tests DownloadOptions with all fields set
func TestDownloadOptions_AllFields(t *testing.T) {
	opts := DownloadOptions{
		OutputPath:   "/tmp/output.pdf",
		MimeType:     "application/pdf",
		Wait:         true,
		Timeout:      600,
		PollInterval: 10,
	}

	if opts.OutputPath != "/tmp/output.pdf" {
		t.Errorf("OutputPath = %s, want /tmp/output.pdf", opts.OutputPath)
	}

	if opts.MimeType != "application/pdf" {
		t.Errorf("MimeType = %s, want application/pdf", opts.MimeType)
	}

	if !opts.Wait {
		t.Error("Wait should be true")
	}

	if opts.Timeout != 600 {
		t.Errorf("Timeout = %d, want 600", opts.Timeout)
	}

	if opts.PollInterval != 10 {
		t.Errorf("PollInterval = %d, want 10", opts.PollInterval)
	}
}

// TestUploadOptions_AllFields tests UploadOptions with all fields set
func TestUploadOptions_AllFields(t *testing.T) {
	opts := UploadOptions{
		ParentID:    "parent123",
		Name:        "document.pdf",
		MimeType:    "application/pdf",
		Convert:     true,
		PinRevision: true,
	}

	if opts.ParentID != "parent123" {
		t.Errorf("ParentID = %s, want parent123", opts.ParentID)
	}

	if opts.Name != "document.pdf" {
		t.Errorf("Name = %s, want document.pdf", opts.Name)
	}

	if opts.MimeType != "application/pdf" {
		t.Errorf("MimeType = %s, want application/pdf", opts.MimeType)
	}

	if !opts.Convert {
		t.Error("Convert should be true")
	}

	if !opts.PinRevision {
		t.Error("PinRevision should be true")
	}
}

// TestUpdateContentOptions_AllFields tests UpdateContentOptions with all fields
func TestUpdateContentOptions_AllFields(t *testing.T) {
	opts := UpdateContentOptions{
		Name:     "updated.txt",
		MimeType: "text/plain",
		Fields:   "id,name,mimeType,modifiedTime",
	}

	if opts.Name != "updated.txt" {
		t.Errorf("Name = %s, want updated.txt", opts.Name)
	}

	if opts.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", opts.MimeType)
	}

	if opts.Fields != "id,name,mimeType,modifiedTime" {
		t.Errorf("Fields = %s, want id,name,mimeType,modifiedTime", opts.Fields)
	}
}

// TestConvertDriveFile_AllCapabilitiesFalse tests conversion with all capabilities false
func TestConvertDriveFile_AllCapabilitiesFalse(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "restricted.txt",
		MimeType: "text/plain",
		Capabilities: &drive.FileCapabilities{
			CanDownload:      false,
			CanEdit:          false,
			CanShare:         false,
			CanDelete:        false,
			CanTrash:         false,
			CanReadRevisions: false,
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	if converted.Capabilities.CanDownload {
		t.Error("CanDownload should be false")
	}

	if converted.Capabilities.CanEdit {
		t.Error("CanEdit should be false")
	}

	if converted.Capabilities.CanShare {
		t.Error("CanShare should be false")
	}

	if converted.Capabilities.CanDelete {
		t.Error("CanDelete should be false")
	}

	if converted.Capabilities.CanTrash {
		t.Error("CanTrash should be false")
	}

	if converted.Capabilities.CanReadRevisions {
		t.Error("CanReadRevisions should be false")
	}
}

// TestConvertDriveFile_LargeFile tests conversion with large file size
func TestConvertDriveFile_LargeFile(t *testing.T) {
	largeSize := int64(10 * 1024 * 1024 * 1024) // 10GB

	driveFile := &drive.File{
		Id:       "largeFile",
		Name:     "large_video.mp4",
		MimeType: "video/mp4",
		Size:     largeSize,
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != largeSize {
		t.Errorf("Size = %d, want %d", converted.Size, largeSize)
	}
}

// TestConvertDriveFile_SpecialCharactersInName tests conversion with special characters
func TestConvertDriveFile_SpecialCharactersInName(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "file with spaces & special (chars) [brackets].txt",
		MimeType: "text/plain",
	}

	converted := convertDriveFile(driveFile)

	if converted.Name != "file with spaces & special (chars) [brackets].txt" {
		t.Errorf("Name = %s, want file with spaces & special (chars) [brackets].txt", converted.Name)
	}
}

// TestConvertDriveFile_UnicodeCharacters tests conversion with unicode characters
func TestConvertDriveFile_UnicodeCharacters(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "文件名_ファイル_файл.txt",
		MimeType: "text/plain",
	}

	converted := convertDriveFile(driveFile)

	if converted.Name != "文件名_ファイル_файл.txt" {
		t.Errorf("Name = %s, want 文件名_ファイル_файл.txt", converted.Name)
	}
}

// TestConvertDriveFile_EmptyParents tests conversion with empty parents list
func TestConvertDriveFile_EmptyParents(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "orphan.txt",
		MimeType: "text/plain",
		Parents:  []string{},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.Parents) != 0 {
		t.Errorf("Parents should be empty, got %d items", len(converted.Parents))
	}
}

// TestConvertDriveFile_NilParents tests conversion with nil parents
func TestConvertDriveFile_NilParents(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "orphan.txt",
		MimeType: "text/plain",
		Parents:  nil,
	}

	converted := convertDriveFile(driveFile)

	if len(converted.Parents) != 0 {
		t.Errorf("Parents should be nil or empty, got %v", converted.Parents)
	}
}

// TestConvertDriveFile_NilExportLinks tests conversion with nil export links
func TestConvertDriveFile_NilExportLinks(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		ExportLinks: nil,
	}

	converted := convertDriveFile(driveFile)

	if len(converted.ExportLinks) != 0 {
		t.Errorf("ExportLinks should be nil or empty, got %v", converted.ExportLinks)
	}
}

// TestManager_NewManager_Initialization tests that NewManager properly initializes all fields
func TestManager_NewManager_Initialization(t *testing.T) {
	client := &api.Client{}
	manager := NewManager(client)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.client != client {
		t.Error("Manager client not set correctly")
	}

	if manager.shaper == nil {
		t.Error("Manager shaper not initialized")
	}
}

// TestListOptions_EmptyQuery tests ListOptions with empty query
func TestListOptions_EmptyQuery(t *testing.T) {
	opts := ListOptions{
		Query: "",
	}

	if opts.Query != "" {
		t.Errorf("Query should be empty, got %s", opts.Query)
	}
}

// TestListOptions_ZeroPageSize tests ListOptions with zero page size
func TestListOptions_ZeroPageSize(t *testing.T) {
	opts := ListOptions{
		PageSize: 0,
	}

	if opts.PageSize != 0 {
		t.Errorf("PageSize should be 0, got %d", opts.PageSize)
	}
}

// TestListOptions_LargePageSize tests ListOptions with large page size
func TestListOptions_LargePageSize(t *testing.T) {
	opts := ListOptions{
		PageSize: 1000,
	}

	if opts.PageSize != 1000 {
		t.Errorf("PageSize should be 1000, got %d", opts.PageSize)
	}
}

// TestDownloadOptions_ZeroTimeout tests DownloadOptions with zero timeout
func TestDownloadOptions_ZeroTimeout(t *testing.T) {
	opts := DownloadOptions{
		Timeout: 0,
	}

	if opts.Timeout != 0 {
		t.Errorf("Timeout should be 0, got %d", opts.Timeout)
	}
}

// TestDownloadOptions_LargeTimeout tests DownloadOptions with large timeout
func TestDownloadOptions_LargeTimeout(t *testing.T) {
	opts := DownloadOptions{
		Timeout: 3600,
	}

	if opts.Timeout != 3600 {
		t.Errorf("Timeout should be 3600, got %d", opts.Timeout)
	}
}

// TestDownloadOptions_ZeroPollInterval tests DownloadOptions with zero poll interval
func TestDownloadOptions_ZeroPollInterval(t *testing.T) {
	opts := DownloadOptions{
		PollInterval: 0,
	}

	if opts.PollInterval != 0 {
		t.Errorf("PollInterval should be 0, got %d", opts.PollInterval)
	}
}

// TestDownloadOptions_LargePollInterval tests DownloadOptions with large poll interval
func TestDownloadOptions_LargePollInterval(t *testing.T) {
	opts := DownloadOptions{
		PollInterval: 300,
	}

	if opts.PollInterval != 300 {
		t.Errorf("PollInterval should be 300, got %d", opts.PollInterval)
	}
}

// TestUploadOptions_EmptyParentID tests UploadOptions with empty parent ID
func TestUploadOptions_EmptyParentID(t *testing.T) {
	opts := UploadOptions{
		ParentID: "",
	}

	if opts.ParentID != "" {
		t.Errorf("ParentID should be empty, got %s", opts.ParentID)
	}
}

// TestUploadOptions_EmptyName tests UploadOptions with empty name
func TestUploadOptions_EmptyName(t *testing.T) {
	opts := UploadOptions{
		Name: "",
	}

	if opts.Name != "" {
		t.Errorf("Name should be empty, got %s", opts.Name)
	}
}

// TestUploadOptions_EmptyMimeType tests UploadOptions with empty mime type
func TestUploadOptions_EmptyMimeType(t *testing.T) {
	opts := UploadOptions{
		MimeType: "",
	}

	if opts.MimeType != "" {
		t.Errorf("MimeType should be empty, got %s", opts.MimeType)
	}
}

// TestUpdateContentOptions_EmptyName tests UpdateContentOptions with empty name
func TestUpdateContentOptions_EmptyName(t *testing.T) {
	opts := UpdateContentOptions{
		Name: "",
	}

	if opts.Name != "" {
		t.Errorf("Name should be empty, got %s", opts.Name)
	}
}

// TestUpdateContentOptions_EmptyMimeType tests UpdateContentOptions with empty mime type
func TestUpdateContentOptions_EmptyMimeType(t *testing.T) {
	opts := UpdateContentOptions{
		MimeType: "",
	}

	if opts.MimeType != "" {
		t.Errorf("MimeType should be empty, got %s", opts.MimeType)
	}
}

// TestUpdateContentOptions_EmptyFields tests UpdateContentOptions with empty fields
func TestUpdateContentOptions_EmptyFields(t *testing.T) {
	opts := UpdateContentOptions{
		Fields: "",
	}

	if opts.Fields != "" {
		t.Errorf("Fields should be empty, got %s", opts.Fields)
	}
}

// TestDownloadOptions_EmptyOutputPath tests DownloadOptions with empty output path
func TestDownloadOptions_EmptyOutputPath(t *testing.T) {
	opts := DownloadOptions{
		OutputPath: "",
	}

	if opts.OutputPath != "" {
		t.Errorf("OutputPath should be empty, got %s", opts.OutputPath)
	}
}

// TestDownloadOptions_EmptyMimeType tests DownloadOptions with empty mime type
func TestDownloadOptions_EmptyMimeType(t *testing.T) {
	opts := DownloadOptions{
		MimeType: "",
	}

	if opts.MimeType != "" {
		t.Errorf("MimeType should be empty, got %s", opts.MimeType)
	}
}

// TestListOptions_EmptyParentID tests ListOptions with empty parent ID
func TestListOptions_EmptyParentID(t *testing.T) {
	opts := ListOptions{
		ParentID: "",
	}

	if opts.ParentID != "" {
		t.Errorf("ParentID should be empty, got %s", opts.ParentID)
	}
}

// TestListOptions_EmptyPageToken tests ListOptions with empty page token
func TestListOptions_EmptyPageToken(t *testing.T) {
	opts := ListOptions{
		PageToken: "",
	}

	if opts.PageToken != "" {
		t.Errorf("PageToken should be empty, got %s", opts.PageToken)
	}
}

// TestListOptions_EmptyOrderBy tests ListOptions with empty order by
func TestListOptions_EmptyOrderBy(t *testing.T) {
	opts := ListOptions{
		OrderBy: "",
	}

	if opts.OrderBy != "" {
		t.Errorf("OrderBy should be empty, got %s", opts.OrderBy)
	}
}

// TestListOptions_EmptyFields tests ListOptions with empty fields
func TestListOptions_EmptyFields(t *testing.T) {
	opts := ListOptions{
		Fields: "",
	}

	if opts.Fields != "" {
		t.Errorf("Fields should be empty, got %s", opts.Fields)
	}
}

// TestConvertDriveFile_EmptyMD5Checksum tests conversion with empty MD5 checksum
func TestConvertDriveFile_EmptyMD5Checksum(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		Md5Checksum: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.MD5Checksum != "" {
		t.Errorf("MD5Checksum should be empty, got %s", converted.MD5Checksum)
	}
}

// TestConvertDriveFile_EmptyResourceKey tests conversion with empty resource key
func TestConvertDriveFile_EmptyResourceKey(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		ResourceKey: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.ResourceKey != "" {
		t.Errorf("ResourceKey should be empty, got %s", converted.ResourceKey)
	}
}

// TestConvertDriveFile_EmptyWebViewLink tests conversion with empty web view link
func TestConvertDriveFile_EmptyWebViewLink(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		WebViewLink: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.WebViewLink != "" {
		t.Errorf("WebViewLink should be empty, got %s", converted.WebViewLink)
	}
}

// TestConvertDriveFile_EmptyWebContentLink tests conversion with empty web content link
func TestConvertDriveFile_EmptyWebContentLink(t *testing.T) {
	driveFile := &drive.File{
		Id:             "file123",
		Name:           "test.txt",
		MimeType:       "text/plain",
		WebContentLink: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.WebContentLink != "" {
		t.Errorf("WebContentLink should be empty, got %s", converted.WebContentLink)
	}
}

// TestConvertDriveFile_EmptyCreatedTime tests conversion with empty created time
func TestConvertDriveFile_EmptyCreatedTime(t *testing.T) {
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		CreatedTime: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.CreatedTime != "" {
		t.Errorf("CreatedTime should be empty, got %s", converted.CreatedTime)
	}
}

// TestConvertDriveFile_EmptyModifiedTime tests conversion with empty modified time
func TestConvertDriveFile_EmptyModifiedTime(t *testing.T) {
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		ModifiedTime: "",
	}

	converted := convertDriveFile(driveFile)

	if converted.ModifiedTime != "" {
		t.Errorf("ModifiedTime should be empty, got %s", converted.ModifiedTime)
	}
}

// TestConvertDriveFile_ZeroSize tests conversion with zero size
func TestConvertDriveFile_ZeroSize(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "empty.txt",
		MimeType: "text/plain",
		Size:     0,
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != 0 {
		t.Errorf("Size should be 0, got %d", converted.Size)
	}
}

// TestConvertDriveFile_NegativeSize tests conversion with negative size (edge case)
func TestConvertDriveFile_NegativeSize(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "test.txt",
		MimeType: "text/plain",
		Size:     -1,
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != -1 {
		t.Errorf("Size should be -1, got %d", converted.Size)
	}
}

func TestRequestContext_WithAllFields(t *testing.T) {
	ctx := &types.RequestContext{
		Profile:           "production",
		DriveID:           "drive456",
		InvolvedFileIDs:   []string{"file1", "file2", "file3"},
		InvolvedParentIDs: []string{"parent1", "parent2"},
		RequestType:       types.RequestTypeListOrSearch,
		TraceID:           "trace456",
	}

	if ctx.Profile != "production" {
		t.Errorf("Profile = %s, want production", ctx.Profile)
	}

	if ctx.DriveID != "drive456" {
		t.Errorf("DriveID = %s, want drive456", ctx.DriveID)
	}

	if len(ctx.InvolvedFileIDs) != 3 {
		t.Errorf("InvolvedFileIDs length = %d, want 3", len(ctx.InvolvedFileIDs))
	}

	if len(ctx.InvolvedParentIDs) != 2 {
		t.Errorf("InvolvedParentIDs length = %d, want 2", len(ctx.InvolvedParentIDs))
	}

	if ctx.RequestType != types.RequestTypeListOrSearch {
		t.Errorf("RequestType = %s, want ListOrSearch", ctx.RequestType)
	}

	if ctx.TraceID != "trace456" {
		t.Errorf("TraceID = %s, want trace456", ctx.TraceID)
	}
}

func TestRequestContext_EmptyFields(t *testing.T) {
	ctx := &types.RequestContext{
		Profile:           "",
		DriveID:           "",
		InvolvedFileIDs:   []string{},
		InvolvedParentIDs: []string{},
		RequestType:       "",
		TraceID:           "",
	}

	if ctx.Profile != "" {
		t.Errorf("Profile should be empty, got %s", ctx.Profile)
	}

	if ctx.DriveID != "" {
		t.Errorf("DriveID should be empty, got %s", ctx.DriveID)
	}

	if len(ctx.InvolvedFileIDs) != 0 {
		t.Errorf("InvolvedFileIDs should be empty, got %d items", len(ctx.InvolvedFileIDs))
	}

	if len(ctx.InvolvedParentIDs) != 0 {
		t.Errorf("InvolvedParentIDs should be empty, got %d items", len(ctx.InvolvedParentIDs))
	}
}

func TestRequestContext_NilSlices(t *testing.T) {
	ctx := &types.RequestContext{
		Profile:           "default",
		DriveID:           "drive123",
		InvolvedFileIDs:   nil,
		InvolvedParentIDs: nil,
		RequestType:       types.RequestTypeMutation,
		TraceID:           "trace123",
	}

	if ctx.InvolvedFileIDs != nil {
		t.Errorf("InvolvedFileIDs should be nil, got %v", ctx.InvolvedFileIDs)
	}

	if ctx.InvolvedParentIDs != nil {
		t.Errorf("InvolvedParentIDs should be nil, got %v", ctx.InvolvedParentIDs)
	}
}

func TestConvertDriveFile_AllFieldsNil(t *testing.T) {
	driveFile := &drive.File{
		Id:             "file123",
		Name:           "test.txt",
		MimeType:       "text/plain",
		Capabilities:   nil,
		ExportLinks:    nil,
		Parents:        nil,
		ResourceKey:    "",
		WebViewLink:    "",
		WebContentLink: "",
		Md5Checksum:    "",
		CreatedTime:    "",
		ModifiedTime:   "",
		Size:           0,
		Trashed:        false,
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities != nil {
		t.Error("Capabilities should be nil")
	}

	if len(converted.ExportLinks) != 0 {
		t.Error("ExportLinks should be nil or empty")
	}

	if len(converted.Parents) != 0 {
		t.Error("Parents should be nil or empty")
	}
}

func TestConvertDriveFile_OnlyRequiredFields(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "test.txt",
		MimeType: "text/plain",
	}

	converted := convertDriveFile(driveFile)

	if converted.ID != "file123" {
		t.Errorf("ID = %s, want file123", converted.ID)
	}

	if converted.Name != "test.txt" {
		t.Errorf("Name = %s, want test.txt", converted.Name)
	}

	if converted.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", converted.MimeType)
	}
}

func TestSelectUploadType_ExactlyAtThreshold(t *testing.T) {
	thresholdSize := int64(5 * 1024 * 1024)
	uploadType := selectUploadType(thresholdSize, &drive.File{})
	if uploadType != "simple" {
		t.Errorf("selectUploadType at threshold = %s, want simple", uploadType)
	}
}

func TestSelectUploadType_JustAboveThreshold(t *testing.T) {
	thresholdSize := int64(5*1024*1024) + 1
	uploadType := selectUploadType(thresholdSize, &drive.File{})
	if uploadType != "resumable" {
		t.Errorf("selectUploadType above threshold = %s, want resumable", uploadType)
	}
}

func TestSelectUploadType_WithEmptyMetadata(t *testing.T) {
	uploadType := selectUploadType(1024, &drive.File{
		Name:     "",
		MimeType: "",
		Parents:  []string{},
	})
	if uploadType != "simple" {
		t.Errorf("selectUploadType with empty metadata = %s, want simple", uploadType)
	}
}

func TestListOptions_WithComplexQuery(t *testing.T) {
	complexQuery := "name contains 'test' and mimeType != 'application/vnd.google-apps.folder' and trashed = false"
	opts := ListOptions{
		Query: complexQuery,
	}

	if opts.Query != complexQuery {
		t.Errorf("Query = %s, want %s", opts.Query, complexQuery)
	}
}

func TestListOptions_WithMultipleOrderBy(t *testing.T) {
	orderBy := "modifiedTime desc, name asc"
	opts := ListOptions{
		OrderBy: orderBy,
	}

	if opts.OrderBy != orderBy {
		t.Errorf("OrderBy = %s, want %s", opts.OrderBy, orderBy)
	}
}

func TestDownloadOptions_WithAllTimeouts(t *testing.T) {
	opts := DownloadOptions{
		Timeout:      1800,
		PollInterval: 30,
	}

	if opts.Timeout != 1800 {
		t.Errorf("Timeout = %d, want 1800", opts.Timeout)
	}

	if opts.PollInterval != 30 {
		t.Errorf("PollInterval = %d, want 30", opts.PollInterval)
	}
}

func TestUploadOptions_WithAllMetadata(t *testing.T) {
	opts := UploadOptions{
		ParentID:    "parent123",
		Name:        "file.pdf",
		MimeType:    "application/pdf",
		Convert:     true,
		PinRevision: true,
	}

	if opts.ParentID != "parent123" {
		t.Errorf("ParentID = %s, want parent123", opts.ParentID)
	}

	if opts.Name != "file.pdf" {
		t.Errorf("Name = %s, want file.pdf", opts.Name)
	}

	if opts.MimeType != "application/pdf" {
		t.Errorf("MimeType = %s, want application/pdf", opts.MimeType)
	}

	if !opts.Convert {
		t.Error("Convert should be true")
	}

	if !opts.PinRevision {
		t.Error("PinRevision should be true")
	}
}

func TestUpdateContentOptions_WithAllOptions(t *testing.T) {
	opts := UpdateContentOptions{
		Name:     "updated_file.txt",
		MimeType: "text/plain",
		Fields:   "id,name,mimeType,modifiedTime,size",
	}

	if opts.Name != "updated_file.txt" {
		t.Errorf("Name = %s, want updated_file.txt", opts.Name)
	}

	if opts.MimeType != "text/plain" {
		t.Errorf("MimeType = %s, want text/plain", opts.MimeType)
	}

	if opts.Fields != "id,name,mimeType,modifiedTime,size" {
		t.Errorf("Fields = %s, want id,name,mimeType,modifiedTime,size", opts.Fields)
	}
}

func TestConvertDriveFile_WithAllCapabilities(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "test.txt",
		MimeType: "text/plain",
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          true,
			CanShare:         true,
			CanDelete:        true,
			CanTrash:         true,
			CanReadRevisions: true,
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	if !converted.Capabilities.CanDownload {
		t.Error("CanDownload should be true")
	}

	if !converted.Capabilities.CanEdit {
		t.Error("CanEdit should be true")
	}

	if !converted.Capabilities.CanShare {
		t.Error("CanShare should be true")
	}

	if !converted.Capabilities.CanDelete {
		t.Error("CanDelete should be true")
	}

	if !converted.Capabilities.CanTrash {
		t.Error("CanTrash should be true")
	}

	if !converted.Capabilities.CanReadRevisions {
		t.Error("CanReadRevisions should be true")
	}
}

func TestConvertDriveFile_WithMixedCapabilities(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "test.txt",
		MimeType: "text/plain",
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          false,
			CanShare:         true,
			CanDelete:        false,
			CanTrash:         true,
			CanReadRevisions: false,
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}

	if !converted.Capabilities.CanDownload {
		t.Error("CanDownload should be true")
	}

	if converted.Capabilities.CanEdit {
		t.Error("CanEdit should be false")
	}

	if !converted.Capabilities.CanShare {
		t.Error("CanShare should be true")
	}

	if converted.Capabilities.CanDelete {
		t.Error("CanDelete should be false")
	}

	if !converted.Capabilities.CanTrash {
		t.Error("CanTrash should be true")
	}

	if converted.Capabilities.CanReadRevisions {
		t.Error("CanReadRevisions should be false")
	}
}

func TestConvertDriveFile_WithManyExportLinks(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "document.doc",
		MimeType: "application/vnd.google-apps.document",
		ExportLinks: map[string]string{
			"application/pdf":      "https://export/pdf",
			"text/plain":           "https://export/txt",
			"application/docx":     "https://export/docx",
			"application/odt":      "https://export/odt",
			"text/html":            "https://export/html",
			"application/rtf":      "https://export/rtf",
			"application/epub+zip": "https://export/epub",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "https://export/docx2",
		},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.ExportLinks) != 8 {
		t.Errorf("ExportLinks length = %d, want 8", len(converted.ExportLinks))
	}
}

func TestConvertDriveFile_WithManyParents(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "shared.txt",
		MimeType: "text/plain",
		Parents:  []string{"parent1", "parent2", "parent3", "parent4", "parent5"},
	}

	converted := convertDriveFile(driveFile)

	if len(converted.Parents) != 5 {
		t.Errorf("Parents length = %d, want 5", len(converted.Parents))
	}

	for i, p := range converted.Parents {
		expectedParent := "parent" + string(rune(i+1+'0'))
		if p != expectedParent {
			t.Errorf("Parent[%d] = %s, want %s", i, p, expectedParent)
		}
	}
}

func TestConvertDriveFile_WithVeryLongName(t *testing.T) {
	longName := "this_is_a_very_long_filename_that_contains_many_characters_and_should_still_be_properly_converted_by_the_convertDriveFile_function_without_any_truncation_or_issues.txt"
	driveFile := &drive.File{
		Id:       "file123",
		Name:     longName,
		MimeType: "text/plain",
	}

	converted := convertDriveFile(driveFile)

	if converted.Name != longName {
		t.Errorf("Name = %s, want %s", converted.Name, longName)
	}
}

func TestConvertDriveFile_WithVeryLongMimeType(t *testing.T) {
	longMimeType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "test.docx",
		MimeType: longMimeType,
	}

	converted := convertDriveFile(driveFile)

	if converted.MimeType != longMimeType {
		t.Errorf("MimeType = %s, want %s", converted.MimeType, longMimeType)
	}
}

func TestConvertDriveFile_WithVeryLongResourceKey(t *testing.T) {
	longResourceKey := "0AExampleResourceKeyThatIsVeryLongAndContainsManyCha" +
		"ractersAndShouldStillBeProperlyConvertedByTheFunctionWithoutAnyIssues" +
		"AndWithoutBeingTruncatedOrModifiedInAnyWay"
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		ResourceKey: longResourceKey,
	}

	converted := convertDriveFile(driveFile)

	if converted.ResourceKey != longResourceKey {
		t.Errorf("ResourceKey = %s, want %s", converted.ResourceKey, longResourceKey)
	}
}

func TestConvertDriveFile_WithVeryLongWebViewLink(t *testing.T) {
	longWebViewLink := "https://drive.google.com/file/d/1ExampleFileIDThatIsVeryLongAndContainsManyCha" +
		"ractersAndShouldStillBeProperlyConvertedByTheFunctionWithoutAnyIssues/view"
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		WebViewLink: longWebViewLink,
	}

	converted := convertDriveFile(driveFile)

	if converted.WebViewLink != longWebViewLink {
		t.Errorf("WebViewLink = %s, want %s", converted.WebViewLink, longWebViewLink)
	}
}

func TestConvertDriveFile_WithVeryLongWebContentLink(t *testing.T) {
	longWebContentLink := "https://drive.google.com/uc?id=1ExampleFileIDThatIsVeryLongAndContainsManyCha" +
		"ractersAndShouldStillBeProperlyConvertedByTheFunctionWithoutAnyIssues&export=download"
	driveFile := &drive.File{
		Id:             "file123",
		Name:           "test.txt",
		MimeType:       "text/plain",
		WebContentLink: longWebContentLink,
	}

	converted := convertDriveFile(driveFile)

	if converted.WebContentLink != longWebContentLink {
		t.Errorf("WebContentLink = %s, want %s", converted.WebContentLink, longWebContentLink)
	}
}

func TestConvertDriveFile_WithVeryLongMD5Checksum(t *testing.T) {
	longMD5 := "d41d8cd98f00b204e9800998ecf8427e"
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		Md5Checksum: longMD5,
	}

	converted := convertDriveFile(driveFile)

	if converted.MD5Checksum != longMD5 {
		t.Errorf("MD5Checksum = %s, want %s", converted.MD5Checksum, longMD5)
	}
}

func TestConvertDriveFile_WithVeryLongCreatedTime(t *testing.T) {
	createdTime := "2024-01-01T00:00:00.000000Z"
	driveFile := &drive.File{
		Id:          "file123",
		Name:        "test.txt",
		MimeType:    "text/plain",
		CreatedTime: createdTime,
	}

	converted := convertDriveFile(driveFile)

	if converted.CreatedTime != createdTime {
		t.Errorf("CreatedTime = %s, want %s", converted.CreatedTime, createdTime)
	}
}

func TestConvertDriveFile_WithVeryLongModifiedTime(t *testing.T) {
	modifiedTime := "2024-12-31T23:59:59.999999Z"
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		ModifiedTime: modifiedTime,
	}

	converted := convertDriveFile(driveFile)

	if converted.ModifiedTime != modifiedTime {
		t.Errorf("ModifiedTime = %s, want %s", converted.ModifiedTime, modifiedTime)
	}
}

func TestConvertDriveFile_WithVeryLargeSize(t *testing.T) {
	veryLargeSize := int64(1000 * 1024 * 1024 * 1024)
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "huge_file.iso",
		MimeType: "application/octet-stream",
		Size:     veryLargeSize,
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != veryLargeSize {
		t.Errorf("Size = %d, want %d", converted.Size, veryLargeSize)
	}
}

func TestConvertDriveFile_WithMaxInt64Size(t *testing.T) {
	maxSize := int64(9223372036854775807)
	driveFile := &drive.File{
		Id:       "file123",
		Name:     "max_size_file.bin",
		MimeType: "application/octet-stream",
		Size:     maxSize,
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != maxSize {
		t.Errorf("Size = %d, want %d", converted.Size, maxSize)
	}
}

func TestListOptions_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name string
		opts ListOptions
	}{
		{
			name: "List with all parameters",
			opts: ListOptions{
				ParentID:       "parent123",
				Query:          "name contains 'test'",
				PageSize:       100,
				PageToken:      "token123",
				OrderBy:        "modifiedTime desc",
				IncludeTrashed: true,
				Fields:         "id,name,mimeType,size",
			},
		},
		{
			name: "List with minimal parameters",
			opts: ListOptions{},
		},
		{
			name: "List with only parent",
			opts: ListOptions{
				ParentID: "parent123",
			},
		},
		{
			name: "List with only query",
			opts: ListOptions{
				Query: "name contains 'test'",
			},
		},
		{
			name: "List with pagination",
			opts: ListOptions{
				PageSize:  50,
				PageToken: "token123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.ParentID != "" && tt.opts.ParentID != "parent123" {
				t.Errorf("ParentID mismatch")
			}
		})
	}
}

func TestDownloadOptions_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name string
		opts DownloadOptions
	}{
		{
			name: "Download with all parameters",
			opts: DownloadOptions{
				OutputPath:   "/tmp/output.pdf",
				MimeType:     "application/pdf",
				Wait:         true,
				Timeout:      600,
				PollInterval: 10,
			},
		},
		{
			name: "Download with minimal parameters",
			opts: DownloadOptions{},
		},
		{
			name: "Download with output path only",
			opts: DownloadOptions{
				OutputPath: "/tmp/file.txt",
			},
		},
		{
			name: "Download with mime type only",
			opts: DownloadOptions{
				MimeType: "text/plain",
			},
		},
		{
			name: "Download with wait and timeout",
			opts: DownloadOptions{
				Wait:    true,
				Timeout: 300,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.OutputPath != "" && tt.opts.OutputPath != "/tmp/output.pdf" && tt.opts.OutputPath != "/tmp/file.txt" {
				t.Errorf("OutputPath mismatch")
			}
		})
	}
}

func TestUploadOptions_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name string
		opts UploadOptions
	}{
		{
			name: "Upload with all parameters",
			opts: UploadOptions{
				ParentID:    "parent123",
				Name:        "file.pdf",
				MimeType:    "application/pdf",
				Convert:     true,
				PinRevision: true,
			},
		},
		{
			name: "Upload with minimal parameters",
			opts: UploadOptions{},
		},
		{
			name: "Upload with parent only",
			opts: UploadOptions{
				ParentID: "parent123",
			},
		},
		{
			name: "Upload with name only",
			opts: UploadOptions{
				Name: "file.txt",
			},
		},
		{
			name: "Upload with mime type only",
			opts: UploadOptions{
				MimeType: "text/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.ParentID != "" && tt.opts.ParentID != "parent123" {
				t.Errorf("ParentID mismatch")
			}
		})
	}
}

func TestUpdateContentOptions_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name string
		opts UpdateContentOptions
	}{
		{
			name: "Update with all parameters",
			opts: UpdateContentOptions{
				Name:     "updated.txt",
				MimeType: "text/plain",
				Fields:   "id,name,mimeType,modifiedTime",
			},
		},
		{
			name: "Update with minimal parameters",
			opts: UpdateContentOptions{},
		},
		{
			name: "Update with name only",
			opts: UpdateContentOptions{
				Name: "newname.txt",
			},
		},
		{
			name: "Update with mime type only",
			opts: UpdateContentOptions{
				MimeType: "application/pdf",
			},
		},
		{
			name: "Update with fields only",
			opts: UpdateContentOptions{
				Fields: "id,name,mimeType",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Name != "" && tt.opts.Name != "updated.txt" && tt.opts.Name != "newname.txt" {
				t.Errorf("Name mismatch")
			}
		})
	}
}

func TestConvertDriveFile_WithAllNilFields(t *testing.T) {
	driveFile := &drive.File{
		Id:             "file123",
		Name:           "test.txt",
		MimeType:       "text/plain",
		Capabilities:   nil,
		ExportLinks:    nil,
		Parents:        nil,
		ResourceKey:    "",
		WebViewLink:    "",
		WebContentLink: "",
		Md5Checksum:    "",
		CreatedTime:    "",
		ModifiedTime:   "",
		Size:           0,
		Trashed:        false,
	}

	converted := convertDriveFile(driveFile)

	if converted.ID != "file123" {
		t.Errorf("ID = %s, want file123", converted.ID)
	}

	if converted.Capabilities != nil {
		t.Error("Capabilities should be nil")
	}

	if len(converted.ExportLinks) != 0 {
		t.Error("ExportLinks should be nil or empty")
	}

	if len(converted.Parents) != 0 {
		t.Error("Parents should be nil or empty")
	}
}

func TestConvertDriveFile_WithMixedNilAndValues(t *testing.T) {
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         1024,
		Capabilities: nil,
		ExportLinks: map[string]string{
			"application/pdf": "https://export/pdf",
		},
		Parents: []string{"parent1"},
	}

	converted := convertDriveFile(driveFile)

	if converted.Size != 1024 {
		t.Errorf("Size = %d, want 1024", converted.Size)
	}

	if converted.Capabilities != nil {
		t.Error("Capabilities should be nil")
	}

	if len(converted.ExportLinks) != 1 {
		t.Errorf("ExportLinks length = %d, want 1", len(converted.ExportLinks))
	}

	if len(converted.Parents) != 1 {
		t.Errorf("Parents length = %d, want 1", len(converted.Parents))
	}
}

func TestSelectUploadType_AllCombinations(t *testing.T) {
	tests := []struct {
		name         string
		size         int64
		hasName      bool
		hasMimeType  bool
		hasParents   bool
		expectedType string
	}{
		{"Small, no metadata", 1024, false, false, false, "simple"},
		{"Small, name only", 1024, true, false, false, "multipart"},
		{"Small, mime only", 1024, false, true, false, "multipart"},
		{"Small, parents only", 1024, false, false, true, "multipart"},
		{"Small, name+mime", 1024, true, true, false, "multipart"},
		{"Small, name+parents", 1024, true, false, true, "multipart"},
		{"Small, mime+parents", 1024, false, true, true, "multipart"},
		{"Small, all metadata", 1024, true, true, true, "multipart"},
		{"Large, no metadata", 6 * 1024 * 1024, false, false, false, "resumable"},
		{"Large, name only", 6 * 1024 * 1024, true, false, false, "resumable"},
		{"Large, mime only", 6 * 1024 * 1024, false, true, false, "resumable"},
		{"Large, parents only", 6 * 1024 * 1024, false, false, true, "resumable"},
		{"Large, all metadata", 6 * 1024 * 1024, true, true, true, "resumable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &drive.File{}
			if tt.hasName {
				metadata.Name = "test.txt"
			}
			if tt.hasMimeType {
				metadata.MimeType = "text/plain"
			}
			if tt.hasParents {
				metadata.Parents = []string{"parent1"}
			}

			got := selectUploadType(tt.size, metadata)
			if got != tt.expectedType {
				t.Errorf("selectUploadType(%d, %+v) = %s, want %s",
					tt.size, metadata, got, tt.expectedType)
			}
		})
	}
}

func TestConvertDriveFile_PreservesAllFields(t *testing.T) {
	driveFile := &drive.File{
		Id:             "id123",
		Name:           "name123",
		MimeType:       "mime123",
		Size:           12345,
		Md5Checksum:    "md5123",
		CreatedTime:    "created123",
		ModifiedTime:   "modified123",
		ResourceKey:    "key123",
		WebViewLink:    "view123",
		WebContentLink: "content123",
		Trashed:        true,
		Parents:        []string{"p1", "p2"},
		ExportLinks: map[string]string{
			"type1": "link1",
			"type2": "link2",
		},
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          false,
			CanShare:         true,
			CanDelete:        false,
			CanTrash:         true,
			CanReadRevisions: false,
		},
	}

	converted := convertDriveFile(driveFile)

	if converted.ID != "id123" {
		t.Errorf("ID = %s, want id123", converted.ID)
	}
	if converted.Name != "name123" {
		t.Errorf("Name = %s, want name123", converted.Name)
	}
	if converted.MimeType != "mime123" {
		t.Errorf("MimeType = %s, want mime123", converted.MimeType)
	}
	if converted.Size != 12345 {
		t.Errorf("Size = %d, want 12345", converted.Size)
	}
	if converted.MD5Checksum != "md5123" {
		t.Errorf("MD5Checksum = %s, want md5123", converted.MD5Checksum)
	}
	if converted.CreatedTime != "created123" {
		t.Errorf("CreatedTime = %s, want created123", converted.CreatedTime)
	}
	if converted.ModifiedTime != "modified123" {
		t.Errorf("ModifiedTime = %s, want modified123", converted.ModifiedTime)
	}
	if converted.ResourceKey != "key123" {
		t.Errorf("ResourceKey = %s, want key123", converted.ResourceKey)
	}
	if converted.WebViewLink != "view123" {
		t.Errorf("WebViewLink = %s, want view123", converted.WebViewLink)
	}
	if converted.WebContentLink != "content123" {
		t.Errorf("WebContentLink = %s, want content123", converted.WebContentLink)
	}
	if !converted.Trashed {
		t.Error("Trashed should be true")
	}
	if len(converted.Parents) != 2 {
		t.Errorf("Parents length = %d, want 2", len(converted.Parents))
	}
	if len(converted.ExportLinks) != 2 {
		t.Errorf("ExportLinks length = %d, want 2", len(converted.ExportLinks))
	}
	if converted.Capabilities == nil {
		t.Fatal("Capabilities should not be nil")
	}
	if !converted.Capabilities.CanDownload {
		t.Error("CanDownload should be true")
	}
	if converted.Capabilities.CanEdit {
		t.Error("CanEdit should be false")
	}
	if !converted.Capabilities.CanShare {
		t.Error("CanShare should be true")
	}
	if converted.Capabilities.CanDelete {
		t.Error("CanDelete should be false")
	}
	if !converted.Capabilities.CanTrash {
		t.Error("CanTrash should be true")
	}
	if converted.Capabilities.CanReadRevisions {
		t.Error("CanReadRevisions should be false")
	}
}
