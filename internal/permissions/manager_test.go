package permissions

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
	"sync/atomic"
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/logging"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type apiErrorResponse struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Errors  []apiErrorEntry `json:"errors"`
}

type apiErrorEntry struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func newTestManager(t *testing.T, handler http.HandlerFunc, maxRetries int) (*Manager, *api.Client, *httptest.Server) {
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

	client := api.NewClient(service, maxRetries, 1, logging.NewNoOpLogger())
	manager := NewManager(client)
	return manager, client, server
}

func decodeJSONBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}
	_ = r.Body.Close()

	if len(body) == 0 {
		return map[string]interface{}{}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("Failed to decode request body: %v", err)
	}
	return payload
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, payload interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("Failed to encode response: %v", err)
	}
}

func writeAPIError(t *testing.T, w http.ResponseWriter, status int, reason, message string) {
	t.Helper()
	if message == "" {
		message = "test error"
	}
	resp := apiErrorResponse{
		Error: apiErrorDetail{
			Code:    status,
			Message: message,
			Errors: []apiErrorEntry{{
				Domain:  "global",
				Reason:  reason,
				Message: message,
			}},
		},
	}
	writeJSON(t, w, status, resp)
}

func requireQueryValue(t *testing.T, query url.Values, key, want string) {
	t.Helper()
	if got := query.Get(key); got != want {
		t.Fatalf("Expected query %s=%q, got %q", key, want, got)
	}
}

func requireQueryMissing(t *testing.T, query url.Values, key string) {
	t.Helper()
	if got := query.Get(key); got != "" {
		t.Fatalf("Expected query %s to be empty, got %q", key, got)
	}
}

func requirePayloadString(t *testing.T, payload map[string]interface{}, key, want string) {
	t.Helper()
	value, ok := payload[key]
	if !ok {
		t.Fatalf("Expected payload to include %s", key)
	}
	got, ok := value.(string)
	if !ok {
		t.Fatalf("Expected payload %s to be string, got %T", key, value)
	}
	if got != want {
		t.Fatalf("Expected payload %s=%q, got %q", key, want, got)
	}
}

func requirePayloadBool(t *testing.T, payload map[string]interface{}, key string, want bool) {
	t.Helper()
	value, ok := payload[key]
	if !ok {
		t.Fatalf("Expected payload to include %s", key)
	}
	got, ok := value.(bool)
	if !ok {
		t.Fatalf("Expected payload %s to be bool, got %T", key, value)
	}
	if got != want {
		t.Fatalf("Expected payload %s=%v, got %v", key, want, got)
	}
}

func assertAppErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error with code %s", wantCode)
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Expected AppError, got %T", err)
	}
	if appErr.CLIError.Code != wantCode {
		t.Fatalf("Expected error code %s, got %s", wantCode, appErr.CLIError.Code)
	}
}

