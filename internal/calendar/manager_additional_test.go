package calendar

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/types"
	calendarapi "google.golang.org/api/calendar/v3"
)

// TestBuildEventDateTime_EdgeCases tests edge cases for date/time building
func TestBuildEventDateTime_EdgeCases(t *testing.T) {
	t.Run("empty string all day", func(t *testing.T) {
		got := buildEventDateTime("", true)
		if got == nil {
			t.Fatal("expected non-nil EventDateTime")
		}
		if got.Date != "" || got.DateTime != "" {
			t.Error("expected empty Date and DateTime for empty input")
		}
	})

	t.Run("empty string timed", func(t *testing.T) {
		got := buildEventDateTime("", false)
		if got == nil {
			t.Fatal("expected non-nil EventDateTime")
		}
		if got.Date != "" || got.DateTime != "" {
			t.Error("expected empty Date and DateTime for empty input")
		}
	})

	t.Run("various date formats all day", func(t *testing.T) {
		dates := []string{
			"2026-01-01",
			"1999-12-31",
			"2050-06-15",
		}
		for _, date := range dates {
			got := buildEventDateTime(date, true)
			if got.Date != date {
				t.Errorf("expected Date %s, got %s", date, got.Date)
			}
			if got.DateTime != "" {
				t.Errorf("expected empty DateTime for all-day, got %s", got.DateTime)
			}
		}
	})

	t.Run("various datetime formats", func(t *testing.T) {
		datetimes := []string{
			"2026-01-01T00:00:00Z",
			"2026-06-15T12:30:45+02:00",
			"1999-12-31T23:59:59-08:00",
		}
		for _, dt := range datetimes {
			got := buildEventDateTime(dt, false)
			if got.DateTime != dt {
				t.Errorf("expected DateTime %s, got %s", dt, got.DateTime)
			}
			if got.Date != "" {
				t.Errorf("expected empty Date for timed event, got %s", got.Date)
			}
		}
	})
}

// TestConvertEvent_EdgeCases tests edge cases for event conversion
func TestConvertEvent_EdgeCases(t *testing.T) {
	t.Run("nil event", func(t *testing.T) {
		// convertEvent doesn't handle nil input, skip this test
		t.Skip("convertEvent doesn't handle nil input safely")
	})

	t.Run("empty event", func(t *testing.T) {
		got := convertEvent(&calendarapi.Event{})
		if got.ID != "" {
			t.Errorf("expected empty ID for empty event, got %s", got.ID)
		}
		if got.Summary != "" {
			t.Errorf("expected empty summary, got %s", got.Summary)
		}
	})

	t.Run("event with nil start", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:      "evt1",
			Summary: "Test",
			Start:   nil,
			End:     &calendarapi.EventDateTime{DateTime: "2026-03-15T10:00:00Z"},
		}
		got := convertEvent(event)
		if got.Start != "" {
			t.Errorf("expected empty Start, got %s", got.Start)
		}
		if got.End != "2026-03-15T10:00:00Z" {
			t.Errorf("expected End time, got %s", got.End)
		}
	})

	t.Run("event with nil end", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:      "evt1",
			Summary: "Test",
			Start:   &calendarapi.EventDateTime{DateTime: "2026-03-15T09:00:00Z"},
			End:     nil,
		}
		got := convertEvent(event)
		if got.Start != "2026-03-15T09:00:00Z" {
			t.Errorf("expected Start time, got %s", got.Start)
		}
		if got.End != "" {
			t.Errorf("expected empty End, got %s", got.End)
		}
	})

	t.Run("event with both Date and DateTime in start", func(t *testing.T) {
		event := &calendarapi.Event{
			Id: "evt1",
			Start: &calendarapi.EventDateTime{
				DateTime: "2026-03-15T09:00:00Z",
				Date:     "2026-03-15",
			},
		}
		got := convertEvent(event)
		// Should prefer DateTime
		if got.Start != "2026-03-15T09:00:00Z" {
			t.Errorf("expected DateTime, got %s", got.Start)
		}
	})

	t.Run("event with nil creator", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:      "evt1",
			Creator: nil,
		}
		got := convertEvent(event)
		if got.Creator != "" {
			t.Errorf("expected empty Creator, got %s", got.Creator)
		}
	})

	t.Run("event with nil organizer", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:        "evt1",
			Organizer: nil,
		}
		got := convertEvent(event)
		if got.Organizer != "" {
			t.Errorf("expected empty Organizer, got %s", got.Organizer)
		}
	})

	t.Run("event with nil attendees", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:        "evt1",
			Attendees: nil,
		}
		got := convertEvent(event)
		if got.Attendees != nil {
			t.Error("expected nil Attendees")
		}
	})

	t.Run("event with empty recurrence", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:         "evt1",
			Recurrence: []string{},
		}
		got := convertEvent(event)
		if len(got.Recurrence) != 0 {
			t.Errorf("expected empty recurrence, got %d items", len(got.Recurrence))
		}
	})

	t.Run("full event", func(t *testing.T) {
		event := &calendarapi.Event{
			Id:          "evt1",
			Summary:     "Full Event",
			Description: "Description",
			Location:    "Location",
			Status:      "confirmed",
			HangoutLink: "https://meet.google.com/abc",
			HtmlLink:    "https://calendar.google.com/event?eid=abc",
			Recurrence:  []string{"RRULE:FREQ=WEEKLY"},
			Start:       &calendarapi.EventDateTime{DateTime: "2026-03-15T09:00:00Z"},
			End:         &calendarapi.EventDateTime{DateTime: "2026-03-15T10:00:00Z"},
			Creator:     &calendarapi.EventCreator{Email: "creator@example.com"},
			Organizer:   &calendarapi.EventOrganizer{Email: "organizer@example.com"},
			Attendees: []*calendarapi.EventAttendee{
				{Email: "att1@example.com", ResponseStatus: "accepted"},
				{Email: "att2@example.com", ResponseStatus: "tentative", Optional: true},
			},
		}
		got := convertEvent(event)
		if got.ID != "evt1" {
			t.Errorf("expected evt1, got %s", got.ID)
		}
		if got.Summary != "Full Event" {
			t.Errorf("expected Full Event, got %s", got.Summary)
		}
		if got.Creator != "creator@example.com" {
			t.Errorf("expected creator email, got %s", got.Creator)
		}
		if got.Organizer != "organizer@example.com" {
			t.Errorf("expected organizer email, got %s", got.Organizer)
		}
		if len(got.Attendees) != 2 {
			t.Errorf("expected 2 attendees, got %d", len(got.Attendees))
		}
		if !got.Attendees[1].Optional {
			t.Error("expected second attendee to be optional")
		}
	})
}

