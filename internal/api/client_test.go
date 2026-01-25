package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/googleapi"
)

func TestExecuteWithRetry_Success(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	callCount := 0
	result, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
		callCount++
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestExecuteWithRetry_RetryableError(t *testing.T) {
	client := NewClient(nil, 3, 10, logging.NewNoOpLogger()) // Short delay for testing
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	callCount := 0
	result, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", &googleapi.Error{Code: 503, Message: "Service Unavailable"}
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	client := NewClient(nil, 3, 10, logging.NewNoOpLogger())
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	callCount := 0
	_, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
		callCount++
		return "", &googleapi.Error{Code: 404, Message: "Not Found"}
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retry), got %d", callCount)
	}

	// Verify error is classified correctly
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("Expected *utils.AppError, got %T", err)
	}
	if appErr.CLIError.Code != utils.ErrCodeFileNotFound {
		t.Errorf("Expected FILE_NOT_FOUND, got %s", appErr.CLIError.Code)
	}
}

func TestExecuteWithRetry_MaxRetriesExceeded(t *testing.T) {
	client := NewClient(nil, 2, 10, logging.NewNoOpLogger()) // Max 2 retries
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	callCount := 0
	_, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
		callCount++
		return "", &googleapi.Error{Code: 503, Message: "Service Unavailable"}
	})

	if err == nil {
		t.Fatal("Expected error after max retries")
	}
	// Should try initial + 2 retries = 3 total
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestExecuteWithRetry_ContextCanceled(t *testing.T) {
	client := NewClient(nil, 5, 100, logging.NewNoOpLogger())
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	callCount := 0
	_, err := ExecuteWithRetry(ctx, client, reqCtx, func() (string, error) {
		callCount++
		return "", &googleapi.Error{Code: 503, Message: "Service Unavailable"}
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call before context check, got %d", callCount)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "429 Too Many Requests",
			err:       &googleapi.Error{Code: 429},
			retryable: true,
		},
		{
			name:      "500 Internal Server Error",
			err:       &googleapi.Error{Code: 500},
			retryable: true,
		},
		{
			name:      "502 Bad Gateway",
			err:       &googleapi.Error{Code: 502},
			retryable: true,
		},
		{
			name:      "503 Service Unavailable",
			err:       &googleapi.Error{Code: 503},
			retryable: true,
		},
		{
			name:      "504 Gateway Timeout",
			err:       &googleapi.Error{Code: 504},
			retryable: true,
		},
		{
			name:      "404 Not Found",
			err:       &googleapi.Error{Code: 404},
			retryable: false,
		},
		{
			name:      "403 Forbidden",
			err:       &googleapi.Error{Code: 403},
			retryable: false,
		},
		{
			name:      "Non-API error",
			err:       errors.New("network error"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryable(tt.err)
			if got != tt.retryable {
				t.Errorf("isRetryable() = %v, want %v", got, tt.retryable)
			}
		})
	}
}

func TestCalculateBackoff_RetryAfterHeader(t *testing.T) {
	baseDelay := 1000 * time.Millisecond

	header := make(map[string][]string)
	header["Retry-After"] = []string{"5"}

	err := &googleapi.Error{
		Code:   429,
		Header: header,
	}

	delay := calculateBackoff(baseDelay, 0, err)

	// Should use Retry-After value
	if delay != 5*time.Second {
		t.Errorf("Expected 5s delay from Retry-After, got %v", delay)
	}
}

func TestCalculateBackoff_ExponentialWithJitter(t *testing.T) {
	baseDelay := 1000 * time.Millisecond

	err := &googleapi.Error{Code: 503}

	// Test exponential growth
	delays := make([]time.Duration, 5)
	for i := 0; i < 5; i++ {
		delays[i] = calculateBackoff(baseDelay, i, err)
	}

	// Verify each delay is in expected range (with jitter)
	// Attempt 0: 1s ± 250ms
	if delays[0] < 750*time.Millisecond || delays[0] > 1250*time.Millisecond {
		t.Errorf("Attempt 0: delay %v out of expected range", delays[0])
	}

	// Attempt 1: 2s ± 500ms
	if delays[1] < 1500*time.Millisecond || delays[1] > 2500*time.Millisecond {
		t.Errorf("Attempt 1: delay %v out of expected range", delays[1])
	}

	// Attempt 2: 4s ± 1s
	if delays[2] < 3000*time.Millisecond || delays[2] > 5000*time.Millisecond {
		t.Errorf("Attempt 2: delay %v out of expected range", delays[2])
	}
}

func TestCalculateBackoff_MaxCap(t *testing.T) {
	baseDelay := 1000 * time.Millisecond
	err := &googleapi.Error{Code: 503}

	// High attempt number should hit max cap
	delay := calculateBackoff(baseDelay, 10, err)

	// Should be capped at MaxRetryDelayMs (32000ms) with jitter
	maxAllowed := time.Duration(utils.MaxRetryDelayMs)*time.Millisecond + 8*time.Second
	if delay > maxAllowed {
		t.Errorf("Delay %v exceeds max cap %v", delay, maxAllowed)
	}
}

func TestClassifyError_NetworkError(t *testing.T) {
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)
	err := errors.New("network timeout")

	appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

	if appErr.CLIError.Code != utils.ErrCodeNetworkError {
		t.Errorf("Expected NETWORK_ERROR, got %s", appErr.CLIError.Code)
	}
	if !appErr.CLIError.Retryable {
		t.Error("Network error should be retryable")
	}
}

func TestClassifyError_Auth(t *testing.T) {
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name     string
		httpCode int
		wantCode string
	}{
		{"Unauthorized", 401, utils.ErrCodeAuthExpired},
		{"Forbidden", 403, utils.ErrCodePermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &googleapi.Error{Code: tt.httpCode, Message: "Auth error"}
			appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if appErr.CLIError.Code != tt.wantCode {
				t.Errorf("Expected %s, got %s", tt.wantCode, appErr.CLIError.Code)
			}
		})
	}
}

