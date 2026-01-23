package utils

import (
	"testing"
)

func TestUploadThresholds(t *testing.T) {
	// Verify upload thresholds match spec
	if UploadSimpleMaxBytes != 5*1024*1024 {
		t.Errorf("UploadSimpleMaxBytes = %d, want 5 MiB", UploadSimpleMaxBytes)
	}
	if UploadChunkSize != 8*1024*1024 {
		t.Errorf("UploadChunkSize = %d, want 8 MiB", UploadChunkSize)
	}
	if ExportMaxBytes != 10*1024*1024 {
		t.Errorf("ExportMaxBytes = %d, want 10 MiB", ExportMaxBytes)
	}
}

func TestIsWorkspaceMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{MimeTypeDocument, true},
		{MimeTypeSpreadsheet, true},
		{MimeTypePresentation, true},
		{MimeTypeDrawing, true},
		{MimeTypeForm, true},
		{MimeTypeScript, true},
		{MimeTypeFolder, false},
		{MimeTypeShortcut, false},
		{"application/pdf", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := IsWorkspaceMimeType(tt.mimeType)
			if got != tt.expected {
				t.Errorf("IsWorkspaceMimeType(%s) = %v, want %v", tt.mimeType, got, tt.expected)
			}
		})
	}
}

func TestFormatMappings(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"pdf", "application/pdf"},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"csv", "text/csv"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := FormatMappings[tt.format]
			if got != tt.expected {
				t.Errorf("FormatMappings[%s] = %s, want %s", tt.format, got, tt.expected)
			}
		})
	}
}
