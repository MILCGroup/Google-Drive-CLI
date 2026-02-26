package errors

import (
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/googleapi"
)

func TestClassifyGoogleAPIError_400Errors(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     *googleapi.Error
		wantCode   string
		wantReason string
	}{
		{
			name: "invalid argument",
			apiErr: &googleapi.Error{
				Code:    400,
				Message: "Invalid request",
			},
			wantCode: utils.ErrCodeInvalidArgument,
		},
		{
			name: "invalid sharing request",
			apiErr: &googleapi.Error{
				Code:    400,
				Message: "Invalid sharing request",
				Errors:  []googleapi.ErrorItem{{Reason: "invalidSharingRequest"}},
			},
			wantCode:   utils.ErrCodeSharingRestricted,
			wantReason: "invalidSharingRequest",
		},
		{
			name: "team drive file limit exceeded",
			apiErr: &googleapi.Error{
				Code:    400,
				Message: "Team drive limit",
				Errors:  []googleapi.ErrorItem{{Reason: "teamDriveFileLimitExceeded"}},
			},
			wantCode:   utils.ErrCodeQuotaExceeded,
			wantReason: "teamDriveFileLimitExceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &types.RequestContext{
				TraceID:     "test-trace",
				RequestType: types.RequestTypeMutation,
			}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", tt.apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if appErr.CLIError.Code != tt.wantCode {
				t.Errorf("Code = %s, want %s", appErr.CLIError.Code, tt.wantCode)
			}

			if tt.wantReason != "" && appErr.CLIError.DriveReason != tt.wantReason {
				t.Errorf("DriveReason = %s, want %s", appErr.CLIError.DriveReason, tt.wantReason)
			}

			if appErr.CLIError.HTTPStatus != tt.apiErr.Code {
				t.Errorf("HTTPStatus = %d, want %d", appErr.CLIError.HTTPStatus, tt.apiErr.Code)
			}
		})
	}
}

func TestClassifyGoogleAPIError_401(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    401,
		Message: "Unauthorized",
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Code != utils.ErrCodeAuthExpired {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeAuthExpired)
	}

	if appErr.CLIError.Context["suggestedAction"] == nil {
		t.Error("Expected suggestedAction in context")
	}
}

func TestClassifyGoogleAPIError_403Errors(t *testing.T) {
	tests := []struct {
		name           string
		reason         string
		wantCode       string
		wantRetryable  bool
		wantSuggestion bool
	}{
		{
			name:     "permission denied",
			reason:   "",
			wantCode: utils.ErrCodePermissionDenied,
		},
		{
			name:           "storage quota exceeded",
			reason:         "storageQuotaExceeded",
			wantCode:       utils.ErrCodeQuotaExceeded,
			wantSuggestion: true,
		},
		{
			name:          "sharing rate limit exceeded",
			reason:        "sharingRateLimitExceeded",
			wantCode:      utils.ErrCodeRateLimited,
			wantRetryable: true,
		},
		{
			name:          "user rate limit exceeded",
			reason:        "userRateLimitExceeded",
			wantCode:      utils.ErrCodeRateLimited,
			wantRetryable: true,
		},
		{
			name:     "daily limit exceeded",
			reason:   "dailyLimitExceeded",
			wantCode: utils.ErrCodeRateLimited,
		},
		{
			name:           "domain policy",
			reason:         "domainPolicy",
			wantCode:       utils.ErrCodePolicyViolation,
			wantSuggestion: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &googleapi.Error{
				Code:    403,
				Message: "Forbidden",
			}
			if tt.reason != "" {
				apiErr.Errors = []googleapi.ErrorItem{{Reason: tt.reason}}
			}

			reqCtx := &types.RequestContext{TraceID: "test-trace"}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if appErr.CLIError.Code != tt.wantCode {
				t.Errorf("Code = %s, want %s", appErr.CLIError.Code, tt.wantCode)
			}

			if appErr.CLIError.Retryable != tt.wantRetryable {
				t.Errorf("Retryable = %v, want %v", appErr.CLIError.Retryable, tt.wantRetryable)
			}

			if tt.wantSuggestion && appErr.CLIError.Context["suggestedAction"] == nil {
				t.Error("Expected suggestedAction in context")
			}
		})
	}
}

