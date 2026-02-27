// Package tasks provides Google Tasks API management functionality.
// It handles task lists and tasks with full CRUD operations.
//
// Rate limits: The Google Tasks API has a quota of 50,000 requests per day.
// There is no batch API available for Tasks; each operation is a single request.
package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	taskapi "google.golang.org/api/tasks/v1"
)

// Manager handles Google Tasks API operations
type Manager struct {
	client  *api.Client
	service *taskapi.Service
}

// NewManager creates a new Tasks manager
func NewManager(client *api.Client, service *taskapi.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// ListTaskLists retrieves all task lists for the authenticated user.
func (m *Manager) ListTaskLists(ctx context.Context, reqCtx *types.RequestContext, maxResults int64, pageToken string) (*types.TaskListResult, string, error) {
	call := m.service.Tasklists.List()
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.TaskLists, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	taskLists := make([]types.TaskList, 0, len(result.Items))
	for _, tl := range result.Items {
		taskLists = append(taskLists, convertTaskList(tl))
	}

	return &types.TaskListResult{TaskLists: taskLists}, result.NextPageToken, nil
}

// CreateTaskList creates a new task list with the given title.
func (m *Manager) CreateTaskList(ctx context.Context, reqCtx *types.RequestContext, title string) (*types.TaskList, error) {
	tl := &taskapi.TaskList{
		Title: title,
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.TaskList, error) {
		return m.service.Tasklists.Insert(tl).Do()
	})
	if err != nil {
		return nil, err
	}

	converted := convertTaskList(result)
	return &converted, nil
}

// ListTasks retrieves tasks from a specific task list.
// showCompleted defaults to false; set to true to include completed tasks.
func (m *Manager) ListTasks(ctx context.Context, reqCtx *types.RequestContext, taskListID string, maxResults int64, pageToken string, showCompleted, showHidden bool, dueMin, dueMax string) (*types.TaskResult, string, error) {
	call := m.service.Tasks.List(taskListID)
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	call = call.ShowCompleted(showCompleted)
	call = call.ShowHidden(showHidden)
	if dueMin != "" {
		call = call.DueMin(dueMin)
	}
	if dueMax != "" {
		call = call.DueMax(dueMax)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Tasks, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	tasks := make([]types.Task, 0, len(result.Items))
	for _, t := range result.Items {
		tasks = append(tasks, convertTask(t))
	}

	return &types.TaskResult{Tasks: tasks}, result.NextPageToken, nil
}

// GetTask retrieves a single task by ID.
func (m *Manager) GetTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, taskID string) (*types.Task, error) {
	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Get(taskListID, taskID).Do()
	})
	if err != nil {
		return nil, err
	}

	converted := convertTask(result)
	return &converted, nil
}

// AddTask creates a new task in the specified task list.
// The due parameter is a date-only string in "YYYY-MM-DD" format; it is formatted
// as RFC 3339 with zero time ("YYYY-MM-DDT00:00:00.000Z") because the Tasks API
// ignores the time portion.
// The parent and previous parameters control task hierarchy and ordering.
func (m *Manager) AddTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, title, notes, due, parent, previous string) (*types.TaskMutationResult, error) {
	task := &taskapi.Task{
		Title: title,
	}
	if notes != "" {
		task.Notes = notes
	}
	if due != "" {
		task.Due = formatDueDate(due)
	}

	call := m.service.Tasks.Insert(taskListID, task)
	if parent != "" {
		call = call.Parent(parent)
	}
	if previous != "" {
		call = call.Previous(previous)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.TaskMutationResult{
		ID:     result.Id,
		Title:  result.Title,
		Status: result.Status,
	}, nil
}

// UpdateTask updates an existing task's title, notes, and/or due date.
// Only non-nil string pointer fields are patched; nil fields are left unchanged.
// The task is fetched first, then fields are patched and the full task is sent back.
func (m *Manager) UpdateTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, taskID string, title, notes, due *string) (*types.TaskMutationResult, error) {
	// Get the existing task first
	existing, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Get(taskListID, taskID).Do()
	})
	if err != nil {
		return nil, err
	}

	// Patch fields
	if title != nil {
		existing.Title = *title
	}
	if notes != nil {
		existing.Notes = *notes
	}
	if due != nil {
		if *due == "" {
			existing.Due = ""
		} else {
			existing.Due = formatDueDate(*due)
		}
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Update(taskListID, taskID, existing).Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.TaskMutationResult{
		ID:     result.Id,
		Title:  result.Title,
		Status: result.Status,
	}, nil
}

