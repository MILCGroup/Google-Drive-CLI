package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
)

// ============================================================
// Globals and GlobalFlags Tests
// ============================================================

func TestGlobalsToGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		globals  Globals
		expected types.GlobalFlags
	}{
		{
			name: "json output",
			globals: Globals{
				Profile:  "default",
				DriveID:  "drive123",
				Output:   "json",
				Quiet:    true,
				Verbose:  false,
				Debug:    false,
				Strict:   true,
				NoCache:  true,
				CacheTTL: 600,
				Config:   "/path/to/config",
				LogFile:  "/path/to/log",
				DryRun:   true,
				Force:    true,
				Yes:      true,
				JSON:     true,
			},
			expected: types.GlobalFlags{
				Profile:      "default",
				DriveID:      "drive123",
				OutputFormat: types.OutputFormatJSON,
				Quiet:        true,
				Verbose:      false,
				Debug:        false,
				Strict:       true,
				NoCache:      true,
				CacheTTL:     600,
				Config:       "/path/to/config",
				LogFile:      "/path/to/log",
				DryRun:       true,
				Force:        true,
				Yes:          true,
				JSON:         true,
			},
		},
		{
			name: "table output",
			globals: Globals{
				Profile: "test",
				Output:  "table",
				JSON:    false,
			},
			expected: types.GlobalFlags{
				Profile:      "test",
				OutputFormat: types.OutputFormatTable,
				JSON:         false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.globals.ToGlobalFlags()
			if result.Profile != tc.expected.Profile {
				t.Errorf("Profile: got %q, want %q", result.Profile, tc.expected.Profile)
			}
			if result.OutputFormat != tc.expected.OutputFormat {
				t.Errorf("OutputFormat: got %v, want %v", result.OutputFormat, tc.expected.OutputFormat)
			}
			if result.DriveID != tc.expected.DriveID {
				t.Errorf("DriveID: got %q, want %q", result.DriveID, tc.expected.DriveID)
			}
			if result.Quiet != tc.expected.Quiet {
				t.Errorf("Quiet: got %v, want %v", result.Quiet, tc.expected.Quiet)
			}
			if result.Strict != tc.expected.Strict {
				t.Errorf("Strict: got %v, want %v", result.Strict, tc.expected.Strict)
			}
		})
	}
}

func TestGlobalsAfterApply_JSONFlag(t *testing.T) {
	g := &Globals{
		JSON:   true,
		Output: "table",
	}
	err := g.AfterApply()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Output != "json" {
		t.Errorf("expected output to be 'json' when JSON flag is true, got %q", g.Output)
	}
}

func TestGlobalsAfterApply_InvalidOutput(t *testing.T) {
	g := &Globals{
		Output: "invalid",
	}
	err := g.AfterApply()
	if err == nil {
		t.Fatal("expected error for invalid output format")
	}
}

func TestGlobalsAfterApply_ValidOutputFormats(t *testing.T) {
	validFormats := []string{"json", "table"}
	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			g := &Globals{
				Output: format,
			}
			err := g.AfterApply()
			if err != nil {
				t.Fatalf("unexpected error for valid format %q: %v", format, err)
			}
		})
	}
}

// ============================================================
// isPath Tests
// ============================================================