func TestClassifyGoogleAPIError_404(t *testing.T) {
	tests := []struct {
		name    string
		driveID string
	}{
		{"file not found", ""},
		{"file not found in shared drive", "shared-drive-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &googleapi.Error{
				Code:    404,
				Message: "File not found",
			}

			reqCtx := &types.RequestContext{
				TraceID: "test-trace",
				DriveID: tt.driveID,
			}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if appErr.CLIError.Code != utils.ErrCodeFileNotFound {
				t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeFileNotFound)
			}

			if tt.driveID != "" {
				if appErr.CLIError.Context["driveId"] != tt.driveID {
					t.Errorf("Expected driveId in context")
				}
			}

			if appErr.CLIError.Context["suggestedAction"] == nil {
				t.Error("Expected suggestedAction in context")
			}
		})
	}
}

func TestClassifyGoogleAPIError_409(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    409,
		Message: "Conflict",
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Code != utils.ErrCodeInvalidArgument {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeInvalidArgument)
	}

	if appErr.CLIError.Context["conflict"] != true {
		t.Error("Expected conflict flag in context")
	}
}

func TestClassifyGoogleAPIError_429(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    429,
		Message: "Rate limit exceeded",
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Code != utils.ErrCodeRateLimited {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeRateLimited)
	}

	if !appErr.CLIError.Retryable {
		t.Error("429 errors should be retryable")
	}
}

func TestClassifyGoogleAPIError_5xxErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"500 Internal Server Error", 500},
		{"502 Bad Gateway", 502},
		{"503 Service Unavailable", 503},
		{"504 Gateway Timeout", 504},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &googleapi.Error{
				Code:    tt.statusCode,
				Message: "Server error",
			}

			reqCtx := &types.RequestContext{TraceID: "test-trace"}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if appErr.CLIError.Code != utils.ErrCodeNetworkError {
				t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeNetworkError)
			}

			if !appErr.CLIError.Retryable {
				t.Error("5xx errors should be retryable")
			}

			if appErr.CLIError.Context["serverError"] != true {
				t.Error("Expected serverError flag in context")
			}
		})
	}
}

func TestClassifyGoogleAPIError_NonAPIError(t *testing.T) {
	plainErr := stderrors.New("network connection failed")

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", plainErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Code != utils.ErrCodeNetworkError {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeNetworkError)
	}

	if !appErr.CLIError.Retryable {
		t.Error("Network errors should be retryable")
	}
}

func TestClassifyGoogleAPIError_ContextPropagation(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    404,
		Message: "Not found",
	}

	reqCtx := &types.RequestContext{
		TraceID:     "test-trace-123",
		RequestType: types.RequestTypeMutation,
		DriveID:     "drive-456",
	}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Context["traceId"] != reqCtx.TraceID {
		t.Error("TraceID not propagated to error context")
	}

	if appErr.CLIError.Context["requestType"] != string(reqCtx.RequestType) {
		t.Error("RequestType not propagated to error context")
	}

	if appErr.CLIError.Context["service"] != "drive" {
		t.Error("Service not propagated to error context")
	}
}

