package api

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/dl-alexandre/gdrive/internal/logging"
	"github.com/dl-alexandre/gdrive/internal/types"
	"github.com/dl-alexandre/gdrive/internal/utils"
	"github.com/google/uuid"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// Client wraps the Drive API with retry logic and request shaping
type Client struct {
	service        *drive.Service
	resourceKeyMgr *ResourceKeyManager
	maxRetries     int
	retryDelay     time.Duration
	logger         logging.Logger
}

// NewClient creates a new Drive API client
func NewClient(service *drive.Service, maxRetries int, retryDelayMs int, logger logging.Logger) *Client {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &Client{
		service:        service,
		resourceKeyMgr: NewResourceKeyManager(),
		maxRetries:     maxRetries,
		retryDelay:     time.Duration(retryDelayMs) * time.Millisecond,
		logger:         logger,
	}
}

// NewRequestContext creates a new request context with trace ID
func NewRequestContext(profile string, driveID string, requestType types.RequestType) *types.RequestContext {
	return &types.RequestContext{
		Profile:           profile,
		DriveID:           driveID,
		InvolvedFileIDs:   []string{},
		InvolvedParentIDs: []string{},
		RequestType:       requestType,
		TraceID:           uuid.New().String(),
	}
}

// WithFileIDs adds file IDs to the request context
func (c *Client) WithFileIDs(ctx *types.RequestContext, fileIDs ...string) *types.RequestContext {
	ctx.InvolvedFileIDs = append(ctx.InvolvedFileIDs, fileIDs...)
	return ctx
}

// WithParentIDs adds parent IDs to the request context
func (c *Client) WithParentIDs(ctx *types.RequestContext, parentIDs ...string) *types.RequestContext {
	ctx.InvolvedParentIDs = append(ctx.InvolvedParentIDs, parentIDs...)
	return ctx
}

// ExecuteWithRetry executes an API call with retry logic
func ExecuteWithRetry[T any](ctx context.Context, client *Client, reqCtx *types.RequestContext, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	// Log the API operation start
	logger := client.logger.WithTraceID(reqCtx.TraceID)
	logger.Info("API operation starting",
		logging.F("requestType", reqCtx.RequestType),
		logging.F("traceId", reqCtx.TraceID),
		logging.F("profile", reqCtx.Profile),
		logging.F("driveId", reqCtx.DriveID),
	)

	start := time.Now()

	for attempt := 0; attempt <= client.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Warn("Retrying API operation",
				logging.F("attempt", attempt),
				logging.F("maxRetries", client.maxRetries),
			)
		}

		result, lastErr = fn()
		if lastErr == nil {
			duration := time.Since(start)
			logger.Info("API operation completed",
				logging.F("duration_ms", duration.Milliseconds()),
				logging.F("attempts", attempt+1),
			)
			return result, nil
		}

		if !isRetryable(lastErr) {
			duration := time.Since(start)
			logger.Error("API operation failed (non-retryable)",
				logging.F("duration_ms", duration.Milliseconds()),
				logging.F("error", lastErr.Error()),
				logging.F("attempts", attempt+1),
			)
			return result, classifyError(lastErr, reqCtx, client.logger)
		}

		if attempt < client.maxRetries {
			delay := calculateBackoff(client.retryDelay, attempt, lastErr)
			logger.Warn("API operation failed (retryable)",
				logging.F("attempt", attempt+1),
				logging.F("delay_ms", delay.Milliseconds()),
				logging.F("error", lastErr.Error()),
			)
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	duration := time.Since(start)
	logger.Error("API operation failed after max retries",
		logging.F("duration_ms", duration.Milliseconds()),
		logging.F("attempts", client.maxRetries+1),
		logging.F("error", lastErr.Error()),
	)

	return result, classifyError(lastErr, reqCtx, client.logger)
}

// isRetryable checks if an error is retryable
func isRetryable(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		switch apiErr.Code {
		case 429, 500, 502, 503, 504:
			return true
		}
	}
	return false
}

// calculateBackoff calculates the retry delay with exponential backoff
func calculateBackoff(baseDelay time.Duration, attempt int, err error) time.Duration {
	// Check for Retry-After header
	if apiErr, ok := err.(*googleapi.Error); ok {
		retryAfter := apiErr.Header.Get("Retry-After")
		if retryAfter != "" {
			// Try to parse as seconds (integer)
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				delay := time.Duration(seconds) * time.Second
				// Cap at max retry delay
				if delay > time.Duration(utils.MaxRetryDelayMs)*time.Millisecond {
					return time.Duration(utils.MaxRetryDelayMs) * time.Millisecond
				}
				return delay
			}
			// Could also parse as HTTP date, but Google typically uses seconds
		}
	}

	// Exponential backoff: base * 2^attempt
	delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))

	// Cap at maximum retry delay
	if delay > time.Duration(utils.MaxRetryDelayMs)*time.Millisecond {
		delay = time.Duration(utils.MaxRetryDelayMs) * time.Millisecond
	}

	// Add jitter (±25% of delay)
	jitterRange := delay / 4
	jitter := time.Duration(rand.Int63n(int64(jitterRange*2))) - jitterRange
	delay = delay + jitter

	// Ensure delay is not negative
	if delay < 0 {
		delay = baseDelay
	}

	return delay
}

