package monitoring

import (
	"context"
	"errors"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
)

// Manager handles Cloud Monitoring API operations via gRPC
type Manager struct {
	metricClient *monitoring.MetricClient
	alertClient  *monitoring.AlertPolicyClient
}

// NewManager creates a new Cloud Monitoring manager
func NewManager(ctx context.Context, opts ...option.ClientOption) (*Manager, error) {
	metricClient, err := monitoring.NewMetricClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric client: %w", err)
	}

	alertClient, err := monitoring.NewAlertPolicyClient(ctx, opts...)
	if err != nil {
		_ = metricClient.Close()
		return nil, fmt.Errorf("failed to create alert client: %w", err)
	}

	return &Manager{
		metricClient: metricClient,
		alertClient:  alertClient,
	}, nil
}

// Close closes all clients
func (m *Manager) Close() error {
	var err error
	if m.metricClient != nil {
		err = m.metricClient.Close()
	}
	if m.alertClient != nil {
		if cerr := m.alertClient.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// ListMetricDescriptors lists metric descriptors
func (m *Manager) ListMetricDescriptors(ctx context.Context, reqCtx *types.RequestContext, filter string, pageSize int32) (*types.MetricDescriptorsListResponse, error) {
	req := &monitoringpb.ListMetricDescriptorsRequest{
		Filter:   filter,
		PageSize: pageSize,
	}

	iter := m.metricClient.ListMetricDescriptors(ctx, req)
	var descriptors []types.MetricDescriptor
	for {
		desc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list metric descriptors: %w", err)
		}
		descriptors = append(descriptors, convertMetricDescriptor(desc))
	}

	return &types.MetricDescriptorsListResponse{
		Descriptors: descriptors,
	}, nil
}

// GetMetricDescriptor gets a specific metric descriptor
func (m *Manager) GetMetricDescriptor(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.MetricDescriptor, error) {
	req := &monitoringpb.GetMetricDescriptorRequest{
		Name: name,
	}

	resp, err := m.metricClient.GetMetricDescriptor(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric descriptor: %w", err)
	}

	desc := convertMetricDescriptor(resp)
	return &desc, nil
}

// ListTimeSeries lists time series data
func (m *Manager) ListTimeSeries(ctx context.Context, reqCtx *types.RequestContext, filter, interval string, pageSize int32) ([]types.TimeSeriesData, error) {
	// Parse interval or use defaults
	req := &monitoringpb.ListTimeSeriesRequest{
		Filter:   filter,
		PageSize: pageSize,
	}

	iter := m.metricClient.ListTimeSeries(ctx, req)
	var series []types.TimeSeriesData
	for {
		ts, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list time series: %w", err)
		}
		series = append(series, convertTimeSeries(ts))
	}

	return series, nil
}

// ListAlertPolicies lists alert policies
func (m *Manager) ListAlertPolicies(ctx context.Context, reqCtx *types.RequestContext, filter string, pageSize int32) (*types.AlertPoliciesListResponse, error) {
	req := &monitoringpb.ListAlertPoliciesRequest{
		Filter:   filter,
		PageSize: pageSize,
	}

	iter := m.alertClient.ListAlertPolicies(ctx, req)
	var policies []types.AlertPolicy
	for {
		policy, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list alert policies: %w", err)
		}
		policies = append(policies, convertAlertPolicy(policy))
	}

	return &types.AlertPoliciesListResponse{
		Policies: policies,
	}, nil
}

// GetAlertPolicy gets a specific alert policy
func (m *Manager) GetAlertPolicy(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.AlertPolicy, error) {
	req := &monitoringpb.GetAlertPolicyRequest{
		Name: name,
	}

	resp, err := m.alertClient.GetAlertPolicy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert policy: %w", err)
	}

	policy := convertAlertPolicy(resp)
	return &policy, nil
}

func convertMetricDescriptor(desc *metricpb.MetricDescriptor) types.MetricDescriptor {
	return types.MetricDescriptor{
		Name:        desc.Name,
		Type:        desc.Type,
		DisplayName: desc.DisplayName,
		Description: desc.Description,
		Unit:        desc.Unit,
	}
}

func convertTimeSeries(ts *monitoringpb.TimeSeries) types.TimeSeriesData {
	t := types.TimeSeriesData{
		Metric: ts.Metric.Type,
	}
	if ts.Resource != nil {
		t.Resource = ts.Resource.Type
	}

	for _, point := range ts.Points {
		p := types.TimeSeriesPoint{
			Timestamp: point.Interval.EndTime.AsTime().Format("2006-01-02 15:04:05"),
		}
		// Extract value based on type
		switch v := point.Value.Value.(type) {
		case *monitoringpb.TypedValue_DoubleValue:
			p.Value = v.DoubleValue
		case *monitoringpb.TypedValue_Int64Value:
			p.Value = float64(v.Int64Value)
		}
		t.Points = append(t.Points, p)
	}

	return t
}

func convertAlertPolicy(policy *monitoringpb.AlertPolicy) types.AlertPolicy {
	return types.AlertPolicy{
		Name:        policy.Name,
		DisplayName: policy.DisplayName,
		Enabled:     policy.Enabled.Value,
	}
}
