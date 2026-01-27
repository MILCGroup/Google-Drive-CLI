package files

import (
	"testing"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
)

func TestManager_Creation(t *testing.T) {
	// Test that manager can be created
	client := &api.Client{}
	manager := NewManager(client)

	if manager == nil {
		t.Error("NewManager returned nil")
	}

	if manager.client != client {
		t.Error("Manager client not set correctly")
	}

	if manager.shaper == nil {
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
