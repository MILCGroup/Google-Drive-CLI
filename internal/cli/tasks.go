package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	tasksmgr "github.com/dl-alexandre/gdrv/internal/tasks"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

// TasksCmd is the top-level command for Google Tasks operations.
type TasksCmd struct {
	Lists  TasksListsCmd    `cmd:"" help:"Task list operations"`
	List   TasksTaskListCmd `cmd:"" help:"List tasks in a task list"`
	Get    TasksGetCmd      `cmd:"" help:"Get a task by ID"`
	Add    TasksAddCmd      `cmd:"" help:"Add a new task"`
	Update TasksUpdateCmd   `cmd:"" help:"Update an existing task"`
	Done   TasksDoneCmd     `cmd:"" help:"Mark a task as completed"`
	Undo   TasksUndoCmd     `cmd:"" help:"Mark a task as not completed"`
	Delete TasksDeleteCmd   `cmd:"" help:"Delete a task"`
	Clear  TasksClearCmd    `cmd:"" help:"Clear completed tasks from a task list"`
}

// ── Task Lists ──────────────────────────────────────────────────────────

type TasksListsCmd struct {
	List   TasksListsListCmd   `cmd:"" help:"List all task lists"`
	Create TasksListsCreateCmd `cmd:"" help:"Create a new task list"`
}

type TasksListsListCmd struct {
	Limit     int    `help:"Maximum task lists to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type TasksListsCreateCmd struct {
	Title string `help:"Title for the new task list" required:"" name:"title"`
}

// ── Tasks ───────────────────────────────────────────────────────────────

type TasksTaskListCmd struct {
	List          string `help:"Task list ID" default:"@default" name:"list"`
	ShowCompleted bool   `help:"Include completed tasks" name:"show-completed"`
	ShowHidden    bool   `help:"Include hidden tasks" name:"show-hidden"`
	DueMin        string `help:"Minimum due date (RFC 3339 timestamp)" name:"due-min"`
	DueMax        string `help:"Maximum due date (RFC 3339 timestamp)" name:"due-max"`
	Limit         int    `help:"Maximum tasks to return per page" default:"100" name:"limit"`
	PageToken     string `help:"Page token for pagination" name:"page-token"`
	Paginate      bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type TasksGetCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task ID"`
	List   string `help:"Task list ID" default:"@default" name:"list"`
}

type TasksAddCmd struct {
	List     string `help:"Task list ID" default:"@default" name:"list"`
	Title    string `help:"Title for the new task" required:"" name:"title"`
	Notes    string `help:"Notes for the task" name:"notes"`
	Due      string `help:"Due date (YYYY-MM-DD)" name:"due"`
	Parent   string `help:"Parent task ID for creating a subtask" name:"parent"`
	Previous string `help:"Previous sibling task ID for ordering" name:"previous"`
}

type TasksUpdateCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task ID"`
	List   string `help:"Task list ID" default:"@default" name:"list"`
	Title  string `help:"New title" name:"title"`
	Notes  string `help:"New notes" name:"notes"`
	Due    string `help:"New due date (YYYY-MM-DD)" name:"due"`
}

type TasksDoneCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task ID"`
	List   string `help:"Task list ID" default:"@default" name:"list"`
}

type TasksUndoCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task ID"`
	List   string `help:"Task list ID" default:"@default" name:"list"`
}

type TasksDeleteCmd struct {
	TaskID string `arg:"" name:"task-id" help:"Task ID"`
	List   string `help:"Task list ID" default:"@default" name:"list"`
}

type TasksClearCmd struct {
	List string `help:"Task list ID" default:"@default" name:"list"`
}

// ── Run methods ─────────────────────────────────────────────────────────

func (cmd *TasksListsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.lists.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allLists []types.TaskList
		pageToken := ""
		for {
			result, nextToken, err := mgr.ListTaskLists(ctx, reqCtx, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "tasks.lists.list", err)
			}
			allLists = append(allLists, result.TaskLists...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("tasks.lists.list", allLists)
		}
		return out.WriteSuccess("tasks.lists.list", map[string]interface{}{
			"taskLists": allLists,
		})
	}

	result, nextToken, err := mgr.ListTaskLists(ctx, reqCtx, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "tasks.lists.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("tasks.lists.list", result.TaskLists)
	}
	resp := map[string]interface{}{
		"taskLists": result.TaskLists,
	}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("tasks.lists.list", resp)
}

