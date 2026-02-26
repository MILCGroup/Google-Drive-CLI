package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	meetmgr "github.com/dl-alexandre/gdrv/internal/meet"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/option"
)

type MeetCmd struct {
	Spaces            MeetSpacesCmd            `cmd:"" help:"Manage Meet spaces"`
	ConferenceRecords MeetConferenceRecordsCmd `cmd:"" help:"Manage conference records"`
}

type MeetSpacesCmd struct {
	Get           MeetSpacesGetCmd           `cmd:"" help:"Get a space"`
	Create        MeetSpacesCreateCmd        `cmd:"" help:"Create a new space"`
	Update        MeetSpacesUpdateCmd        `cmd:"" help:"Update a space"`
	EndConference MeetSpacesEndConferenceCmd `cmd:"" help:"End active conference"`
}

type MeetConferenceRecordsCmd struct {
	List         MeetConferenceRecordsListCmd         `cmd:"" help:"List conference records"`
	Get          MeetConferenceRecordsGetCmd          `cmd:"" help:"Get a conference record"`
	Participants MeetConferenceRecordsParticipantsCmd `cmd:"" help:"List participants"`
}

type MeetSpacesGetCmd struct {
	Name string `arg:"" name:"name" help:"Space name (e.g., spaces/ABC123)"`
}

type MeetSpacesCreateCmd struct {
	AccessType string `help:"Access type (OPEN, TRUSTED, or RESTRICTED)" default:"OPEN" name:"access-type"`
}

type MeetSpacesUpdateCmd struct {
	Name       string `arg:"" name:"name" help:"Space name (e.g., spaces/ABC123)"`
	AccessType string `help:"Access type (OPEN, TRUSTED, or RESTRICTED)" name:"access-type"`
}

type MeetSpacesEndConferenceCmd struct {
	Name string `arg:"" name:"name" help:"Space name (e.g., spaces/ABC123)"`
}

type MeetConferenceRecordsListCmd struct {
	Filter   string `help:"Filter expression" name:"filter"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

type MeetConferenceRecordsGetCmd struct {
	Name string `arg:"" name:"name" help:"Conference record name"`
}

type MeetConferenceRecordsParticipantsCmd struct {
	Parent   string `arg:"" name:"parent" help:"Parent resource name"`
	PageSize int32  `help:"Page size" default:"100" name:"page-size"`
}

func (cmd *MeetSpacesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMeetManager(ctx, flags)
	if err != nil {
		return out.WriteError("meet.spaces.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetSpace(ctx, reqCtx, cmd.Name)
	if err != nil {
		return handleCLIError(out, "meet.spaces.get", err)
	}

	return out.WriteSuccess("meet.spaces.get", result)
}

func (cmd *MeetSpacesCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMeetManager(ctx, flags)
	if err != nil {
		return out.WriteError("meet.spaces.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateSpace(ctx, reqCtx, cmd.AccessType)
	if err != nil {
		return handleCLIError(out, "meet.spaces.create", err)
	}

	return out.WriteSuccess("meet.spaces.create", result)
}

func (cmd *MeetSpacesEndConferenceCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMeetManager(ctx, flags)
	if err != nil {
		return out.WriteError("meet.spaces.end-conference", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.EndActiveConference(ctx, reqCtx, cmd.Name); err != nil {
		return handleCLIError(out, "meet.spaces.end-conference", err)
	}

	return out.WriteSuccess("meet.spaces.end-conference", map[string]string{
		"name":   cmd.Name,
		"status": "conference ended",
	})
}

func (cmd *MeetConferenceRecordsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getMeetManager(ctx, flags)
	if err != nil {
		return out.WriteError("meet.conference-records.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	defer func() { _ = mgr.Close() }()

	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.ListConferenceRecords(ctx, reqCtx, cmd.PageSize, "", cmd.Filter)
	if err != nil {
		return handleCLIError(out, "meet.conference-records.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("meet.conference-records.list", result.ConferenceRecords)
	}
	return out.WriteSuccess("meet.conference-records.list", result)
}

func getMeetManager(ctx context.Context, flags types.GlobalFlags) (*meetmgr.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	tokenSource := authMgr.GetTokenSource(ctx, creds)
	opt := option.WithTokenSource(tokenSource)

	mgr, err := meetmgr.NewManager(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, nil
}