func TestClassifyError_QuotaExceeded(t *testing.T) {
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	err := &googleapi.Error{
		Code:    403,
		Message: "Quota exceeded",
		Errors: []googleapi.ErrorItem{
			{Reason: "storageQuotaExceeded"},
		},
	}

	appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

	if appErr.CLIError.Code != utils.ErrCodeQuotaExceeded {
		t.Errorf("Expected QUOTA_EXCEEDED, got %s", appErr.CLIError.Code)
	}
	if appErr.CLIError.DriveReason != "storageQuotaExceeded" {
		t.Errorf("Expected driveReason 'storageQuotaExceeded', got %s", appErr.CLIError.DriveReason)
	}
}

func TestClassifyError_RateLimited(t *testing.T) {
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name      string
		httpCode  int
		reason    string
		retryable bool
	}{
		{"429 Rate Limit", 429, "", true},
		{"403 Sharing Rate Limit", 403, "sharingRateLimitExceeded", true},
		{"403 User Rate Limit", 403, "userRateLimitExceeded", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &googleapi.Error{
				Code:    tt.httpCode,
				Message: "Rate limit",
			}
			if tt.reason != "" {
				err.Errors = []googleapi.ErrorItem{{Reason: tt.reason}}
			}

			appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if appErr.CLIError.Code != utils.ErrCodeRateLimited {
				t.Errorf("Expected RATE_LIMITED, got %s", appErr.CLIError.Code)
			}
			if appErr.CLIError.Retryable != tt.retryable {
				t.Errorf("Expected retryable=%v, got %v", tt.retryable, appErr.CLIError.Retryable)
			}
		})
	}
}

func TestClassifyError_PolicyViolation(t *testing.T) {
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	err := &googleapi.Error{
		Code:    403,
		Message: "Domain policy",
		Errors: []googleapi.ErrorItem{
			{Reason: "domainPolicy"},
		},
	}

	appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

	if appErr.CLIError.Code != utils.ErrCodePolicyViolation {
		t.Errorf("Expected POLICY_VIOLATION, got %s", appErr.CLIError.Code)
	}
}

func TestClassifyError_ContextInclusion(t *testing.T) {
	reqCtx := NewRequestContext("test-profile", "drive-123", types.RequestTypeListOrSearch)
	reqCtx.TraceID = "trace-456"

	err := &googleapi.Error{Code: 404, Message: "Not found"}

	appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

	// Verify trace ID is included
	if traceID, ok := appErr.CLIError.Context["traceId"].(string); !ok || traceID != "trace-456" {
		t.Errorf("Expected traceId 'trace-456' in context, got %v", appErr.CLIError.Context["traceId"])
	}

	// Verify request type is included
	if reqType, ok := appErr.CLIError.Context["requestType"].(string); !ok || reqType != "ListOrSearch" {
		t.Errorf("Expected requestType 'ListOrSearch' in context, got %v", appErr.CLIError.Context["requestType"])
	}

	// Verify drive context for 404 with driveID
	if driveID, ok := appErr.CLIError.Context["driveId"].(string); !ok || driveID != "drive-123" {
		t.Errorf("Expected driveId 'drive-123' in context, got %v", appErr.CLIError.Context["driveId"])
	}
}

func TestNewRequestContext(t *testing.T) {
	ctx := NewRequestContext("my-profile", "drive-abc", types.RequestTypeMutation)

	if ctx.Profile != "my-profile" {
		t.Errorf("Expected profile 'my-profile', got %s", ctx.Profile)
	}
	if ctx.DriveID != "drive-abc" {
		t.Errorf("Expected driveID 'drive-abc', got %s", ctx.DriveID)
	}
	if ctx.RequestType != types.RequestTypeMutation {
		t.Errorf("Expected request type Mutation, got %s", ctx.RequestType)
	}
	if ctx.TraceID == "" {
		t.Error("Expected non-empty trace ID")
	}
	if ctx.InvolvedFileIDs == nil {
		t.Error("Expected initialized InvolvedFileIDs")
	}
	if ctx.InvolvedParentIDs == nil {
		t.Error("Expected initialized InvolvedParentIDs")
	}
}

func TestClient_WithFileIDs(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	ctx := NewRequestContext("default", "", types.RequestTypeGetByID)

	client.WithFileIDs(ctx, "file1", "file2")

	if len(ctx.InvolvedFileIDs) != 2 {
		t.Errorf("Expected 2 file IDs, got %d", len(ctx.InvolvedFileIDs))
	}
	if ctx.InvolvedFileIDs[0] != "file1" || ctx.InvolvedFileIDs[1] != "file2" {
		t.Errorf("Expected [file1, file2], got %v", ctx.InvolvedFileIDs)
	}

	// Add more
	client.WithFileIDs(ctx, "file3")
	if len(ctx.InvolvedFileIDs) != 3 {
		t.Errorf("Expected 3 file IDs, got %d", len(ctx.InvolvedFileIDs))
	}
}

func TestClient_WithParentIDs(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	ctx := NewRequestContext("default", "", types.RequestTypeGetByID)

	client.WithParentIDs(ctx, "parent1", "parent2")

	if len(ctx.InvolvedParentIDs) != 2 {
		t.Errorf("Expected 2 parent IDs, got %d", len(ctx.InvolvedParentIDs))
	}
	if ctx.InvolvedParentIDs[0] != "parent1" || ctx.InvolvedParentIDs[1] != "parent2" {
		t.Errorf("Expected [parent1, parent2], got %v", ctx.InvolvedParentIDs)
	}
}
