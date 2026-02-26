package files

import (
	"testing"

	"github.com/dl-alexandre/gdrv/internal/utils"
)

// Property 2: API Method Selection by File Type
// Validates: Requirements 2.4, 2.5, 2.6
// Property: Download method selection must be correct based on MIME type

func TestProperty_APIMethodSelection_WorkspaceFiles(t *testing.T) {
	// Property: All Google Workspace MIME types require export, not download

	workspaceMimeTypes := []string{
		utils.MimeTypeDocument,
		utils.MimeTypeSpreadsheet,
		utils.MimeTypePresentation,
		utils.MimeTypeDrawing,
		utils.MimeTypeForm,
		utils.MimeTypeScript,
	}

	for _, mimeType := range workspaceMimeTypes {
		t.Run(mimeType, func(t *testing.T) {
			isWorkspace := utils.IsWorkspaceMimeType(mimeType)
			if !isWorkspace {
				t.Errorf("IsWorkspaceMimeType(%s) = false, want true", mimeType)
			}
		})
	}
}

func TestProperty_APIMethodSelection_NonWorkspaceFiles(t *testing.T) {
	// Property: All non-Workspace MIME types use direct download

	nonWorkspaceMimeTypes := []string{
		"application/pdf",
		"text/plain",
		"image/png",
		"image/jpeg",
		"video/mp4",
		"audio/mp3",
		"application/zip",
		"application/json",
		"text/csv",
		"text/html",
		"application/octet-stream",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-excel",
		"application/msword",
	}

	for _, mimeType := range nonWorkspaceMimeTypes {
		t.Run(mimeType, func(t *testing.T) {
			isWorkspace := utils.IsWorkspaceMimeType(mimeType)
			if isWorkspace {
				t.Errorf("IsWorkspaceMimeType(%s) = true, want false", mimeType)
			}
		})
	}
}

func TestProperty_APIMethodSelection_EdgeCases(t *testing.T) {
	// Property: Edge cases are handled correctly

	tests := []struct {
		name          string
		mimeType      string
		wantWorkspace bool
	}{
		{"Empty string", "", false},
		{"Folder", utils.MimeTypeFolder, false},
		{"Shortcut", utils.MimeTypeShortcut, false},
		{"Case variation (lowercase)", "application/vnd.google-apps.document", true},
		{"Partial match", "application/vnd.google-apps.doc", false},
		{"With charset", "text/plain; charset=utf-8", false},
		{"Unknown MIME", "application/x-unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsWorkspaceMimeType(tt.mimeType)
			if got != tt.wantWorkspace {
				t.Errorf("IsWorkspaceMimeType(%s) = %v, want %v", tt.mimeType, got, tt.wantWorkspace)
			}
		})
	}
}

func TestProperty_APIMethodSelection_Consistency(t *testing.T) {
	// Property: Same MIME type always produces same result (deterministic)

	testMimeTypes := []string{
		utils.MimeTypeDocument,
		"application/pdf",
		"text/plain",
		utils.MimeTypeSpreadsheet,
		"image/png",
	}

	for _, mimeType := range testMimeTypes {
		firstResult := utils.IsWorkspaceMimeType(mimeType)

		// Run 100 iterations
		for i := 0; i < 100; i++ {
			result := utils.IsWorkspaceMimeType(mimeType)
			if result != firstResult {
				t.Errorf("IsWorkspaceMimeType not deterministic for %s: iteration %d gave %v, expected %v",
					mimeType, i, result, firstResult)
			}
		}
	}
}

func TestProperty_APIMethodSelection_AllWorkspaceTypes(t *testing.T) {
	// Property: Exactly these 6 types are Workspace types, no more, no less

	allWorkspaceTypes := map[string]bool{
		utils.MimeTypeDocument:     true,
		utils.MimeTypeSpreadsheet:  true,
		utils.MimeTypePresentation: true,
		utils.MimeTypeDrawing:      true,
		utils.MimeTypeForm:         true,
		utils.MimeTypeScript:       true,
	}

	// Test all workspace types return true
	for mimeType := range allWorkspaceTypes {
		if !utils.IsWorkspaceMimeType(mimeType) {
			t.Errorf("Expected %s to be a workspace type", mimeType)
		}
	}

	// Test that the count is exactly 6
	count := 0
	testTypes := []string{
		utils.MimeTypeDocument,
		utils.MimeTypeSpreadsheet,
		utils.MimeTypePresentation,
		utils.MimeTypeDrawing,
		utils.MimeTypeForm,
		utils.MimeTypeScript,
		utils.MimeTypeFolder,
		utils.MimeTypeShortcut,
		"application/pdf",
		"text/plain",
	}

	for _, mimeType := range testTypes {
		if utils.IsWorkspaceMimeType(mimeType) {
			count++
		}
	}

	if count != 6 {
		t.Errorf("Expected exactly 6 workspace types, found %d", count)
	}
}

// Property 17: Export Limit Handling
// Validates: Requirements 2.13, 2.14
func TestProperty_ExportLimitHandling_Threshold(t *testing.T) {
	// Property: Export limit is exactly 10 MiB

	expectedLimit := 10 * 1024 * 1024 // 10 MiB

	if utils.ExportMaxBytes != expectedLimit {
		t.Errorf("ExportMaxBytes = %d, want %d", utils.ExportMaxBytes, expectedLimit)
	}
}

func TestProperty_ExportLimitHandling_Comparison(t *testing.T) {
	// Property: Files below limit can export, files at or above cannot

	tests := []struct {
		name      string
		size      int64
		canExport bool
	}{
		{"Zero size", 0, true},
		{"1 byte", 1, true},
		{"1 KB", 1024, true},
		{"1 MB", 1024 * 1024, true},
		{"9 MB", 9 * 1024 * 1024, true},
		{"10 MB minus 1", int64(utils.ExportMaxBytes) - 1, true},
		{"10 MB exact", int64(utils.ExportMaxBytes), false},
		{"10 MB plus 1", int64(utils.ExportMaxBytes) + 1, false},
		{"20 MB", 20 * 1024 * 1024, false},
		{"100 MB", 100 * 1024 * 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canExport := tt.size < int64(utils.ExportMaxBytes)
			if canExport != tt.canExport {
				t.Errorf("size %d: canExport = %v, want %v", tt.size, canExport, tt.canExport)
			}
		})
	}
}
