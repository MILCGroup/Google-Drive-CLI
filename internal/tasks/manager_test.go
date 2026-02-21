package tasks

import (
	"testing"

	"github.com/dl-alexandre/gdrv/internal/types"
	taskapi "google.golang.org/api/tasks/v1"
)

func TestConvertTaskList(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := convertTaskList(nil)
		if got.ID != "" || got.Title != "" || got.Updated != "" {
			t.Fatalf("expected zero-value TaskList, got %+v", got)
		}
	})

	t.Run("fully populated", func(t *testing.T) {
		tl := &taskapi.TaskList{
			Id:      "list-abc123",
			Title:   "Work Tasks",
			Updated: "2026-02-20T10:30:00.000Z",
		}
		got := convertTaskList(tl)
		if got.ID != "list-abc123" {
			t.Fatalf("expected ID %q, got %q", "list-abc123", got.ID)
		}
		if got.Title != "Work Tasks" {
			t.Fatalf("expected Title %q, got %q", "Work Tasks", got.Title)
		}
		if got.Updated != "2026-02-20T10:30:00.000Z" {
			t.Fatalf("expected Updated %q, got %q", "2026-02-20T10:30:00.000Z", got.Updated)
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		tl := &taskapi.TaskList{
			Id: "list-empty",
		}
		got := convertTaskList(tl)
		if got.ID != "list-empty" {
			t.Fatalf("expected ID %q, got %q", "list-empty", got.ID)
		}
		if got.Title != "" {
			t.Fatalf("expected empty Title, got %q", got.Title)
		}
		if got.Updated != "" {
			t.Fatalf("expected empty Updated, got %q", got.Updated)
		}
	})
}

