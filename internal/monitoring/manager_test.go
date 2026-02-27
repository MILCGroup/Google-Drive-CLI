package monitoring

import (
	"testing"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/milcgroup/gdrv/internal/types"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConvertMetricDescriptor(t *testing.T) {
	tests := []struct {
		name     string
		input    *metricpb.MetricDescriptor
		expected types.MetricDescriptor
	}{
		{
			name: "full descriptor",
			input: &metricpb.MetricDescriptor{
				Name:        "projects/my-project/metricDescriptors/custom.googleapis.com/my_metric",
				Type:        "custom.googleapis.com/my_metric",
				DisplayName: "My Custom Metric",
				Description: "A custom metric for testing",
				Unit:        "1",
			},
			expected: types.MetricDescriptor{
				Name:        "projects/my-project/metricDescriptors/custom.googleapis.com/my_metric",
				Type:        "custom.googleapis.com/my_metric",
				DisplayName: "My Custom Metric",
				Description: "A custom metric for testing",
				Unit:        "1",
			},
		},
		{
			name: "minimal descriptor",
			input: &metricpb.MetricDescriptor{
				Name: "projects/my-project/metricDescriptors/custom.googleapis.com/minimal",
				Type: "custom.googleapis.com/minimal",
			},
			expected: types.MetricDescriptor{
				Name: "projects/my-project/metricDescriptors/custom.googleapis.com/minimal",
				Type: "custom.googleapis.com/minimal",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertMetricDescriptor(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Type != tc.expected.Type {
				t.Errorf("Type: got %q, want %q", result.Type, tc.expected.Type)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
			if result.Description != tc.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tc.expected.Description)
			}
			if result.Unit != tc.expected.Unit {
				t.Errorf("Unit: got %q, want %q", result.Unit, tc.expected.Unit)
			}
		})
	}
}

func TestConvertTimeSeries(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *monitoringpb.TimeSeries
		expected types.TimeSeriesData
	}{
		{
			name: "time series with double values",
			input: &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type: "custom.googleapis.com/my_metric",
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type: "gce_instance",
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(timestamp),
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 42.5},
						},
					},
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(timestamp.Add(time.Minute)),
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 43.0},
						},
					},
				},
			},
			expected: types.TimeSeriesData{
				Metric:   "custom.googleapis.com/my_metric",
				Resource: "gce_instance",
				Points: []types.TimeSeriesPoint{
					{Timestamp: "2024-01-15 10:00:00", Value: 42.5},
					{Timestamp: "2024-01-15 10:01:00", Value: 43.0},
				},
			},
		},
		{
			name: "time series with int64 values",
			input: &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type: "custom.googleapis.com/counter",
				},
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(timestamp),
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{Int64Value: 100},
						},
					},
				},
			},
			expected: types.TimeSeriesData{
				Metric: "custom.googleapis.com/counter",
				Points: []types.TimeSeriesPoint{
					{Timestamp: "2024-01-15 10:00:00", Value: 100.0},
				},
			},
		},
		{
			name: "time series without resource",
			input: &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type: "custom.googleapis.com/simple",
				},
				Points: []*monitoringpb.Point{},
			},
			expected: types.TimeSeriesData{
				Metric:   "custom.googleapis.com/simple",
				Resource: "",
				Points:   []types.TimeSeriesPoint{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertTimeSeries(tc.input)
			if result.Metric != tc.expected.Metric {
				t.Errorf("Metric: got %q, want %q", result.Metric, tc.expected.Metric)
			}
			if result.Resource != tc.expected.Resource {
				t.Errorf("Resource: got %q, want %q", result.Resource, tc.expected.Resource)
			}
			if len(result.Points) != len(tc.expected.Points) {
				t.Fatalf("Points length: got %d, want %d", len(result.Points), len(tc.expected.Points))
			}
			for i, point := range result.Points {
				if point.Timestamp != tc.expected.Points[i].Timestamp {
					t.Errorf("Points[%d].Timestamp: got %q, want %q", i, point.Timestamp, tc.expected.Points[i].Timestamp)
				}
				if point.Value != tc.expected.Points[i].Value {
					t.Errorf("Points[%d].Value: got %f, want %f", i, point.Value, tc.expected.Points[i].Value)
				}
			}
		})
	}
}

func TestConvertAlertPolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    *monitoringpb.AlertPolicy
		expected types.AlertPolicy
	}{
		{
			name: "enabled policy",
			input: &monitoringpb.AlertPolicy{
				Name:        "projects/my-project/alertPolicies/alert123",
				DisplayName: "High CPU Alert",
				Enabled:     wrapperspb.Bool(true),
			},
			expected: types.AlertPolicy{
				Name:        "projects/my-project/alertPolicies/alert123",
				DisplayName: "High CPU Alert",
				Enabled:     true,
			},
		},
		{
			name: "disabled policy",
			input: &monitoringpb.AlertPolicy{
				Name:        "projects/my-project/alertPolicies/alert456",
				DisplayName: "Memory Alert",
				Enabled:     wrapperspb.Bool(false),
			},
			expected: types.AlertPolicy{
				Name:        "projects/my-project/alertPolicies/alert456",
				DisplayName: "Memory Alert",
				Enabled:     false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertAlertPolicy(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
			if result.Enabled != tc.expected.Enabled {
				t.Errorf("Enabled: got %v, want %v", result.Enabled, tc.expected.Enabled)
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