func TestIsPath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/folder/file.txt", true},
		{"folder/file.txt", true},
		{"file with spaces", true},
		{"file.txt", true},
		{"document.pdf", true},
		{"1ABC123xyz789", false}, // Looks like a file ID
		{"abc123def456ghi789", false},
		{"just_a_file", false},
		{"", false}, // Empty is not a path
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := isPath(tc.input)
			if result != tc.expected {
				t.Errorf("isPath(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

// ============================================================
// GetResolveOptions Tests
// ============================================================

func TestGetResolveOptions(t *testing.T) {
	flags := types.GlobalFlags{
		DriveID:             "drive123",
		IncludeSharedWithMe: true,
		NoCache:             false,
		Strict:              true,
	}

	opts := GetResolveOptions(flags)

	if opts.DriveID != "drive123" {
		t.Errorf("DriveID: got %q, want %q", opts.DriveID, "drive123")
	}
	if !opts.IncludeSharedWithMe {
		t.Error("IncludeSharedWithMe should be true")
	}
	if !opts.UseCache {
		t.Error("UseCache should be true when NoCache is false")
	}
	if !opts.StrictMode {
		t.Error("StrictMode should be true")
	}
}

func TestGetResolveOptions_NoCache(t *testing.T) {
	flags := types.GlobalFlags{
		NoCache: true,
	}

	opts := GetResolveOptions(flags)

	if opts.UseCache {
		t.Error("UseCache should be false when NoCache is true")
	}
}

// ============================================================
// getConfigDir Tests
// ============================================================

func TestGetConfigDir(t *testing.T) {
	// Save original env var
	original := os.Getenv("GDRV_CONFIG_DIR")
	defer func() { _ = os.Setenv("GDRV_CONFIG_DIR", original) }()

	tests := []struct {
		name     string
		envValue string
		wantErr  bool
	}{
		{
			name:     "from env var",
			envValue: "/custom/config/dir",
			wantErr:  false,
		},
		{
			name:     "default location",
			envValue: "",
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				_ = os.Setenv("GDRV_CONFIG_DIR", tc.envValue)
			} else {
				_ = os.Unsetenv("GDRV_CONFIG_DIR")
			}

			dir := getConfigDir()

			if tc.envValue != "" {
				if dir != tc.envValue {
					t.Errorf("got %q, want %q", dir, tc.envValue)
				}
			} else {
				// Should contain gdrv (either in .config or Application Support)
				if !contains(dir, "gdrv") {
					t.Errorf("expected path to contain gdrv, got %q", dir)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================
// isTruthyEnv Tests
// ============================================================

func TestIsTruthyEnv(t *testing.T) {
	original := os.Getenv("TEST_VAR")
	defer func() { _ = os.Setenv("TEST_VAR", original) }()

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"on", "on", true},
		{"TRUE uppercase", "TRUE", true},
		{"YES uppercase", "YES", true},
		{"ON uppercase", "ON", true},
		{"TrUe mixed", "TrUe", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"off", "off", false},
		{"empty", "", false},
		{"random", "random", false},
		{"unset", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue == "" && tc.name == "unset" {
				_ = os.Unsetenv("TEST_VAR")
			} else {
				_ = os.Setenv("TEST_VAR", tc.envValue)
			}
			result := isTruthyEnv("TEST_VAR")
			if result != tc.expected {
				t.Errorf("isTruthyEnv(%q=%q) = %v, want %v", "TEST_VAR", tc.envValue, result, tc.expected)
			}
		})
	}
}

// ============================================================
// resolveAuthScopes Tests
// ============================================================

func TestResolveAuthScopes_WithPreset(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)

	scopes, err := resolveAuthScopes(out, "workspace-basic", false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scopes) == 0 {
		t.Error("expected scopes to be returned")
	}
}

func TestResolveAuthScopes_WithWide(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)

	scopes, err := resolveAuthScopes(out, "", true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scopes) != 1 {
		t.Errorf("expected 1 scope for wide, got %d", len(scopes))
	}

	if scopes[0] != utils.ScopeFull {
		t.Errorf("expected %q, got %q", utils.ScopeFull, scopes[0])
	}
}

func TestResolveAuthScopes_WithCommandScopes(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)
	commandScopes := []string{"scope1", "scope2"}

	scopes, err := resolveAuthScopes(out, "", false, commandScopes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(scopes))
	}
}

func TestResolveAuthScopes_DefaultPreset(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)

	scopes, err := resolveAuthScopes(out, "", false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should default to workspace-complete
	if len(scopes) == 0 {
		t.Error("expected default scopes to be returned")
	}
}

func TestResolveAuthScopes_InvalidPreset(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)

	_, err := resolveAuthScopes(out, "invalid-preset", false, nil)
	if err == nil {
		t.Fatal("expected error for invalid preset")
	}
}

// ============================================================
// scopesForPreset Tests
// ============================================================

