package types

import "fmt"

// LogEntry represents a Cloud Logging entry
type LogEntry struct {
	LogName      string                 `json:"logName"`
	Timestamp    string                 `json:"timestamp"`
	Severity     string                 `json:"severity"`
	TextPayload  string                 `json:"textPayload,omitempty"`
	JSONPayload  map[string]interface{} `json:"jsonPayload,omitempty"`
	ProtoPayload map[string]interface{} `json:"protoPayload,omitempty"`
	Labels       map[string]string      `json:"labels,omitempty"`
	Resource     *LogResource           `json:"resource,omitempty"`
}

func (e *LogEntry) Headers() []string {
	return []string{"Timestamp", "Severity", "Log Name", "Payload"}
}

func (e *LogEntry) Rows() [][]string {
	payload := e.TextPayload
	if payload == "" && len(e.JSONPayload) > 0 {
		payload = "[JSON]"
	}
	if len(payload) > 50 {
		payload = payload[:47] + "..."
	}
	return [][]string{{
		e.Timestamp,
		e.Severity,
		truncateID(e.LogName, 30),
		payload,
	}}
}

func (e *LogEntry) EmptyMessage() string {
	return "No log entry found"
}

// LogResource represents a monitored resource
type LogResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels,omitempty"`
}

// LogSink represents a logging sink
type LogSink struct {
	Name        string            `json:"name"`
	Destination string            `json:"destination"`
	Filter      string            `json:"filter,omitempty"`
	Description string            `json:"description,omitempty"`
	Disabled    bool              `json:"disabled"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func (s *LogSink) Headers() []string {
	return []string{"Sink Name", "Destination", "Filter", "Disabled"}
}

func (s *LogSink) Rows() [][]string {
	filter := s.Filter
	if len(filter) > 30 {
		filter = filter[:27] + "..."
	}
	return [][]string{{
		truncateID(s.Name, 30),
		truncateID(s.Destination, 40),
		filter,
		fmt.Sprintf("%v", s.Disabled),
	}}
}

func (s *LogSink) EmptyMessage() string {
	return "No sink information available"
}

// LogMetric represents a logs-based metric
type LogMetric struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Filter      string `json:"filter"`
}

func (m *LogMetric) Headers() []string {
	return []string{"Metric Name", "Description", "Filter"}
}

func (m *LogMetric) Rows() [][]string {
	filter := m.Filter
	if len(filter) > 40 {
		filter = filter[:37] + "..."
	}
	return [][]string{{
		truncateID(m.Name, 30),
		m.Description,
		filter,
	}}
}

func (m *LogMetric) EmptyMessage() string {
	return "No metric information available"
}

// LogEntriesListResponse represents a list of log entries
type LogEntriesListResponse struct {
	Entries       []LogEntry `json:"entries"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}

func (r *LogEntriesListResponse) Headers() []string {
	return []string{"Timestamp", "Severity", "Log Name", "Payload"}
}

func (r *LogEntriesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Entries))
	for i, entry := range r.Entries {
		payload := entry.TextPayload
		if payload == "" && len(entry.JSONPayload) > 0 {
			payload = "[JSON]"
		}
		if len(payload) > 50 {
			payload = payload[:47] + "..."
		}
		rows[i] = []string{
			entry.Timestamp,
			entry.Severity,
			truncateID(entry.LogName, 30),
			payload,
		}
	}
	return rows
}

func (r *LogEntriesListResponse) EmptyMessage() string {
	return "No log entries found"
}

// LogSinksListResponse represents a list of log sinks
type LogSinksListResponse struct {
	Sinks         []LogSink `json:"sinks"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

func (r *LogSinksListResponse) Headers() []string {
	return []string{"Sink Name", "Destination", "Filter", "Disabled"}
}

func (r *LogSinksListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Sinks))
	for i, sink := range r.Sinks {
		filter := sink.Filter
		if len(filter) > 30 {
			filter = filter[:27] + "..."
		}
		rows[i] = []string{
			truncateID(sink.Name, 30),
			truncateID(sink.Destination, 40),
			filter,
			fmt.Sprintf("%v", sink.Disabled),
		}
	}
	return rows
}

func (r *LogSinksListResponse) EmptyMessage() string {
	return "No sinks found"
}

// LogMetricsListResponse represents a list of log metrics
type LogMetricsListResponse struct {
	Metrics       []LogMetric `json:"metrics"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

func (r *LogMetricsListResponse) Headers() []string {
	return []string{"Metric Name", "Description", "Filter"}
}

func (r *LogMetricsListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Metrics))
	for i, metric := range r.Metrics {
		filter := metric.Filter
		if len(filter) > 40 {
			filter = filter[:37] + "..."
		}
		rows[i] = []string{
			truncateID(metric.Name, 30),
			metric.Description,
			filter,
		}
	}
	return rows
}

func (r *LogMetricsListResponse) EmptyMessage() string {
	return "No metrics found"
}
