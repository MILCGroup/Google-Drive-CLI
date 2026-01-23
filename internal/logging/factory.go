package logging

import (
	"fmt"
	"os"
)

// LogConfig contains configuration for creating a logger
type LogConfig struct {
	// Level is the minimum log level
	Level LogLevel

	// OutputFile is the path to the log file (empty means no file logging)
	OutputFile string

	// EnableConsole enables console logging
	EnableConsole bool

	// EnableDebug enables debug mode with HTTP request/response logging
	EnableDebug bool

	// RedactSensitive enables redaction of sensitive data
	RedactSensitive bool

	// MaxFileSize is the maximum log file size before rotation (in bytes)
	MaxFileSize int64

	// EnableColor enables colored console output
	EnableColor bool

	// EnableTimestamp enables timestamp in console output
	EnableTimestamp bool
}

// DefaultLogConfig returns a default log configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:           INFO,
		EnableConsole:   true,
		RedactSensitive: true,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
		EnableColor:     true,
		EnableTimestamp: true,
	}
}

// NewLogger creates a new logger based on the configuration
func NewLogger(config LogConfig) (Logger, error) {
	var loggers []Logger

	// Create file logger if output file is specified
	if config.OutputFile != "" {
		fileLogger, err := NewFileLogger(FileLoggerConfig{
			FilePath:      config.OutputFile,
			Level:         config.Level,
			MaxFileSize:   config.MaxFileSize,
			RotateEnabled: config.MaxFileSize > 0,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create file logger: %w", err)
		}
		loggers = append(loggers, fileLogger)
	}

	// Create console logger if enabled
	if config.EnableConsole {
		consoleLogger := NewConsoleLogger(ConsoleLoggerConfig{
			Writer:           os.Stderr,
			Level:            config.Level,
			ColorEnabled:     config.EnableColor,
			TimestampEnabled: config.EnableTimestamp,
			RedactSensitive:  config.RedactSensitive,
		})
		loggers = append(loggers, consoleLogger)
	}

	// If no loggers are configured, return a no-op logger
	if len(loggers) == 0 {
		return NewNoOpLogger(), nil
	}

	// If only one logger is configured, return it directly
	if len(loggers) == 1 {
		return loggers[0], nil
	}

	// Otherwise, return a multi-logger
	return NewMultiLogger(loggers...), nil
}

// NewDebugLoggerWithTransport creates a logger with debug transport wrapper
func NewDebugLoggerWithTransport(config LogConfig) (Logger, *DebugTransport, error) {
	logger, err := NewLogger(config)
	if err != nil {
		return nil, nil, err
	}

	// Create debug transport if debug mode is enabled
	var debugTransport *DebugTransport
	if config.EnableDebug {
		debugTransport = NewDebugTransport(nil, logger)
	}

	return logger, debugTransport, nil
}
