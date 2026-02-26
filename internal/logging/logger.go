package logging

import (
	"context"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DEBUG level for detailed diagnostic information
	DEBUG LogLevel = iota
	// INFO level for general informational messages
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to a LogLevel
func ParseLogLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new Field (convenience function)
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// LogEntry represents a complete log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"traceId,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger defines the interface for logging operations
type Logger interface {
	// Debug logs a debug-level message with optional fields
	Debug(msg string, fields ...Field)

	// Info logs an info-level message with optional fields
	Info(msg string, fields ...Field)

	// Warn logs a warning-level message with optional fields
	Warn(msg string, fields ...Field)

	// Error logs an error-level message with optional fields
	Error(msg string, fields ...Field)

	// WithTraceID returns a new logger with the trace ID set
	WithTraceID(traceID string) Logger

	// WithContext returns a new logger that extracts trace ID from context
	WithContext(ctx context.Context) Logger

	// SetLevel sets the minimum log level
	SetLevel(level LogLevel)

	// Close flushes and closes the logger
	Close() error
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, fields ...Field)      {}
func (l *NoOpLogger) Info(msg string, fields ...Field)       {}
func (l *NoOpLogger) Warn(msg string, fields ...Field)       {}
func (l *NoOpLogger) Error(msg string, fields ...Field)      {}
func (l *NoOpLogger) WithTraceID(traceID string) Logger      { return l }
func (l *NoOpLogger) WithContext(ctx context.Context) Logger { return l }
func (l *NoOpLogger) SetLevel(level LogLevel)                {}
func (l *NoOpLogger) Close() error                           { return nil }

// NewNoOpLogger creates a new no-op logger
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
}

// Context key for trace ID
type contextKey string

const traceIDKey contextKey = "traceID"

// ContextWithTraceID adds a trace ID to the context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from the context
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
