package logging

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// MaxBodyLogSize is the maximum size of request/response body to log
	MaxBodyLogSize = 10 * 1024 // 10KB
)

// DebugTransport wraps an http.RoundTripper to log HTTP requests and responses
type DebugTransport struct {
	transport http.RoundTripper
	logger    Logger
}

// NewDebugTransport creates a new debug transport wrapper
func NewDebugTransport(transport http.RoundTripper, logger Logger) *DebugTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &DebugTransport{
		transport: transport,
		logger:    logger,
	}
}

// RoundTrip implements http.RoundTripper interface with debug logging
func (dt *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log request
	dt.logRequest(req)

	// Execute request with timing
	start := time.Now()
	resp, err := dt.transport.RoundTrip(req)
	duration := time.Since(start)

	// Log response
	if err != nil {
		dt.logger.Error("HTTP request failed",
			F("method", req.Method),
			F("url", req.URL.String()),
			F("duration_ms", duration.Milliseconds()),
			F("error", err.Error()),
		)
		return resp, err
	}

	dt.logResponse(resp, duration)
	return resp, nil
}

// logRequest logs HTTP request details
func (dt *DebugTransport) logRequest(req *http.Request) {
	fields := []Field{
		F("method", req.Method),
		F("url", req.URL.String()),
		F("proto", req.Proto),
	}

	// Log headers (redact Authorization)
	headers := make(map[string]string)
	for k, v := range req.Header {
		if strings.EqualFold(k, "Authorization") {
			headers[k] = "[REDACTED]"
		} else {
			headers[k] = strings.Join(v, ", ")
		}
	}
	if len(headers) > 0 {
		fields = append(fields, F("headers", headers))
	}

	// Log request body if present
	if req.Body != nil && req.ContentLength > 0 {
		bodyData, err := io.ReadAll(io.LimitReader(req.Body, MaxBodyLogSize))
		if err == nil {
			// Restore body for actual request
			req.Body = io.NopCloser(io.MultiReader(
				bytes.NewReader(bodyData),
				req.Body,
			))

			// Log body (truncate if too large)
			body := string(bodyData)
			if req.ContentLength > MaxBodyLogSize {
				body = body + fmt.Sprintf("\n... (truncated, %d bytes total)", req.ContentLength)
			}
			fields = append(fields, F("body", body))
		}
		fields = append(fields, F("content_length", req.ContentLength))
	}

	dt.logger.Debug("HTTP Request", fields...)
}

// logResponse logs HTTP response details
func (dt *DebugTransport) logResponse(resp *http.Response, duration time.Duration) {
	fields := []Field{
		F("status", resp.Status),
		F("status_code", resp.StatusCode),
		F("proto", resp.Proto),
		F("duration_ms", duration.Milliseconds()),
	}

	// Log response headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ", ")
	}
	if len(headers) > 0 {
		fields = append(fields, F("headers", headers))
	}

	// Log response body if present
	if resp.Body != nil && resp.ContentLength > 0 {
		bodyData, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyLogSize))
		if err == nil {
			// Restore body for actual reading
			resp.Body = io.NopCloser(io.MultiReader(
				bytes.NewReader(bodyData),
				resp.Body,
			))

			// Log body (truncate if too large)
			body := string(bodyData)
			if resp.ContentLength > MaxBodyLogSize {
				body = body + fmt.Sprintf("\n... (truncated, %d bytes total)", resp.ContentLength)
			}
			fields = append(fields, F("body", body))
		}
		fields = append(fields, F("content_length", resp.ContentLength))
	}

	// Determine log level based on status code
	if resp.StatusCode >= 400 {
		dt.logger.Error("HTTP Response", fields...)
	} else {
		dt.logger.Debug("HTTP Response", fields...)
	}
}

// DebugLogger wraps a logger with debug-specific functionality
type DebugLogger struct {
	logger Logger
}

// NewDebugLogger creates a new debug logger
func NewDebugLogger(logger Logger) *DebugLogger {
	return &DebugLogger{logger: logger}
}

// LogAPICall logs a high-level API call
func (dl *DebugLogger) LogAPICall(operation string, fields ...Field) {
	dl.logger.Debug(fmt.Sprintf("API Call: %s", operation), fields...)
}

// LogAPIResponse logs a high-level API response
func (dl *DebugLogger) LogAPIResponse(operation string, duration time.Duration, fields ...Field) {
	allFields := append([]Field{F("duration_ms", duration.Milliseconds())}, fields...)
	dl.logger.Debug(fmt.Sprintf("API Response: %s", operation), allFields...)
}

// LogRetry logs a retry attempt
func (dl *DebugLogger) LogRetry(operation string, attempt int, delay time.Duration, reason string) {
	dl.logger.Warn(fmt.Sprintf("Retry attempt %d for %s", attempt, operation),
		F("attempt", attempt),
		F("delay_ms", delay.Milliseconds()),
		F("reason", reason),
	)
}

// LogRateLimit logs a rate limit event
func (dl *DebugLogger) LogRateLimit(operation string, retryAfter time.Duration) {
	dl.logger.Warn(fmt.Sprintf("Rate limited: %s", operation),
		F("retry_after_ms", retryAfter.Milliseconds()),
	)
}
