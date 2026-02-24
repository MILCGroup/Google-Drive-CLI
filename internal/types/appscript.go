package types

import (
	"fmt"
	"strings"
)

type ScriptProject struct {
	ScriptID   string `json:"scriptId"`
	Title      string `json:"title"`
	ParentID   string `json:"parentId,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	UpdateTime string `json:"updateTime,omitempty"`
}

func (p *ScriptProject) Headers() []string {
	return []string{"ID", "Title", "Created", "Updated"}
}

func (p *ScriptProject) Rows() [][]string {
	return [][]string{{
		p.ScriptID,
		p.Title,
		p.CreateTime,
		p.UpdateTime,
	}}
}

func (p *ScriptProject) EmptyMessage() string {
	return "No script project found"
}

type ScriptFile struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Source      string   `json:"source,omitempty"`
	CreateTime  string   `json:"createTime,omitempty"`
	UpdateTime  string   `json:"updateTime,omitempty"`
	FunctionSet []string `json:"functionSet,omitempty"`
}

func (f *ScriptFile) Headers() []string {
	return []string{"Name", "Type", "Functions"}
}

func (f *ScriptFile) Rows() [][]string {
	return [][]string{{
		f.Name,
		f.Type,
		fmt.Sprintf("%d", len(f.FunctionSet)),
	}}
}

func (f *ScriptFile) EmptyMessage() string {
	return "No script file found"
}

type ScriptContent struct {
	ScriptID string       `json:"scriptId"`
	Files    []ScriptFile `json:"files"`
}

func (c *ScriptContent) Headers() []string {
	return []string{"Name", "Type", "Functions"}
}

func (c *ScriptContent) Rows() [][]string {
	rows := make([][]string, len(c.Files))
	for i, f := range c.Files {
		rows[i] = []string{
			f.Name,
			f.Type,
			fmt.Sprintf("%d", len(f.FunctionSet)),
		}
	}
	return rows
}

func (c *ScriptContent) EmptyMessage() string {
	return "No script files found"
}

type ScriptError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *ScriptError) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("script error %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("script error %d: %s (%s)", e.Code, e.Message, strings.Join(e.Details, "; "))
}

type ScriptRunResult struct {
	Done     bool                   `json:"done"`
	Error    *ScriptError           `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

func (r *ScriptRunResult) Headers() []string {
	return []string{"Done", "Error"}
}

func (r *ScriptRunResult) Rows() [][]string {
	errMsg := ""
	if r.Error != nil {
		errMsg = r.Error.Message
	}
	return [][]string{{
		fmt.Sprintf("%t", r.Done),
		errMsg,
	}}
}

func (r *ScriptRunResult) EmptyMessage() string {
	return "No run result"
}

type ScriptCreateResult struct {
	ScriptID string `json:"scriptId"`
	Title    string `json:"title"`
}

func (r *ScriptCreateResult) Headers() []string {
	return []string{"ID", "Title"}
}

func (r *ScriptCreateResult) Rows() [][]string {
	return [][]string{{
		r.ScriptID,
		r.Title,
	}}
}

func (r *ScriptCreateResult) EmptyMessage() string {
	return "No result"
}