func TestClassifyGoogleAPIError_SuggestedActions(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     *googleapi.Error
		wantHint   bool
		contextKey string
		contextVal string
	}{
		{
			name: "storage quota exceeded",
			apiErr: &googleapi.Error{
				Code:    403,
				Message: "Quota exceeded",
				Errors:  []googleapi.ErrorItem{{Reason: "storageQuotaExceeded"}},
			},
			wantHint:   true,
			contextKey: "suggestedAction",
		},
		{
			name: "app not authorized to file",
			apiErr: &googleapi.Error{
				Code:    403,
				Message: "Not authorized",
				Errors:  []googleapi.ErrorItem{{Reason: "appNotAuthorizedToFile"}},
			},
			wantHint:   true,
			contextKey: "suggestedAction",
		},
		{
			name: "insufficient file permissions",
			apiErr: &googleapi.Error{
				Code:    403,
				Message: "Permission denied",
				Errors:  []googleapi.ErrorItem{{Reason: "insufficientFilePermissions"}},
			},
			wantHint:   true,
			contextKey: "capability",
			contextVal: "write_access_required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &types.RequestContext{TraceID: "test-trace"}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", tt.apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if tt.wantHint {
				if appErr.CLIError.Context[tt.contextKey] == nil {
					t.Errorf("Expected %s in context", tt.contextKey)
				}
				if tt.contextVal != "" && appErr.CLIError.Context[tt.contextKey] != tt.contextVal {
					t.Errorf("Expected %s = %s, got %v", tt.contextKey, tt.contextVal, appErr.CLIError.Context[tt.contextKey])
				}
			}
		})
	}
}

func TestClassifyGoogleAPIError_UnknownStatusCode(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    418, // I'm a teapot
		Message: "Unknown status",
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	if appErr.CLIError.Code != utils.ErrCodeUnknown {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodeUnknown)
	}

	if appErr.CLIError.Retryable {
		t.Error("Unknown 4xx errors should not be retryable")
	}
}

func TestClassifyGoogleAPIError_MultipleErrorReasons(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    403,
		Message: "Multiple errors",
		Errors: []googleapi.ErrorItem{
			{Reason: "domainPolicy"},
			{Reason: "insufficientPermissions"}, // Only first should be used
		},
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	// Should use the first error reason
	if appErr.CLIError.Code != utils.ErrCodePolicyViolation {
		t.Errorf("Code = %s, want %s", appErr.CLIError.Code, utils.ErrCodePolicyViolation)
	}

	if appErr.CLIError.DriveReason != "domainPolicy" {
		t.Errorf("DriveReason = %s, want domainPolicy", appErr.CLIError.DriveReason)
	}
}

func TestClassifyGoogleAPIError_NonDriveService(t *testing.T) {
	apiErr := &googleapi.Error{
		Code:    404,
		Message: "Not found",
		Errors:  []googleapi.ErrorItem{{Reason: "notFound"}},
	}

	reqCtx := &types.RequestContext{TraceID: "test-trace"}
	logger := logging.NewNoOpLogger()

	// Test with "sheets" service
	err := ClassifyGoogleAPIError("sheets", apiErr, reqCtx, logger)

	var appErr *utils.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected *utils.AppError, got %T", err)
	}

	// DriveReason should only be set for "drive" service
	if appErr.CLIError.DriveReason != "" {
		t.Error("DriveReason should not be set for non-drive service")
	}

	if appErr.CLIError.Context["service"] != "sheets" {
		t.Error("Service should be set to sheets")
	}
}

func TestClassifyGoogleAPIError_HTTPStatusPropagation(t *testing.T) {
	tests := []int{400, 401, 403, 404, 409, 429, 500, 502, 503, 504}

	for _, status := range tests {
		t.Run(http.StatusText(status), func(t *testing.T) {
			apiErr := &googleapi.Error{
				Code:    status,
				Message: "Test error",
			}

			reqCtx := &types.RequestContext{TraceID: "test-trace"}
			logger := logging.NewNoOpLogger()

			err := ClassifyGoogleAPIError("drive", apiErr, reqCtx, logger)

			var appErr *utils.AppError
			if !stderrors.As(err, &appErr) {
				t.Fatalf("expected *utils.AppError, got %T", err)
			}

			if appErr.CLIError.HTTPStatus != status {
				t.Errorf("HTTPStatus = %d, want %d", appErr.CLIError.HTTPStatus, status)
			}
		})
	}
}
