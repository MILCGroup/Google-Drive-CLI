package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestConsoleLogger_Creation(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            INFO,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	if logger == nil {
		t.Fatal("NewConsoleLogger() returned nil")
	}
}

func TestConsoleLogger_Logging(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            DEBUG,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "DEBUG") || !strings.Contains(lines[0], "debug message") {
		t.Errorf("First line doesn't contain expected content: %s", lines[0])
	}
	if !strings.Contains(lines[1], "INFO") || !strings.Contains(lines[1], "info message") {
		t.Errorf("Second line doesn't contain expected content: %s", lines[1])
	}
}

func TestConsoleLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            WARN,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)

	logger.Debug("debug message") // Filtered
	logger.Info("info message")   // Filtered
	logger.Warn("warn message")   // Logged
	logger.Error("error message") // Logged

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
}

func TestConsoleLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            INFO,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	logger.Info("test message", F("key1", "value1"), F("key2", 123))

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Output doesn't contain key1=value1: %s", output)
	}
	if !strings.Contains(output, "key2=123") {
		t.Errorf("Output doesn't contain key2=123: %s", output)
	}
}

func TestConsoleLogger_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            INFO,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	traceID := "trace-123-456"
	tracedLogger := logger.WithTraceID(traceID)
	tracedLogger.Info("test message")

	output := buf.String()
	// Should contain first 8 chars of trace ID
	if !strings.Contains(output, "[trace-12]") {
		t.Errorf("Output doesn't contain trace ID prefix: %s", output)
	}
}

func TestConsoleLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            INFO,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	ctx := context.Background()
	traceID := "ctx-trace-789"
	ctx = ContextWithTraceID(ctx, traceID)

	tracedLogger := logger.WithContext(ctx)
	tracedLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[ctx-trac]") {
		t.Errorf("Output doesn't contain trace ID prefix: %s", output)
	}
}

func TestConsoleLogger_RedactSensitive(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Bearer token",
			message:  "Authorization: Bearer ya29.a0AfH6SMBx",
			expected: "Authorization: [REDACTED]",
		},
		{
			name:     "Access token",
			message:  "access_token=ya29.a0AfH6SMBx",
			expected: "access_token=[REDACTED]",
		},
		{
			name:     "API key",
			message:  "api_key=AIzaSyD1234567890",
			expected: "api_key=[REDACTED]",
		},
		{
			name:     "Authorization header",
			message:  "Authorization: Bearer token123",
			expected: "Authorization: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := ConsoleLoggerConfig{
				Writer:           &buf,
				Level:            INFO,
				ColorEnabled:     false,
				TimestampEnabled: false,
				RedactSensitive:  true,
			}

			logger := NewConsoleLogger(config)
			logger.Info(tt.message)

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expected, output)
			}
			if strings.Contains(output, "ya29") || strings.Contains(output, "token123") {
				t.Errorf("Output contains sensitive data: %s", output)
			}
		})
	}
}

func TestConsoleLogger_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            INFO,
		ColorEnabled:     false,
		TimestampEnabled: true,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	logger.Info("test message")

	output := buf.String()
	// Should contain timestamp in format YYYY-MM-DD HH:MM:SS
	if !strings.Contains(output, "-") || !strings.Contains(output, ":") {
		t.Errorf("Output doesn't contain timestamp: %s", output)
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	config := ConsoleLoggerConfig{
		Writer:           &buf,
		Level:            DEBUG,
		ColorEnabled:     false,
		TimestampEnabled: false,
		RedactSensitive:  false,
	}

	logger := NewConsoleLogger(config)
	logger.Debug("debug 1") // Should be logged

	logger.SetLevel(ERROR)

	logger.Debug("debug 2") // Should be filtered
	logger.Info("info 2")   // Should be filtered
	logger.Error("error 1") // Should be logged

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}
}

func TestRedactSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mustHave string
		mustNot  string
	}{
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer ya29.a0AfH6SMBx_1234567890",
			mustHave: "Authorization: [REDACTED]",
			mustNot:  "ya29",
		},
		{
			name:     "OAuth access token",
			input:    "access_token: ya29.a0AfH6SMBx",
			mustHave: "access_token=[REDACTED]",
			mustNot:  "ya29",
		},
		{
			name:     "Refresh token",
			input:    "refresh_token=1//0gK1234567890",
			mustHave: "refresh_token=[REDACTED]",
			mustNot:  "1//0gK",
		},
		{
			name:     "API key",
			input:    "apiKey: AIzaSyD1234567890",
			mustHave: "apiKey=[REDACTED]",
			mustNot:  "AIzaSy",
		},
		{
			name:     "Multiple sensitive values",
			input:    "Bearer ya29.token, api_key=AIzaSy123",
			mustHave: "[REDACTED]",
			mustNot:  "ya29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactSensitiveData(tt.input)
			if !strings.Contains(result, tt.mustHave) {
				t.Errorf("Result doesn't contain %q: %s", tt.mustHave, result)
			}
			if strings.Contains(result, tt.mustNot) {
				t.Errorf("Result still contains sensitive data %q: %s", tt.mustNot, result)
			}
		})
	}
}
