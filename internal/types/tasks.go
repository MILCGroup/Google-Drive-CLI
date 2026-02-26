package types

// TaskLink represents a link associated with a task.
type TaskLink struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

// TaskList represents a Google Tasks task list.
type TaskList struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated"`
}

func (t *TaskList) Headers() []string {
	return []string{"ID", "Title", "Updated"}
}

func (t *TaskList) Rows() [][]string {
	return [][]string{{
		t.ID,
		t.Title,
		t.Updated,
	}}
}

func (t *TaskList) EmptyMessage() string {
	return "No task list found"
}

// TaskListResult represents a collection of task lists.
type TaskListResult struct {
	TaskLists []TaskList `json:"taskLists"`
}

func (r *TaskListResult) Headers() []string {
	return []string{"ID", "Title", "Updated"}
}

func (r *TaskListResult) Rows() [][]string {
	rows := make([][]string, len(r.TaskLists))
	for i, tl := range r.TaskLists {
		rows[i] = []string{
			tl.ID,
			tl.Title,
			tl.Updated,
		}
	}
	return rows
}

func (r *TaskListResult) EmptyMessage() string {
	return "No task lists found"
}

// Task represents a Google Tasks task.
// Status is "needsAction" or "completed".
// Due is a date-only string in "YYYY-MM-DD" format.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes,omitempty"`
	Status    string     `json:"status"`
	Due       string     `json:"due,omitempty"`
	Completed string     `json:"completed,omitempty"`
	Parent    string     `json:"parent,omitempty"`
	Position  string     `json:"position,omitempty"`
	Links     []TaskLink `json:"links,omitempty"`
}

func (t *Task) Headers() []string {
	return []string{"ID", "Title", "Status", "Due"}
}

func (t *Task) Rows() [][]string {
	return [][]string{{
		t.ID,
		t.Title,
		t.Status,
		t.Due,
	}}
}

func (t *Task) EmptyMessage() string {
	return "No task found"
}

// TaskResult represents a collection of tasks.
type TaskResult struct {
	Tasks []Task `json:"tasks"`
}

func (r *TaskResult) Headers() []string {
	return []string{"ID", "Title", "Status", "Due"}
}

func (r *TaskResult) Rows() [][]string {
	rows := make([][]string, len(r.Tasks))
	for i, task := range r.Tasks {
		rows[i] = []string{
			task.ID,
			task.Title,
			task.Status,
			task.Due,
		}
	}
	return rows
}

func (r *TaskResult) EmptyMessage() string {
	return "No tasks found"
}

// TaskMutationResult represents the result of a task mutation operation.
type TaskMutationResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func (r *TaskMutationResult) Headers() []string {
	return []string{"ID", "Title", "Status"}
}

func (r *TaskMutationResult) Rows() [][]string {
	return [][]string{{
		r.ID,
		r.Title,
		r.Status,
	}}
}

func (r *TaskMutationResult) EmptyMessage() string {
	return "No result"
}
