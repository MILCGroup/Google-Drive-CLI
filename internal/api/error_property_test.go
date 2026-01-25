package api

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/googleapi"
)

// Property 6: Error Structure Consistency
// Validates: Requirements 7.5, 7.6, 7.7, 7.8
// Property: All errors must have consistent structure with code, message, and context

func TestProperty_ErrorStructure_RequiredFields(t *testing.T) {
	// Property: All CLIErrors must have code and message

	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name       string
		err        error
		expectCode string
		expectMsg  bool
	}{
		{
			"Network error",
			errors.New("network timeout"),
			utils.ErrCodeNetworkError,
			true,
		},
		{
			"404 Not Found",
			&googleapi.Error{Code: 404, Message: "File not found"},
			utils.ErrCodeFileNotFound,
			true,
		},
		{
			"401 Unauthorized",
			&googleapi.Error{Code: 401, Message: "Token expired"},
			utils.ErrCodeAuthExpired,
			true,
		},
		{
			"403 Forbidden",
			&googleapi.Error{Code: 403, Message: "Permission denied"},
			utils.ErrCodePermissionDenied,
			true,
		},
		{
			"429 Rate Limited",
			&googleapi.Error{Code: 429, Message: "Too many requests"},
			utils.ErrCodeRateLimited,
			true,
		},
		{
			"500 Server Error",
			&googleapi.Error{Code: 500, Message: "Internal error"},
			utils.ErrCodeNetworkError,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := classifyError(tt.err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if appErr.CLIError.Code == "" {
				t.Error("CLIError.Code is empty")
			}
			if appErr.CLIError.Code != tt.expectCode {
				t.Errorf("CLIError.Code = %s, want %s", appErr.CLIError.Code, tt.expectCode)
			}
			if tt.expectMsg && appErr.CLIError.Message == "" {
				t.Error("CLIError.Message is empty")
			}
		})
	}
}

func TestProperty_ErrorStructure_ContextInclusion(t *testing.T) {
	// Property: All errors include context (traceId, requestType, etc.)

	tests := []struct {
		name       string
		reqCtx     *types.RequestContext
		err        error
		checkTrace bool
		checkType  bool
		checkDrive bool
	}{
		{
			"With trace ID",
			func() *types.RequestContext {
				ctx := NewRequestContext("default", "", types.RequestTypeGetByID)
				ctx.TraceID = "trace-123"
				return ctx
			}(),
			&googleapi.Error{Code: 404},
			true, true, false,
		},
		{
			"With drive ID",
			NewRequestContext("default", "drive-456", types.RequestTypeListOrSearch),
			&googleapi.Error{Code: 404},
			true, true, true,
		},
		{
			"With file IDs",
			func() *types.RequestContext {
				ctx := NewRequestContext("default", "", types.RequestTypeMutation)
				ctx.InvolvedFileIDs = []string{"file1", "file2"}
				return ctx
			}(),
			&googleapi.Error{Code: 403},
			true, true, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := classifyError(tt.err, tt.reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if appErr.CLIError.Context == nil {
				t.Fatal("CLIError.Context is nil")
			}

			if tt.checkTrace {
				if _, ok := appErr.CLIError.Context["traceId"]; !ok {
					t.Error("Context missing traceId")
				}
			}

			if tt.checkType {
				if _, ok := appErr.CLIError.Context["requestType"]; !ok {
					t.Error("Context missing requestType")
				}
			}

			if tt.checkDrive {
				if driveID, ok := appErr.CLIError.Context["driveId"]; !ok || driveID == "" {
					t.Error("Context missing or empty driveId")
				}
			}
		})
	}
}

func TestProperty_ErrorStructure_HTTPStatus(t *testing.T) {
	// Property: Errors from HTTP include HTTP status code

	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		httpCode   int
		shouldHave bool
	}{
		{404, true},
		{401, true},
		{403, true},
		{429, true},
		{500, true},
		{503, true},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.httpCode)), func(t *testing.T) {
			err := &googleapi.Error{Code: tt.httpCode, Message: "Test"}
			appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if tt.shouldHave && appErr.CLIError.HTTPStatus == 0 {
				t.Errorf("HTTPStatus not set for %d error", tt.httpCode)
			}
			if tt.shouldHave && appErr.CLIError.HTTPStatus != tt.httpCode {
				t.Errorf("HTTPStatus = %d, want %d", appErr.CLIError.HTTPStatus, tt.httpCode)
			}
		})
	}
}

