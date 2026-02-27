package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	calendarmgr "github.com/milcgroup/gdrv/internal/calendar"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	calendarapi "google.golang.org/api/calendar/v3"
)

// CalendarCmd is the top-level command group for Google Calendar operations.
type CalendarCmd struct {
	Events    CalendarEventsCmd    `cmd:"" help:"Manage calendar events"`
	FreeBusy  CalendarFreeBusyCmd  `cmd:"free-busy" help:"Query free/busy information"`
	Conflicts CalendarConflictsCmd `cmd:"" help:"Find conflicting events in a time range"`
}

// CalendarEventsCmd groups all event subcommands.
type CalendarEventsCmd struct {
	List    CalendarEventsListCmd    `cmd:"" help:"List calendar events"`
	Get     CalendarEventsGetCmd     `cmd:"" help:"Get a calendar event by ID"`
	Search  CalendarEventsSearchCmd  `cmd:"" help:"Search calendar events"`
	Create  CalendarEventsCreateCmd  `cmd:"" help:"Create a calendar event"`
	Update  CalendarEventsUpdateCmd  `cmd:"" help:"Update an existing calendar event"`
	Delete  CalendarEventsDeleteCmd  `cmd:"" help:"Delete a calendar event"`
	Respond CalendarEventsRespondCmd `cmd:"" help:"Respond to a calendar event invitation"`
}

