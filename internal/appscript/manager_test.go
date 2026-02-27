package appscript

import (
	"testing"

	"google.golang.org/api/googleapi"
	script "google.golang.org/api/script/v1"
)

func TestConvertProject(t *testing.T) {
	tests := []struct {
		name   string
		input  *script.Project
		wantID string
		wantT  string
		wantP  string
		wantCT string
		wantUT string
	}{
		{
			name:   "nil project",
			input:  nil,
			wantID: "",
			wantT:  "",
			wantP:  "",
			wantCT: "",
			wantUT: "",
		},
		{
			name: "all fields populated",
			input: &script.Project{
				ScriptId:   "script-123",
				Title:      "My Script",
				ParentId:   "parent-456",
				CreateTime: "2025-01-01T00:00:00Z",
				UpdateTime: "2025-06-15T12:30:00Z",
			},
			wantID: "script-123",
			wantT:  "My Script",
			wantP:  "parent-456",
			wantCT: "2025-01-01T00:00:00Z",
			wantUT: "2025-06-15T12:30:00Z",
		},
		{
			name: "standalone project without parent",
			input: &script.Project{
				ScriptId:   "script-789",
				Title:      "Standalone",
				CreateTime: "2025-03-10T08:00:00Z",
				UpdateTime: "2025-03-10T08:00:00Z",
			},
			wantID: "script-789",
			wantT:  "Standalone",
			wantP:  "",
			wantCT: "2025-03-10T08:00:00Z",
			wantUT: "2025-03-10T08:00:00Z",
		},
		{
			name:   "empty project",
			input:  &script.Project{},
			wantID: "",
			wantT:  "",
			wantP:  "",
			wantCT: "",
			wantUT: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertProject(tt.input)
			if got.ScriptID != tt.wantID {
				t.Errorf("ScriptID = %q, want %q", got.ScriptID, tt.wantID)
			}
			if got.Title != tt.wantT {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantT)
			}
			if got.ParentID != tt.wantP {
				t.Errorf("ParentID = %q, want %q", got.ParentID, tt.wantP)
			}
			if got.CreateTime != tt.wantCT {
				t.Errorf("CreateTime = %q, want %q", got.CreateTime, tt.wantCT)
			}
			if got.UpdateTime != tt.wantUT {
				t.Errorf("UpdateTime = %q, want %q", got.UpdateTime, tt.wantUT)
			}
		})
	}
}

func TestConvertFile(t *testing.T) {
	tests := []struct {
		name          string
		input         *script.File
		wantName      string
		wantType      string
		wantSource    string
		wantCreate    string
		wantUpdate    string
		wantFuncCount int
		wantFuncs     []string
	}{
		{
			name:          "nil file",
			input:         nil,
			wantName:      "",
			wantType:      "",
			wantSource:    "",
			wantFuncCount: 0,
		},
		{
			name: "server js file with functions",
			input: &script.File{
				Name:       "Code",
				Type:       "SERVER_JS",
				Source:     "function doGet() { return 'hello'; }",
				CreateTime: "2025-01-01T00:00:00Z",
				UpdateTime: "2025-06-15T12:30:00Z",
				FunctionSet: &script.GoogleAppsScriptTypeFunctionSet{
					Values: []*script.GoogleAppsScriptTypeFunction{
						{Name: "doGet"},
						{Name: "doPost"},
						{Name: "processData"},
					},
				},
			},
			wantName:      "Code",
			wantType:      "SERVER_JS",
			wantSource:    "function doGet() { return 'hello'; }",
			wantCreate:    "2025-01-01T00:00:00Z",
			wantUpdate:    "2025-06-15T12:30:00Z",
			wantFuncCount: 3,
			wantFuncs:     []string{"doGet", "doPost", "processData"},
		},
		{
			name: "html file without functions",
			input: &script.File{
				Name:   "Index",
				Type:   "HTML",
				Source: "<html><body>Hello</body></html>",
			},
			wantName:      "Index",
			wantType:      "HTML",
			wantSource:    "<html><body>Hello</body></html>",
			wantFuncCount: 0,
		},
		{
			name: "json manifest file",
			input: &script.File{
				Name:   "appsscript",
				Type:   "JSON",
				Source: `{"timeZone":"America/New_York"}`,
			},
			wantName:      "appsscript",
			wantType:      "JSON",
			wantSource:    `{"timeZone":"America/New_York"}`,
			wantFuncCount: 0,
		},
		{
			name: "file with empty function set",
			input: &script.File{
				Name: "Empty",
				Type: "SERVER_JS",
				FunctionSet: &script.GoogleAppsScriptTypeFunctionSet{
					Values: []*script.GoogleAppsScriptTypeFunction{},
				},
			},
			wantName:      "Empty",
			wantType:      "SERVER_JS",
			wantFuncCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertFile(tt.input)
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Source != tt.wantSource {
				t.Errorf("Source = %q, want %q", got.Source, tt.wantSource)
			}
			if got.CreateTime != tt.wantCreate {
				t.Errorf("CreateTime = %q, want %q", got.CreateTime, tt.wantCreate)
			}
			if got.UpdateTime != tt.wantUpdate {
				t.Errorf("UpdateTime = %q, want %q", got.UpdateTime, tt.wantUpdate)
			}
			if len(got.FunctionSet) != tt.wantFuncCount {
				t.Errorf("FunctionSet count = %d, want %d", len(got.FunctionSet), tt.wantFuncCount)
			}
			for i, wantFn := range tt.wantFuncs {
				if i >= len(got.FunctionSet) {
					t.Errorf("missing function at index %d: want %q", i, wantFn)
					continue
				}
				if got.FunctionSet[i] != wantFn {
					t.Errorf("FunctionSet[%d] = %q, want %q", i, got.FunctionSet[i], wantFn)
				}
			}
		})
	}
}

func TestConvertRunResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *script.Operation
		wantDone bool
		wantErr  bool
		wantCode int
		wantMsg  string
		wantResp bool
	}{
		{
			name:     "nil operation",
			input:    nil,
			wantDone: false,
			wantErr:  false,
			wantResp: false,
		},
		{
			name: "successful completion with response",
			input: &script.Operation{
				Done:     true,
				Response: googleapi.RawMessage(`{"result":"success","value":42}`),
			},
			wantDone: true,
			wantErr:  false,
			wantResp: true,
		},
		{
			name: "completed with error",
			input: &script.Operation{
				Done: true,
				Error: &script.Status{
					Code:    3,
					Message: "INVALID_ARGUMENT",
					Details: []googleapi.RawMessage{
						googleapi.RawMessage(`{"errorType":"TypeError","errorMessage":"Cannot read property"}`),
					},
				},
			},
			wantDone: true,
			wantErr:  true,
			wantCode: 3,
			wantMsg:  "INVALID_ARGUMENT",
			wantResp: false,
		},
		{
			name: "timeout error",
			input: &script.Operation{
				Done: true,
				Error: &script.Status{
					Code:    10,
					Message: "SCRIPT_TIMEOUT",
				},
			},
			wantDone: true,
			wantErr:  true,
			wantCode: 10,
			wantMsg:  "SCRIPT_TIMEOUT",
			wantResp: false,
		},
		{
			name: "not yet done",
			input: &script.Operation{
				Done: false,
			},
			wantDone: false,
			wantErr:  false,
			wantResp: false,
		},
		{
			name: "nil error field explicitly",
			input: &script.Operation{
				Done:     true,
				Error:    nil,
				Response: googleapi.RawMessage(`{"result":"ok"}`),
			},
			wantDone: true,
			wantErr:  false,
			wantResp: true,
		},
		{
			name: "error with no details",
			input: &script.Operation{
				Done: true,
				Error: &script.Status{
					Code:    1,
					Message: "CANCELLED",
				},
			},
			wantDone: true,
			wantErr:  true,
			wantCode: 1,
			wantMsg:  "CANCELLED",
			wantResp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertRunResponse(tt.input)

			if got.Done != tt.wantDone {
				t.Errorf("Done = %v, want %v", got.Done, tt.wantDone)
			}

			if tt.wantErr {
				if got.Error == nil {
					t.Fatal("expected non-nil Error")
				}
				if got.Error.Code != tt.wantCode {
					t.Errorf("Error.Code = %d, want %d", got.Error.Code, tt.wantCode)
				}
				if got.Error.Message != tt.wantMsg {
					t.Errorf("Error.Message = %q, want %q", got.Error.Message, tt.wantMsg)
				}
			} else {
				if got.Error != nil {
					t.Errorf("expected nil Error, got %+v", got.Error)
				}
			}

			if tt.wantResp {
				if got.Response == nil {
					t.Fatal("expected non-nil Response")
				}
			} else {
				if got.Response != nil {
					t.Errorf("expected nil Response, got %v", got.Response)
				}
			}
		})
	}
}

func TestConvertRunResponseDetails(t *testing.T) {
	op := &script.Operation{
		Done: true,
		Error: &script.Status{
			Code:    3,
			Message: "error",
			Details: []googleapi.RawMessage{
				googleapi.RawMessage(`{"errorType":"ReferenceError"}`),
				googleapi.RawMessage(`{"stackTrace":"at line 5"}`),
			},
		},
	}

	got := convertRunResponse(op)

	if got.Error == nil {
		t.Fatal("expected error")
	}
	if len(got.Error.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(got.Error.Details))
	}
	if got.Error.Details[0] != `{"errorType":"ReferenceError"}` {
		t.Errorf("Details[0] = %q, unexpected", got.Error.Details[0])
	}
	if got.Error.Details[1] != `{"stackTrace":"at line 5"}` {
		t.Errorf("Details[1] = %q, unexpected", got.Error.Details[1])
	}
}

func TestConvertRunResponseMalformedJSON(t *testing.T) {
	op := &script.Operation{
		Done:     true,
		Response: googleapi.RawMessage(`not valid json`),
	}

	got := convertRunResponse(op)

	if got.Response == nil {
		t.Fatal("expected non-nil Response for malformed JSON fallback")
	}
	raw, ok := got.Response["raw"]
	if !ok {
		t.Fatal("expected 'raw' key in Response")
	}
	if raw != "not valid json" {
		t.Errorf("raw = %q, want %q", raw, "not valid json")
	}
}
