package logging

import (
	"context"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"ERROR", ERROR},
		{"INVALID", INFO}, // Default to INFO
		{"", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseLogLevel(tt.input); got != tt.expected {
				t.Errorf("ParseLogLevel(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestField_Creation(t *testing.T) {
	field := F("key", "value")
	if field.Key != "key" {
		t.Errorf("Field.Key = %v, want %v", field.Key, "key")
	}
	if field.Value != "value" {
		t.Errorf("Field.Value = %v, want %v", field.Value, "value")
	}
}

func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	// These should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.SetLevel(DEBUG)

	newLogger := logger.WithTraceID("trace-123")
	if newLogger == nil {
		t.Error("WithTraceID() returned nil")
	}

	ctx := context.Background()
	newLogger = logger.WithContext(ctx)
	if newLogger == nil {
		t.Error("WithContext() returned nil")
	}

	if err := logger.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestContextTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "test-trace-123"

	// Add trace ID to context
	ctx = ContextWithTraceID(ctx, traceID)

	// Extract trace ID from context
	extractedID := TraceIDFromContext(ctx)
	if extractedID != traceID {
		t.Errorf("TraceIDFromContext() = %v, want %v", extractedID, traceID)
	}

	// Test with empty context
	emptyCtx := context.Background()
	extractedID = TraceIDFromContext(emptyCtx)
	if extractedID != "" {
		t.Errorf("TraceIDFromContext() = %v, want empty string", extractedID)
	}
}