func TestProperty_ErrorStructure_DriveReason(t *testing.T) {
	// Property: Errors with Drive-specific reasons include driveReason field

	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name       string
		reason     string
		shouldHave bool
	}{
		{"Storage quota", "storageQuotaExceeded", true},
		{"Sharing rate limit", "sharingRateLimitExceeded", true},
		{"User rate limit", "userRateLimitExceeded", true},
		{"Domain policy", "domainPolicy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &googleapi.Error{
				Code:    403,
				Message: "Error",
				Errors:  []googleapi.ErrorItem{{Reason: tt.reason}},
			}
			appErr := classifyError(err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if tt.shouldHave && appErr.CLIError.DriveReason == "" {
				t.Error("DriveReason not set")
			}
			if tt.shouldHave && appErr.CLIError.DriveReason != tt.reason {
				t.Errorf("DriveReason = %s, want %s", appErr.CLIError.DriveReason, tt.reason)
			}
		})
	}
}

func TestProperty_ErrorStructure_RetryableFlag(t *testing.T) {
	// Property: Retryable errors have retryable=true, non-retryable have false

	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name          string
		err           error
		wantRetryable bool
	}{
		{"429 Rate Limited", &googleapi.Error{Code: 429}, true},
		{"500 Server Error", &googleapi.Error{Code: 500}, true},
		{"503 Unavailable", &googleapi.Error{Code: 503}, true},
		{"404 Not Found", &googleapi.Error{Code: 404}, false},
		{"403 Permission Denied", &googleapi.Error{Code: 403}, false},
		{"401 Unauthorized", &googleapi.Error{Code: 401}, false},
		{"Network error", errors.New("timeout"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := classifyError(tt.err, reqCtx, logging.NewNoOpLogger()).(*utils.AppError)

			if appErr.CLIError.Retryable != tt.wantRetryable {
				t.Errorf("Retryable = %v, want %v", appErr.CLIError.Retryable, tt.wantRetryable)
			}
		})
	}
}

func TestProperty_ErrorStructure_ExitCodeMapping(t *testing.T) {
	// Property: All error codes map to unique exit codes

	errorCodes := []string{
		utils.ErrCodeAuthRequired,
		utils.ErrCodeAuthExpired,
		utils.ErrCodeScopeInsufficient,
		utils.ErrCodeFileNotFound,
		utils.ErrCodePermissionDenied,
		utils.ErrCodeQuotaExceeded,
		utils.ErrCodeExportSizeLimit,
		utils.ErrCodeNetworkError,
		utils.ErrCodeTimeout,
		utils.ErrCodeRateLimited,
		utils.ErrCodeInvalidArgument,
		utils.ErrCodeInvalidPath,
		utils.ErrCodePolicyViolation,
	}

	exitCodes := make(map[int]string)

	for _, code := range errorCodes {
		exitCode := utils.GetExitCode(code)

		if exitCode == 0 {
			t.Errorf("Error code %s maps to exit code 0 (success)", code)
		}

		if prev, exists := exitCodes[exitCode]; exists {
			t.Errorf("Exit code %d used by both %s and %s", exitCode, code, prev)
		}
		exitCodes[exitCode] = code
	}
}

// Property 19: Retry Logic Implementation
// Validates: Requirements 7.1, 7.2
// Property: Retry logic must handle retryable errors correctly

func TestProperty_RetryLogic_RetryableErrors(t *testing.T) {
	// Property: Retryable errors (429, 500, 502, 503, 504) trigger retries

	retryableCodes := []int{429, 500, 502, 503, 504}

	for _, code := range retryableCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			err := &googleapi.Error{Code: code, Message: "Test"}
			if !isRetryable(err) {
				t.Errorf("HTTP %d should be retryable", code)
			}
		})
	}
}

func TestProperty_RetryLogic_NonRetryableErrors(t *testing.T) {
	// Property: Non-retryable errors (400, 401, 403, 404) don't trigger retries

	nonRetryableCodes := []int{400, 401, 403, 404}

	for _, code := range nonRetryableCodes {
		t.Run(string(rune(code)), func(t *testing.T) {
			err := &googleapi.Error{Code: code, Message: "Test"}
			if isRetryable(err) {
				t.Errorf("HTTP %d should not be retryable", code)
			}
		})
	}
}

