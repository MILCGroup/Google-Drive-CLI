package types

// CLIOutput is the standard JSON envelope for all CLI responses
type CLIOutput struct {
	SchemaVersion string       `json:"schemaVersion"`
	TraceID       string       `json:"traceId"`
	Command       string       `json:"command"`
	Data          interface{}  `json:"data"`
	Warnings      []CLIWarning `json:"warnings"`
	Errors        []CLIError   `json:"errors"`
}

// CLIWarning represents a non-fatal warning
type CLIWarning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // low, medium, high
}

// CLIError represents a structured error
type CLIError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	HTTPStatus  int                    `json:"httpStatus,omitempty"`
	DriveReason string                 `json:"driveReason,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// OutputFormat defines supported output formats
type OutputFormat string

const (
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatTable OutputFormat = "table"
)

// GlobalFlags represents CLI global options
type GlobalFlags struct {
	Profile             string
	DriveID             string
	OutputFormat        OutputFormat
	Quiet               bool
	Verbose             bool
	Debug               bool
	Strict              bool
	NoCache             bool
	CacheTTL            int
	IncludeSharedWithMe bool
	Config              string
	LogFile             string
	DryRun              bool
	Force               bool
	Yes                 bool
	JSON                bool
}
