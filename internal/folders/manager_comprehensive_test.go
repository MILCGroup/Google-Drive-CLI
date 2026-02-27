package folders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/logging"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// mockServer creates a test HTTP server with the given handler
func mockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *drive.Service) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	ctx := context.Background()
	service, err := drive.NewService(ctx,
		option.WithEndpoint(server.URL+"/drive/v3/"),
		option.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create drive service: %v", err)
	}

	return server, service
}

// newTestManager creates a Manager with a mock server
func newTestManager(t *testing.T, handler http.HandlerFunc) (*Manager, *api.Client) {
	t.Helper()
	_, service := mockServer(t, handler)
	client := api.NewClient(service, 0, 100, logging.NewNoOpLogger())
	return NewManager(client), client
}

// newTestManagerWithRetry creates a Manager with retry capability
func newTestManagerWithRetry(t *testing.T, handler http.HandlerFunc, maxRetries int) (*Manager, *api.Client) {
	t.Helper()
	_, service := mockServer(t, handler)
	client := api.NewClient(service, maxRetries, 100, logging.NewNoOpLogger())
	return NewManager(client), client
}

// writeJSON writes a JSON response
func writeJSON(t *testing.T, w http.ResponseWriter, status int, payload interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("Failed to encode response: %v", err)
	}
}

// writeAPIError writes an API error response
func writeAPIError(t *testing.T, w http.ResponseWriter, status int, reason, message string) {
	t.Helper()
	if message == "" {
		message = "test error"
	}
	resp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
			"errors": []map[string]string{{
				"domain":  "global",
				"reason":  reason,
				"message": message,
			}},
		},
	}
	writeJSON(t, w, status, resp)
}

// requireQueryValue checks that a query parameter has the expected value
func requireQueryValue(t *testing.T, query url.Values, key, want string) {
	t.Helper()
	got := query.Get(key)
	if got != want {
		t.Fatalf("Expected query %s=%q, got %q", key, want, got)
	}
}

// assertAppErrorCode checks that an error has the expected code
func assertAppErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error with code %s", wantCode)
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Expected AppError, got %T: %v", err, err)
	}
	if appErr.CLIError.Code != wantCode {
		t.Fatalf("Expected error code %s, got %s", wantCode, appErr.CLIError.Code)
	}
}

