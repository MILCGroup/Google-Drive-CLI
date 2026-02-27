package appscript

import (
	"context"
	"encoding/json"
	"time"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	script "google.golang.org/api/script/v1"
)

// Manager wraps the Apps Script API with retry logic and type conversion.
type Manager struct {
	client  *api.Client
	service *script.Service
}

// NewManager creates a new Apps Script manager.
func NewManager(client *api.Client, service *script.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// GetProject retrieves metadata for an Apps Script project by its script ID.
// Note: the script ID is not the same as the Drive file ID.
func (m *Manager) GetProject(ctx context.Context, reqCtx *types.RequestContext, scriptID string) (*types.ScriptProject, error) {
	call := m.service.Projects.Get(scriptID)

	project, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*script.Project, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertProject(project), nil
}

// GetContent retrieves the file content of an Apps Script project, including
// source code and function names extracted from each file's function set.
func (m *Manager) GetContent(ctx context.Context, reqCtx *types.RequestContext, scriptID string) (*types.ScriptContent, error) {
	call := m.service.Projects.GetContent(scriptID)

	content, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*script.Content, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	files := make([]types.ScriptFile, len(content.Files))
	for i, f := range content.Files {
		files[i] = convertFile(f)
	}

	return &types.ScriptContent{
		ScriptID: content.ScriptId,
		Files:    files,
	}, nil
}

// CreateProject creates a new Apps Script project. If parentID is non-empty,
// the project is bound to that parent (typically a Google Doc, Sheet, Form, or
// Slides file). Otherwise a standalone script project is created.
func (m *Manager) CreateProject(ctx context.Context, reqCtx *types.RequestContext, title, parentID string) (*types.ScriptCreateResult, error) {
	req := &script.CreateProjectRequest{
		Title: title,
	}
	if parentID != "" {
		req.ParentId = parentID
	}

	call := m.service.Projects.Create(req)

	project, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*script.Project, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.ScriptCreateResult{
		ScriptID: project.ScriptId,
		Title:    project.Title,
	}, nil
}

// EXPERIMENTAL: RunFunction executes a function in an Apps Script project.
//
// This method has significant constraints:
//   - The script must be deployed as an API executable (not a web app).
//   - The caller must have been granted access to the script.
//   - The caller must provide all OAuth scopes that the script requires.
//   - Both the script and the calling application must share the same
//     Google Cloud Platform (GCP) project.
//   - There is a hard 6-minute execution timeout enforced by the Apps Script
//     runtime; this method applies a matching context deadline.
//   - Parameters must be JSON-serializable primitive types only (string,
//     number, bool, array, or plain object). Apps Script-specific types
//     such as Document or SpreadsheetApp are not supported.
func (m *Manager) RunFunction(ctx context.Context, reqCtx *types.RequestContext, scriptID, functionName string, parameters []interface{}) (*types.ScriptRunResult, error) {
	// Enforce the 6-minute Apps Script execution timeout.
	const scriptTimeout = 6 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()

	execReq := &script.ExecutionRequest{
		Function:   functionName,
		Parameters: parameters,
	}

	call := m.service.Scripts.Run(scriptID, execReq)

	op, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*script.Operation, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertRunResponse(op), nil
}

// convertProject maps a script.Project to a types.ScriptProject.
func convertProject(p *script.Project) *types.ScriptProject {
	if p == nil {
		return &types.ScriptProject{}
	}
	return &types.ScriptProject{
		ScriptID:   p.ScriptId,
		Title:      p.Title,
		ParentID:   p.ParentId,
		CreateTime: p.CreateTime,
		UpdateTime: p.UpdateTime,
	}
}

// convertFile maps a script.File to a types.ScriptFile, extracting function
// names from the file's FunctionSet when present.
func convertFile(f *script.File) types.ScriptFile {
	if f == nil {
		return types.ScriptFile{}
	}

	sf := types.ScriptFile{
		Name:       f.Name,
		Type:       f.Type,
		Source:     f.Source,
		CreateTime: f.CreateTime,
		UpdateTime: f.UpdateTime,
	}

	if f.FunctionSet != nil {
		names := make([]string, len(f.FunctionSet.Values))
		for i, fn := range f.FunctionSet.Values {
			names[i] = fn.Name
		}
		sf.FunctionSet = names
	}

	return sf
}

// convertRunResponse maps a script.Operation to a types.ScriptRunResult,
// handling the cases where Error or Response may be nil.
func convertRunResponse(op *script.Operation) *types.ScriptRunResult {
	if op == nil {
		return &types.ScriptRunResult{}
	}

	result := &types.ScriptRunResult{
		Done: op.Done,
	}

	if op.Error != nil {
		details := make([]string, 0, len(op.Error.Details))
		for _, d := range op.Error.Details {
			details = append(details, string(d))
		}
		result.Error = &types.ScriptError{
			Code:    int(op.Error.Code),
			Message: op.Error.Message,
			Details: details,
		}
	}

	if len(op.Response) > 0 {
		var resp map[string]interface{}
		if err := json.Unmarshal(op.Response, &resp); err == nil {
			result.Response = resp
		} else {
			// If unmarshal fails, store the raw JSON as a string under "raw".
			result.Response = map[string]interface{}{
				"raw": string(op.Response),
			}
		}
	}

	return result
}