func TestScopesForPreset(t *testing.T) {
	tests := []struct {
		name          string
		preset        string
		expectError   bool
		expectedScope string // checks if this scope is in the result
	}{
		{"workspace-basic", "workspace-basic", false, "drive"},
		{"workspace-full", "workspace-full", false, "drive"},
		{"admin", "admin", false, "admin"},
		{"workspace-with-admin", "workspace-with-admin", false, "admin"},
		{"workspace-activity", "workspace-activity", false, "activity"},
		{"workspace-labels", "workspace-labels", false, "labels"},
		{"workspace-sync", "workspace-sync", false, "drive"},
		{"workspace-complete", "workspace-complete", false, "drive"},
		{"gmail", "gmail", false, "gmail"},
		{"gmail-readonly", "gmail-readonly", false, "gmail.readonly"},
		{"calendar", "calendar", false, "calendar"},
		{"calendar-readonly", "calendar-readonly", false, "calendar.readonly"},
		{"people", "people", false, "contacts"},
		{"tasks", "tasks", false, "tasks"},
		{"forms", "forms", false, "forms"},
		{"appscript", "appscript", false, "script"},
		{"groups", "groups", false, "groups"},
		{"suite-complete", "suite-complete", false, "drive"},
		{"invalid", "invalid-preset", true, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scopes, err := scopesForPreset(tc.preset)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(scopes) == 0 {
				t.Error("expected non-empty scopes")
			}

			// Check if expected scope is present
			if tc.expectedScope != "" {
				found := false
				for _, s := range scopes {
					if contains(s, tc.expectedScope) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected scope containing %q not found in %v", tc.expectedScope, scopes)
				}
			}
		})
	}
}

// ============================================================
// validateAdminScopesRequireImpersonation Tests
// ============================================================