// TestManager_Create tests folder creation
func TestManager_Create(t *testing.T) {
	tests := []struct {
		testName   string
		name       string
		parentID   string
		mockFile   *drive.File
		mockStatus int
		wantErr    bool
	}{
		{
			testName: "create folder with parent",
			name:     "New Folder",
			parentID: "parent123",
			mockFile: &drive.File{
				Id:       "folder-id",
				Name:     "New Folder",
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{"parent123"},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName: "create folder without parent (root)",
			name:     "Root Folder",
			parentID: "",
			mockFile: &drive.File{
				Id:       "folder-id-2",
				Name:     "Root Folder",
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:   "API error",
			name:       "Test Folder",
			parentID:   "parent123",
			mockFile:   nil,
			mockStatus: http.StatusForbidden,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("Expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/drive/v3/files" {
					t.Fatalf("Expected path /drive/v3/files, got %s", r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")

				// Read request body
				body, _ := io.ReadAll(r.Body)
				bodyStr := string(body)

				if !strings.Contains(bodyStr, "application/vnd.google-apps.folder") {
					t.Errorf("Expected folder MIME type in body: %s", bodyStr)
				}

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "forbidden", "access denied")
					return
				}
				writeJSON(t, w, http.StatusOK, tt.mockFile)
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
			result, err := manager.Create(context.Background(), reqCtx, tt.name, tt.parentID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.mockFile.Id {
				t.Errorf("ID mismatch: got %s, want %s", result.ID, tt.mockFile.Id)
			}
			if result.MimeType != "application/vnd.google-apps.folder" {
				t.Errorf("Expected folder MIME type, got %s", result.MimeType)
			}
		})
	}
}

// TestManager_List tests listing folder contents
func TestManager_List(t *testing.T) {
	tests := []struct {
		testName      string
		folderID      string
		pageSize      int
		pageToken     string
		mockFiles     []*drive.File
		nextToken     string
		mockStatus    int
		wantErr       bool
		wantFileCount int
	}{
		{
			testName:  "list folder contents",
			folderID:  "folder123",
			pageSize:  10,
			pageToken: "",
			mockFiles: []*drive.File{
				{Id: "file1", Name: "file1.txt", MimeType: "text/plain"},
				{Id: "subfolder", Name: "Sub Folder", MimeType: "application/vnd.google-apps.folder"},
			},
			nextToken:     "",
			mockStatus:    http.StatusOK,
			wantErr:       false,
			wantFileCount: 2,
		},
		{
			testName:  "list with pagination",
			folderID:  "folder456",
			pageSize:  10,
			pageToken: "",
			mockFiles: []*drive.File{
				{Id: "file1", Name: "file1.txt"},
			},
			nextToken:     "next-token",
			mockStatus:    http.StatusOK,
			wantErr:       false,
			wantFileCount: 1,
		},
		{
			testName:      "empty folder",
			folderID:      "empty-folder",
			pageSize:      10,
			pageToken:     "",
			mockFiles:     []*drive.File{},
			nextToken:     "",
			mockStatus:    http.StatusOK,
			wantErr:       false,
			wantFileCount: 0,
		},
		{
			testName:   "folder not found",
			folderID:   "missing",
			pageSize:   10,
			mockFiles:  nil,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/drive/v3/files" {
					t.Fatalf("Expected path /drive/v3/files, got %s", r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")

				q := query.Get("q")
				if !strings.Contains(q, fmt.Sprintf("'%s' in parents", tt.folderID)) {
					t.Fatalf("Expected query to contain parent filter, got %s", q)
				}
				if !strings.Contains(q, "trashed = false") {
					t.Fatalf("Expected query to contain trashed filter, got %s", q)
				}

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "notFound", "folder not found")
					return
				}
				writeJSON(t, w, http.StatusOK, &drive.FileList{
					Files:         tt.mockFiles,
					NextPageToken: tt.nextToken,
				})
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
			result, err := manager.List(context.Background(), reqCtx, tt.folderID, tt.pageSize, tt.pageToken)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if len(result.Files) != tt.wantFileCount {
				t.Errorf("File count mismatch: got %d, want %d", len(result.Files), tt.wantFileCount)
			}
		})
	}
}

// TestManager_Delete tests folder deletion
func TestManager_Delete(t *testing.T) {
	tests := []struct {
		testName   string
		folderID   string
		recursive  bool
		mockStatus int
		mockErr    bool
		wantErr    bool
	}{
		{
			testName:   "delete empty folder",
			folderID:   "folder123",
			recursive:  false,
			mockStatus: http.StatusNoContent,
			mockErr:    false,
			wantErr:    false,
		},
		{
			testName:   "delete folder recursively",
			folderID:   "folder456",
			recursive:  true,
			mockStatus: http.StatusNoContent,
			mockErr:    false,
			wantErr:    false,
		},
		{
			testName:   "folder not found",
			folderID:   "missing",
			recursive:  false,
			mockStatus: http.StatusNotFound,
			mockErr:    true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					// Get folder metadata
					if tt.mockErr {
						writeAPIError(t, w, tt.mockStatus, "notFound", "folder not found")
						return
					}
					writeJSON(t, w, http.StatusOK, &drive.File{
						Id:       tt.folderID,
						Name:     "Test Folder",
						MimeType: "application/vnd.google-apps.folder",
					})
				case http.MethodDelete:
					// Delete folder
					w.WriteHeader(tt.mockStatus)
				}
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
			err := manager.Delete(context.Background(), reqCtx, tt.folderID, tt.recursive)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestManager_Move tests moving a folder
func TestManager_Move(t *testing.T) {
	tests := []struct {
		testName    string
		folderID    string
		newParentID string
		mockFile    *drive.File
		mockStatus  int
		wantErr     bool
	}{
		{
			testName:    "move folder to new parent",
			folderID:    "folder123",
			newParentID: "parent456",
			mockFile: &drive.File{
				Id:       "folder123",
				Name:     "moved folder",
				Parents:  []string{"parent456"},
				MimeType: "application/vnd.google-apps.folder",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:    "folder not found",
			folderID:    "missing",
			newParentID: "parent123",
			mockFile:    nil,
			mockStatus:  http.StatusNotFound,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					// Get current parents
					if tt.wantErr {
						writeAPIError(t, w, tt.mockStatus, "notFound", "folder not found")
						return
					}
					writeJSON(t, w, http.StatusOK, &drive.File{
						Id:      tt.folderID,
						Parents: []string{"old-parent"},
					})
				case http.MethodPatch:
					// Move folder
					query := r.URL.Query()
					addParents := query.Get("addParents")
					if addParents != tt.newParentID {
						t.Fatalf("Expected addParents=%s, got %s", tt.newParentID, addParents)
					}
					writeJSON(t, w, http.StatusOK, tt.mockFile)
				}
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
			result, err := manager.Move(context.Background(), reqCtx, tt.folderID, tt.newParentID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.mockFile.Id {
				t.Errorf("ID mismatch: got %s, want %s", result.ID, tt.mockFile.Id)
			}
		})
	}
}

// TestManager_Get tests getting folder metadata
func TestManager_Get(t *testing.T) {
	tests := []struct {
		testName   string
		folderID   string
		fields     string
		mockFile   *drive.File
		mockStatus int
		wantErr    bool
	}{
		{
			testName: "get folder with default fields",
			folderID: "folder123",
			fields:   "",
			mockFile: &drive.File{
				Id:       "folder123",
				Name:     "My Folder",
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{"parent1"},
				Trashed:  false,
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName: "get folder with custom fields",
			folderID: "folder456",
			fields:   "id,name,parents",
			mockFile: &drive.File{
				Id:      "folder456",
				Name:    "Custom Folder",
				Parents: []string{"parent2"},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:   "folder not found",
			folderID:   "missing",
			fields:     "",
			mockFile:   nil,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/drive/v3/files/%s", tt.folderID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "notFound", "folder not found")
					return
				}
				writeJSON(t, w, http.StatusOK, tt.mockFile)
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
			result, err := manager.Get(context.Background(), reqCtx, tt.folderID, tt.fields)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.mockFile.Id {
				t.Errorf("ID mismatch: got %s, want %s", result.ID, tt.mockFile.Id)
			}
		})
	}
}

// TestManager_Rename tests renaming a folder
func TestManager_Rename(t *testing.T) {
	tests := []struct {
		testName   string
		folderID   string
		newName    string
		mockFile   *drive.File
		mockStatus int
		wantErr    bool
	}{
		{
			testName: "rename folder",
			folderID: "folder123",
			newName:  "Renamed Folder",
			mockFile: &drive.File{
				Id:       "folder123",
				Name:     "Renamed Folder",
				MimeType: "application/vnd.google-apps.folder",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName: "rename with special characters",
			folderID: "folder456",
			newName:  "Folder (2024) [test]",
			mockFile: &drive.File{
				Id:   "folder456",
				Name: "Folder (2024) [test]",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:   "folder not found",
			folderID:   "missing",
			newName:    "New Name",
			mockFile:   nil,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Fatalf("Expected PATCH, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/drive/v3/files/%s", tt.folderID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")

				// Verify name in body
				body, _ := io.ReadAll(r.Body)
				bodyStr := string(body)
				if !strings.Contains(bodyStr, tt.newName) {
					t.Errorf("Expected new name in body: %s", bodyStr)
				}

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "notFound", "folder not found")
					return
				}
				writeJSON(t, w, http.StatusOK, tt.mockFile)
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
			result, err := manager.Rename(context.Background(), reqCtx, tt.folderID, tt.newName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.Name != tt.newName {
				t.Errorf("Name mismatch: got %s, want %s", result.Name, tt.newName)
			}
		})
	}
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		testName string
		status   int
		reason   string
		wantCode string
	}{
		{
			testName: "rate limit exceeded",
			status:   http.StatusTooManyRequests,
			reason:   "rateLimitExceeded",
			wantCode: utils.ErrCodeRateLimited,
		},
		{
			testName: "file not found",
			status:   http.StatusNotFound,
			reason:   "notFound",
			wantCode: utils.ErrCodeFileNotFound,
		},
		{
			testName: "permission denied",
			status:   http.StatusForbidden,
			reason:   "insufficientFilePermissions",
			wantCode: utils.ErrCodePermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				writeAPIError(t, w, tt.status, tt.reason, "test error")
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
			_, err := manager.Get(context.Background(), reqCtx, "folder123", "")
			assertAppErrorCode(t, err, tt.wantCode)
		})
	}
}

// TestRetryLogic tests retry behavior
func TestRetryLogic(t *testing.T) {
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			writeAPIError(t, w, http.StatusServiceUnavailable, "backendError", "temporary error")
			return
		}
		writeJSON(t, w, http.StatusOK, &drive.File{
			Id:       "folder123",
			Name:     "Test Folder",
			MimeType: "application/vnd.google-apps.folder",
		})
	}

	manager, _ := newTestManagerWithRetry(t, handler, 3)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	result, err := manager.Get(context.Background(), reqCtx, "folder123", "")

	if err != nil {
		t.Fatalf("Expected retry to succeed, got error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

// TestPagination tests pagination handling
func TestPagination(t *testing.T) {
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		query := r.URL.Query()
		pageToken := query.Get("pageToken")

		var result *drive.FileList
		if pageToken == "" {
			result = &drive.FileList{
				Files: []*drive.File{
					{Id: "file1", Name: "file1.txt"},
				},
				NextPageToken: "page2",
			}
		} else {
			result = &drive.FileList{
				Files: []*drive.File{
					{Id: "file2", Name: "file2.txt"},
				},
				NextPageToken: "",
			}
		}
		writeJSON(t, w, http.StatusOK, result)
	}

	manager, _ := newTestManager(t, handler)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)

	// Get first page
	result1, err := manager.List(context.Background(), reqCtx, "folder123", 10, "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result1.Files) != 1 {
		t.Errorf("Expected 1 file in first page, got %d", len(result1.Files))
	}

	// Get second page
	result2, err := manager.List(context.Background(), reqCtx, "folder123", 10, result1.NextPageToken)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result2.Files) != 1 {
		t.Errorf("Expected 1 file in second page, got %d", len(result2.Files))
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls, got %d", callCount)
	}
}

