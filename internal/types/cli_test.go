package types

import (
	"encoding/json"
	"testing"
)

func TestCLIOutput_JSONMarshaling(t *testing.T) {
	output := CLIOutput{
		SchemaVersion: "1.0",
		TraceID:       "trace-123",
		Command:       "files list",
		Data:          map[string]string{"result": "success"},
		Warnings:      []CLIWarning{{Code: "W001", Message: "test warning", Severity: "low"}},
		Errors:        []CLIError{{Code: "E001", Message: "test error", Retryable: false}},
	}

	// Marshal to JSON
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal CLIOutput: %v", err)
	}

	// Unmarshal back
	var decoded CLIOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CLIOutput: %v", err)
	}

	// Verify fields
	if decoded.SchemaVersion != output.SchemaVersion {
		t.Errorf("SchemaVersion = %s, want %s", decoded.SchemaVersion, output.SchemaVersion)
	}

	if decoded.TraceID != output.TraceID {
		t.Errorf("TraceID = %s, want %s", decoded.TraceID, output.TraceID)
	}

	if decoded.Command != output.Command {
		t.Errorf("Command = %s, want %s", decoded.Command, output.Command)
	}

	if len(decoded.Warnings) != len(output.Warnings) {
		t.Errorf("Warnings length = %d, want %d", len(decoded.Warnings), len(output.Warnings))
	}

	if len(decoded.Errors) != len(output.Errors) {
		t.Errorf("Errors length = %d, want %d", len(decoded.Errors), len(output.Errors))
	}
}

func TestCLIWarning_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		warning  CLIWarning
		severity string
	}{
		{
			name:     "Low severity",
			warning:  CLIWarning{Code: "W001", Message: "Low severity warning", Severity: "low"},
			severity: "low",
		},
		{
			name:     "Medium severity",
			warning:  CLIWarning{Code: "W002", Message: "Medium severity warning", Severity: "medium"},
			severity: "medium",
		},
		{
			name:     "High severity",
			warning:  CLIWarning{Code: "W003", Message: "High severity warning", Severity: "high"},
			severity: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.warning)
			if err != nil {
				t.Fatalf("Failed to marshal CLIWarning: %v", err)
			}

			var decoded CLIWarning
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal CLIWarning: %v", err)
			}

			if decoded.Code != tt.warning.Code {
				t.Errorf("Code = %s, want %s", decoded.Code, tt.warning.Code)
			}

			if decoded.Message != tt.warning.Message {
				t.Errorf("Message = %s, want %s", decoded.Message, tt.warning.Message)
			}

			if decoded.Severity != tt.severity {
				t.Errorf("Severity = %s, want %s", decoded.Severity, tt.severity)
			}
		})
	}
}

func TestCLIError_JSONMarshaling(t *testing.T) {
	cliError := CLIError{
		Code:        "E001",
		Message:     "Test error",
		HTTPStatus:  404,
		DriveReason: "notFound",
		Retryable:   true,
		Context:     map[string]interface{}{"file_id": "123"},
	}

	data, err := json.Marshal(cliError)
	if err != nil {
		t.Fatalf("Failed to marshal CLIError: %v", err)
	}

	var decoded CLIError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CLIError: %v", err)
	}

	if decoded.Code != cliError.Code {
		t.Errorf("Code = %s, want %s", decoded.Code, cliError.Code)
	}

	if decoded.Message != cliError.Message {
		t.Errorf("Message = %s, want %s", decoded.Message, cliError.Message)
	}

	if decoded.HTTPStatus != cliError.HTTPStatus {
		t.Errorf("HTTPStatus = %d, want %d", decoded.HTTPStatus, cliError.HTTPStatus)
	}

	if decoded.DriveReason != cliError.DriveReason {
		t.Errorf("DriveReason = %s, want %s", decoded.DriveReason, cliError.DriveReason)
	}

	if decoded.Retryable != cliError.Retryable {
		t.Errorf("Retryable = %v, want %v", decoded.Retryable, cliError.Retryable)
	}

	if len(decoded.Context) != len(cliError.Context) {
		t.Errorf("Context length = %d, want %d", len(decoded.Context), len(cliError.Context))
	}
}

func TestCLIError_OmitEmpty(t *testing.T) {
	// Test that empty optional fields are omitted
	cliError := CLIError{
		Code:      "E001",
		Message:   "Test error",
		Retryable: false,
	}

	data, err := json.Marshal(cliError)
	if err != nil {
		t.Fatalf("Failed to marshal CLIError: %v", err)
	}

	jsonStr := string(data)

	// These fields should not be present when empty
	if contains(jsonStr, "httpStatus") {
		t.Error("httpStatus should be omitted when zero")
	}

	if contains(jsonStr, "driveReason") {
		t.Error("driveReason should be omitted when empty")
	}

	if contains(jsonStr, "context") {
		t.Error("context should be omitted when nil")
	}
}

