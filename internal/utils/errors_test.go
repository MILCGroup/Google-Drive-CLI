package utils

import (
	"testing"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{ErrCodeAuthRequired, ExitAuthRequired},
		{ErrCodeFileNotFound, ExitFileNotFound},
		{ErrCodePermissionDenied, ExitPermissionDenied},
		{ErrCodeQuotaExceeded, ExitQuotaExceeded},
		{ErrCodeRateLimited, ExitRateLimited},
		{"UNKNOWN_CODE", ExitUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := GetExitCode(tt.code)
			if got != tt.expected {
				t.Errorf("GetExitCode(%s) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestNewCLIError(t *testing.T) {
	err := NewCLIError(ErrCodeFileNotFound, "File not found").
		WithHTTPStatus(404).
		WithDriveReason("notFound").
		WithRetryable(false).
		WithContext("fileId", "abc123").
		Build()

	if err.Code != ErrCodeFileNotFound {
		t.Errorf("Code = %s, want %s", err.Code, ErrCodeFileNotFound)
	}
	if err.HTTPStatus != 404 {
		t.Errorf("HTTPStatus = %d, want 404", err.HTTPStatus)
	}
	if err.DriveReason != "notFound" {
		t.Errorf("DriveReason = %s, want notFound", err.DriveReason)
	}
	if err.Retryable {
		t.Error("Retryable should be false")
	}
	if err.Context["fileId"] != "abc123" {
		t.Errorf("Context[fileId] = %v, want abc123", err.Context["fileId"])
	}
}

func TestAppError(t *testing.T) {
	cliErr := NewCLIError(ErrCodeFileNotFound, "test message").Build()
	appErr := NewAppError(cliErr)

	expected := "FILE_NOT_FOUND: test message"
	if appErr.Error() != expected {
		t.Errorf("Error() = %s, want %s", appErr.Error(), expected)
	}
}
