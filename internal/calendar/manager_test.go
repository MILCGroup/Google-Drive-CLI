package calendar

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/types"
	calendarapi "google.golang.org/api/calendar/v3"
)

func TestBuildEventDateTime(t *testing.T) {
	tests := []struct {
		name        string
		dateTimeStr string
		allDay      bool
		wantDate    string
		wantDT      string
	}{
		{
			name:        "all-day event sets Date only",
			dateTimeStr: "2026-03-15",
			allDay:      true,
			wantDate:    "2026-03-15",
			wantDT:      "",
		},
		{
			name:        "timed event sets DateTime only",
			dateTimeStr: "2026-03-15T10:00:00Z",
			allDay:      false,
			wantDate:    "",
			wantDT:      "2026-03-15T10:00:00Z",
		},
		{
			name:        "all-day with empty string",
			dateTimeStr: "",
			allDay:      true,
			wantDate:    "",
			wantDT:      "",
		},
		{
			name:        "timed with empty string",
			dateTimeStr: "",
			allDay:      false,
			wantDate:    "",
			wantDT:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildEventDateTime(tt.dateTimeStr, tt.allDay)
			if got == nil {
				t.Fatal("expected non-nil EventDateTime")
			}
			if got.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", got.Date, tt.wantDate)
			}
			if got.DateTime != tt.wantDT {
				t.Errorf("DateTime = %q, want %q", got.DateTime, tt.wantDT)
			}
			// Critical invariant: Date and DateTime must never both be set
			if got.Date != "" && got.DateTime != "" {
				t.Error("Date and DateTime are both set — these are mutually exclusive")
			}
		})
	}
}

