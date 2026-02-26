package export

import (
	"strings"
	"testing"
)

func TestIsGoogleWorkspaceFile(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		{"Google Docs", MimeTypeGoogleDocs, true},
		{"Google Sheets", MimeTypeGoogleSheets, true},
		{"Google Slides", MimeTypeGoogleSlides, true},
		{"PDF", "application/pdf", false},
		{"Plain text", "text/plain", false},
		{"DOCX", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGoogleWorkspaceFile(tt.mimeType)
			if got != tt.want {
				t.Errorf("IsGoogleWorkspaceFile(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestGetConvenienceFormat(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
		want      string
		wantErr   bool
	}{
		{"PDF", "pdf", "application/pdf", false},
		{"DOCX", "docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"XLSX", "xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", false},
		{"PPTX", "pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation", false},
		{"TXT", "txt", "text/plain", false},
		{"CSV", "csv", "text/csv", false},
		{"MIME type passthrough", "application/pdf", "application/pdf", false},
		{"Unknown format", "unknown", "", true},
		{"Case insensitive", "PDF", "application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConvenienceFormat(tt.shorthand)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConvenienceFormat(%q) error = %v, wantErr %v", tt.shorthand, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetConvenienceFormat(%q) = %q, want %q", tt.shorthand, got, tt.want)
			}
		})
	}
}

func TestGetAvailableFormats(t *testing.T) {
	tests := []struct {
		name      string
		mimeType  string
		wantErr   bool
		wantCount int
	}{
		{"Google Docs", MimeTypeGoogleDocs, false, 8},
		{"Google Sheets", MimeTypeGoogleSheets, false, 6},
		{"Google Slides", MimeTypeGoogleSlides, false, 4},
		{"Google Drawing", MimeTypeGoogleDrawing, false, 4},
		{"Non-Workspace file", "application/pdf", true, 0},
		{"Unknown Workspace type", "application/vnd.google-apps.unknown", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAvailableFormats(tt.mimeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAvailableFormats(%q) error = %v, wantErr %v", tt.mimeType, err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetAvailableFormats(%q) returned %d formats, want %d", tt.mimeType, len(got), tt.wantCount)
			}
		})
	}
}

func TestValidateExportFormat(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		target  string
		wantErr bool
	}{
		{"Docs to PDF", MimeTypeGoogleDocs, "application/pdf", false},
		{"Docs to DOCX", MimeTypeGoogleDocs, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"Sheets to XLSX", MimeTypeGoogleSheets, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", false},
		{"Sheets to CSV", MimeTypeGoogleSheets, "text/csv", false},
		{"Slides to PPTX", MimeTypeGoogleSlides, "application/vnd.openxmlformats-officedocument.presentationml.presentation", false},
		{"Invalid: Docs to XLSX", MimeTypeGoogleDocs, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
		{"Invalid: Non-Workspace source", "application/pdf", "text/plain", true},
		{"Invalid: Unsupported target", MimeTypeGoogleDocs, "application/unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExportFormat(tt.source, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportFormat(%q, %q) error = %v, wantErr %v", tt.source, tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestGetMimeTypeForWorkspaceType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
		wantErr  bool
	}{
		{"Document", "document", MimeTypeGoogleDocs, false},
		{"Doc", "doc", MimeTypeGoogleDocs, false},
		{"Spreadsheet", "spreadsheet", MimeTypeGoogleSheets, false},
		{"Sheet", "sheet", MimeTypeGoogleSheets, false},
		{"Presentation", "presentation", MimeTypeGoogleSlides, false},
		{"Slide", "slide", MimeTypeGoogleSlides, false},
		{"Drawing", "drawing", MimeTypeGoogleDrawing, false},
		{"Form", "form", MimeTypeGoogleForm, false},
		{"Script", "script", MimeTypeGoogleScript, false},
		{"Unknown", "unknown", "", true},
		{"Case insensitive", "DOCUMENT", MimeTypeGoogleDocs, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMimeTypeForWorkspaceType(tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMimeTypeForWorkspaceType(%q) error = %v, wantErr %v", tt.typeName, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMimeTypeForWorkspaceType(%q) = %q, want %q", tt.typeName, got, tt.want)
			}
		})
	}
}

// Property tests

func TestMimeTypeValidation_Property(t *testing.T) {
	// Property 23: MIME Type Validation
	// Validates that export format validation rejects invalid MIME type combinations

	workspaceTypes := []string{
		MimeTypeGoogleDocs,
		MimeTypeGoogleSheets,
		MimeTypeGoogleSlides,
		MimeTypeGoogleDrawing,
	}

	invalidTargets := []string{
		"application/unknown",
		"text/invalid",
		"image/invalid",
	}

	for _, source := range workspaceTypes {
		for _, target := range invalidTargets {
			err := ValidateExportFormat(source, target)
			if err == nil {
				t.Errorf("ValidateExportFormat(%q, %q) should have returned error for invalid target", source, target)
			}
		}
	}
}

func TestFormatConvenienceMapping_Property(t *testing.T) {
	// Property 24: Format Convenience Mapping
	// Validates that convenience format mappings are correctly applied

	tests := []struct {
		shorthand string
		mimeType  string
	}{
		{"pdf", "application/pdf"},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	}

	for _, tt := range tests {
		got, err := GetConvenienceFormat(tt.shorthand)
		if err != nil {
			t.Errorf("GetConvenienceFormat(%q) returned error: %v", tt.shorthand, err)
			continue
		}
		if got != tt.mimeType {
			t.Errorf("GetConvenienceFormat(%q) = %q, want %q", tt.shorthand, got, tt.mimeType)
		}

		// Test case insensitivity
		upperShorthand := strings.ToUpper(tt.shorthand)
		got, err = GetConvenienceFormat(upperShorthand)
		if err != nil {
			t.Errorf("GetConvenienceFormat(%q) returned error: %v", upperShorthand, err)
			continue
		}
		if got != tt.mimeType {
			t.Errorf("GetConvenienceFormat(%q) = %q, want %q (case insensitive)", upperShorthand, got, tt.mimeType)
		}
	}
}
