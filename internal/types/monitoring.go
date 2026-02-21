package types

import "fmt"

// MetricDescriptor represents a Cloud Monitoring metric
type MetricDescriptor struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit,omitempty"`
}

func (m *MetricDescriptor) Headers() []string {
	return []string{"Metric Name", "Type", "Display Name", "Unit"}
}

func (m *MetricDescriptor) Rows() [][]string {
	return [][]string{{
		truncateID(m.Name, 40),
		m.Type,
		m.DisplayName,
		m.Unit,
	}}
}

func (m *MetricDescriptor) EmptyMessage() string {
	return "No metric information available"
}

// TimeSeriesData represents time series data points
type TimeSeriesData struct {
	Metric   string            `json:"metric"`
	Resource string            `json:"resource,omitempty"`
	Points   []TimeSeriesPoint `json:"points"`
}

func (t *TimeSeriesData) Headers() []string {
	return []string{"Metric", "Resource", "Points Count"}
}

func (t *TimeSeriesData) Rows() [][]string {
	return [][]string{{
		truncateID(t.Metric, 40),
		truncateID(t.Resource, 40),
		fmt.Sprintf("%d", len(t.Points)),
	}}
}

func (t *TimeSeriesData) EmptyMessage() string {
	return "No time series data available"
}

// TimeSeriesPoint represents a single data point
type TimeSeriesPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// AlertPolicy represents a Cloud Monitoring alert policy
type AlertPolicy struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Enabled     bool   `json:"enabled"`
}

func (a *AlertPolicy) Headers() []string {
	return []string{"Policy Name", "Display Name", "Enabled"}
}

func (a *AlertPolicy) Rows() [][]string {
	return [][]string{{
		truncateID(a.Name, 40),
		a.DisplayName,
		fmt.Sprintf("%v", a.Enabled),
	}}
}

func (a *AlertPolicy) EmptyMessage() string {
	return "No alert policy information available"
}

// MetricDescriptorsListResponse represents a list of metric descriptors
type MetricDescriptorsListResponse struct {
	Descriptors   []MetricDescriptor `json:"descriptors"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
}

func (r *MetricDescriptorsListResponse) Headers() []string {
	return []string{"Metric Name", "Type", "Display Name", "Unit"}
}

func (r *MetricDescriptorsListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Descriptors))
	for i, desc := range r.Descriptors {
		rows[i] = []string{
			truncateID(desc.Name, 40),
			desc.Type,
			desc.DisplayName,
			desc.Unit,
		}
	}
	return rows
}

func (r *MetricDescriptorsListResponse) EmptyMessage() string {
	return "No metric descriptors found"
}

// AlertPoliciesListResponse represents a list of alert policies
type AlertPoliciesListResponse struct {
	Policies      []AlertPolicy `json:"policies"`
	NextPageToken string        `json:"nextPageToken,omitempty"`
}

func (r *AlertPoliciesListResponse) Headers() []string {
	return []string{"Policy Name", "Display Name", "Enabled"}
}

func (r *AlertPoliciesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Policies))
	for i, policy := range r.Policies {
		rows[i] = []string{
			truncateID(policy.Name, 40),
			policy.DisplayName,
			fmt.Sprintf("%v", policy.Enabled),
		}
	}
	return rows
}

func (r *AlertPoliciesListResponse) EmptyMessage() string {
	return "No alert policies found"
}
