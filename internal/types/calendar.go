package types

import "fmt"

type CalendarAttendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
	Self           bool   `json:"self,omitempty"`
	Optional       bool   `json:"optional,omitempty"`
}

type CalendarEvent struct {
	ID          string             `json:"id"`
	Summary     string             `json:"summary"`
	Description string             `json:"description,omitempty"`
	Location    string             `json:"location,omitempty"`
	Start       string             `json:"start"`
	End         string             `json:"end"`
	Status      string             `json:"status,omitempty"`
	Creator     string             `json:"creator,omitempty"`
	Organizer   string             `json:"organizer,omitempty"`
	Attendees   []CalendarAttendee `json:"attendees,omitempty"`
	HangoutLink string             `json:"hangoutLink,omitempty"`
	HtmlLink    string             `json:"htmlLink,omitempty"`
	Recurrence  []string           `json:"recurrence,omitempty"`
}

func (e *CalendarEvent) Headers() []string {
	return []string{"ID", "Summary", "Start", "End", "Location", "Status"}
}

func (e *CalendarEvent) Rows() [][]string {
	return [][]string{{
		e.ID,
		e.Summary,
		e.Start,
		e.End,
		e.Location,
		e.Status,
	}}
}

func (e *CalendarEvent) EmptyMessage() string {
	return "No event found"
}

type CalendarEventList struct {
	Events  []CalendarEvent `json:"events"`
	Summary string          `json:"summary,omitempty"`
}

func (l *CalendarEventList) Headers() []string {
	return []string{"ID", "Summary", "Start", "End", "Location", "Status"}
}

func (l *CalendarEventList) Rows() [][]string {
	rows := make([][]string, len(l.Events))
	for i, event := range l.Events {
		rows[i] = []string{
			event.ID,
			event.Summary,
			event.Start,
			event.End,
			event.Location,
			event.Status,
		}
	}
	return rows
}

func (l *CalendarEventList) EmptyMessage() string {
	return "No events found"
}

type CalendarBusyPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type CalendarFreeBusy struct {
	Calendars map[string][]CalendarBusyPeriod `json:"calendars"`
}

func (fb *CalendarFreeBusy) Headers() []string {
	return []string{"Calendar", "Start", "End"}
}

func (fb *CalendarFreeBusy) Rows() [][]string {
	var rows [][]string
	for calendar, periods := range fb.Calendars {
		for _, period := range periods {
			rows = append(rows, []string{
				calendar,
				period.Start,
				period.End,
			})
		}
	}
	return rows
}

func (fb *CalendarFreeBusy) EmptyMessage() string {
	return "No busy periods found"
}

type CalendarConflict struct {
	Event         CalendarEvent   `json:"event"`
	ConflictsWith []CalendarEvent `json:"conflictsWith"`
}

func (c *CalendarConflict) Headers() []string {
	return []string{"Event", "Conflicts With"}
}

func (c *CalendarConflict) Rows() [][]string {
	return [][]string{{
		c.Event.Summary,
		fmt.Sprintf("%d", len(c.ConflictsWith)),
	}}
}

func (c *CalendarConflict) EmptyMessage() string {
	return "No conflicts found"
}

type CalendarEventResult struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	HtmlLink string `json:"htmlLink,omitempty"`
}

func (r *CalendarEventResult) Headers() []string {
	return []string{"ID", "Summary", "Link"}
}

func (r *CalendarEventResult) Rows() [][]string {
	return [][]string{{
		r.ID,
		r.Summary,
		r.HtmlLink,
	}}
}

func (r *CalendarEventResult) EmptyMessage() string {
	return "No event result"
}
