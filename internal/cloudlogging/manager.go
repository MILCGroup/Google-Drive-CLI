package cloudlogging

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Manager handles Cloud Logging API operations via gRPC
type Manager struct {
	client        *logging.Client
	configClient  *logging.ConfigClient
	metricsClient *logging.MetricsClient
}

// NewManager creates a new Cloud Logging manager
func NewManager(ctx context.Context, opts ...option.ClientOption) (*Manager, error) {
	client, err := logging.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	configClient, err := logging.NewConfigClient(ctx, opts...)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	metricsClient, err := logging.NewMetricsClient(ctx, opts...)
	if err != nil {
		client.Close()
		configClient.Close()
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &Manager{
		client:        client,
		configClient:  configClient,
		metricsClient: metricsClient,
	}, nil
}

// Close closes all clients
func (m *Manager) Close() error {
	var err error
	if m.client != nil {
		err = m.client.Close()
	}
	if m.configClient != nil {
		if cerr := m.configClient.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	if m.metricsClient != nil {
		if cerr := m.metricsClient.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// ListLogEntries lists log entries
func (m *Manager) ListLogEntries(ctx context.Context, reqCtx *types.RequestContext, projectID, filter, pageToken string, pageSize int32) (*types.LogEntriesListResponse, error) {
	resourceNames := []string{fmt.Sprintf("projects/%s", projectID)}

	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: resourceNames,
		Filter:        filter,
		PageToken:     pageToken,
		PageSize:      pageSize,
	}

	iter := m.client.ListLogEntries(ctx, req)
	var entries []types.LogEntry
	for {
		entry, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list log entries: %w", err)
		}
		entries = append(entries, convertLogEntry(entry))
	}

	return &types.LogEntriesListResponse{
		Entries: entries,
	}, nil
}

// ListLogs lists log names
func (m *Manager) ListLogs(ctx context.Context, reqCtx *types.RequestContext, parent string, pageSize int32) ([]string, error) {
	req := &loggingpb.ListLogsRequest{
		Parent:   parent,
		PageSize: pageSize,
	}

	iter := m.client.ListLogs(ctx, req)
	var logs []string
	for {
		log, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list logs: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// DeleteLog deletes a log and all its entries
func (m *Manager) DeleteLog(ctx context.Context, reqCtx *types.RequestContext, logName string) error {
	req := &loggingpb.DeleteLogRequest{
		LogName: logName,
	}

	if err := m.client.DeleteLog(ctx, req); err != nil {
		return fmt.Errorf("failed to delete log: %w", err)
	}

	return nil
}

// ListSinks lists logging sinks
func (m *Manager) ListSinks(ctx context.Context, reqCtx *types.RequestContext, parent string, pageSize int32) (*types.LogSinksListResponse, error) {
	req := &loggingpb.ListSinksRequest{
		Parent:   parent,
		PageSize: pageSize,
	}

	iter := m.configClient.ListSinks(ctx, req)
	var sinks []types.LogSink
	for {
		sink, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list sinks: %w", err)
		}
		sinks = append(sinks, convertLogSink(sink))
	}

	return &types.LogSinksListResponse{
		Sinks: sinks,
	}, nil
}

// GetSink gets a logging sink
func (m *Manager) GetSink(ctx context.Context, reqCtx *types.RequestContext, sinkName string) (*types.LogSink, error) {
	req := &loggingpb.GetSinkRequest{
		SinkName: sinkName,
	}

	resp, err := m.configClient.GetSink(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sink: %w", err)
	}

	sink := convertLogSink(resp)
	return &sink, nil
}

// ListMetrics lists logs-based metrics
func (m *Manager) ListMetrics(ctx context.Context, reqCtx *types.RequestContext, parent string, pageSize int32) (*types.LogMetricsListResponse, error) {
	req := &loggingpb.ListLogMetricsRequest{
		Parent:   parent,
		PageSize: pageSize,
	}

	iter := m.metricsClient.ListLogMetrics(ctx, req)
	var metrics []types.LogMetric
	for {
		metric, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list metrics: %w", err)
		}
		metrics = append(metrics, convertLogMetric(metric))
	}

	return &types.LogMetricsListResponse{
		Metrics: metrics,
	}, nil
}

// GetMetric gets a logs-based metric
func (m *Manager) GetMetric(ctx context.Context, reqCtx *types.RequestContext, metricName string) (*types.LogMetric, error) {
	req := &loggingpb.GetLogMetricRequest{
		MetricName: metricName,
	}

	resp, err := m.metricsClient.GetLogMetric(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	metric := convertLogMetric(resp)
	return &metric, nil
}

func convertLogEntry(entry *loggingpb.LogEntry) types.LogEntry {
	e := types.LogEntry{
		LogName:  entry.LogName,
		Severity: entry.Severity.String(),
		Labels:   entry.Labels,
	}

	if entry.Timestamp != nil {
		e.Timestamp = entry.Timestamp.AsTime().Format("2006-01-02 15:04:05")
	}

	switch payload := entry.Payload.(type) {
	case *loggingpb.LogEntry_TextPayload:
		e.TextPayload = payload.TextPayload
	case *loggingpb.LogEntry_JsonPayload:
		e.JSONPayload = payload.JsonPayload.AsMap()
	case *loggingpb.LogEntry_ProtoPayload:
		// ProtoPayload is *anypb.Any, can't easily convert to map
		e.TextPayload = "[Proto]"
	}

	if entry.Resource != nil {
		e.Resource = &types.LogResource{
			Type:   entry.Resource.Type,
			Labels: entry.Resource.Labels,
		}
	}

	return e
}

func convertLogSink(sink *loggingpb.LogSink) types.LogSink {
	return types.LogSink{
		Name:        sink.Name,
		Destination: sink.Destination,
		Filter:      sink.Filter,
		Description: sink.Description,
		Disabled:    sink.Disabled,
	}
}

func convertLogMetric(metric *loggingpb.LogMetric) types.LogMetric {
	return types.LogMetric{
		Name:        metric.Name,
		Description: metric.Description,
		Filter:      metric.Filter,
	}
}