// classifyError converts API errors to CLI errors
func classifyError(err error, reqCtx *types.RequestContext, logger logging.Logger) error {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		logger.Error("Non-API error",
			logging.F("error", err.Error()),
			logging.F("traceId", reqCtx.TraceID),
		)
		return utils.NewAppError(utils.NewCLIError(utils.ErrCodeNetworkError, err.Error()).
			WithRetryable(true).
			WithContext("traceId", reqCtx.TraceID).
			Build())
	}

	var code string
	var retryable bool

	// Classify based on HTTP status code and Drive-specific reasons
	code = utils.ErrCodeUnknown
	switch apiErr.Code {
	case 400:
		code = utils.ErrCodeInvalidArgument
		// Check for specific bad request reasons
		for _, e := range apiErr.Errors {
			switch e.Reason {
			case "invalidSharingRequest":
				code = utils.ErrCodeSharingRestricted
			case "teamDriveFileLimitExceeded":
				code = utils.ErrCodeQuotaExceeded
			}
		}
	case 401:
		code = utils.ErrCodeAuthExpired
	case 403:
		code = utils.ErrCodePermissionDenied
		// Check for specific 403 reasons
		for _, e := range apiErr.Errors {
			switch e.Reason {
			case "storageQuotaExceeded":
				code = utils.ErrCodeQuotaExceeded
			case "sharingRateLimitExceeded", "userRateLimitExceeded", "rateLimitExceeded":
				code = utils.ErrCodeRateLimited
				retryable = true
			case "dailyLimitExceeded":
				code = utils.ErrCodeRateLimited
			case "domainPolicy":
				code = utils.ErrCodePolicyViolation
			}
		}
	case 404:
		code = utils.ErrCodeFileNotFound
	case 409:
		code = utils.ErrCodeInvalidArgument
	case 429:
		code = utils.ErrCodeRateLimited
		retryable = true
	case 500, 502, 503, 504:
		code = utils.ErrCodeNetworkError
		retryable = true
	default:
		code = utils.ErrCodeUnknown
		retryable = apiErr.Code >= 500
	}

	// Log the error classification
	logger.Error("API error classified",
		logging.F("httpStatus", apiErr.Code),
		logging.F("errorCode", code),
		logging.F("retryable", retryable),
		logging.F("message", apiErr.Message),
		logging.F("traceId", reqCtx.TraceID),
	)

	// Build error with all context
	builder := utils.NewCLIError(code, apiErr.Message).
		WithHTTPStatus(apiErr.Code).
		WithRetryable(retryable).
		WithContext("traceId", reqCtx.TraceID).
		WithContext("requestType", string(reqCtx.RequestType))

	// Add Drive API error reasons if available
	if len(apiErr.Errors) > 0 {
		builder.WithDriveReason(apiErr.Errors[0].Reason)

		// Add specific suggested actions based on reason
		switch apiErr.Errors[0].Reason {
		case "storageQuotaExceeded":
			builder.WithContext("suggestedAction", "free up space in Google Drive or upgrade storage")
		case "sharingRateLimitExceeded", "userRateLimitExceeded", "rateLimitExceeded":
			builder.WithContext("suggestedAction", "wait before retrying")
		case "dailyLimitExceeded":
			builder.WithContext("suggestedAction", "quota will reset in 24 hours")
		case "appNotAuthorizedToFile":
			builder.WithContext("suggestedAction", "file may require access via web interface first")
		case "insufficientFilePermissions":
			builder.WithContext("capability", "write_access_required")
		case "domainPolicy":
			builder.WithContext("suggestedAction", "contact domain administrator")
		}
	}

	// Add specific context based on error code
	switch code {
	case utils.ErrCodeAuthExpired:
		builder.WithContext("suggestedAction", "run 'gdrive auth login' to re-authenticate")
	case utils.ErrCodeFileNotFound:
		if reqCtx.DriveID != "" {
			builder.WithContext("searchDomain", "sharedDrive").
				WithContext("driveId", reqCtx.DriveID)
		}
		builder.WithContext("suggestedAction", "verify file ID or path is correct and accessible")
	case utils.ErrCodeRateLimited:
		builder.WithContext("suggestedAction", "rate limit exceeded, retrying with backoff")
	}

	if apiErr.Code == 409 {
		builder.WithContext("conflict", true)
	}

	if apiErr.Code >= 500 && apiErr.Code <= 504 {
		builder.WithContext("serverError", true).
			WithContext("suggestedAction", "temporary server error, retrying")
	}

	return utils.NewAppError(builder.Build())
}

// Service returns the underlying Drive service
func (c *Client) Service() *drive.Service {
	return c.service
}

// ResourceKeys returns the resource key manager
func (c *Client) ResourceKeys() *ResourceKeyManager {
	return c.resourceKeyMgr
}