func TestValidateAdminScopesRequireImpersonation(t *testing.T) {
	tests := []struct {
		name            string
		scopes          []string
		impersonateUser string
		expectError     bool
	}{
		{
			name:            "no admin scopes",
			scopes:          []string{"https://www.googleapis.com/auth/drive"},
			impersonateUser: "",
			expectError:     false,
		},
		{
			name:            "admin scope with impersonation",
			scopes:          []string{utils.ScopeAdminDirectoryUser},
			impersonateUser: "admin@example.com",
			expectError:     false,
		},
		{
			name:            "admin scope without impersonation",
			scopes:          []string{utils.ScopeAdminDirectoryUser},
			impersonateUser: "",
			expectError:     true,
		},
		{
			name:            "readonly admin scope without impersonation",
			scopes:          []string{utils.ScopeAdminDirectoryUserReadonly},
			impersonateUser: "",
			expectError:     true,
		},
		{
			name:            "group admin scope without impersonation",
			scopes:          []string{utils.ScopeAdminDirectoryGroup},
			impersonateUser: "",
			expectError:     true,
		},
		{
			name:            "mixed scopes with admin and impersonation",
			scopes:          []string{"https://www.googleapis.com/auth/drive", utils.ScopeAdminDirectoryUser},
			impersonateUser: "admin@example.com",
			expectError:     false,
		},
		{
			name:            "mixed scopes with admin no impersonation",
			scopes:          []string{"https://www.googleapis.com/auth/drive", utils.ScopeAdminDirectoryUser},
			impersonateUser: "",
			expectError:     true,
		},
		{
			name:            "empty scopes",
			scopes:          []string{},
			impersonateUser: "",
			expectError:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAdminScopesRequireImpersonation(tc.scopes, tc.impersonateUser)
			if tc.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ============================================================
// splitCSV Tests (from people.go)
// ============================================================

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single value",
			input:    "value1",
			expected: []string{"value1"},
		},
		{
			name:     "multiple values",
			input:    "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "values with spaces",
			input:    " value1 , value2 , value3 ",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "single value with spaces",
			input:    "  value1  ",
			expected: []string{"value1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := splitCSV(tc.input)

			if tc.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tc.expected) {
				t.Errorf("expected %d items, got %d", len(tc.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tc.expected[i] {
					t.Errorf("item %d: got %q, want %q", i, v, tc.expected[i])
				}
			}
		})
	}
}

// ============================================================
// OutputWriter WriteSuccess/WriteError Tests
// ============================================================

func TestOutputWriterWriteSuccess_JSON(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, true, false)
	data := map[string]string{"key": "value"}

	// This should not error when writing JSON
	err := w.WriteSuccess("test.command", data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutputWriterWriteError(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, true, false)
	cliErr := utils.NewCLIError(utils.ErrCodeFileNotFound, "resource not found").Build()

	err := w.WriteError("test.command", cliErr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutputWriterWriteSuccess_TableWithFiles(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatTable, true, false)
	files := []*types.DriveFile{
		{
			ID:           "file123",
			Name:         "test.txt",
			MimeType:     "text/plain",
			Size:         1024,
			ModifiedTime: "2024-01-01T00:00:00Z",
		},
	}

	// This should not error when writing table
	err := w.WriteSuccess("files.list", files)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutputWriterWriteSuccess_TableWithPermissions(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatTable, true, false)
	perms := []*types.Permission{
		{
			ID:           "perm123",
			Type:         "user",
			Role:         "writer",
			EmailAddress: "user@example.com",
		},
	}

	err := w.WriteSuccess("permissions.list", perms)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================
// handleCLIError Tests
// ============================================================

func TestHandleCLIError_WithAppError(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, true, false)
	cliErr := utils.NewCLIError(utils.ErrCodeFileNotFound, "not found").WithContext("key", "value").Build()
	appErr := utils.NewAppError(cliErr)

	err := handleCLIError(w, "test.command", appErr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleCLIError_WithRegularError(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, true, false)
	regularErr := &utils.AppError{
		CLIError: utils.NewCLIError(utils.ErrCodeUnknown, "unknown error").Build(),
	}

	err := handleCLIError(w, "test.command", regularErr)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================
// File Path Helpers
// ============================================================

func TestFilepathJoin(t *testing.T) {
	result := filepath.Join("path", "to", "file")
	if result == "" {
		t.Error("filepath.Join returned empty string")
	}
}

// ============================================================
// resolveTimeRange Tests
// ============================================================

func TestResolveTimeRange_Today(t *testing.T) {
	timeMin, timeMax := resolveTimeRange(true, false, 0)

	if timeMin == "" {
		t.Error("expected timeMin to be set for today")
	}
	if timeMax == "" {
		t.Error("expected timeMax to be set for today")
	}

	// Verify RFC3339 format
	if !isRFC3339(timeMin) {
		t.Errorf("timeMin is not RFC3339: %q", timeMin)
	}
	if !isRFC3339(timeMax) {
		t.Errorf("timeMax is not RFC3339: %q", timeMax)
	}
}

func TestResolveTimeRange_Week(t *testing.T) {
	timeMin, timeMax := resolveTimeRange(false, true, 0)

	if timeMin == "" {
		t.Error("expected timeMin to be set for week")
	}
	if timeMax == "" {
		t.Error("expected timeMax to be set for week")
	}

	if !isRFC3339(timeMin) {
		t.Errorf("timeMin is not RFC3339: %q", timeMin)
	}
	if !isRFC3339(timeMax) {
		t.Errorf("timeMax is not RFC3339: %q", timeMax)
	}
}

func TestResolveTimeRange_Days(t *testing.T) {
	timeMin, timeMax := resolveTimeRange(false, false, 5)

	if timeMin == "" {
		t.Error("expected timeMin to be set for days")
	}
	if timeMax == "" {
		t.Error("expected timeMax to be set for days")
	}

	if !isRFC3339(timeMin) {
		t.Errorf("timeMin is not RFC3339: %q", timeMin)
	}
	if !isRFC3339(timeMax) {
		t.Errorf("timeMax is not RFC3339: %q", timeMax)
	}
}

func TestResolveTimeRange_None(t *testing.T) {
	timeMin, timeMax := resolveTimeRange(false, false, 0)

	if timeMin != "" {
		t.Errorf("expected empty timeMin, got %q", timeMin)
	}
	if timeMax != "" {
		t.Errorf("expected empty timeMax, got %q", timeMax)
	}
}

func isRFC3339(s string) bool {
	// Basic check for RFC3339 format (contains T and ends with Z or timezone)
	return len(s) > 10 && (contains(s, "T") || contains(s, "t"))
}

// ============================================================
// convertDriveFile Tests
// ============================================================

func TestConvertDriveFile(t *testing.T) {
	driveFile := &drive.File{
		Id:           "file123",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         1024,
		Md5Checksum:  "abc123",
		CreatedTime:  "2024-01-01T00:00:00Z",
		ModifiedTime: "2024-01-02T00:00:00Z",
		Parents:      []string{"folder123"},
		ResourceKey:  "key123",
		Trashed:      false,
		Capabilities: &drive.FileCapabilities{
			CanDownload:      true,
			CanEdit:          true,
			CanShare:         false,
			CanDelete:        true,
			CanTrash:         true,
			CanReadRevisions: false,
		},
	}

	result := convertDriveFile(driveFile)

	if result.ID != "file123" {
		t.Errorf("ID: got %q, want %q", result.ID, "file123")
	}
	if result.Name != "test.txt" {
		t.Errorf("Name: got %q, want %q", result.Name, "test.txt")
	}
	if result.MimeType != "text/plain" {
		t.Errorf("MimeType: got %q, want %q", result.MimeType, "text/plain")
	}
	if result.Size != 1024 {
		t.Errorf("Size: got %d, want %d", result.Size, 1024)
	}
	if result.MD5Checksum != "abc123" {
		t.Errorf("MD5Checksum: got %q, want %q", result.MD5Checksum, "abc123")
	}
	if result.Trashed != false {
		t.Errorf("Trashed: got %v, want %v", result.Trashed, false)
	}

	// Test capabilities
	if result.Capabilities == nil {
		t.Fatal("expected Capabilities to be set")
	}
	if !result.Capabilities.CanDownload {
		t.Error("expected CanDownload to be true")
	}
	if result.Capabilities.CanShare {
		t.Error("expected CanShare to be false")
	}
}

func TestConvertDriveFile_NoCapabilities(t *testing.T) {
	driveFile := &drive.File{
		Id:   "file456",
		Name: "minimal.txt",
	}

	result := convertDriveFile(driveFile)

	if result.ID != "file456" {
		t.Errorf("ID: got %q, want %q", result.ID, "file456")
	}
	if result.Name != "minimal.txt" {
		t.Errorf("Name: got %q, want %q", result.Name, "minimal.txt")
	}
	if result.Capabilities != nil {
		t.Error("expected Capabilities to be nil when not provided")
	}
}

// ============================================================
// buildOAuthClientError Tests
// ============================================================

func TestBuildOAuthClientError(t *testing.T) {
	builder := buildOAuthClientError(utils.ErrCodeAuthClientMissing, "/test/config", "test message")
	cliErr := builder.Build()

	if cliErr.Code != utils.ErrCodeAuthClientMissing {
		t.Errorf("Code: got %q, want %q", cliErr.Code, utils.ErrCodeAuthClientMissing)
	}

	if !contains(cliErr.Message, "test message") {
		t.Errorf("expected message to contain 'test message', got %q", cliErr.Message)
	}

	if !contains(cliErr.Message, "Config path") {
		t.Errorf("expected message to contain 'Config path', got %q", cliErr.Message)
	}

	// Check context
	if cliErr.Context == nil {
		t.Fatal("expected context to be set")
	}
	if _, ok := cliErr.Context["configPath"]; !ok {
		t.Error("expected configPath in context")
	}
	if _, ok := cliErr.Context["tokenLocation"]; !ok {
		t.Error("expected tokenLocation in context")
	}
}

// ============================================================
// ActivityQueryResult Table Tests
// ============================================================

func TestActivityQueryResult_Headers(t *testing.T) {
	result := &ActivityQueryResult{}
	headers := result.Headers()
	expected := []string{"Timestamp", "Action", "Actor", "Target"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
	for i, h := range headers {
		if h != expected[i] {
			t.Errorf("header %d: got %q, want %q", i, h, expected[i])
		}
	}
}

func TestActivityQueryResult_Rows(t *testing.T) {
	result := &ActivityQueryResult{
		Activities: []types.Activity{
			{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				PrimaryActionDetail: types.ActionDetail{
					Type: "edit",
				},
				Actors: []types.Actor{
					{
						Type: "user",
						User: &types.ActivityUser{Email: "user@example.com"},
					},
				},
				Targets: []types.Target{
					{
						Type:      "driveItem",
						DriveItem: &types.DriveItem{Title: "Document.txt"},
					},
				},
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// Check timestamp format
	if !contains(rows[0][0], "2024") {
		t.Errorf("expected year in timestamp, got %q", rows[0][0])
	}
	if rows[0][1] != "edit" {
		t.Errorf("action: got %q, want %q", rows[0][1], "edit")
	}
	if rows[0][2] != "user@example.com" {
		t.Errorf("actor: got %q, want %q", rows[0][2], "user@example.com")
	}
	if rows[0][3] != "Document.txt" {
		t.Errorf("target: got %q, want %q", rows[0][3], "Document.txt")
	}
}

func TestActivityQueryResult_Rows_NoUser(t *testing.T) {
	result := &ActivityQueryResult{
		Activities: []types.Activity{
			{
				Timestamp: time.Now(),
				PrimaryActionDetail: types.ActionDetail{
					Type: "system",
				},
				Actors: []types.Actor{
					{Type: "system"},
				},
				Targets: []types.Target{
					{Type: "drive"},
				},
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][2] != "system" {
		t.Errorf("actor: got %q, want %q", rows[0][2], "system")
	}
	if rows[0][3] != "drive" {
		t.Errorf("target: got %q, want %q", rows[0][3], "drive")
	}
}

func TestActivityQueryResult_EmptyMessage(t *testing.T) {
	result := &ActivityQueryResult{}
	if result.EmptyMessage() != "No activity found" {
		t.Errorf("expected 'No activity found', got %q", result.EmptyMessage())
	}
}

func TestActivityQueryResult_Rows_Empty(t *testing.T) {
	result := &ActivityQueryResult{
		Activities: []types.Activity{},
	}
	rows := result.Rows()
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// ============================================================
// Labels Table Result Tests
// ============================================================

func TestLabelsListResult_Headers(t *testing.T) {
	result := &LabelsListResult{}
	headers := result.Headers()
	expected := []string{"ID", "Name", "Type", "State"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestLabelsListResult_Rows(t *testing.T) {
	result := &LabelsListResult{
		Labels: []*types.Label{
			{
				ID:        "label1",
				Name:      "Test Label",
				LabelType: "SHARED",
				Lifecycle: &types.LabelLifecycle{
					State: "PUBLISHED",
				},
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "label1" {
		t.Errorf("ID: got %q, want %q", rows[0][0], "label1")
	}
	if rows[0][1] != "Test Label" {
		t.Errorf("Name: got %q, want %q", rows[0][1], "Test Label")
	}
	if rows[0][2] != "SHARED" {
		t.Errorf("Type: got %q, want %q", rows[0][2], "SHARED")
	}
	if rows[0][3] != "PUBLISHED" {
		t.Errorf("State: got %q, want %q", rows[0][3], "PUBLISHED")
	}
}

func TestLabelsListResult_Rows_NoLifecycle(t *testing.T) {
	result := &LabelsListResult{
		Labels: []*types.Label{
			{
				ID:        "label2",
				Name:      "No Lifecycle Label",
				LabelType: "ADMIN",
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][3] != "" {
		t.Errorf("State should be empty, got %q", rows[0][3])
	}
}

func TestLabelsListResult_EmptyMessage(t *testing.T) {
	result := &LabelsListResult{}
	if result.EmptyMessage() != "No labels found" {
		t.Errorf("expected 'No labels found', got %q", result.EmptyMessage())
	}
}

func TestLabelResult_Headers(t *testing.T) {
	result := &LabelResult{}
	headers := result.Headers()
	expected := []string{"ID", "Name", "Type", "State", "Fields"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestLabelResult_Rows(t *testing.T) {
	result := &LabelResult{
		Label: &types.Label{
			ID:        "label3",
			Name:      "Single Label",
			LabelType: "SHARED",
			Lifecycle: &types.LabelLifecycle{
				State: "PUBLISHED",
			},
			Fields: []*types.LabelField{
				{ID: "field1"},
				{ID: "field2"},
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][4] != "2" {
		t.Errorf("Fields count: got %q, want %q", rows[0][4], "2")
	}
}

func TestFileLabelsListResult_Headers(t *testing.T) {
	result := &FileLabelsListResult{}
	headers := result.Headers()
	expected := []string{"Label ID", "Revision ID", "Fields"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestFileLabelsListResult_Rows(t *testing.T) {
	result := &FileLabelsListResult{
		FileLabels: []*types.FileLabel{
			{
				ID:         "fl1",
				RevisionID: "rev1",
				Fields: map[string]*types.LabelFieldValue{
					"f1": {},
					"f2": {},
					"f3": {},
				},
			},
		},
	}

	rows := result.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "fl1" {
		t.Errorf("Label ID: got %q, want %q", rows[0][0], "fl1")
	}
	if rows[0][1] != "rev1" {
		t.Errorf("Revision ID: got %q, want %q", rows[0][1], "rev1")
	}
	if rows[0][2] != "3" {
		t.Errorf("Fields count: got %q, want %q", rows[0][2], "3")
	}
}

func TestFileLabelResult_Headers(t *testing.T) {
	result := &FileLabelResult{}
	headers := result.Headers()
	expected := []string{"Label ID", "Revision ID", "Fields"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestSuccessResult_Headers(t *testing.T) {
	result := &SuccessResult{Message: "Success!"}
	headers := result.Headers()
	expected := []string{"Status"}

	if len(headers) != len(expected) {
		t.Errorf("expected %d headers, got %d", len(expected), len(headers))
	}
}

func TestSuccessResult_Rows(t *testing.T) {
	result := &SuccessResult{Message: "Operation completed"}
	rows := result.Rows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][0] != "Operation completed" {
		t.Errorf("got %q, want %q", rows[0][0], "Operation completed")
	}
}

func TestSuccessResult_EmptyMessage(t *testing.T) {
	result := &SuccessResult{}
	if result.EmptyMessage() != "" {
		t.Errorf("expected empty string, got %q", result.EmptyMessage())
	}
}

// ============================================================
// OutputWriter writeTable with renderable types Tests
// ============================================================

func TestOutputWriterWriteSuccess_EmptyTable(t *testing.T) {
	// Test that empty table result renders empty message
	result := &LabelsListResult{
		Labels: []*types.Label{},
	}

	w := NewOutputWriter(types.OutputFormatTable, false, false)
	err := w.WriteSuccess("labels.list", result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutputWriterWriteSuccess_TableQuietMode(t *testing.T) {
	result := &LabelsListResult{
		Labels: []*types.Label{},
	}

	w := NewOutputWriter(types.OutputFormatTable, true, false) // quiet mode
	err := w.WriteSuccess("labels.list", result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================
// truncate Tests
// ============================================================

func TestTruncate_ExactLength(t *testing.T) {
	input := "exactly10!"
	result := truncate(input, 10)
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

func TestTruncate_LongString(t *testing.T) {
	input := "this is a very long string"
	result := truncate(input, 10)
	expected := "this is..."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
	if len(result) != 10 {
		t.Errorf("expected length 10, got %d", len(result))
	}
}

func TestTruncate_ShortString(t *testing.T) {
	input := "hi"
	result := truncate(input, 10)
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

// ============================================================
// formatSize Tests - Edge Cases
// ============================================================

func TestFormatSize_Zero(t *testing.T) {
	result := formatSize(0)
	if result != "0 B" {
		t.Errorf("expected '0 B', got %q", result)
	}
}

func TestFormatSize_Bytes(t *testing.T) {
	result := formatSize(512)
	if result != "512 B" {
		t.Errorf("expected '512 B', got %q", result)
	}
}

func TestFormatSize_KB(t *testing.T) {
	result := formatSize(1024)
	if result != "1.0 KB" {
		t.Errorf("expected '1.0 KB', got %q", result)
	}
}

func TestFormatSize_MB(t *testing.T) {
	result := formatSize(1024 * 1024)
	if result != "1.0 MB" {
		t.Errorf("expected '1.0 MB', got %q", result)
	}
}

func TestFormatSize_GB(t *testing.T) {
	result := formatSize(1024 * 1024 * 1024)
	if result != "1.0 GB" {
		t.Errorf("expected '1.0 GB', got %q", result)
	}
}

// ============================================================
// OutputWriter AddWarning Tests - Edge Cases
// ============================================================

func TestOutputWriterAddWarning_Multiple(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, false, false)
	w.AddWarning("WARN1", "first warning", "low")
	w.AddWarning("WARN2", "second warning", "medium")
	w.AddWarning("WARN3", "third warning", "high")

	if len(w.warnings) != 3 {
		t.Errorf("expected 3 warnings, got %d", len(w.warnings))
	}

	if w.warnings[0].Code != "WARN1" {
		t.Errorf("expected first warning code WARN1, got %s", w.warnings[0].Code)
	}

	if w.warnings[2].Severity != "high" {
		t.Errorf("expected third warning severity high, got %s", w.warnings[2].Severity)
	}
}

// ============================================================
// Additional Edge Case Tests
// ============================================================

func TestFormatSize_LargeValues(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{1024 * 1024 * 1024 * 1024, "1.0 TB"}, // 1 TB
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
	}

	for _, tc := range tests {
		result := formatSize(tc.bytes)
		if result != tc.expected {
			t.Errorf("formatSize(%d) = %q, want %q", tc.bytes, result, tc.expected)
		}
	}
}

func TestTruncate_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"", 10, ""},
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"12345678901", 10, "1234567..."},
		{"x", 10, "x"},
		{"", 0, ""},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.max)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.max, result, tc.expected)
		}
	}
}

func TestResolveAuthScopes_AllPresets(t *testing.T) {
	out := NewOutputWriter(types.OutputFormatJSON, true, false)

	presets := []string{
		"workspace-basic",
		"workspace-full",
		"admin",
		"workspace-with-admin",
		"workspace-activity",
		"workspace-labels",
		"workspace-sync",
		"workspace-complete",
		"gmail",
		"gmail-readonly",
		"calendar",
		"calendar-readonly",
		"people",
		"tasks",
		"forms",
		"appscript",
		"groups",
		"suite-complete",
	}

	for _, preset := range presets {
		t.Run(preset, func(t *testing.T) {
			scopes, err := scopesForPreset(preset)
			if err != nil {
				t.Errorf("unexpected error for preset %s: %v", preset, err)
			}
			if len(scopes) == 0 {
				t.Errorf("expected scopes for preset %s", preset)
			}
		})
	}

	// Test with OutputWriter
	for _, preset := range presets {
		t.Run("resolve_"+preset, func(t *testing.T) {
			scopes, err := resolveAuthScopes(out, preset, false, nil)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(scopes) == 0 {
				t.Error("expected scopes")
			}
		})
	}
}

func TestScopesForPreset_Unknown(t *testing.T) {
	_, err := scopesForPreset("unknown-preset")
	if err == nil {
		t.Error("expected error for unknown preset")
	}
}

func TestIsPath_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"a", false},
		{"ab", false},
		{"abc", false},
		{"a/b", true},
		{"file name", true},
		{"file.name", true},
		{"12345", false},      // short, no spaces/slashes
		{"abcdefghij", false}, // 10 chars, looks like ID
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := isPath(tc.input)
			if result != tc.expected {
				t.Errorf("isPath(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestSplitCSV_EmptyAndWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"   ", nil},
		{"a", []string{"a"}},
		{"  a  ", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{"  a  ,  b  ,  c  ", []string{"a", "b", "c"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := splitCSV(tc.input)
			if tc.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d items, got %d", len(tc.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("item %d: got %q, want %q", i, result[i], tc.expected[i])
				}
			}
		})
	}
}

func TestBuildAuthFlowError_VariousSources(t *testing.T) {
	sources := []oauthClientSource{
		oauthClientSourceFlags,
		oauthClientSourceEnv,
		oauthClientSourceConfig,
		oauthClientSourceBundled,
	}

	testErr := fmt.Errorf("test error")

	for _, source := range sources {
		t.Run(string(source), func(t *testing.T) {
			err := buildAuthFlowError(testErr, source, "client-id", "secret")
			if err == nil {
				t.Error("expected error builder")
			}
		})
	}
}
