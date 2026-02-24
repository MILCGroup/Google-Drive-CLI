package cli

import (
	"context"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	monitoringmgr "github.com/milcgroup/gdrv/internal/monitoring"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/option"
)

type MonitoringCmd struct {
	Metrics       MonitoringMetricsCmd       `cmd:"" help:"Manage metrics"`
	TimeSeries    MonitoringTimeSeriesCmd    `cmd:"" help:"Manage time series"`
	AlertPolicies MonitoringAlertPoliciesCmd `cmd:"" help:"Manage alert policies"`
}

type MonitoringMetricsCmd struct {
	List MonitoringMetricsListCmd `cmd:"" help:"List metric descriptors"`
	Get  MonitoringMetricsGetCmd  `cmd:"" help:"Get a metric descriptor"`
}

type MonitoringTimeSeriesCmd struct {
	List MonitoringTimeSeriesListCmd `cmd:"" help:"List time series"`
}

type MonitoringAlertPoliciesCmd struct {
	List MonitoringAlertPoliciesListCmd `cmd:"" help:"List alert policies"`
	Get  MonitoringAlertPoliciesGetCmd  `cmd:"" help:"Get an alert policy"`
}

type MonitoringMetricsListCmd struct {
	Filter   string `help:"Filter expression" name:"filter"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type MonitoringMetricsGetCmd struct {
	Name string `arg:"" name:"name" help:"Metric descriptor name"`
}

type MonitoringTimeSeriesListCmd struct {
	Filter   string `arg:"" name:"filter" help:"Filter expression"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type MonitoringAlertPoliciesListCmd struct {
	Filter   string `help:"Filter expression" name:"filter"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type MonitoringAlertPoliciesGetCmd struct {
	Name string `arg:"" name:"name" help:"Alert policy name"`
}

func (cmd *MonitoringMetricsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMonitoringManager(ctx, flags)
	if err != nil {
		return out.WriteError("monitoring.metrics.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer mgr.Close()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListMetricDescriptors(ctx, reqCtx, cmd.Filter, cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "monitoring.metrics.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("monitoring.metrics.list", result.Descriptors)
	}
	return out.WriteSuccess("monitoring.metrics.list", result)
}

func (cmd *MonitoringAlertPoliciesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMonitoringManager(ctx, flags)
	if err != nil {
		return out.WriteError("monitoring.alert-policies.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer mgr.Close()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListAlertPolicies(ctx, reqCtx, cmd.Filter, cmd.PageSize)
	if err != nil {
		return handleCLIError(out, "monitoring.alert-policies.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("monitoring.alert-policies.list", result.Policies)
	}
	return out.WriteSuccess("monitoring.alert-policies.list", result)
}

func getMonitoringManager(ctx context.Context, flags types.GlobalFlags) (*monitoringmgr.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	tokenSource := authMgr.GetTokenSource(ctx, creds)
	opt := option.WithTokenSource(tokenSource)

	mgr, err := monitoringmgr.NewManager(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, nil
}