func TestConvertEvent(t *testing.T) {
	tests := []struct {
		name string
		in   *calendarapi.Event
		want types.CalendarEvent
	}{
		{
			name: "minimal event",
			in: &calendarapi.Event{
				Id:      "evt1",
				Summary: "Meeting",
			},
			want: types.CalendarEvent{
				ID:      "evt1",
				Summary: "Meeting",
			},
		},
		{
			name: "timed event with DateTime",
			in: &calendarapi.Event{
				Id:      "evt2",
				Summary: "Standup",
				Start:   &calendarapi.EventDateTime{DateTime: "2026-03-15T09:00:00Z"},
				End:     &calendarapi.EventDateTime{DateTime: "2026-03-15T09:30:00Z"},
			},
			want: types.CalendarEvent{
				ID:      "evt2",
				Summary: "Standup",
				Start:   "2026-03-15T09:00:00Z",
				End:     "2026-03-15T09:30:00Z",
			},
		},
		{
			name: "all-day event with Date",
			in: &calendarapi.Event{
				Id:      "evt3",
				Summary: "Holiday",
				Start:   &calendarapi.EventDateTime{Date: "2026-12-25"},
				End:     &calendarapi.EventDateTime{Date: "2026-12-26"},
			},
			want: types.CalendarEvent{
				ID:      "evt3",
				Summary: "Holiday",
				Start:   "2026-12-25",
				End:     "2026-12-26",
			},
		},
		{
			name: "prefers DateTime over Date when both present",
			in: &calendarapi.Event{
				Id:    "evt4",
				Start: &calendarapi.EventDateTime{DateTime: "2026-03-15T10:00:00Z", Date: "2026-03-15"},
				End:   &calendarapi.EventDateTime{DateTime: "2026-03-15T11:00:00Z", Date: "2026-03-15"},
			},
			want: types.CalendarEvent{
				ID:    "evt4",
				Start: "2026-03-15T10:00:00Z",
				End:   "2026-03-15T11:00:00Z",
			},
		},
		{
			name: "full event with all fields",
			in: &calendarapi.Event{
				Id:          "evt5",
				Summary:     "All Hands",
				Description: "Quarterly meeting",
				Location:    "Building 42",
				Status:      "confirmed",
				HangoutLink: "https://meet.google.com/abc",
				HtmlLink:    "https://calendar.google.com/event?eid=abc",
				Recurrence:  []string{"RRULE:FREQ=WEEKLY"},
				Start:       &calendarapi.EventDateTime{DateTime: "2026-03-15T14:00:00Z"},
				End:         &calendarapi.EventDateTime{DateTime: "2026-03-15T15:00:00Z"},
				Creator:     &calendarapi.EventCreator{Email: "creator@example.com"},
				Organizer:   &calendarapi.EventOrganizer{Email: "organizer@example.com"},
				Attendees: []*calendarapi.EventAttendee{
					{Email: "alice@example.com", ResponseStatus: "accepted", Self: true},
				},
			},
			want: types.CalendarEvent{
				ID:          "evt5",
				Summary:     "All Hands",
				Description: "Quarterly meeting",
				Location:    "Building 42",
				Status:      "confirmed",
				HangoutLink: "https://meet.google.com/abc",
				HtmlLink:    "https://calendar.google.com/event?eid=abc",
				Recurrence:  []string{"RRULE:FREQ=WEEKLY"},
				Start:       "2026-03-15T14:00:00Z",
				End:         "2026-03-15T15:00:00Z",
				Creator:     "creator@example.com",
				Organizer:   "organizer@example.com",
				Attendees: []types.CalendarAttendee{
					{Email: "alice@example.com", ResponseStatus: "accepted", Self: true},
				},
			},
		},
		{
			name: "nil start and end",
			in: &calendarapi.Event{
				Id: "evt6",
			},
			want: types.CalendarEvent{
				ID: "evt6",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEvent(tt.in)
			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.Summary != tt.want.Summary {
				t.Errorf("Summary = %q, want %q", got.Summary, tt.want.Summary)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Location != tt.want.Location {
				t.Errorf("Location = %q, want %q", got.Location, tt.want.Location)
			}
			if got.Start != tt.want.Start {
				t.Errorf("Start = %q, want %q", got.Start, tt.want.Start)
			}
			if got.End != tt.want.End {
				t.Errorf("End = %q, want %q", got.End, tt.want.End)
			}
			if got.Status != tt.want.Status {
				t.Errorf("Status = %q, want %q", got.Status, tt.want.Status)
			}
			if got.Creator != tt.want.Creator {
				t.Errorf("Creator = %q, want %q", got.Creator, tt.want.Creator)
			}
			if got.Organizer != tt.want.Organizer {
				t.Errorf("Organizer = %q, want %q", got.Organizer, tt.want.Organizer)
			}
			if got.HangoutLink != tt.want.HangoutLink {
				t.Errorf("HangoutLink = %q, want %q", got.HangoutLink, tt.want.HangoutLink)
			}
			if got.HtmlLink != tt.want.HtmlLink {
				t.Errorf("HtmlLink = %q, want %q", got.HtmlLink, tt.want.HtmlLink)
			}
			if len(got.Recurrence) != len(tt.want.Recurrence) {
				t.Errorf("Recurrence length = %d, want %d", len(got.Recurrence), len(tt.want.Recurrence))
			} else {
				for i := range got.Recurrence {
					if got.Recurrence[i] != tt.want.Recurrence[i] {
						t.Errorf("Recurrence[%d] = %q, want %q", i, got.Recurrence[i], tt.want.Recurrence[i])
					}
				}
			}
			if len(got.Attendees) != len(tt.want.Attendees) {
				t.Errorf("Attendees length = %d, want %d", len(got.Attendees), len(tt.want.Attendees))
			} else {
				for i := range got.Attendees {
					if got.Attendees[i].Email != tt.want.Attendees[i].Email {
						t.Errorf("Attendees[%d].Email = %q, want %q", i, got.Attendees[i].Email, tt.want.Attendees[i].Email)
					}
					if got.Attendees[i].ResponseStatus != tt.want.Attendees[i].ResponseStatus {
						t.Errorf("Attendees[%d].ResponseStatus = %q, want %q", i, got.Attendees[i].ResponseStatus, tt.want.Attendees[i].ResponseStatus)
					}
					if got.Attendees[i].Self != tt.want.Attendees[i].Self {
						t.Errorf("Attendees[%d].Self = %v, want %v", i, got.Attendees[i].Self, tt.want.Attendees[i].Self)
					}
				}
			}
		})
	}
}

func TestConvertAttendees(t *testing.T) {
	tests := []struct {
		name string
		in   []*calendarapi.EventAttendee
		want []types.CalendarAttendee
	}{
		{
			name: "nil attendees",
			in:   nil,
			want: []types.CalendarAttendee{},
		},
		{
			name: "empty attendees",
			in:   []*calendarapi.EventAttendee{},
			want: []types.CalendarAttendee{},
		},
		{
			name: "single attendee with all fields",
			in: []*calendarapi.EventAttendee{
				{
					Email:          "alice@example.com",
					DisplayName:    "Alice",
					ResponseStatus: "accepted",
					Self:           true,
					Optional:       false,
				},
			},
			want: []types.CalendarAttendee{
				{
					Email:          "alice@example.com",
					DisplayName:    "Alice",
					ResponseStatus: "accepted",
					Self:           true,
					Optional:       false,
				},
			},
		},
		{
			name: "multiple attendees",
			in: []*calendarapi.EventAttendee{
				{Email: "alice@example.com", ResponseStatus: "accepted"},
				{Email: "bob@example.com", ResponseStatus: "tentative", Optional: true},
				{Email: "carol@example.com", DisplayName: "Carol", ResponseStatus: "declined"},
			},
			want: []types.CalendarAttendee{
				{Email: "alice@example.com", ResponseStatus: "accepted"},
				{Email: "bob@example.com", ResponseStatus: "tentative", Optional: true},
				{Email: "carol@example.com", DisplayName: "Carol", ResponseStatus: "declined"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertAttendees(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Email != tt.want[i].Email {
					t.Errorf("[%d].Email = %q, want %q", i, got[i].Email, tt.want[i].Email)
				}
				if got[i].DisplayName != tt.want[i].DisplayName {
					t.Errorf("[%d].DisplayName = %q, want %q", i, got[i].DisplayName, tt.want[i].DisplayName)
				}
				if got[i].ResponseStatus != tt.want[i].ResponseStatus {
					t.Errorf("[%d].ResponseStatus = %q, want %q", i, got[i].ResponseStatus, tt.want[i].ResponseStatus)
				}
				if got[i].Self != tt.want[i].Self {
					t.Errorf("[%d].Self = %v, want %v", i, got[i].Self, tt.want[i].Self)
				}
				if got[i].Optional != tt.want[i].Optional {
					t.Errorf("[%d].Optional = %v, want %v", i, got[i].Optional, tt.want[i].Optional)
				}
			}
		})
	}
}

func TestDetectConflicts(t *testing.T) {
	tests := []struct {
		name       string
		events     []types.CalendarEvent
		wantCount  int
		wantFirst  string
		wantOverlap int
	}{
		{
			name:      "no events",
			events:    nil,
			wantCount: 0,
		},
		{
			name: "single event — no conflicts",
			events: []types.CalendarEvent{
				{ID: "e1", Summary: "Meeting", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
			},
			wantCount: 0,
		},
		{
			name: "non-overlapping events",
			events: []types.CalendarEvent{
				{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
				{ID: "e2", Summary: "Second", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T11:00:00Z"},
			},
			wantCount: 0,
		},
		{
			name: "two overlapping events",
			events: []types.CalendarEvent{
				{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:30:00Z"},
				{ID: "e2", Summary: "Second", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T11:00:00Z"},
			},
			wantCount:   1,
			wantFirst:   "e1",
			wantOverlap: 1,
		},
		{
			name: "three-way overlap",
			events: []types.CalendarEvent{
				{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T11:00:00Z"},
				{ID: "e2", Summary: "Second", Start: "2026-03-15T09:30:00Z", End: "2026-03-15T10:30:00Z"},
				{ID: "e3", Summary: "Third", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T10:45:00Z"},
			},
			wantCount:   2,
			wantFirst:   "e1",
			wantOverlap: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectConflicts(tt.events)
			if len(got) != tt.wantCount {
				t.Fatalf("conflict count = %d, want %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if got[0].Event.ID != tt.wantFirst {
					t.Errorf("first conflict event ID = %q, want %q", got[0].Event.ID, tt.wantFirst)
				}
				if len(got[0].ConflictsWith) != tt.wantOverlap {
					t.Errorf("first conflict overlaps = %d, want %d", len(got[0].ConflictsWith), tt.wantOverlap)
				}
			}
		})
	}
}
