package chat

import (
	"context"
	stderrors "errors"

	"github.com/milcgroup/gdrv/internal/logging"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClassifyGRPCError converts gRPC errors to CLI errors
func ClassifyGRPCError(service string, err error, reqCtx *types.RequestContext, logger logging.Logger) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC status error - could be network error, context error, etc.
		// Check for common non-gRPC errors
		if stderrors.Is(err, context.DeadlineExceeded) {
			return utils.NewAppError(utils.NewCLIError(utils.ErrCodeTimeout, err.Error()).
				WithRetryable(true).
				WithContext("traceId", reqCtx.TraceID).
				WithContext("requestType", string(reqCtx.RequestType)).
				WithContext("service", service).
				Build())
		}
		// For other non-gRPC errors, treat as network errors
		return utils.NewAppError(utils.NewCLIError(utils.ErrCodeNetworkError, err.Error()).
			WithRetryable(true).
			WithContext("traceId", reqCtx.TraceID).
			WithContext("requestType", string(reqCtx.RequestType)).
			WithContext("service", service).
			Build())
	}

	var code string
	var retryable bool

	switch st.Code() {
	case codes.OK:
		return nil
	case codes.Canceled:
		code = utils.ErrCodeNetworkError
		retryable = true
	case codes.DeadlineExceeded:
		code = utils.ErrCodeTimeout
		retryable = true
	case codes.InvalidArgument:
		code = utils.ErrCodeInvalidArgument
	case codes.NotFound:
		code = utils.ErrCodeFileNotFound
	case codes.PermissionDenied:
		code = utils.ErrCodePermissionDenied
	case codes.ResourceExhausted:
		code = utils.ErrCodeRateLimited
		retryable = true
	case codes.Unauthenticated:
		code = utils.ErrCodeAuthExpired
	case codes.Unavailable:
		code = utils.ErrCodeNetworkError
		retryable = true
	case codes.AlreadyExists:
		code = utils.ErrCodeInvalidArgument
	default:
		code = utils.ErrCodeUnknown
		retryable = st.Code() >= codes.Internal
	}

	// Build error using existing patterns
	builder := utils.NewCLIError(code, st.Message()).
		WithRetryable(retryable).
		WithContext("traceId", reqCtx.TraceID).
		WithContext("requestType", string(reqCtx.RequestType)).
		WithContext("service", service).
		WithContext("grpcCode", st.Code().String())

	return utils.NewAppError(builder.Build())
}

// IsRetryableGRPCError determines if a gRPC error is retryable
func IsRetryableGRPCError(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Canceled,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Unavailable:
		return true
	default:
		return false
	}
}

// GRPCCodeToExitCode converts gRPC status code to CLI exit code
func GRPCCodeToExitCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return utils.ExitSuccess
	case codes.NotFound:
		return utils.ExitFileNotFound
	case codes.PermissionDenied:
		return utils.ExitPermissionDenied
	case codes.Unauthenticated:
		return utils.ExitAuthExpired
	case codes.InvalidArgument:
		return utils.ExitInvalidArgument
	case codes.ResourceExhausted:
		return utils.ExitRateLimited
	default:
		return utils.ExitUnknown
	}
}