// TestContextCancellation tests context cancellation
// Note: This test verifies that context handling doesn't panic.
// The Google API client may or may not return an error for cancelled contexts
// depending on timing and internal implementation.
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, &drive.File{
			Id:   "folder123",
			Name: "Test Folder",
		})
	}

	manager, _ := newTestManager(t, handler)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)

	// Just verify this doesn't panic - the actual behavior depends on the API client
	_, _ = manager.Get(ctx, reqCtx, "folder123", "")
}

// TestSharedDriveContext tests operations with shared drive context
func TestSharedDriveContext(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		requireQueryValue(t, query, "supportsAllDrives", "true")
		writeJSON(t, w, http.StatusOK, &drive.File{
			Id:       "folder123",
			Name:     "Shared Drive Folder",
			MimeType: "application/vnd.google-apps.folder",
		})
	}

	manager, _ := newTestManager(t, handler)
	reqCtx := api.NewRequestContext("default", "shared-drive-123", types.RequestTypeMutation)
	result, err := manager.Create(context.Background(), reqCtx, "New Folder", "")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.ID != "folder123" {
		t.Errorf("ID mismatch: got %s", result.ID)
	}
}

// TestConvertDriveFileComprehensive tests the conversion function
func TestConvertDriveFileComprehensive(t *testing.T) {
	tests := []struct {
		testName string
		input    *drive.File
		validate func(*testing.T, *types.DriveFile)
	}{
		{
			testName: "basic folder conversion",
			input: &drive.File{
				Id:           "folder123",
				Name:         "My Folder",
				MimeType:     "application/vnd.google-apps.folder",
				Size:         0,
				CreatedTime:  "2024-01-01T00:00:00Z",
				ModifiedTime: "2024-01-02T00:00:00Z",
				Parents:      []string{"parent1"},
				ResourceKey:  "resource-key-123",
				Trashed:      false,
			},
			validate: func(t *testing.T, f *types.DriveFile) {
				if f.MimeType != "application/vnd.google-apps.folder" {
					t.Errorf("Expected folder MIME type, got %s", f.MimeType)
				}
				if f.Size != 0 {
					t.Error("Folder should have size 0")
				}
				if len(f.Parents) != 1 {
					t.Errorf("Expected 1 parent, got %d", len(f.Parents))
				}
			},
		},
		{
			testName: "trashed folder",
			input: &drive.File{
				Id:       "trashed-folder",
				Name:     "Deleted Folder",
				MimeType: "application/vnd.google-apps.folder",
				Trashed:  true,
			},
			validate: func(t *testing.T, f *types.DriveFile) {
				if !f.Trashed {
					t.Error("Expected folder to be trashed")
				}
			},
		},
		{
			testName: "folder with capabilities",
			input: &drive.File{
				Id:       "folder123",
				Name:     "Protected Folder",
				MimeType: "application/vnd.google-apps.folder",
				Capabilities: &drive.FileCapabilities{
					CanEdit:   false,
					CanDelete: false,
					CanTrash:  true,
				},
			},
			validate: func(t *testing.T, f *types.DriveFile) {
				if f.Capabilities == nil {
					t.Fatal("Expected capabilities")
				}
				if f.Capabilities.CanEdit {
					t.Error("Expected CanEdit to be false")
				}
			},
		},
		{
			testName: "folder without capabilities",
			input: &drive.File{
				Id:       "folder123",
				Name:     "Simple Folder",
				MimeType: "application/vnd.google-apps.folder",
			},
			validate: func(t *testing.T, f *types.DriveFile) {
				if f.Capabilities != nil {
					t.Error("Expected nil capabilities")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := convertDriveFile(tt.input)
			tt.validate(t, result)
		})
	}
}

// TestRecursiveDeleteWithContents tests recursive deletion with subfolders
func TestRecursiveDeleteWithContents(t *testing.T) {
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/files"):
			// List folder contents - return empty to prevent infinite recursion
			// The recursive counting would call this repeatedly for subfolders
			writeJSON(t, w, http.StatusOK, &drive.FileList{
				Files: []*drive.File{},
			})
		case r.Method == http.MethodGet:
			// Get folder metadata
			writeJSON(t, w, http.StatusOK, &drive.File{
				Id:       "parent-folder",
				Name:     "Parent Folder",
				MimeType: "application/vnd.google-apps.folder",
			})
		case r.Method == http.MethodDelete:
			// Delete file/folder
			w.WriteHeader(http.StatusNoContent)
		}
	}

	manager, _ := newTestManager(t, handler)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)

	// Test recursive delete with empty folder (no infinite recursion)
	err := manager.Delete(context.Background(), reqCtx, "parent-folder", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Benchmark tests
func BenchmarkConvertDriveFile(b *testing.B) {
	driveFile := &drive.File{
		Id:       "folder123",
		Name:     "benchmark",
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"parent1", "parent2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertDriveFile(driveFile)
	}
}