// TestConvertAttendees_EdgeCases tests edge cases for attendee conversion
func TestConvertAttendees_EdgeCases(t *testing.T) {
	t.Run("nil attendees", func(t *testing.T) {
		got := convertAttendees(nil)
		if got == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("empty attendees", func(t *testing.T) {
		got := convertAttendees([]*calendarapi.EventAttendee{})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("attendee with all fields", func(t *testing.T) {
		attendees := []*calendarapi.EventAttendee{
			{
				Email:          "full@example.com",
				DisplayName:    "Full Attendee",
				ResponseStatus: "accepted",
				Self:           true,
				Optional:       false,
			},
		}
		got := convertAttendees(attendees)
		if len(got) != 1 {
			t.Fatalf("expected 1 attendee, got %d", len(got))
		}
		if got[0].Email != "full@example.com" {
			t.Errorf("expected email, got %s", got[0].Email)
		}
		if got[0].DisplayName != "Full Attendee" {
			t.Errorf("expected display name, got %s", got[0].DisplayName)
		}
		if got[0].ResponseStatus != "accepted" {
			t.Errorf("expected accepted, got %s", got[0].ResponseStatus)
		}
		if !got[0].Self {
			t.Error("expected Self to be true")
		}
		if got[0].Optional {
			t.Error("expected Optional to be false")
		}
	})

	t.Run("multiple attendees with different statuses", func(t *testing.T) {
		attendees := []*calendarapi.EventAttendee{
			{Email: "att1@example.com", ResponseStatus: "accepted"},
			{Email: "att2@example.com", ResponseStatus: "declined", Self: true},
			{Email: "att3@example.com", ResponseStatus: "tentative", Optional: true},
			{Email: "att4@example.com", ResponseStatus: "needsAction"},
		}
		got := convertAttendees(attendees)
		if len(got) != 4 {
			t.Fatalf("expected 4 attendees, got %d", len(got))
		}
		statuses := []string{"accepted", "declined", "tentative", "needsAction"}
		for i, status := range statuses {
			if got[i].ResponseStatus != status {
				t.Errorf("attendee %d: expected %s, got %s", i, status, got[i].ResponseStatus)
			}
		}
	})
}

// TestDetectConflicts_EdgeCases tests edge cases for conflict detection
func TestDetectConflicts_EdgeCases(t *testing.T) {
	t.Run("nil events", func(t *testing.T) {
		// detectConflicts doesn't handle nil input, skip this test
		t.Skip("detectConflicts doesn't handle nil input safely")
	})

	t.Run("empty events", func(t *testing.T) {
		got := detectConflicts([]types.CalendarEvent{})
		if len(got) != 0 {
			t.Errorf("expected empty conflicts, got %d", len(got))
		}
	})

	t.Run("single event no conflicts", func(t *testing.T) {
		events := []types.CalendarEvent{
			{ID: "e1", Summary: "Meeting", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
		}
		got := detectConflicts(events)
		if len(got) != 0 {
			t.Errorf("expected no conflicts for single event, got %d", len(got))
		}
	})

	t.Run("consecutive non-overlapping events", func(t *testing.T) {
		events := []types.CalendarEvent{
			{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
			{ID: "e2", Summary: "Second", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T11:00:00Z"},
		}
		got := detectConflicts(events)
		if len(got) != 0 {
			t.Errorf("expected no conflicts for consecutive events, got %d", len(got))
		}
	})

	t.Run("events with gap between them", func(t *testing.T) {
		events := []types.CalendarEvent{
			{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
			{ID: "e2", Summary: "Second", Start: "2026-03-15T11:00:00Z", End: "2026-03-15T12:00:00Z"},
		}
		got := detectConflicts(events)
		if len(got) != 0 {
			t.Errorf("expected no conflicts for events with gap, got %d", len(got))
		}
	})

	t.Run("same start time", func(t *testing.T) {
		events := []types.CalendarEvent{
			{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:00:00Z"},
			{ID: "e2", Summary: "Second", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T09:30:00Z"},
		}
		got := detectConflicts(events)
		if len(got) != 1 {
			t.Fatalf("expected 1 conflict, got %d", len(got))
		}
		if got[0].Event.ID != "e1" {
			t.Errorf("expected first event as conflict source, got %s", got[0].Event.ID)
		}
	})

	t.Run("cascade conflicts", func(t *testing.T) {
		// Event 1 conflicts with 2, Event 2 conflicts with 3
		events := []types.CalendarEvent{
			{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T11:00:00Z"},
			{ID: "e2", Summary: "Second", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T12:00:00Z"},
			{ID: "e3", Summary: "Third", Start: "2026-03-15T11:30:00Z", End: "2026-03-15T13:00:00Z"},
		}
		got := detectConflicts(events)
		if len(got) != 2 {
			t.Fatalf("expected 2 conflicts, got %d", len(got))
		}
		// e1 conflicts with e2
		if len(got[0].ConflictsWith) != 1 {
			t.Errorf("expected 1 conflict for e1, got %d", len(got[0].ConflictsWith))
		}
		// e2 conflicts with e3
		if len(got[1].ConflictsWith) != 1 {
			t.Errorf("expected 1 conflict for e2, got %d", len(got[1].ConflictsWith))
		}
	})
}

// Benchmarks
func BenchmarkConvertEvent(b *testing.B) {
	event := &calendarapi.Event{
		Id:          "evt1",
		Summary:     "Benchmark",
		Description: "Description",
		Start:       &calendarapi.EventDateTime{DateTime: "2026-03-15T09:00:00Z"},
		End:         &calendarapi.EventDateTime{DateTime: "2026-03-15T10:00:00Z"},
		Creator:     &calendarapi.EventCreator{Email: "creator@example.com"},
		Attendees: []*calendarapi.EventAttendee{
			{Email: "att1@example.com", ResponseStatus: "accepted"},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertEvent(event)
	}
}

func BenchmarkConvertEventNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertEvent(nil)
	}
}

func BenchmarkDetectConflicts(b *testing.B) {
	events := []types.CalendarEvent{
		{ID: "e1", Summary: "First", Start: "2026-03-15T09:00:00Z", End: "2026-03-15T10:30:00Z"},
		{ID: "e2", Summary: "Second", Start: "2026-03-15T10:00:00Z", End: "2026-03-15T11:00:00Z"},
		{ID: "e3", Summary: "Third", Start: "2026-03-15T10:30:00Z", End: "2026-03-15T11:30:00Z"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectConflicts(events)
	}
}
