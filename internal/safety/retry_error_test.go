package safety

import (
	"errors"
	"testing"
)

func TestIsRetryableError(t *testing.T) {
	// Test nil error - should return false
	result := isRetryableError(nil)
	if result {
		t.Error("isRetryableError(nil) should return false")
	}

	// Test non-nil error - should return true
	err := errors.New("test error")
	result = isRetryableError(err)
	if !result {
		t.Error("isRetryableError(error) should return true")
	}
}

func TestIsRetryableError_WithNilCheck(t *testing.T) {
	// Ensure nil check works correctly
	if isRetryableError(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryableError_WithRealError(t *testing.T) {
	// Ensure any non-nil error is considered retryable
	testCases := []error{
		errors.New("network error"),
		errors.New("temporary failure"),
		errors.New("rate limited"),
		errors.New("timeout"),
		errors.New("context canceled"),
		errors.New("connection refused"),
	}

	for _, err := range testCases {
		if !isRetryableError(err) {
			t.Errorf("isRetryableError(%v) should return true", err)
		}
	}
}
