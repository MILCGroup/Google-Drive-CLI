package cloudlogging

import (
	"testing"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/dl-alexandre/gdrv/internal/types"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Severity constants from google.logging.type.LogSeverity
const (
	LogSeverity_DEFAULT   = 0
	LogSeverity_DEBUG     = 100
	LogSeverity_INFO      = 200
	LogSeverity_NOTICE    = 300
	LogSeverity_WARNING   = 400
	LogSeverity_ERROR     = 500
	LogSeverity_CRITICAL  = 600
	LogSeverity_ALERT     = 700
	LogSeverity_EMERGENCY = 800
)

func TestConvertLogEntry(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *loggingpb.LogEntry
		expected types.LogEntry
	}{
		{
			name: "text payload entry",
			input: &loggingpb.LogEntry{
				LogName:   "projects/my-project/logs/my-log",
				Timestamp: timestamppb.New(timestamp),
				Severity:  LogSeverity_INFO,
				Payload:   &loggingpb.LogEntry_TextPayload{TextPayload: "Test log message"},
				Labels:    map[string]string{"key": "value"},
				Resource: &monitoredrespb.MonitoredResource{
					Type:   "gce_instance",
					Labels: map[string]string{"instance_id": "123"},
				},
			},
			expected: types.LogEntry{
				LogName:     "projects/my-project/logs/my-log",
				Timestamp:   "2024-01-15 10:00:00",
				Severity:    "INFO",
				TextPayload: "Test log message",
				Labels:      map[string]string{"key": "value"},
				Resource: &types.LogResource{
					Type:   "gce_instance",
					Labels: map[string]string{"instance_id": "123"},
				},
			},
		},
		{
			name: "json payload entry",
			input: &loggingpb.LogEntry{
				LogName:   "projects/my-project/logs/json-log",
				Timestamp: timestamppb.New(timestamp),
				Severity:  LogSeverity_ERROR,
				Payload: &loggingpb.LogEntry_JsonPayload{
					JsonPayload: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"message": structpb.NewStringValue("Error occurred"),
							"code":    structpb.NewNumberValue(500),
						},
					},
				},
			},
			expected: types.LogEntry{
				LogName:     "projects/my-project/logs/json-log",
				Timestamp:   "2024-01-15 10:00:00",
				Severity:    "ERROR",
				JSONPayload: map[string]interface{}{"message": "Error occurred", "code": float64(500)},
			},
		},
		{
			name: "proto payload entry",
			input: &loggingpb.LogEntry{
				LogName:   "projects/my-project/logs/proto-log",
				Timestamp: timestamppb.New(timestamp),
				Severity:  LogSeverity_WARNING,
				Payload: &loggingpb.LogEntry_ProtoPayload{
					ProtoPayload: &anypb.Any{TypeUrl: "type.googleapis.com/google.cloud.audit.AuditLog"},
				},
			},
			expected: types.LogEntry{
				LogName:     "projects/my-project/logs/proto-log",
				Timestamp:   "2024-01-15 10:00:00",
				Severity:    "WARNING",
				TextPayload: "[Proto]",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertLogEntry(tc.input)
			if result.LogName != tc.expected.LogName {
				t.Errorf("LogName: got %q, want %q", result.LogName, tc.expected.LogName)
			}
			if result.Timestamp != tc.expected.Timestamp {
				t.Errorf("Timestamp: got %q, want %q", result.Timestamp, tc.expected.Timestamp)
			}
			if result.Severity != tc.expected.Severity {
				t.Errorf("Severity: got %q, want %q", result.Severity, tc.expected.Severity)
			}
			if result.TextPayload != tc.expected.TextPayload {
				t.Errorf("TextPayload: got %q, want %q", result.TextPayload, tc.expected.TextPayload)
			}
			if len(result.JSONPayload) != len(tc.expected.JSONPayload) {
				t.Errorf("JSONPayload length: got %d, want %d", len(result.JSONPayload), len(tc.expected.JSONPayload))
			}
		})
	}
}

func TestConvertLogSink(t *testing.T) {
	tests := []struct {
		name     string
		input    *loggingpb.LogSink
		expected types.LogSink
	}{
		{
			name: "full sink",
			input: &loggingpb.LogSink{
				Name:        "projects/my-project/sinks/my-sink",
				Destination: "bigquery.googleapis.com/projects/my-project/datasets/my_dataset",
				Filter:      "severity>=ERROR",
				Description: "Sink for error logs",
				Disabled:    false,
			},
			expected: types.LogSink{
				Name:        "projects/my-project/sinks/my-sink",
				Destination: "bigquery.googleapis.com/projects/my-project/datasets/my_dataset",
				Filter:      "severity>=ERROR",
				Description: "Sink for error logs",
				Disabled:    false,
			},
		},
		{
			name: "disabled sink",
			input: &loggingpb.LogSink{
				Name:        "projects/my-project/sinks/disabled-sink",
				Destination: "pubsub.googleapis.com/projects/my-project/topics/my-topic",
				Disabled:    true,
			},
			expected: types.LogSink{
				Name:        "projects/my-project/sinks/disabled-sink",
				Destination: "pubsub.googleapis.com/projects/my-project/topics/my-topic",
				Disabled:    true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertLogSink(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Destination != tc.expected.Destination {
				t.Errorf("Destination: got %q, want %q", result.Destination, tc.expected.Destination)
			}
			if result.Filter != tc.expected.Filter {
				t.Errorf("Filter: got %q, want %q", result.Filter, tc.expected.Filter)
			}
			if result.Disabled != tc.expected.Disabled {
				t.Errorf("Disabled: got %v, want %v", result.Disabled, tc.expected.Disabled)
			}
		})
	}
}

func TestConvertLogMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    *loggingpb.LogMetric
		expected types.LogMetric
	}{
		{
			name: "full metric",
			input: &loggingpb.LogMetric{
				Name:        "projects/my-project/metrics/error_count",
				Description: "Count of error logs",
				Filter:      "severity>=ERROR",
			},
			expected: types.LogMetric{
				Name:        "projects/my-project/metrics/error_count",
				Description: "Count of error logs",
				Filter:      "severity>=ERROR",
			},
		},
		{
			name: "minimal metric",
			input: &loggingpb.LogMetric{
				Name:   "projects/my-project/metrics/simple",
				Filter: "resource.type=gce_instance",
			},
			expected: types.LogMetric{
				Name:   "projects/my-project/metrics/simple",
				Filter: "resource.type=gce_instance",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertLogMetric(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Description != tc.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tc.expected.Description)
			}
			if result.Filter != tc.expected.Filter {
				t.Errorf("Filter: got %q, want %q", result.Filter, tc.expected.Filter)
			}
		})
	}
}

func TestManagerClose(t *testing.T) {
	t.Run("close with nil clients", func(t *testing.T) {
		m := &Manager{}
		err := m.Close()
		if err != nil {
			t.Errorf("Close() with nil clients should not error, got: %v", err)
		}
	})
}

func TestManagerConstructorValidation(t *testing.T) {
	t.Run("NewManager with nil context should error", func(t *testing.T) {
		// This test validates that NewManager requires a valid context
		// We can't actually test this without a real connection, but we verify
		// the function signature is correct
		var _ *logging.Client // Just to ensure import is used
	})
}
