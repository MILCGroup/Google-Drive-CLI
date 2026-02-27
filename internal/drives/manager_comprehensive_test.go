package drives

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// TestManager_List tests listing shared drives
func TestManager_List(t *testing.T) {
	tests := []struct {
		testName       string
		pageSize       int
		pageToken      string
		mockDrives     []*drive.Drive
		nextToken      string
		mockStatus     int
		wantErr        bool
		wantDriveCount int
	}{
		{
			testName:  "list shared drives",
			pageSize:  10,
			pageToken: "",
			mockDrives: []*drive.Drive{
				{Id: "drive1", Name: "Drive 1"},
				{Id: "drive2", Name: "Drive 2"},
			},
			nextToken:      "",
			mockStatus:     http.StatusOK,
			wantErr:        false,
			wantDriveCount: 2,
		},
		{
			testName:  "list with pagination",
			pageSize:  10,
			pageToken: "",
			mockDrives: []*drive.Drive{
				{Id: "drive3", Name: "Drive 3"},
			},
			nextToken:      "next-token",
			mockStatus:     http.StatusOK,
			wantErr:        false,
			wantDriveCount: 1,
		},
		{
			testName:   "empty list",
			pageSize:   10,
			mockDrives: []*drive.Drive{},
			nextToken:  "",
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:   "API error",
			pageSize:   10,
			mockDrives: nil,
			mockStatus: http.StatusForbidden,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/drive/v3/drives" {
					t.Fatalf("Expected path /drive/v3/drives, got %s", r.URL.Path)
				}

				query := r.URL.Query()
				if tt.pageSize > 0 {
					requireQueryValue(t, query, "pageSize", fmt.Sprintf("%d", tt.pageSize))
				}

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "forbidden", "access denied")
					return
				}
				writeJSON(t, w, http.StatusOK, &drive.DriveList{
					Drives:        tt.mockDrives,
					NextPageToken: tt.nextToken,
				})
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
			result, err := manager.List(context.Background(), reqCtx, tt.pageSize, tt.pageToken)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if len(result.Drives) != tt.wantDriveCount {
				t.Errorf("Drive count mismatch: got %d, want %d", len(result.Drives), tt.wantDriveCount)
			}
		})
	}
}