// CompleteTask marks a task as completed by setting its status to "completed"
// and recording the current time as the completion timestamp.
func (m *Manager) CompleteTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, taskID string) (*types.TaskMutationResult, error) {
	// Get the existing task
	existing, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Get(taskListID, taskID).Do()
	})
	if err != nil {
		return nil, err
	}

	existing.Status = "completed"
	completedTime := time.Now().UTC().Format(time.RFC3339)
	existing.Completed = &completedTime

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Update(taskListID, taskID, existing).Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.TaskMutationResult{
		ID:     result.Id,
		Title:  result.Title,
		Status: result.Status,
	}, nil
}

// UncompleteTask marks a task as not completed by setting its status to "needsAction"
// and clearing the completion timestamp.
func (m *Manager) UncompleteTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, taskID string) (*types.TaskMutationResult, error) {
	// Get the existing task
	existing, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Get(taskListID, taskID).Do()
	})
	if err != nil {
		return nil, err
	}

	existing.Status = "needsAction"
	existing.Completed = nil
	// Use NullFields to ensure the null Completed field is sent to the API
	// so the server clears the completion timestamp.
	existing.NullFields = append(existing.NullFields, "Completed")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*taskapi.Task, error) {
		return m.service.Tasks.Update(taskListID, taskID, existing).Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.TaskMutationResult{
		ID:     result.Id,
		Title:  result.Title,
		Status: result.Status,
	}, nil
}

// DeleteTask permanently removes a task from a task list.
func (m *Manager) DeleteTask(ctx context.Context, reqCtx *types.RequestContext, taskListID, taskID string) error {
	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (interface{}, error) {
		return nil, m.service.Tasks.Delete(taskListID, taskID).Do()
	})
	return err
}

// ClearCompleted removes all completed tasks from the specified task list.
func (m *Manager) ClearCompleted(ctx context.Context, reqCtx *types.RequestContext, taskListID string) error {
	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (interface{}, error) {
		return nil, m.service.Tasks.Clear(taskListID).Do()
	})
	return err
}

// convertTaskList converts a Tasks API TaskList to the internal types.TaskList.
func convertTaskList(tl *taskapi.TaskList) types.TaskList {
	if tl == nil {
		return types.TaskList{}
	}
	return types.TaskList{
		ID:      tl.Id,
		Title:   tl.Title,
		Updated: tl.Updated,
	}
}

// convertTask converts a Tasks API Task to the internal types.Task.
// The Due field is kept as-is from the API (already an RFC 3339 string).
func convertTask(t *taskapi.Task) types.Task {
	if t == nil {
		return types.Task{}
	}

	completed := ""
	if t.Completed != nil {
		completed = *t.Completed
	}

	task := types.Task{
		ID:        t.Id,
		Title:     t.Title,
		Notes:     t.Notes,
		Status:    t.Status,
		Due:       t.Due,
		Completed: completed,
		Parent:    t.Parent,
		Position:  t.Position,
	}

	if len(t.Links) > 0 {
		task.Links = make([]types.TaskLink, len(t.Links))
		for i, link := range t.Links {
			task.Links[i] = types.TaskLink{
				Type:        link.Type,
				Description: link.Description,
				Link:        link.Link,
			}
		}
	}

	return task
}

// formatDueDate converts a date-only string "YYYY-MM-DD" to RFC 3339 format
// with zero time: "YYYY-MM-DDT00:00:00.000Z".
// The Tasks API ignores the time portion of due dates, so we always use midnight UTC.
func formatDueDate(dateStr string) string {
	return fmt.Sprintf("%sT00:00:00.000Z", dateStr)
}
