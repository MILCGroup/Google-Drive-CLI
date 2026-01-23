package logging

import (
	"context"
)

// MultiLogger combines multiple loggers
type MultiLogger struct {
	loggers []Logger
	traceID string
}

// NewMultiLogger creates a new multi-logger
func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// Debug logs a debug-level message to all loggers
func (ml *MultiLogger) Debug(msg string, fields ...Field) {
	for _, logger := range ml.loggers {
		logger.Debug(msg, fields...)
	}
}

// Info logs an info-level message to all loggers
func (ml *MultiLogger) Info(msg string, fields ...Field) {
	for _, logger := range ml.loggers {
		logger.Info(msg, fields...)
	}
}

// Warn logs a warning-level message to all loggers
func (ml *MultiLogger) Warn(msg string, fields ...Field) {
	for _, logger := range ml.loggers {
		logger.Warn(msg, fields...)
	}
}

// Error logs an error-level message to all loggers
func (ml *MultiLogger) Error(msg string, fields ...Field) {
	for _, logger := range ml.loggers {
		logger.Error(msg, fields...)
	}
}

// WithTraceID returns a new logger with the trace ID set
func (ml *MultiLogger) WithTraceID(traceID string) Logger {
	newLoggers := make([]Logger, len(ml.loggers))
	for i, logger := range ml.loggers {
		newLoggers[i] = logger.WithTraceID(traceID)
	}
	return &MultiLogger{
		loggers: newLoggers,
		traceID: traceID,
	}
}

// WithContext returns a new logger that extracts trace ID from context
func (ml *MultiLogger) WithContext(ctx context.Context) Logger {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return ml
	}
	return ml.WithTraceID(traceID)
}

// SetLevel sets the minimum log level for all loggers
func (ml *MultiLogger) SetLevel(level LogLevel) {
	for _, logger := range ml.loggers {
		logger.SetLevel(level)
	}
}

// Close closes all loggers
func (ml *MultiLogger) Close() error {
	var lastErr error
	for _, logger := range ml.loggers {
		if err := logger.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