// TestManager_Get tests getting a shared drive by ID
func TestManager_Get(t *testing.T) {
	tests := []struct {
		testName   string
		driveID    string
		fields     string
		mockDrive  *drive.Drive
		mockStatus int
		wantErr    bool
	}{
		{
			testName: "get shared drive with full metadata",
			driveID:  "drive123",
			fields:   "",
			mockDrive: &drive.Drive{
				Id:                  "drive123",
				Name:                "Test Drive",
				Kind:                "drive#drive",
				ColorRgb:            "#ff0000",
				BackgroundImageLink: "https://example.com/bg.jpg",
				ThemeId:             "theme1",
				CreatedTime:         "2024-01-01T00:00:00.000Z",
				Hidden:              false,
				OrgUnitId:           "org123",
				Capabilities: &drive.DriveCapabilities{
					CanAddChildren:   true,
					CanManageMembers: true,
					CanDeleteDrive:   false,
				},
				Restrictions: &drive.DriveRestrictions{
					AdminManagedRestrictions: true,
					DomainUsersOnly:          false,
				},
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName: "get shared drive with custom fields",
			driveID:  "drive456",
			fields:   "id,name",
			mockDrive: &drive.Drive{
				Id:   "drive456",
				Name: "Custom Drive",
			},
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			testName:   "drive not found",
			driveID:    "missing",
			fields:     "",
			mockDrive:  nil,
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
				expectedPath := fmt.Sprintf("/drive/v3/drives/%s", tt.driveID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				if tt.fields != "" {
					fields := query.Get("fields")
					if !strings.Contains(fields, tt.fields) {
						t.Fatalf("Expected fields to include %s, got %s", tt.fields, fields)
					}
				}

				if tt.mockStatus != http.StatusOK {
					writeAPIError(t, w, tt.mockStatus, "notFound", "drive not found")
					return
				}
				writeJSON(t, w, http.StatusOK, tt.mockDrive)
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
			result, err := manager.Get(context.Background(), reqCtx, tt.driveID, tt.fields)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.mockDrive.Id {
				t.Errorf("ID mismatch: got %s, want %s", result.ID, tt.mockDrive.Id)
			}
			if result.Name != tt.mockDrive.Name {
				t.Errorf("Name mismatch: got %s, want %s", result.Name, tt.mockDrive.Name)
			}
		})
	}
}

// TestManager_GetDriveRootID tests getting drive root ID
func TestManager_GetDriveRootID(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	manager, _ := newTestManager(t, handler)

	driveID := "drive123"
	rootID := manager.GetDriveRootID(driveID)

	if rootID != driveID {
		t.Errorf("Expected root ID to equal drive ID: got %s, want %s", rootID, driveID)
	}
}

// TestMapDriveToSharedDrive tests drive conversion
func TestMapDriveToSharedDrive(t *testing.T) {
	tests := []struct {
		testName string
		input    *drive.Drive
		validate func(*testing.T, *SharedDrive)
	}{
		{
			testName: "full drive conversion",
			input: &drive.Drive{
				Id:                  "drive123",
				Name:                "Full Drive",
				Kind:                "drive#drive",
				ColorRgb:            "#00ff00",
				BackgroundImageLink: "https://example.com/image.png",
				ThemeId:             "theme456",
				CreatedTime:         "2024-01-15T12:00:00.000Z",
				Hidden:              true,
				OrgUnitId:           "org456",
				Capabilities: &drive.DriveCapabilities{
					CanAddChildren:   true,
					CanManageMembers: true,
					CanDeleteDrive:   false,
					CanChangeCopyRequiresWriterPermissionRestriction: true,
					CanChangeDomainUsersOnlyRestriction:              false,
					CanChangeDriveMembersOnlyRestriction:             true,
					CanTrashChildren:                                 true,
				},
				Restrictions: &drive.DriveRestrictions{
					AdminManagedRestrictions:     true,
					CopyRequiresWriterPermission: false,
					DomainUsersOnly:              false,
					DriveMembersOnly:             true,
				},
			},
			validate: func(t *testing.T, d *SharedDrive) {
				if d.ID != "drive123" {
					t.Errorf("ID mismatch: got %s", d.ID)
				}
				if d.Name != "Full Drive" {
					t.Errorf("Name mismatch: got %s", d.Name)
				}
				if d.Kind != "drive#drive" {
					t.Errorf("Kind mismatch: got %s", d.Kind)
				}
				if d.ColorRgb != "#00ff00" {
					t.Errorf("ColorRgb mismatch: got %s", d.ColorRgb)
				}
				if d.Capabilities == nil {
					t.Fatal("Expected capabilities")
				}
				if !d.Capabilities.CanAddChildren {
					t.Error("Expected CanAddChildren to be true")
				}
				if d.Capabilities.CanDeleteDrive {
					t.Error("Expected CanDeleteDrive to be false")
				}
				if d.Restrictions == nil {
					t.Fatal("Expected restrictions")
				}
				if !d.Restrictions.AdminManagedRestrictions {
					t.Error("Expected AdminManagedRestrictions to be true")
				}
			},
		},
		{
			testName: "minimal drive conversion",
			input: &drive.Drive{
				Id:   "drive789",
				Name: "Minimal Drive",
			},
			validate: func(t *testing.T, d *SharedDrive) {
				if d.ID != "drive789" {
					t.Errorf("ID mismatch: got %s", d.ID)
				}
				if d.Name != "Minimal Drive" {
					t.Errorf("Name mismatch: got %s", d.Name)
				}
				if d.Capabilities != nil {
					t.Error("Expected nil capabilities for minimal drive")
				}
				if d.Restrictions != nil {
					t.Error("Expected nil restrictions for minimal drive")
				}
			},
		},
		{
			testName: "hidden drive",
			input: &drive.Drive{
				Id:     "hidden-drive",
				Name:   "Hidden Drive",
				Hidden: true,
			},
			validate: func(t *testing.T, d *SharedDrive) {
				if !d.Hidden {
					t.Error("Expected Hidden to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := mapDriveToSharedDrive(tt.input)
			tt.validate(t, result)
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
			testName: "drive not found",
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
		{
			testName: "rate limit exceeded",
			status:   http.StatusTooManyRequests,
			reason:   "rateLimitExceeded",
			wantCode: utils.ErrCodeRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				writeAPIError(t, w, tt.status, tt.reason, "test error")
			}

			manager, _ := newTestManager(t, handler)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
			_, err := manager.Get(context.Background(), reqCtx, "drive123", "")
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
		writeJSON(t, w, http.StatusOK, &drive.Drive{
			Id:   "drive123",
			Name: "Test Drive",
		})
	}

	_, service := mockServer(t, handler)
	client := api.NewClient(service, 3, 100, logging.NewNoOpLogger())
	manager := NewManager(client)

	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	result, err := manager.Get(context.Background(), reqCtx, "drive123", "")

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

// TestSharedDriveStructures tests struct fields
func TestSharedDriveStructures(t *testing.T) {
	t.Run("SharedDrive struct fields", func(t *testing.T) {
		drive := &SharedDrive{
			ID:                  "drive123",
			Name:                "Test Drive",
			Kind:                "drive#drive",
			ColorRgb:            "#ff0000",
			BackgroundImageLink: "https://example.com/bg.jpg",
			ThemeID:             "theme1",
			CreatedTime:         "2024-01-01T00:00:00Z",
			Hidden:              false,
			OrgUnitID:           "org123",
		}

		if drive.ID != "drive123" {
			t.Errorf("ID mismatch")
		}
		if drive.Name != "Test Drive" {
			t.Errorf("Name mismatch")
		}
	})

	t.Run("DriveCapabilities struct fields", func(t *testing.T) {
		caps := &DriveCapabilities{
			CanAddChildren:            true,
			CanManageMembers:          true,
			CanEdit:                   true,
			CanShare:                  false,
			CanDeleteDrive:            false,
			CanRenameDrive:            true,
			CanListChildren:           true,
			CanDownload:               true,
			CanCopy:                   true,
			CanComment:                true,
			CanReadRevisions:          true,
			CanRename:                 true,
			CanChangeDriveBackground:  true,
			CanResetDriveRestrictions: false,
		}

		if !caps.CanAddChildren {
			t.Error("CanAddChildren should be true")
		}
		if caps.CanShare {
			t.Error("CanShare should be false")
		}
	})

	t.Run("DriveRestrictions struct fields", func(t *testing.T) {
		restrictions := &DriveRestrictions{
			AdminManagedRestrictions:     true,
			CopyRequiresWriterPermission: false,
			DomainUsersOnly:              true,
			DriveMembersOnly:             false,
		}

		if !restrictions.AdminManagedRestrictions {
			t.Error("AdminManagedRestrictions should be true")
		}
		if restrictions.CopyRequiresWriterPermission {
			t.Error("CopyRequiresWriterPermission should be false")
		}
	})

	t.Run("ListResult struct fields", func(t *testing.T) {
		result := &ListResult{
			Drives: []*SharedDrive{
				{ID: "drive1", Name: "Drive 1"},
				{ID: "drive2", Name: "Drive 2"},
			},
			NextPageToken: "token123",
		}

		if len(result.Drives) != 2 {
			t.Errorf("Expected 2 drives, got %d", len(result.Drives))
		}
		if result.NextPageToken != "token123" {
			t.Errorf("NextPageToken mismatch")
		}
	})
}

// Benchmark tests
func BenchmarkMapDriveToSharedDrive(b *testing.B) {
	apiDrive := &drive.Drive{
		Id:       "drive123",
		Name:     "Benchmark Drive",
		Kind:     "drive#drive",
		ColorRgb: "#0000ff",
		Capabilities: &drive.DriveCapabilities{
			CanAddChildren:   true,
			CanManageMembers: true,
		},
		Restrictions: &drive.DriveRestrictions{
			AdminManagedRestrictions: true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapDriveToSharedDrive(apiDrive)
	}
}