func (cmd *TasksListsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.lists.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateTaskList(ctx, reqCtx, cmd.Title)
	if err != nil {
		return handleCLIError(out, "tasks.lists.create", err)
	}

	return out.WriteSuccess("tasks.lists.create", result)
}

func (cmd *TasksTaskListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allTasks []types.Task
		pageToken := ""
		for {
			result, nextToken, err := mgr.ListTasks(ctx, reqCtx, cmd.List, int64(cmd.Limit), pageToken, cmd.ShowCompleted, cmd.ShowHidden, cmd.DueMin, cmd.DueMax)
			if err != nil {
				return handleCLIError(out, "tasks.list", err)
			}
			allTasks = append(allTasks, result.Tasks...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("tasks.list", allTasks)
		}
		return out.WriteSuccess("tasks.list", map[string]interface{}{
			"tasks": allTasks,
		})
	}

	result, nextToken, err := mgr.ListTasks(ctx, reqCtx, cmd.List, int64(cmd.Limit), cmd.PageToken, cmd.ShowCompleted, cmd.ShowHidden, cmd.DueMin, cmd.DueMax)
	if err != nil {
		return handleCLIError(out, "tasks.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("tasks.list", result.Tasks)
	}
	resp := map[string]interface{}{
		"tasks": result.Tasks,
	}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("tasks.list", resp)
}

func (cmd *TasksGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetTask(ctx, reqCtx, cmd.List, cmd.TaskID)
	if err != nil {
		return handleCLIError(out, "tasks.get", err)
	}

	return out.WriteSuccess("tasks.get", result)
}

func (cmd *TasksAddCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.add", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.AddTask(ctx, reqCtx, cmd.List, cmd.Title, cmd.Notes, cmd.Due, cmd.Parent, cmd.Previous)
	if err != nil {
		return handleCLIError(out, "tasks.add", err)
	}

	return out.WriteSuccess("tasks.add", result)
}

func (cmd *TasksUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	// Only pass non-empty values as pointers; empty strings mean "not specified"
	var titlePtr, notesPtr, duePtr *string
	if cmd.Title != "" {
		titlePtr = &cmd.Title
	}
	if cmd.Notes != "" {
		notesPtr = &cmd.Notes
	}
	if cmd.Due != "" {
		duePtr = &cmd.Due
	}

	result, err := mgr.UpdateTask(ctx, reqCtx, cmd.List, cmd.TaskID, titlePtr, notesPtr, duePtr)
	if err != nil {
		return handleCLIError(out, "tasks.update", err)
	}

	return out.WriteSuccess("tasks.update", result)
}

func (cmd *TasksDoneCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.done", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CompleteTask(ctx, reqCtx, cmd.List, cmd.TaskID)
	if err != nil {
		return handleCLIError(out, "tasks.done", err)
	}

	return out.WriteSuccess("tasks.done", result)
}

func (cmd *TasksUndoCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.undo", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.UncompleteTask(ctx, reqCtx, cmd.List, cmd.TaskID)
	if err != nil {
		return handleCLIError(out, "tasks.undo", err)
	}

	return out.WriteSuccess("tasks.undo", result)
}

func (cmd *TasksDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteTask(ctx, reqCtx, cmd.List, cmd.TaskID); err != nil {
		return handleCLIError(out, "tasks.delete", err)
	}

	return out.WriteSuccess("tasks.delete", map[string]string{
		"taskId": cmd.TaskID,
		"status": "deleted",
	})
}

func (cmd *TasksClearCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	mgr, reqCtx, err := getTasksManager(ctx, flags)
	if err != nil {
		return out.WriteError("tasks.clear", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.ClearCompleted(ctx, reqCtx, cmd.List); err != nil {
		return handleCLIError(out, "tasks.clear", err)
	}

	return out.WriteSuccess("tasks.clear", map[string]string{
		"listId": cmd.List,
		"status": "cleared",
	})
}

// ── Helper ──────────────────────────────────────────────────────────────

func getTasksManager(ctx context.Context, flags types.GlobalFlags) (*tasksmgr.Manager, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceTasks); err != nil {
		return nil, nil, err
	}

	svc, err := authMgr.GetTasksService(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	mgr := tasksmgr.NewManager(client, svc)
	return mgr, reqCtx, nil
}