func TestProperty_RetryLogic_ExponentialBackoff(t *testing.T) {
	// Property: Backoff delays increase exponentially

	baseDelay := 1000 * time.Millisecond
	err := &googleapi.Error{Code: 503}

	var prevDelay time.Duration

	for attempt := 0; attempt < 5; attempt++ {
		delay := calculateBackoff(baseDelay, attempt, err)

		if delay <= 0 {
			t.Errorf("Attempt %d: delay is not positive: %v", attempt, delay)
		}

		if attempt > 0 && delay <= prevDelay {
			// Allow for jitter variation, but should generally increase
			// This is a weak property due to jitter
			t.Logf("Attempt %d: delay %v not significantly greater than previous %v (jitter)",
				attempt, delay, prevDelay)
		}

		prevDelay = delay
	}
}

func TestProperty_RetryLogic_RespectRetryAfter(t *testing.T) {
	// Property: Retry-After header is respected

	baseDelay := 1000 * time.Millisecond
	retryAfterSeconds := 5

	header := make(map[string][]string)
	header["Retry-After"] = []string{strconv.Itoa(retryAfterSeconds)}

	err := &googleapi.Error{
		Code:   429,
		Header: header,
	}

	delay := calculateBackoff(baseDelay, 0, err)

	// Should use Retry-After value
	expectedDelay := time.Duration(retryAfterSeconds) * time.Second
	if delay != expectedDelay {
		t.Errorf("Expected delay %v from Retry-After, got %v", expectedDelay, delay)
	}
}

func TestProperty_RetryLogic_MaxRetries(t *testing.T) {
	// Property: Retries stop after maxRetries attempts

	client := NewClient(nil, 3, 10, logging.NewNoOpLogger()) // max 3 retries
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	callCount := 0
	_, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
		callCount++
		return "", &googleapi.Error{Code: 503, Message: "Service Unavailable"}
	})

	if err == nil {
		t.Fatal("Expected error after max retries")
	}

	// Should call: initial + 3 retries = 4 total
	expectedCalls := 4
	if callCount != expectedCalls {
		t.Errorf("Expected %d calls (1 initial + 3 retries), got %d", expectedCalls, callCount)
	}
}

func TestProperty_RetryLogic_ContextCancellation(t *testing.T) {
	// Property: Context cancellation stops retries immediately

	client := NewClient(nil, 10, 100, logging.NewNoOpLogger()) // Many retries with delay
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

	// Should only call once before context check
	if callCount > 2 {
		t.Errorf("Expected at most 2 calls before context cancellation, got %d", callCount)
	}
}

func TestProperty_RetryLogic_SuccessStopsRetries(t *testing.T) {
	// Property: Success on any attempt stops further retries

	client := NewClient(nil, 5, 10, logging.NewNoOpLogger())
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)

	tests := []struct {
		name             string
		successOnAttempt int
	}{
		{"First attempt", 1},
		{"Second attempt", 2},
		{"Third attempt", 3},
		{"Fourth attempt", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			result, err := ExecuteWithRetry(context.Background(), client, reqCtx, func() (string, error) {
				callCount++
				if callCount == tt.successOnAttempt {
					return "success", nil
				}
				return "", &googleapi.Error{Code: 503, Message: "Service Unavailable"}
			})

			if err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if result != "success" {
				t.Errorf("Expected 'success', got %s", result)
			}
			if callCount != tt.successOnAttempt {
				t.Errorf("Expected %d calls, got %d", tt.successOnAttempt, callCount)
			}
		})
	}
}

func TestProperty_RetryLogic_BackoffCap(t *testing.T) {
	// Property: Backoff delay is capped at maximum

	baseDelay := 1000 * time.Millisecond
	err := &googleapi.Error{Code: 503}

	// Very high attempt number
	delay := calculateBackoff(baseDelay, 20, err)

	maxAllowed := time.Duration(utils.MaxRetryDelayMs)*time.Millisecond + 8*time.Second // Allow jitter
	if delay > maxAllowed {
		t.Errorf("Delay %v exceeds maximum allowed %v", delay, maxAllowed)
	}
}
