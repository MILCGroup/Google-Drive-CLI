package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	loggingmgr "github.com/dl-alexandre/gdrv/internal/cloudlogging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/option"
)

type CloudLoggingCmd struct {
	Entries CloudLoggingEntriesCmd `cmd:"" help:"Manage log entries"`
	Logs    CloudLoggingLogsCmd    `cmd:"" help:"Manage logs"`
	Sinks   CloudLoggingSinksCmd   `cmd:"" help:"Manage sinks"`
	Metrics CloudLoggingMetricsCmd `cmd:"" help:"Manage log metrics"`
}

type CloudLoggingEntriesCmd struct {
	List CloudLoggingEntriesListCmd `cmd:"" help:"List log entries"`
}

type CloudLoggingLogsCmd struct {
	List   CloudLoggingLogsListCmd   `cmd:"" help:"List logs"`
	Delete CloudLoggingLogsDeleteCmd `cmd:"" help:"Delete a log"`
}

type CloudLoggingSinksCmd struct {
	List CloudLoggingSinksListCmd `cmd:"" help:"List sinks"`
	Get  CloudLoggingSinksGetCmd  `cmd:"" help:"Get a sink"`
}

type CloudLoggingMetricsCmd struct {
	List CloudLoggingMetricsListCmd `cmd:"" help:"List metrics"`
	Get  CloudLoggingMetricsGetCmd  `cmd:"" help:"Get a metric"`
}

type CloudLoggingEntriesListCmd struct {
	ProjectID string `arg:"" name:"project-id" help:"Project ID"`
	Filter    string `help:"Filter expression" name:"filter"`
	PageSize  int32  `help:"Page size" default:"100" name:"page-size"`
}

type CloudLoggingLogsListCmd struct {
	Parent   string `arg:"" name:"parent" help:"Parent resource (e.g., projects/my-project)"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type CloudLoggingLogsDeleteCmd struct {
	LogName string `arg:"" name:"log-name" help:"Full log name to delete"`
}

type CloudLoggingSinksListCmd struct {
	Parent   string `arg:"" name:"parent" help:"Parent resource (e.g., projects/my-project)"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type CloudLoggingSinksGetCmd struct {
	SinkName string `arg:"" name:"sink-name" help:"Full sink name"`
}

type CloudLoggingMetricsListCmd struct {
	Parent   string `arg:"" name:"parent" help:"Parent resource (e.g., projects/my-project)"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type CloudLoggingMetricsGetCmd struct {
	MetricName string `arg:"" name:"metric-name" help:"Full metric name"`
}

func (cmd *CloudLoggingEntriesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.entries.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListLogEntries(ctx, reqCtx, cmd.ProjectID, cmd.Filter, "", cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "logging.entries.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("logging.entries.list", result.Entries)
	}
	return out.WriteSuccess("logging.entries.list", result)
}

func (cmd *CloudLoggingSinksListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.sinks.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListSinks(ctx, reqCtx, cmd.Parent, cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "logging.sinks.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("logging.sinks.list", result.Sinks)
	}
	return out.WriteSuccess("logging.sinks.list", result)
}

func (cmd *CloudLoggingLogsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.logs.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListLogs(ctx, reqCtx, cmd.Parent, cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "logging.logs.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("logging.logs.list", result)
	}
	return out.WriteSuccess("logging.logs.list", result)
}

func (cmd *CloudLoggingLogsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.logs.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteLog(ctx, reqCtx, cmd.LogName); err != nil {
		return handleCLIError(out, "logging.logs.delete", err)
	}

	return out.WriteSuccess("logging.logs.delete", map[string]string{"logName": cmd.LogName, "status": "deleted"})
}

func (cmd *CloudLoggingSinksGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.sinks.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetSink(ctx, reqCtx, cmd.SinkName)
	if err != nil {
		return handleCLIError(out, "logging.sinks.get", err)
	}

	return out.WriteSuccess("logging.sinks.get", result)
}

func (cmd *CloudLoggingMetricsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.metrics.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListMetrics(ctx, reqCtx, cmd.Parent, cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "logging.metrics.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("logging.metrics.list", result.Metrics)
	}
	return out.WriteSuccess("logging.metrics.list", result)
}

func (cmd *CloudLoggingMetricsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getCloudLoggingManager(ctx, flags)
	if err != nil {
		return out.WriteError("logging.metrics.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetMetric(ctx, reqCtx, cmd.MetricName)
	if err != nil {
		return handleCLIError(out, "logging.metrics.get", err)
	}

	return out.WriteSuccess("logging.metrics.get", result)
}

func getCloudLoggingManager(ctx context.Context, flags types.GlobalFlags) (*loggingmgr.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	tokenSource := authMgr.GetTokenSource(ctx, creds)
	opt := option.WithTokenSource(tokenSource)

	mgr, err := loggingmgr.NewManager(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, nil
}