func TestConvertTask(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := convertTask(nil)
		if got.ID != "" || got.Title != "" || got.Status != "" {
			t.Fatalf("expected zero-value Task, got %+v", got)
		}
	})

	t.Run("minimal task", func(t *testing.T) {
		task := &taskapi.Task{
			Id:     "task-001",
			Title:  "Buy groceries",
			Status: "needsAction",
		}
		got := convertTask(task)
		if got.ID != "task-001" {
			t.Fatalf("expected ID %q, got %q", "task-001", got.ID)
		}
		if got.Title != "Buy groceries" {
			t.Fatalf("expected Title %q, got %q", "Buy groceries", got.Title)
		}
		if got.Status != "needsAction" {
			t.Fatalf("expected Status %q, got %q", "needsAction", got.Status)
		}
		if got.Notes != "" {
			t.Fatalf("expected empty Notes, got %q", got.Notes)
		}
		if got.Due != "" {
			t.Fatalf("expected empty Due, got %q", got.Due)
		}
		if got.Completed != "" {
			t.Fatalf("expected empty Completed, got %q", got.Completed)
		}
		if got.Parent != "" {
			t.Fatalf("expected empty Parent, got %q", got.Parent)
		}
		if got.Position != "" {
			t.Fatalf("expected empty Position, got %q", got.Position)
		}
		if len(got.Links) != 0 {
			t.Fatalf("expected no links, got %d", len(got.Links))
		}
	})

	t.Run("fully populated task", func(t *testing.T) {
		completedTime := "2026-02-20T14:00:00.000Z"
		task := &taskapi.Task{
			Id:        "task-002",
			Title:     "Review PR",
			Notes:     "Check the test coverage",
			Status:    "completed",
			Due:       "2026-02-25T00:00:00.000Z",
			Completed: &completedTime,
			Parent:    "task-parent",
			Position:  "00000000000000000001",
			Links: []*taskapi.TaskLinks{
				{
					Type:        "email",
					Description: "Related email",
					Link:        "https://mail.google.com/123",
				},
				{
					Type:        "related",
					Description: "PR link",
					Link:        "https://github.com/example/pr/1",
				},
			},
		}
		got := convertTask(task)
		if got.ID != "task-002" {
			t.Fatalf("expected ID %q, got %q", "task-002", got.ID)
		}
		if got.Title != "Review PR" {
			t.Fatalf("expected Title %q, got %q", "Review PR", got.Title)
		}
		if got.Notes != "Check the test coverage" {
			t.Fatalf("expected Notes %q, got %q", "Check the test coverage", got.Notes)
		}
		if got.Status != "completed" {
			t.Fatalf("expected Status %q, got %q", "completed", got.Status)
		}
		if got.Due != "2026-02-25T00:00:00.000Z" {
			t.Fatalf("expected Due %q, got %q", "2026-02-25T00:00:00.000Z", got.Due)
		}
		if got.Completed != "2026-02-20T14:00:00.000Z" {
			t.Fatalf("expected Completed %q, got %q", "2026-02-20T14:00:00.000Z", got.Completed)
		}
		if got.Parent != "task-parent" {
			t.Fatalf("expected Parent %q, got %q", "task-parent", got.Parent)
		}
		if got.Position != "00000000000000000001" {
			t.Fatalf("expected Position %q, got %q", "00000000000000000001", got.Position)
		}
		if len(got.Links) != 2 {
			t.Fatalf("expected 2 links, got %d", len(got.Links))
		}
		if got.Links[0].Type != "email" {
			t.Fatalf("expected link[0] Type %q, got %q", "email", got.Links[0].Type)
		}
		if got.Links[0].Description != "Related email" {
			t.Fatalf("expected link[0] Description %q, got %q", "Related email", got.Links[0].Description)
		}
		if got.Links[0].Link != "https://mail.google.com/123" {
			t.Fatalf("expected link[0] Link %q, got %q", "https://mail.google.com/123", got.Links[0].Link)
		}
		if got.Links[1].Type != "related" {
			t.Fatalf("expected link[1] Type %q, got %q", "related", got.Links[1].Type)
		}
		if got.Links[1].Description != "PR link" {
			t.Fatalf("expected link[1] Description %q, got %q", "PR link", got.Links[1].Description)
		}
		if got.Links[1].Link != "https://github.com/example/pr/1" {
			t.Fatalf("expected link[1] Link %q, got %q", "https://github.com/example/pr/1", got.Links[1].Link)
		}
	})

	t.Run("task with no links", func(t *testing.T) {
		task := &taskapi.Task{
			Id:     "task-003",
			Title:  "No links",
			Status: "needsAction",
			Links:  nil,
		}
		got := convertTask(task)
		if got.Links != nil {
			t.Fatalf("expected nil links, got %v", got.Links)
		}
	})

	t.Run("task with empty links slice", func(t *testing.T) {
		task := &taskapi.Task{
			Id:     "task-004",
			Title:  "Empty links",
			Status: "needsAction",
			Links:  []*taskapi.TaskLinks{},
		}
		got := convertTask(task)
		if got.Links != nil {
			t.Fatalf("expected nil links for empty slice, got %v", got.Links)
		}
	})

	t.Run("due date preserved as-is", func(t *testing.T) {
		task := &taskapi.Task{
			Id:     "task-005",
			Title:  "Due date test",
			Status: "needsAction",
			Due:    "2026-12-31T00:00:00.000Z",
		}
		got := convertTask(task)
		if got.Due != "2026-12-31T00:00:00.000Z" {
			t.Fatalf("expected Due to be preserved as-is, got %q", got.Due)
		}
	})
}

func TestFormatDueDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard date",
			input:    "2026-02-20",
			expected: "2026-02-20T00:00:00.000Z",
		},
		{
			name:     "end of year",
			input:    "2026-12-31",
			expected: "2026-12-31T00:00:00.000Z",
		},
		{
			name:     "beginning of year",
			input:    "2026-01-01",
			expected: "2026-01-01T00:00:00.000Z",
		},
		{
			name:     "leap year date",
			input:    "2028-02-29",
			expected: "2028-02-29T00:00:00.000Z",
		},
		{
			name:     "arbitrary date",
			input:    "2025-06-15",
			expected: "2025-06-15T00:00:00.000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDueDate(tt.input)
			if got != tt.expected {
				t.Fatalf("formatDueDate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestConvertTaskListFieldMapping(t *testing.T) {
	// Verify that convertTaskList maps all fields correctly from API type to internal type
	tl := &taskapi.TaskList{
		Id:      "id-field",
		Title:   "title-field",
		Updated: "updated-field",
	}
	got := convertTaskList(tl)

	fields := map[string]struct{ got, want string }{
		"ID":      {got.ID, tl.Id},
		"Title":   {got.Title, tl.Title},
		"Updated": {got.Updated, tl.Updated},
	}
	for name, v := range fields {
		if v.got != v.want {
			t.Errorf("field %s: got %q, want %q", name, v.got, v.want)
		}
	}
}

func TestConvertTaskFieldMapping(t *testing.T) {
	// Verify that convertTask maps all scalar fields correctly
	completedVal := "completed-field"
	task := &taskapi.Task{
		Id:        "id-field",
		Title:     "title-field",
		Notes:     "notes-field",
		Status:    "status-field",
		Due:       "due-field",
		Completed: &completedVal,
		Parent:    "parent-field",
		Position:  "position-field",
	}
	got := convertTask(task)

	fields := map[string]struct{ got, want string }{
		"ID":        {got.ID, task.Id},
		"Title":     {got.Title, task.Title},
		"Notes":     {got.Notes, task.Notes},
		"Status":    {got.Status, task.Status},
		"Due":       {got.Due, task.Due},
		"Completed": {got.Completed, *task.Completed},
		"Parent":    {got.Parent, task.Parent},
		"Position":  {got.Position, task.Position},
	}
	for name, v := range fields {
		if v.got != v.want {
			t.Errorf("field %s: got %q, want %q", name, v.got, v.want)
		}
	}
}

// TestNewManager verifies the constructor wires dependencies correctly.
func TestNewManager(t *testing.T) {
	m := NewManager(nil, nil)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.client != nil {
		t.Fatal("expected nil client")
	}
	if m.service != nil {
		t.Fatal("expected nil service")
	}
}

// TestTaskTypesInterface ensures the types implement expected interface contracts
// by verifying Headers/Rows/EmptyMessage return sensible values.
func TestTaskTypesInterface(t *testing.T) {
	t.Run("TaskList", func(t *testing.T) {
		tl := &types.TaskList{ID: "id", Title: "title", Updated: "updated"}
		headers := tl.Headers()
		if len(headers) != 3 {
			t.Fatalf("expected 3 headers, got %d", len(headers))
		}
		rows := tl.Rows()
		if len(rows) != 1 || len(rows[0]) != 3 {
			t.Fatalf("expected 1 row with 3 cols, got %d rows", len(rows))
		}
		if tl.EmptyMessage() == "" {
			t.Fatal("expected non-empty EmptyMessage")
		}
	})

	t.Run("TaskListResult", func(t *testing.T) {
		r := &types.TaskListResult{TaskLists: []types.TaskList{
			{ID: "a", Title: "A", Updated: "u1"},
			{ID: "b", Title: "B", Updated: "u2"},
		}}
		headers := r.Headers()
		if len(headers) != 3 {
			t.Fatalf("expected 3 headers, got %d", len(headers))
		}
		rows := r.Rows()
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		if r.EmptyMessage() == "" {
			t.Fatal("expected non-empty EmptyMessage")
		}
	})

	t.Run("Task", func(t *testing.T) {
		task := &types.Task{ID: "id", Title: "title", Status: "needsAction", Due: "2026-01-01"}
		headers := task.Headers()
		if len(headers) != 4 {
			t.Fatalf("expected 4 headers, got %d", len(headers))
		}
		rows := task.Rows()
		if len(rows) != 1 || len(rows[0]) != 4 {
			t.Fatalf("expected 1 row with 4 cols, got %d rows", len(rows))
		}
		if task.EmptyMessage() == "" {
			t.Fatal("expected non-empty EmptyMessage")
		}
	})

	t.Run("TaskResult", func(t *testing.T) {
		r := &types.TaskResult{Tasks: []types.Task{
			{ID: "t1", Title: "T1", Status: "needsAction", Due: ""},
		}}
		rows := r.Rows()
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		if r.EmptyMessage() == "" {
			t.Fatal("expected non-empty EmptyMessage")
		}
	})

	t.Run("TaskMutationResult", func(t *testing.T) {
		r := &types.TaskMutationResult{ID: "id", Title: "title", Status: "completed"}
		headers := r.Headers()
		if len(headers) != 3 {
			t.Fatalf("expected 3 headers, got %d", len(headers))
		}
		rows := r.Rows()
		if len(rows) != 1 || len(rows[0]) != 3 {
			t.Fatalf("expected 1 row with 3 cols, got %d rows", len(rows))
		}
		if r.EmptyMessage() == "" {
			t.Fatal("expected non-empty EmptyMessage")
		}
	})
}