// CalendarEventsListCmd lists events from a calendar.
type CalendarEventsListCmd struct {
	Calendar  string `help:"Calendar ID" default:"primary" name:"calendar"`
	Today     bool   `help:"Show only today's events" name:"today"`
	Week      bool   `help:"Show events for the next 7 days" name:"week"`
	Days      int    `help:"Show events for the next N days" name:"days"`
	Limit     int    `help:"Maximum events to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
	Query     string `help:"Free-text search query" name:"query"`
	OrderBy   string `help:"Sort order (startTime or updated)" default:"startTime" name:"order-by"`
}

// CalendarEventsGetCmd retrieves a single event by ID.
type CalendarEventsGetCmd struct {
	EventID  string `arg:"" name:"event-id" help:"Event ID"`
	Calendar string `help:"Calendar ID" default:"primary" name:"calendar"`
}

// CalendarEventsSearchCmd searches events by query.
type CalendarEventsSearchCmd struct {
	Calendar string `help:"Calendar ID" default:"primary" name:"calendar"`
	Query    string `help:"Search query" required:"" name:"query"`
	Limit    int    `help:"Maximum events to return" default:"100" name:"limit"`
}

// CalendarEventsCreateCmd creates a new calendar event.
type CalendarEventsCreateCmd struct {
	Calendar    string `help:"Calendar ID" default:"primary" name:"calendar"`
	Summary     string `help:"Event title" required:"" name:"summary"`
	Start       string `help:"Start time (RFC3339 or YYYY-MM-DD for all-day)" required:"" name:"start"`
	End         string `help:"End time (RFC3339 or YYYY-MM-DD for all-day)" required:"" name:"end"`
	Description string `help:"Event description" name:"description"`
	Location    string `help:"Event location" name:"location"`
	Attendees   string `help:"Comma-separated attendee email addresses" name:"attendees"`
	AllDay      bool   `help:"Create an all-day event" name:"all-day"`
	Recurrence  string `help:"Recurrence rule (e.g. RRULE:FREQ=WEEKLY;COUNT=5)" name:"recurrence"`
	SendUpdates string `help:"Send update notifications (none, all, externalOnly)" default:"none" name:"send-updates"`
}

// CalendarEventsUpdateCmd updates an existing calendar event.
type CalendarEventsUpdateCmd struct {
	EventID     string `arg:"" name:"event-id" help:"Event ID"`
	Calendar    string `help:"Calendar ID" default:"primary" name:"calendar"`
	Summary     string `help:"New event title" name:"summary"`
	Description string `help:"New event description" name:"description"`
	Location    string `help:"New event location" name:"location"`
	Start       string `help:"New start time (RFC3339)" name:"start"`
	End         string `help:"New end time (RFC3339)" name:"end"`
	SendUpdates string `help:"Send update notifications (none, all, externalOnly)" default:"none" name:"send-updates"`
}

// CalendarEventsDeleteCmd deletes a calendar event.
type CalendarEventsDeleteCmd struct {
	EventID     string `arg:"" name:"event-id" help:"Event ID"`
	Calendar    string `help:"Calendar ID" default:"primary" name:"calendar"`
	SendUpdates string `help:"Send update notifications (none, all, externalOnly)" default:"none" name:"send-updates"`
}

// CalendarEventsRespondCmd responds to a calendar event invitation.
type CalendarEventsRespondCmd struct {
	EventID     string `arg:"" name:"event-id" help:"Event ID"`
	Calendar    string `help:"Calendar ID" default:"primary" name:"calendar"`
	Response    string `help:"Response status (accepted, declined, tentative)" required:"" name:"response"`
	SendUpdates string `help:"Send update notifications (none, all, externalOnly)" default:"none" name:"send-updates"`
}

// CalendarFreeBusyCmd queries free/busy information.
type CalendarFreeBusyCmd struct {
	Calendars string `help:"Comma-separated calendar IDs" default:"primary" name:"calendars"`
	Start     string `help:"Start time (RFC3339)" required:"" name:"start"`
	End       string `help:"End time (RFC3339)" required:"" name:"end"`
}

// CalendarConflictsCmd finds conflicting events.
type CalendarConflictsCmd struct {
	Calendar string `help:"Calendar ID" default:"primary" name:"calendar"`
	Start    string `help:"Start time (RFC3339)" required:"" name:"start"`
	End      string `help:"End time (RFC3339)" required:"" name:"end"`
}

// getCalendarService returns a Calendar service, API client, and request context.
func getCalendarService(ctx context.Context, flags types.GlobalFlags) (*calendarapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceCalendar); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetCalendarService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return svc, client, reqCtx, nil
}

// resolveTimeRange converts --today, --week, or --days flags to RFC3339 timeMin/timeMax strings.
func resolveTimeRange(today, week bool, days int) (string, string) {
	now := time.Now().UTC()
	if today {
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)
		return startOfDay.Format(time.RFC3339), endOfDay.Format(time.RFC3339)
	}
	if week {
		return now.Format(time.RFC3339), now.Add(7 * 24 * time.Hour).Format(time.RFC3339)
	}
	if days > 0 {
		return now.Format(time.RFC3339), now.Add(time.Duration(days) * 24 * time.Hour).Format(time.RFC3339)
	}
	return "", ""
}

// Run executes the calendar events list command.
func (cmd *CalendarEventsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	_ = client // calendar commands do not use the Drive client directly

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	timeMin, timeMax := resolveTimeRange(cmd.Today, cmd.Week, cmd.Days)

	// When using time-based ordering, singleEvents must be true so recurring
	// events are expanded to their individual instances.
	singleEvents := cmd.OrderBy == "startTime"

	if cmd.Paginate {
		var allEvents []types.CalendarEvent
		var summary string
		pageToken := cmd.PageToken
		for {
			result, nextToken, listErr := mgr.ListEvents(ctx, reqCtx, cmd.Calendar, timeMin, timeMax, int64(cmd.Limit), pageToken, cmd.Query, singleEvents, cmd.OrderBy)
			if listErr != nil {
				return handleCLIError(out, "calendar.events.list", listErr)
			}
			allEvents = append(allEvents, result.Events...)
			if summary == "" {
				summary = result.Summary
			}
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("calendar.events.list", map[string]interface{}{
			"summary": summary,
			"events":  allEvents,
		})
	}

	result, nextPageToken, err := mgr.ListEvents(ctx, reqCtx, cmd.Calendar, timeMin, timeMax, int64(cmd.Limit), cmd.PageToken, cmd.Query, singleEvents, cmd.OrderBy)
	if err != nil {
		return handleCLIError(out, "calendar.events.list", err)
	}

	response := map[string]interface{}{
		"summary": result.Summary,
		"events":  result.Events,
	}
	if nextPageToken != "" {
		response["nextPageToken"] = nextPageToken
	}
	return out.WriteSuccess("calendar.events.list", response)
}

// Run executes the calendar events get command.
func (cmd *CalendarEventsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	event, err := mgr.GetEvent(ctx, reqCtx, cmd.Calendar, cmd.EventID)
	if err != nil {
		return handleCLIError(out, "calendar.events.get", err)
	}

	return out.WriteSuccess("calendar.events.get", event)
}

// Run executes the calendar events search command.
func (cmd *CalendarEventsSearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.search", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	// Use ListEvents with query for richer control over result count
	result, _, err := mgr.ListEvents(ctx, reqCtx, cmd.Calendar, "", "", int64(cmd.Limit), "", cmd.Query, true, "startTime")
	if err != nil {
		return handleCLIError(out, "calendar.events.search", err)
	}

	return out.WriteSuccess("calendar.events.search", result)
}

// Run executes the calendar events create command.
func (cmd *CalendarEventsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	var attendees []string
	if cmd.Attendees != "" {
		attendees = strings.Split(cmd.Attendees, ",")
		for i := range attendees {
			attendees[i] = strings.TrimSpace(attendees[i])
		}
	}

	var recurrence []string
	if cmd.Recurrence != "" {
		recurrence = []string{cmd.Recurrence}
	}

	result, err := mgr.CreateEvent(ctx, reqCtx, cmd.Calendar, cmd.Summary, cmd.Description, cmd.Location, cmd.Start, cmd.End, attendees, cmd.AllDay, recurrence, cmd.SendUpdates)
	if err != nil {
		return handleCLIError(out, "calendar.events.create", err)
	}

	out.Log("Created event: %s", result.ID)
	return out.WriteSuccess("calendar.events.create", result)
}

// Run executes the calendar events update command.
func (cmd *CalendarEventsUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	// Build optional fields: only pass non-empty values as pointers
	var summary, description, location, start, end *string
	if cmd.Summary != "" {
		summary = &cmd.Summary
	}
	if cmd.Description != "" {
		description = &cmd.Description
	}
	if cmd.Location != "" {
		location = &cmd.Location
	}
	if cmd.Start != "" {
		start = &cmd.Start
	}
	if cmd.End != "" {
		end = &cmd.End
	}

	if summary == nil && description == nil && location == nil && start == nil && end == nil {
		return out.WriteError("calendar.events.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one field to update is required").Build())
	}

	result, err := mgr.UpdateEvent(ctx, reqCtx, cmd.Calendar, cmd.EventID, summary, description, location, start, end, cmd.SendUpdates)
	if err != nil {
		return handleCLIError(out, "calendar.events.update", err)
	}

	out.Log("Updated event: %s", result.ID)
	return out.WriteSuccess("calendar.events.update", result)
}

// Run executes the calendar events delete command.
func (cmd *CalendarEventsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteEvent(ctx, reqCtx, cmd.Calendar, cmd.EventID, cmd.SendUpdates); err != nil {
		return handleCLIError(out, "calendar.events.delete", err)
	}

	out.Log("Deleted event: %s", cmd.EventID)
	return out.WriteSuccess("calendar.events.delete", map[string]string{
		"id":     cmd.EventID,
		"status": "deleted",
	})
}

// Run executes the calendar events respond command.
func (cmd *CalendarEventsRespondCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	validResponses := map[string]bool{"accepted": true, "declined": true, "tentative": true}
	if !validResponses[cmd.Response] {
		return out.WriteError("calendar.events.respond", utils.NewCLIError(utils.ErrCodeInvalidArgument, fmt.Sprintf("invalid response %q: must be accepted, declined, or tentative", cmd.Response)).Build())
	}

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.events.respond", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.RespondToEvent(ctx, reqCtx, cmd.Calendar, cmd.EventID, cmd.Response, cmd.SendUpdates)
	if err != nil {
		return handleCLIError(out, "calendar.events.respond", err)
	}

	out.Log("Responded %s to event: %s", cmd.Response, result.ID)
	return out.WriteSuccess("calendar.events.respond", result)
}

// Run executes the calendar free-busy command.
func (cmd *CalendarFreeBusyCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.free-busy", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	calendarIDs := strings.Split(cmd.Calendars, ",")
	for i := range calendarIDs {
		calendarIDs[i] = strings.TrimSpace(calendarIDs[i])
	}

	result, err := mgr.FreeBusy(ctx, reqCtx, calendarIDs, cmd.Start, cmd.End)
	if err != nil {
		return handleCLIError(out, "calendar.free-busy", err)
	}

	return out.WriteSuccess("calendar.free-busy", result)
}

// Run executes the calendar conflicts command.
func (cmd *CalendarConflictsCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getCalendarService(ctx, flags)
	if err != nil {
		return out.WriteError("calendar.conflicts", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := calendarmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	conflicts, err := mgr.FindConflicts(ctx, reqCtx, cmd.Calendar, cmd.Start, cmd.End)
	if err != nil {
		return handleCLIError(out, "calendar.conflicts", err)
	}

	return out.WriteSuccess("calendar.conflicts", map[string]interface{}{
		"conflicts": conflicts,
		"count":     len(conflicts),
	})
}
