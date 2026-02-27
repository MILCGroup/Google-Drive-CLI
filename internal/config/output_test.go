package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/milcgroup/gdrv/internal/types"
)

func TestOutputFormatterJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format:  types.OutputFormatJSON,
		Quiet:   false,
		Verbose: false,
	})
	formatter.writer = &buf

	// Test data
	files := []*types.DriveFile{
		{
			ID:           "file123",
			Name:         "test.txt",
			MimeType:     "text/plain",
			Size:         1024,
			ModifiedTime: "2024-01-15T10:30:00Z",
		},
	}

	// Write success
	err := formatter.WriteSuccess("files list", files)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	// Parse JSON output
	var output types.CLIOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify structure
	if output.SchemaVersion == "" {
		t.Error("SchemaVersion should not be empty")
	}

	if output.Command != "files list" {
		t.Errorf("Expected command 'files list', got '%s'", output.Command)
	}

	if len(output.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(output.Errors))
	}

	// Verify data can be unmarshaled back to files
	dataBytes, err := json.Marshal(output.Data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	var parsedFiles []*types.DriveFile
	if err := json.Unmarshal(dataBytes, &parsedFiles); err != nil {
		t.Fatalf("Failed to unmarshal files: %v", err)
	}

	if len(parsedFiles) != 1 {
		t.Errorf("Expected 1 file, got %d", len(parsedFiles))
	}

	if parsedFiles[0].ID != files[0].ID {
		t.Errorf("Expected file ID '%s', got '%s'", files[0].ID, parsedFiles[0].ID)
	}
}

func TestOutputFormatterWithWarnings(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatJSON,
	})
	formatter.writer = &buf

	// Add warnings
	formatter.AddWarning("CACHE_MISS", "Path resolution cache miss", "low")
	formatter.AddWarning("DEPRECATED", "Feature is deprecated", "medium")

	// Write success
	err := formatter.WriteSuccess("test command", map[string]string{"result": "ok"})
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	// Parse output
	var output types.CLIOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify warnings
	if len(output.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(output.Warnings))
	}

	if output.Warnings[0].Code != "CACHE_MISS" {
		t.Errorf("Expected warning code 'CACHE_MISS', got '%s'", output.Warnings[0].Code)
	}

	if output.Warnings[1].Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", output.Warnings[1].Severity)
	}
}

func TestOutputFormatterError(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatJSON,
	})
	formatter.writer = &buf

	// Create error
	cliErr := types.CLIError{
		Code:       "NOT_FOUND",
		Message:    "File not found",
		HTTPStatus: 404,
		Retryable:  false,
	}

	// Write error
	err := formatter.WriteError("files get", cliErr)
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	// Parse output
	var output types.CLIOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify error structure
	if len(output.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(output.Errors))
	}

	if output.Errors[0].Code != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got '%s'", output.Errors[0].Code)
	}

	if output.Errors[0].HTTPStatus != 404 {
		t.Errorf("Expected HTTP status 404, got %d", output.Errors[0].HTTPStatus)
	}

	// Trace ID should always be present for errors
	if output.TraceID == "" {
		t.Error("Expected trace ID for error output")
	}
}

func TestOutputFormatterTable(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatTable,
	})
	formatter.writer = &buf

	files := []*types.DriveFile{
		{
			ID:           "file123",
			Name:         "test.txt",
			MimeType:     "text/plain",
			Size:         1024,
			ModifiedTime: "2024-01-15T10:30:00Z",
		},
		{
			ID:           "file456",
			Name:         "document.pdf",
			MimeType:     "application/pdf",
			Size:         2048000,
			ModifiedTime: "2024-01-16T14:20:00Z",
		},
	}

	err := formatter.WriteSuccess("files list", files)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	output := buf.String()

	// Verify table contains file information
	if !strings.Contains(output, "test.txt") {
		t.Error("Expected output to contain 'test.txt'")
	}

	if !strings.Contains(output, "document.pdf") {
		t.Error("Expected output to contain 'document.pdf'")
	}

	// File IDs should be present (possibly truncated)
	if !strings.Contains(output, "file123") && !strings.Contains(output, "file1") {
		t.Error("Expected output to contain file ID")
	}
}

func TestOutputFormatterQuietMode(t *testing.T) {
	var buf, errBuf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatJSON,
		Quiet:  true,
	})
	formatter.writer = &buf
	formatter.errorWriter = &errBuf

	// Log should not produce output in quiet mode
	formatter.Log("This should not appear")

	if errBuf.Len() > 0 {
		t.Errorf("Expected no output in quiet mode, got: %s", errBuf.String())
	}
}