// Test permission creation
func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		fileID  string
		opts    CreateOptions
		want    *types.Permission
		wantErr bool
	}{
		{
			name:   "create user permission",
			fileID: "file123",
			opts: CreateOptions{
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
			},
			want: &types.Permission{
				ID:           "perm123",
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
		{
			name:   "create group permission",
			fileID: "file123",
			opts: CreateOptions{
				Type:         "group",
				Role:         "writer",
				EmailAddress: "group@example.com",
			},
			want: &types.Permission{
				ID:           "perm124",
				Type:         "group",
				Role:         "writer",
				EmailAddress: "group@example.com",
			},
			wantErr: false,
		},
		{
			name:   "create domain permission",
			fileID: "file123",
			opts: CreateOptions{
				Type:   "domain",
				Role:   "reader",
				Domain: "example.com",
			},
			want: &types.Permission{
				ID:     "perm125",
				Type:   "domain",
				Role:   "reader",
				Domain: "example.com",
			},
			wantErr: false,
		},
		{
			name:   "create public link with discovery",
			fileID: "file123",
			opts: CreateOptions{
				Type:               "anyone",
				Role:               "reader",
				AllowFileDiscovery: true,
			},
			want: &types.Permission{
				ID:   "perm126",
				Type: "anyone",
				Role: "reader",
			},
			wantErr: false,
		},
		{
			name:   "transfer ownership",
			fileID: "file123",
			opts: CreateOptions{
				Type:              "user",
				Role:              "owner",
				EmailAddress:      "newowner@example.com",
				TransferOwnership: true,
			},
			want: &types.Permission{
				ID:           "perm127",
				Type:         "user",
				Role:         "owner",
				EmailAddress: "newowner@example.com",
			},
			wantErr: false,
		},
		{
			name:   "create with notification",
			fileID: "file123",
			opts: CreateOptions{
				Type:                  "user",
				Role:                  "writer",
				EmailAddress:          "user@example.com",
				SendNotificationEmail: true,
				EmailMessage:          "Shared with you",
			},
			want: &types.Permission{
				ID:           "perm128",
				Type:         "user",
				Role:         "writer",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
		{
			name:   "create with domain admin access",
			fileID: "file123",
			opts: CreateOptions{
				Type:                 "user",
				Role:                 "reader",
				EmailAddress:         "user@example.com",
				UseDomainAdminAccess: true,
			},
			want: &types.Permission{
				ID:           "perm129",
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("Expected POST, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/drive/v3/files/%s/permissions", tt.fileID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")
				fields := query.Get("fields")
				if !strings.Contains(fields, "id,type,role") {
					t.Fatalf("Expected fields to include permission fields, got %q", fields)
				}
				if tt.opts.SendNotificationEmail {
					requireQueryValue(t, query, "sendNotificationEmail", "true")
				}
				if tt.opts.EmailMessage != "" {
					requireQueryValue(t, query, "emailMessage", tt.opts.EmailMessage)
				}
				if tt.opts.TransferOwnership {
					requireQueryValue(t, query, "transferOwnership", "true")
				}
				if tt.opts.UseDomainAdminAccess {
					requireQueryValue(t, query, "useDomainAdminAccess", "true")
				}

				payload := decodeJSONBody(t, r)
				requirePayloadString(t, payload, "type", tt.opts.Type)
				requirePayloadString(t, payload, "role", tt.opts.Role)
				if tt.opts.EmailAddress != "" {
					requirePayloadString(t, payload, "emailAddress", tt.opts.EmailAddress)
				}
				if tt.opts.Domain != "" {
					requirePayloadString(t, payload, "domain", tt.opts.Domain)
				}
				if tt.opts.Type == "anyone" && tt.opts.AllowFileDiscovery {
					requirePayloadBool(t, payload, "allowFileDiscovery", true)
				}

				response := &drive.Permission{
					Id:           tt.want.ID,
					Type:         tt.want.Type,
					Role:         tt.want.Role,
					EmailAddress: tt.want.EmailAddress,
					Domain:       tt.want.Domain,
					DisplayName:  tt.want.DisplayName,
				}
				writeJSON(t, w, http.StatusOK, response)
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			result, err := manager.Create(context.Background(), reqCtx, tt.fileID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result == nil {
				t.Fatal("Expected permission result")
			}
			if result.ID != tt.want.ID || result.Type != tt.want.Type || result.Role != tt.want.Role || result.EmailAddress != tt.want.EmailAddress || result.Domain != tt.want.Domain {
				t.Fatalf("Unexpected permission result: %+v", result)
			}
		})
	}
}

// Test permission listing
func TestList(t *testing.T) {
	samplePermissions := []*drive.Permission{
		{Id: "perm1", Type: "user", Role: "reader", EmailAddress: "user@example.com"},
		{Id: "perm2", Type: "group", Role: "writer", EmailAddress: "group@example.com"},
		{Id: "perm3", Type: "anyone", Role: "reader"},
	}

	tests := []struct {
		name    string
		fileID  string
		opts    ListOptions
		want    int
		wantErr bool
	}{
		{
			name:    "list all permissions",
			fileID:  "file123",
			opts:    ListOptions{},
			want:    3,
			wantErr: false,
		},
		{
			name:   "list with domain admin access",
			fileID: "file123",
			opts: ListOptions{
				UseDomainAdminAccess: true,
			},
			want:    3,
			wantErr: false,
		},
		{
			name:   "list with page size",
			fileID: "file123",
			opts: ListOptions{
				PageSize: 10,
			},
			want:    3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/drive/v3/files/%s/permissions", tt.fileID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")
				if tt.opts.UseDomainAdminAccess {
					requireQueryValue(t, query, "useDomainAdminAccess", "true")
				}
				if tt.opts.PageSize > 0 {
					requireQueryValue(t, query, "pageSize", fmt.Sprintf("%d", tt.opts.PageSize))
				}
				fields := query.Get("fields")
				if !strings.Contains(fields, "permissions(") {
					t.Fatalf("Expected fields to include permissions, got %q", fields)
				}

				writeJSON(t, w, http.StatusOK, &drive.PermissionList{Permissions: samplePermissions})
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			result, err := manager.List(context.Background(), reqCtx, tt.fileID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if len(result) != tt.want {
				t.Fatalf("Expected %d permissions, got %d", tt.want, len(result))
			}
		})
	}
}

// Test permission update
func TestUpdate(t *testing.T) {
	tests := []struct {
		name         string
		fileID       string
		permissionID string
		opts         UpdateOptions
		want         *types.Permission
		wantErr      bool
	}{
		{
			name:         "update role",
			fileID:       "file123",
			permissionID: "perm123",
			opts: UpdateOptions{
				Role: "writer",
			},
			want: &types.Permission{
				ID:   "perm123",
				Type: "user",
				Role: "writer",
			},
			wantErr: false,
		},
		{
			name:         "update with domain admin access",
			fileID:       "file123",
			permissionID: "perm123",
			opts: UpdateOptions{
				Role:                 "reader",
				UseDomainAdminAccess: true,
			},
			want: &types.Permission{
				ID:   "perm123",
				Type: "user",
				Role: "reader",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Fatalf("Expected PATCH, got %s", r.Method)
				}
				expectedPath := fmt.Sprintf("/drive/v3/files/%s/permissions/%s", tt.fileID, tt.permissionID)
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")
				if tt.opts.UseDomainAdminAccess {
					requireQueryValue(t, query, "useDomainAdminAccess", "true")
				}
				fields := query.Get("fields")
				if !strings.Contains(fields, "id,type,role") {
					t.Fatalf("Expected fields to include permission fields, got %q", fields)
				}

				payload := decodeJSONBody(t, r)
				requirePayloadString(t, payload, "role", tt.opts.Role)

				response := &drive.Permission{
					Id:           tt.want.ID,
					Type:         tt.want.Type,
					Role:         tt.want.Role,
					EmailAddress: tt.want.EmailAddress,
					Domain:       tt.want.Domain,
				}
				writeJSON(t, w, http.StatusOK, response)
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			result, err := manager.Update(context.Background(), reqCtx, tt.fileID, tt.permissionID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.want.ID || result.Type != tt.want.Type || result.Role != tt.want.Role {
				t.Fatalf("Unexpected permission result: %+v", result)
			}
		})
	}
}

// Test permission deletion
func TestDelete(t *testing.T) {
	tests := []struct {
		name         string
		fileID       string
		permissionID string
		opts         DeleteOptions
		wantErr      bool
	}{
		{
			name:         "delete permission",
			fileID:       "file123",
			permissionID: "perm123",
			opts:         DeleteOptions{},
			wantErr:      false,
		},
		{
			name:         "delete with domain admin access",
			fileID:       "file123",
			permissionID: "perm123",
			opts: DeleteOptions{
				UseDomainAdminAccess: true,
			},
			wantErr: false,
		},
		{
			name:         "delete non-existent permission",
			fileID:       "file123",
			permissionID: "invalid",
			opts:         DeleteOptions{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deleteCalls int32
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != fmt.Sprintf("/drive/v3/files/%s/permissions/%s", tt.fileID, tt.permissionID) {
					t.Fatalf("Unexpected path %s", r.URL.Path)
				}

				switch r.Method {
				case http.MethodGet:
					if tt.wantErr {
						writeAPIError(t, w, http.StatusNotFound, "notFound", "permission not found")
						return
					}
					writeJSON(t, w, http.StatusOK, &drive.Permission{Id: tt.permissionID, Type: "user", Role: "reader", EmailAddress: "user@example.com"})
				case http.MethodDelete:
					atomic.AddInt32(&deleteCalls, 1)
					query := r.URL.Query()
					requireQueryValue(t, query, "supportsAllDrives", "true")
					if tt.opts.UseDomainAdminAccess {
						requireQueryValue(t, query, "useDomainAdminAccess", "true")
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Fatalf("Unexpected method %s", r.Method)
				}
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			err := manager.Delete(context.Background(), reqCtx, tt.fileID, tt.permissionID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				if deleteCalls > 0 {
					t.Fatalf("Did not expect delete call when get failed")
				}
				return
			}
			if deleteCalls == 0 {
				t.Fatalf("Expected delete call")
			}
		})
	}
}

// Test public link creation
func TestCreatePublicLink(t *testing.T) {
	tests := []struct {
		name           string
		fileID         string
		role           string
		allowDiscovery bool
		want           *types.Permission
		wantErr        bool
	}{
		{
			name:           "create public link reader",
			fileID:         "file123",
			role:           "reader",
			allowDiscovery: false,
			want: &types.Permission{
				ID:   "anyoneWithLink",
				Type: "anyone",
				Role: "reader",
			},
			wantErr: false,
		},
		{
			name:           "create public link with discovery",
			fileID:         "file123",
			role:           "reader",
			allowDiscovery: true,
			want: &types.Permission{
				ID:   "anyoneWithLink",
				Type: "anyone",
				Role: "reader",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("Expected POST, got %s", r.Method)
				}
				payload := decodeJSONBody(t, r)
				requirePayloadString(t, payload, "type", "anyone")
				requirePayloadString(t, payload, "role", tt.role)
				if tt.allowDiscovery {
					requirePayloadBool(t, payload, "allowFileDiscovery", true)
				}
				writeJSON(t, w, http.StatusOK, &drive.Permission{Id: tt.want.ID, Type: tt.want.Type, Role: tt.want.Role})
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			result, err := manager.CreatePublicLink(context.Background(), reqCtx, tt.fileID, tt.role, tt.allowDiscovery)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.ID != tt.want.ID || result.Type != tt.want.Type || result.Role != tt.want.Role {
				t.Fatalf("Unexpected permission result: %+v", result)
			}
		})
	}
}

// Test error handling for policy violations
func TestPolicyViolationErrors(t *testing.T) {
	tests := []struct {
		name        string
		apiStatus   int
		reason      string
		wantErrCode string
	}{
		{
			name:        "sharing restricted by policy",
			apiStatus:   http.StatusForbidden,
			reason:      "domainPolicy",
			wantErrCode: utils.ErrCodePolicyViolation,
		},
		{
			name:        "invalid sharing request",
			apiStatus:   http.StatusBadRequest,
			reason:      "invalidSharingRequest",
			wantErrCode: utils.ErrCodeSharingRestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				writeAPIError(t, w, tt.apiStatus, tt.reason, "policy error")
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			_, err := manager.Create(context.Background(), reqCtx, "file123", CreateOptions{
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
			})
			assertAppErrorCode(t, err, tt.wantErrCode)
		})
	}
}

// Test Shared Drive permission handling
func TestSharedDrivePermissions(t *testing.T) {
	tests := []struct {
		name    string
		fileID  string
		driveID string
		opts    CreateOptions
		wantErr bool
	}{
		{
			name:    "create permission on shared drive file",
			fileID:  "file123",
			driveID: "drive456",
			opts: CreateOptions{
				Type:         "user",
				Role:         "writer",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
		{
			name:    "organizer role on shared drive",
			fileID:  "file123",
			driveID: "drive456",
			opts: CreateOptions{
				Type:         "user",
				Role:         "organizer",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()
				requireQueryValue(t, query, "supportsAllDrives", "true")
				writeJSON(t, w, http.StatusOK, &drive.Permission{Id: "perm123", Type: tt.opts.Type, Role: tt.opts.Role, EmailAddress: tt.opts.EmailAddress})
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", tt.driveID, types.RequestTypePermissionOp)
			result, err := manager.Create(context.Background(), reqCtx, tt.fileID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Expected error=%v, got %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result == nil || result.ID == "" {
				t.Fatalf("Expected permission result")
			}
		})
	}
}

// Test permission type validation
func TestPermissionTypeValidation(t *testing.T) {
	validTypes := []string{"user", "group", "domain", "anyone"}
	validRoles := []string{"reader", "commenter", "writer", "organizer", "owner"}

	t.Run("valid permission types", func(t *testing.T) {
		for _, permType := range validTypes {
			_ = CreateOptions{
				Type: permType,
				Role: "reader",
			}
		}
	})

	t.Run("valid permission roles", func(t *testing.T) {
		for _, role := range validRoles {
			_ = CreateOptions{
				Type: "user",
				Role: role,
			}
		}
	})
}

// Test ownership transfer validation
func TestOwnershipTransfer(t *testing.T) {
	tests := []struct {
		name    string
		opts    CreateOptions
		wantErr bool
	}{
		{
			name: "valid ownership transfer",
			opts: CreateOptions{
				Type:              "user",
				Role:              "owner",
				EmailAddress:      "newowner@example.com",
				TransferOwnership: true,
			},
			wantErr: false,
		},
		{
			name: "ownership transfer on shared drive",
			opts: CreateOptions{
				Type:              "user",
				Role:              "organizer", // Shared Drives don't have owners
				EmailAddress:      "organizer@example.com",
				TransferOwnership: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the structure, actual behavior tests require mocking
			if tt.opts.Type == "" || tt.opts.Role == "" {
				t.Error("Invalid permission options")
			}
		})
	}
}

// Test resource key handling in permissions
func TestResourceKeyHandling(t *testing.T) {
	t.Run("permissions with resource keys", func(t *testing.T) {
		fileID := "file123"
		permissionID := "perm123"
		resourceKey := "resource-key-1"

		handler := func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("X-Goog-Drive-Resource-Keys")
			expected := fmt.Sprintf("%s/%s", fileID, resourceKey)
			if header != expected {
				t.Fatalf("Expected resource key header %q, got %q", expected, header)
			}
			writeJSON(t, w, http.StatusOK, &drive.Permission{Id: permissionID, Type: "user", Role: "reader"})
		}

		manager, client, _ := newTestManager(t, handler, 0)
		client.ResourceKeys().AddKey(fileID, resourceKey, "api")

		reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
		_, err := manager.Get(context.Background(), reqCtx, fileID, permissionID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

// Test pagination in permission listing
func TestPermissionPagination(t *testing.T) {
	t.Run("multiple pages", func(t *testing.T) {
		fileID := "file123"
		var callCount int32

		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("Expected GET, got %s", r.Method)
			}
			expectedPath := fmt.Sprintf("/drive/v3/files/%s/permissions", fileID)
			if r.URL.Path != expectedPath {
				t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}

			current := atomic.AddInt32(&callCount, 1)
			if current == 1 {
				requireQueryMissing(t, r.URL.Query(), "pageToken")
				writeJSON(t, w, http.StatusOK, &drive.PermissionList{
					Permissions:   []*drive.Permission{{Id: "perm1", Type: "user", Role: "reader"}},
					NextPageToken: "next-page",
				})
				return
			}

			requireQueryValue(t, r.URL.Query(), "pageToken", "next-page")
			writeJSON(t, w, http.StatusOK, &drive.PermissionList{
				Permissions: []*drive.Permission{{Id: "perm2", Type: "user", Role: "writer"}},
			})
		}

		manager, _, _ := newTestManager(t, handler, 0)
		reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
		perms, err := manager.List(context.Background(), reqCtx, fileID, ListOptions{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(perms) != 2 {
			t.Fatalf("Expected 2 permissions, got %d", len(perms))
		}
		if callCount != 2 {
			t.Fatalf("Expected 2 API calls, got %d", callCount)
		}
	})
}

// Test error scenarios
func TestErrorScenarios(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		reason   string
		wantCode string
		fileID   string
		permID   string
	}{
		{
			name:     "file not found",
			status:   http.StatusNotFound,
			reason:   "notFound",
			wantCode: utils.ErrCodeFileNotFound,
			fileID:   "missing-file",
			permID:   "perm123",
		},
		{
			name:     "permission not found",
			status:   http.StatusNotFound,
			reason:   "notFound",
			wantCode: utils.ErrCodeFileNotFound,
			fileID:   "file123",
			permID:   "missing-perm",
		},
		{
			name:     "insufficient permissions",
			status:   http.StatusForbidden,
			reason:   "insufficientFilePermissions",
			wantCode: utils.ErrCodePermissionDenied,
			fileID:   "file123",
			permID:   "perm123",
		},
		{
			name:     "rate limit exceeded",
			status:   http.StatusTooManyRequests,
			reason:   "rateLimitExceeded",
			wantCode: utils.ErrCodeRateLimited,
			fileID:   "file123",
			permID:   "perm123",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				writeAPIError(t, w, tt.status, tt.reason, "test error")
			}

			manager, _, _ := newTestManager(t, handler, 0)
			reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
			_, err := manager.Get(context.Background(), reqCtx, tt.fileID, tt.permID)
			assertAppErrorCode(t, err, tt.wantCode)
		})
	}
}

// TestConvertPermission tests the permission conversion function
func TestConvertPermission(t *testing.T) {
	tests := []struct {
		name  string
		input *drive.Permission
		want  *types.Permission
	}{
		{
			name: "user permission",
			input: &drive.Permission{
				Id:           "perm123",
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
				DisplayName:  "Test User",
			},
			want: &types.Permission{
				ID:           "perm123",
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
				DisplayName:  "Test User",
			},
		},
		{
			name: "domain permission",
			input: &drive.Permission{
				Id:     "perm124",
				Type:   "domain",
				Role:   "reader",
				Domain: "example.com",
			},
			want: &types.Permission{
				ID:     "perm124",
				Type:   "domain",
				Role:   "reader",
				Domain: "example.com",
			},
		},
		{
			name: "anyone permission",
			input: &drive.Permission{
				Id:   "anyoneWithLink",
				Type: "anyone",
				Role: "reader",
			},
			want: &types.Permission{
				ID:   "anyoneWithLink",
				Type: "anyone",
				Role: "reader",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPermission(tt.input)
			if got.ID != tt.want.ID ||
				got.Type != tt.want.Type ||
				got.Role != tt.want.Role ||
				got.EmailAddress != tt.want.EmailAddress ||
				got.Domain != tt.want.Domain ||
				got.DisplayName != tt.want.DisplayName {
				t.Errorf("convertPermission() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkConvertPermission(b *testing.B) {
	perm := &drive.Permission{
		Id:           "perm123",
		Type:         "user",
		Role:         "reader",
		EmailAddress: "user@example.com",
		DisplayName:  "Test User",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertPermission(perm)
	}
}

// Integration test helpers
func TestIntegrationPermissionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("full permission lifecycle", func(t *testing.T) {
		t.Skip("Skipping - requires actual Drive API connection")
		// TODO: Implement integration test that:
		// 1. Creates a test file
		// 2. Adds various permission types
		// 3. Lists permissions
		// 4. Updates permissions
		// 5. Removes permissions
		// 6. Cleans up
	})
}

// Example test for documentation
func ExampleManager_Create() {
	// This example shows how to create a permission
	// Note: Requires proper setup with Drive API client

	// ctx := context.Background()
	// manager := NewManager(client)
	// reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
	//
	// opts := CreateOptions{
	// 	Type:         "user",
	// 	Role:         "reader",
	// 	EmailAddress: "user@example.com",
	// }
	//
	// perm, err := manager.Create(ctx, reqCtx, "fileId123", opts)
	// if err != nil {
	// 	// handle error
	// }
	// _ = perm
}

// Test context cancellation handling
func TestContextCancellation(t *testing.T) {
	t.Run("create with cancelled context", func(t *testing.T) {
		fileID := "file123"
		handler := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			writeJSON(t, w, http.StatusOK, &drive.Permission{Id: "perm123", Type: "user", Role: "reader"})
		}

		manager, _, _ := newTestManager(t, handler, 0)
		reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		_, err := manager.Create(ctx, reqCtx, fileID, CreateOptions{Type: "user", Role: "reader", EmailAddress: "user@example.com"})
		if err == nil {
			t.Fatal("Expected error from cancelled context")
		}
		assertAppErrorCode(t, err, utils.ErrCodeNetworkError)
		if !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("Expected context cancellation error, got %v", err)
		}
	})
}

// Test retry logic for permissions
func TestRetryLogic(t *testing.T) {
	t.Run("retry on rate limit", func(t *testing.T) {
		fileID := "file123"
		permissionID := "perm123"
		var callCount int32

		handler := func(w http.ResponseWriter, r *http.Request) {
			current := atomic.AddInt32(&callCount, 1)
			if current == 1 {
				writeAPIError(t, w, http.StatusTooManyRequests, "rateLimitExceeded", "rate limit")
				return
			}
			writeJSON(t, w, http.StatusOK, &drive.Permission{Id: permissionID, Type: "user", Role: "reader"})
		}

		manager, _, _ := newTestManager(t, handler, 1)
		reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
		_, err := manager.Get(context.Background(), reqCtx, fileID, permissionID)
		if err != nil {
			t.Fatalf("Expected retry to succeed, got %v", err)
		}
		if callCount != 2 {
			t.Fatalf("Expected 2 calls, got %d", callCount)
		}
	})

	t.Run("no retry on 400 errors", func(t *testing.T) {
		fileID := "file123"
		permissionID := "perm123"
		var callCount int32

		handler := func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			writeAPIError(t, w, http.StatusBadRequest, "badRequest", "invalid")
		}

		manager, _, _ := newTestManager(t, handler, 1)
		reqCtx := api.NewRequestContext("default", "", types.RequestTypePermissionOp)
		_, err := manager.Get(context.Background(), reqCtx, fileID, permissionID)
		assertAppErrorCode(t, err, utils.ErrCodeInvalidArgument)
		if callCount != 1 {
			t.Fatalf("Expected 1 call, got %d", callCount)
		}
	})
}

// Utility function to create test error
// Test helper for validating permission options
func validateCreateOptions(opts CreateOptions) error {
	validTypes := map[string]bool{
		"user": true, "group": true, "domain": true, "anyone": true,
	}
	validRoles := map[string]bool{
		"reader": true, "commenter": true, "writer": true, "organizer": true, "owner": true,
	}

	if !validTypes[opts.Type] {
		return errors.New("invalid permission type")
	}
	if !validRoles[opts.Role] {
		return errors.New("invalid permission role")
	}

	// Validate email for user/group types
	if (opts.Type == "user" || opts.Type == "group") && opts.EmailAddress == "" {
		return errors.New("email address required for user/group permissions")
	}

	// Validate domain for domain type
	if opts.Type == "domain" && opts.Domain == "" {
		return errors.New("domain required for domain permissions")
	}

	return nil
}

func TestValidateCreateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    CreateOptions
		wantErr bool
	}{
		{
			name: "valid user permission",
			opts: CreateOptions{
				Type:         "user",
				Role:         "reader",
				EmailAddress: "user@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			opts: CreateOptions{
				Type: "invalid",
				Role: "reader",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			opts: CreateOptions{
				Type: "user",
				Role: "invalid",
			},
			wantErr: true,
		},
		{
			name: "user without email",
			opts: CreateOptions{
				Type: "user",
				Role: "reader",
			},
			wantErr: true,
		},
		{
			name: "domain without domain field",
			opts: CreateOptions{
				Type: "domain",
				Role: "reader",
			},
			wantErr: true,
		},
		{
			name: "valid anyone permission",
			opts: CreateOptions{
				Type: "anyone",
				Role: "reader",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test CreateOptions structure and defaults
func TestCreateOptions_Struct(t *testing.T) {
	opts := CreateOptions{
		Type:                  "user",
		Role:                  "reader",
		EmailAddress:          "user@example.com",
		Domain:                "",
		SendNotificationEmail: true,
		EmailMessage:          "Shared with you",
		TransferOwnership:     false,
		AllowFileDiscovery:    false,
		UseDomainAdminAccess:  false,
	}

	if opts.Type != "user" {
		t.Errorf("Type = %s, want user", opts.Type)
	}
	if opts.Role != "reader" {
		t.Errorf("Role = %s, want reader", opts.Role)
	}
	if opts.EmailAddress != "user@example.com" {
		t.Errorf("EmailAddress = %s, want user@example.com", opts.EmailAddress)
	}
	if !opts.SendNotificationEmail {
		t.Error("SendNotificationEmail should be true")
	}
	if opts.EmailMessage != "Shared with you" {
		t.Errorf("EmailMessage = %s, want 'Shared with you'", opts.EmailMessage)
	}
}

// Test UpdateOptions structure
func TestUpdateOptions_Struct(t *testing.T) {
	opts := UpdateOptions{
		Role:                 "writer",
		UseDomainAdminAccess: true,
	}

	if opts.Role != "writer" {
		t.Errorf("Role = %s, want writer", opts.Role)
	}
	if !opts.UseDomainAdminAccess {
		t.Error("UseDomainAdminAccess should be true")
	}
}

// Test DeleteOptions structure
func TestDeleteOptions_Struct(t *testing.T) {
	opts := DeleteOptions{
		UseDomainAdminAccess: true,
	}

	if !opts.UseDomainAdminAccess {
		t.Error("UseDomainAdminAccess should be true")
	}
}

// Test ListOptions structure
func TestListOptions_Struct(t *testing.T) {
	opts := ListOptions{
		UseDomainAdminAccess: true,
		PageSize:             50,
	}

	if !opts.UseDomainAdminAccess {
		t.Error("UseDomainAdminAccess should be true")
	}
	if opts.PageSize != 50 {
		t.Errorf("PageSize = %d, want 50", opts.PageSize)
	}
}