func TestOutputFormat_String(t *testing.T) {
	tests := []struct {
		format OutputFormat
		want   string
	}{
		{OutputFormatJSON, "json"},
		{OutputFormatTable, "table"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if string(tt.format) != tt.want {
				t.Errorf("OutputFormat = %s, want %s", tt.format, tt.want)
			}
		})
	}
}

func TestOutputFormat_Constants(t *testing.T) {
	if OutputFormatJSON != "json" {
		t.Errorf("OutputFormatJSON = %s, want json", OutputFormatJSON)
	}

	if OutputFormatTable != "table" {
		t.Errorf("OutputFormatTable = %s, want table", OutputFormatTable)
	}
}

func TestGlobalFlags_DefaultValues(t *testing.T) {
	flags := GlobalFlags{}

	// Test default zero values
	if flags.Profile != "" {
		t.Errorf("Default Profile = %s, want empty", flags.Profile)
	}

	if flags.DriveID != "" {
		t.Errorf("Default DriveID = %s, want empty", flags.DriveID)
	}

	if flags.Quiet {
		t.Error("Default Quiet should be false")
	}

	if flags.Verbose {
		t.Error("Default Verbose should be false")
	}

	if flags.Debug {
		t.Error("Default Debug should be false")
	}

	if flags.DryRun {
		t.Error("Default DryRun should be false")
	}

	if flags.CacheTTL != 0 {
		t.Errorf("Default CacheTTL = %d, want 0", flags.CacheTTL)
	}
}

func TestGlobalFlags_AllFields(t *testing.T) {
	flags := GlobalFlags{
		Profile:             "test-profile",
		DriveID:             "drive-123",
		OutputFormat:        OutputFormatJSON,
		Quiet:               true,
		Verbose:             true,
		Debug:               true,
		Strict:              true,
		NoCache:             true,
		CacheTTL:            300,
		IncludeSharedWithMe: true,
		Config:              "/path/to/config",
		LogFile:             "/path/to/log",
		DryRun:              true,
		Force:               true,
		Yes:                 true,
		JSON:                true,
	}

	// Verify all fields are set
	if flags.Profile != "test-profile" {
		t.Errorf("Profile = %s, want test-profile", flags.Profile)
	}

	if flags.DriveID != "drive-123" {
		t.Errorf("DriveID = %s, want drive-123", flags.DriveID)
	}

	if flags.OutputFormat != OutputFormatJSON {
		t.Errorf("OutputFormat = %s, want json", flags.OutputFormat)
	}

	if !flags.Quiet {
		t.Error("Quiet should be true")
	}

	if !flags.Verbose {
		t.Error("Verbose should be true")
	}

	if !flags.Debug {
		t.Error("Debug should be true")
	}

	if !flags.Strict {
		t.Error("Strict should be true")
	}

	if !flags.NoCache {
		t.Error("NoCache should be true")
	}

	if flags.CacheTTL != 300 {
		t.Errorf("CacheTTL = %d, want 300", flags.CacheTTL)
	}

	if !flags.IncludeSharedWithMe {
		t.Error("IncludeSharedWithMe should be true")
	}

	if flags.Config != "/path/to/config" {
		t.Errorf("Config = %s, want /path/to/config", flags.Config)
	}

	if flags.LogFile != "/path/to/log" {
		t.Errorf("LogFile = %s, want /path/to/log", flags.LogFile)
	}

	if !flags.DryRun {
		t.Error("DryRun should be true")
	}

	if !flags.Force {
		t.Error("Force should be true")
	}

	if !flags.Yes {
		t.Error("Yes should be true")
	}

	if !flags.JSON {
		t.Error("JSON should be true")
	}
}

func TestCLIOutput_EmptyWarningsAndErrors(t *testing.T) {
	output := CLIOutput{
		SchemaVersion: "1.0",
		TraceID:       "trace-123",
		Command:       "files list",
		Data:          nil,
		Warnings:      []CLIWarning{},
		Errors:        []CLIError{},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal CLIOutput: %v", err)
	}

	var decoded CLIOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CLIOutput: %v", err)
	}

	if decoded.Warnings == nil {
		t.Error("Warnings should be empty slice, not nil")
	}

	if decoded.Errors == nil {
		t.Error("Errors should be empty slice, not nil")
	}

	if len(decoded.Warnings) != 0 {
		t.Errorf("Warnings length = %d, want 0", len(decoded.Warnings))
	}

	if len(decoded.Errors) != 0 {
		t.Errorf("Errors length = %d, want 0", len(decoded.Errors))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