func TestOutputFormatterVerboseMode(t *testing.T) {
	var buf, errBuf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format:  types.OutputFormatJSON,
		Verbose: true,
	})
	formatter.writer = &buf
	formatter.errorWriter = &errBuf

	// Verbose should produce output
	formatter.Verbose("Verbose message")

	if errBuf.Len() == 0 {
		t.Error("Expected verbose output")
	}

	if !strings.Contains(errBuf.String(), "Verbose message") {
		t.Errorf("Expected verbose message, got: %s", errBuf.String())
	}

	if !strings.Contains(errBuf.String(), "[VERBOSE]") {
		t.Error("Expected [VERBOSE] prefix")
	}
}

func TestOutputFormatterPermissionTable(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatTable,
	})
	formatter.writer = &buf

	perms := []*types.Permission{
		{
			ID:           "perm123",
			Type:         "user",
			Role:         "writer",
			EmailAddress: "user@example.com",
			DisplayName:  "Test User",
		},
		{
			ID:     "perm456",
			Type:   "domain",
			Role:   "reader",
			Domain: "example.com",
		},
	}

	err := formatter.WriteSuccess("permissions list", perms)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	output := buf.String()

	// Verify table contains permission information
	if !strings.Contains(output, "user@example.com") {
		t.Error("Expected output to contain email")
	}

	if !strings.Contains(output, "example.com") {
		t.Error("Expected output to contain domain")
	}

	if !strings.Contains(output, "writer") {
		t.Error("Expected output to contain role 'writer'")
	}

	if !strings.Contains(output, "reader") {
		t.Error("Expected output to contain role 'reader'")
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		mimeType string
		want     string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			mimeType: "text/plain",
			want:     "-",
		},
		{
			name:     "bytes",
			bytes:    512,
			mimeType: "text/plain",
			want:     "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536,
			mimeType: "text/plain",
			want:     "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    2048000,
			mimeType: "application/pdf",
			want:     "2.0 MB",
		},
		{
			name:     "google doc",
			bytes:    1024,
			mimeType: "application/vnd.google-apps.document",
			want:     "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFileSize(tt.bytes, tt.mimeType)
			if got != tt.want {
				t.Errorf("formatFileSize(%d, %s) = %s, want %s",
					tt.bytes, tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestFormatMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		want     string
	}{
		{
			mimeType: "application/vnd.google-apps.folder",
			want:     "Folder",
		},
		{
			mimeType: "application/vnd.google-apps.document",
			want:     "Doc",
		},
		{
			mimeType: "application/vnd.google-apps.spreadsheet",
			want:     "Sheet",
		},
		{
			mimeType: "application/pdf",
			want:     "PDF",
		},
		{
			mimeType: "text/plain",
			want:     "Text",
		},
		{
			mimeType: "image/jpeg",
			want:     "JPEG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := formatMimeType(tt.mimeType)
			if got != tt.want {
				t.Errorf("formatMimeType(%s) = %s, want %s", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "truncate with ellipsis",
			input:  "this is a very long string",
			maxLen: 10,
			want:   "this is...",
		},
		{
			name:   "exact length",
			input:  "exactly",
			maxLen: 7,
			want:   "exactly",
		},
		{
			name:   "very short maxLen",
			input:  "test",
			maxLen: 2,
			want:   "te",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFileListResultTable(t *testing.T) {
	var buf, errBuf bytes.Buffer
	formatter := NewOutputFormatter(OutputOptions{
		Format: types.OutputFormatTable,
	})
	formatter.writer = &buf
	formatter.errorWriter = &errBuf

	result := &types.FileListResult{
		Files: []*types.DriveFile{
			{
				ID:       "file123",
				Name:     "test.txt",
				MimeType: "text/plain",
			},
		},
		NextPageToken:    "token123",
		IncompleteSearch: true,
	}

	err := formatter.WriteSuccess("files list", result)
	if err != nil {
		t.Fatalf("WriteSuccess failed: %v", err)
	}

	// Check main output
	output := buf.String()
	if !strings.Contains(output, "test.txt") {
		t.Error("Expected output to contain file name")
	}

	// Check error output for warnings
	errOutput := errBuf.String()
	if !strings.Contains(errOutput, "More results available") {
		t.Error("Expected pagination message")
	}

	if !strings.Contains(errOutput, "incomplete") {
		t.Error("Expected incomplete search warning")
	}
}

func TestGetPermissionIdentity(t *testing.T) {
	tests := []struct {
		name string
		perm *types.Permission
		want string
	}{
		{
			name: "user with email",
			perm: &types.Permission{
				Type:         "user",
				EmailAddress: "user@example.com",
			},
			want: "user@example.com",
		},
		{
			name: "domain permission",
			perm: &types.Permission{
				Type:   "domain",
				Domain: "example.com",
			},
			want: "example.com",
		},
		{
			name: "anyone permission",
			perm: &types.Permission{
				Type: "anyone",
			},
			want: "Anyone",
		},
		{
			name: "no identity",
			perm: &types.Permission{
				Type: "group",
			},
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPermissionIdentity(tt.perm)
			if got != tt.want {
				t.Errorf("getPermissionIdentity() = %s, want %s", got, tt.want)
			}
		})
	}
}
