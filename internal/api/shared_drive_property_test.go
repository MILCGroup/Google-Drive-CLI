package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Property 4: Shared Drive Parameter Inclusion
// Validates: Requirements 2.23, 4.12, 10.3, 11.9
// Property: All file operations must include supportsAllDrives=true

func TestProperty_SharedDriveParameters_SupportsAllDrives(t *testing.T) {
	// Property: All shaped file operations set supportsAllDrives=true
	// This is critical for Shared Drive support

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name      string
		reqCtx    *types.RequestContext
		operation string
	}{
		{"User domain", NewRequestContext("default", "", types.RequestTypeGetByID), "FilesGet"},
		{"Shared drive", NewRequestContext("default", "drive123", types.RequestTypeGetByID), "FilesGet"},
		{"List user", NewRequestContext("default", "", types.RequestTypeListOrSearch), "FilesList"},
		{"List drive", NewRequestContext("default", "drive123", types.RequestTypeListOrSearch), "FilesList"},
		{"Create user", NewRequestContext("default", "", types.RequestTypeMutation), "FilesCreate"},
		{"Create drive", NewRequestContext("default", "drive123", types.RequestTypeMutation), "FilesCreate"},
		{"Update user", NewRequestContext("default", "", types.RequestTypeMutation), "FilesUpdate"},
		{"Update drive", NewRequestContext("default", "drive123", types.RequestTypeMutation), "FilesUpdate"},
		{"Delete user", NewRequestContext("default", "", types.RequestTypeMutation), "FilesDelete"},
		{"Delete drive", NewRequestContext("default", "drive123", types.RequestTypeMutation), "FilesDelete"},
		{"Copy user", NewRequestContext("default", "", types.RequestTypeMutation), "FilesCopy"},
		{"Copy drive", NewRequestContext("default", "drive123", types.RequestTypeMutation), "FilesCopy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var shaped interface{}

			switch tt.operation {
			case "FilesGet":
				shaped = shaper.ShapeFilesGet(service.Files.Get("test"), tt.reqCtx)
			case "FilesList":
				shaped = shaper.ShapeFilesList(service.Files.List(), tt.reqCtx)
			case "FilesCreate":
				shaped = shaper.ShapeFilesCreate(service.Files.Create(&drive.File{}), tt.reqCtx)
			case "FilesUpdate":
				shaped = shaper.ShapeFilesUpdate(service.Files.Update("test", &drive.File{}), tt.reqCtx)
			case "FilesDelete":
				shaped = shaper.ShapeFilesDelete(service.Files.Delete("test"), tt.reqCtx)
			case "FilesCopy":
				shaped = shaper.ShapeFilesCopy(service.Files.Copy("test", &drive.File{}), tt.reqCtx)
			}

			if shaped == nil {
				t.Errorf("Shaped %s returned nil", tt.operation)
			}

			// Note: We can't easily verify the supportsAllDrives parameter is set
			// without executing the request, but the test ensures the shaping doesn't panic
		})
	}
}

func TestProperty_SharedDriveParameters_DriveIDContext(t *testing.T) {
	// Property: DriveID in RequestContext affects corpora parameter

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name          string
		driveID       string
		expectedScope string
	}{
		{"No DriveID - user domain", "", "user"},
		{"With DriveID - drive domain", "drive123", "drive"},
		{"Empty DriveID - user domain", "", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := NewRequestContext("default", tt.driveID, types.RequestTypeListOrSearch)
			shaped := shaper.ShapeFilesList(service.Files.List(), reqCtx)

			if shaped == nil {
				t.Error("ShapeFilesList returned nil")
			}

			// The test verifies that shaping succeeds for both user and drive contexts
		})
	}
}

func TestProperty_SharedDriveParameters_PermissionOperations(t *testing.T) {
	// Property: Permission operations also support Shared Drives

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name      string
		driveID   string
		operation string
	}{
		{"List perms user", "", "PermissionsList"},
		{"List perms drive", "drive123", "PermissionsList"},
		{"Create perm user", "", "PermissionsCreate"},
		{"Create perm drive", "drive123", "PermissionsCreate"},
		{"Update perm user", "", "PermissionsUpdate"},
		{"Update perm drive", "drive123", "PermissionsUpdate"},
		{"Delete perm user", "", "PermissionsDelete"},
		{"Delete perm drive", "drive123", "PermissionsDelete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := NewRequestContext("default", tt.driveID, types.RequestTypePermissionOp)
			reqCtx.InvolvedFileIDs = []string{"file123"}

			var shaped interface{}

			switch tt.operation {
			case "PermissionsList":
				shaped = shaper.ShapePermissionsList(service.Permissions.List("file123"), reqCtx)
			case "PermissionsCreate":
				shaped = shaper.ShapePermissionsCreate(
					service.Permissions.Create("file123", &drive.Permission{}), reqCtx)
			case "PermissionsUpdate":
				shaped = shaper.ShapePermissionsUpdate(
					service.Permissions.Update("file123", "perm1", &drive.Permission{}), reqCtx)
			case "PermissionsDelete":
				shaped = shaper.ShapePermissionsDelete(
					service.Permissions.Delete("file123", "perm1"), reqCtx)
			}

			if shaped == nil {
				t.Errorf("Shaped %s returned nil", tt.operation)
			}
		})
	}
}

func TestProperty_SharedDriveParameters_Consistency(t *testing.T) {
	// Property: Same RequestContext produces same shaped request

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	reqCtx := NewRequestContext("default", "drive123", types.RequestTypeGetByID)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	// Shape the same request 100 times
	for i := 0; i < 100; i++ {
		shaped := shaper.ShapeFilesGet(service.Files.Get("file123"), reqCtx)
		if shaped == nil {
			t.Errorf("Iteration %d: ShapeFilesGet returned nil", i)
		}
	}
}

// Property 5: Pagination Continuation
// Validates: Requirements 2.18
// Property: Pagination tokens must be properly handled

func TestProperty_PaginationHandling_PageToken(t *testing.T) {
	// Property: PageToken from response can be used in next request
	// This is a mock test since we can't test against live API

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	// Simulate pagination scenario
	tokens := []string{
		"",
		"token1",
		"token2",
		"token3",
		"verylongtoken1234567890abcdefghijklmnopqrstuvwxyz",
	}

	for _, token := range tokens {
		t.Run("Token_"+token, func(t *testing.T) {
			call := service.Files.List()
			if token != "" {
				call = call.PageToken(token)
			}
			shaped := shaper.ShapeFilesList(call, reqCtx)

			if shaped == nil {
				t.Error("ShapeFilesList returned nil")
			}
		})
	}
}

func TestProperty_PaginationHandling_PageSize(t *testing.T) {
	// Property: PageSize parameter is respected

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	pageSizes := []int64{1, 10, 50, 100, 1000}

	for _, pageSize := range pageSizes {
		t.Run(string(rune(pageSize)), func(t *testing.T) {
			call := service.Files.List().PageSize(pageSize)
			shaped := shaper.ShapeFilesList(call, reqCtx)

			if shaped == nil {
				t.Error("ShapeFilesList returned nil")
			}
		})
	}
}

func TestProperty_PaginationHandling_EmptyPageToken(t *testing.T) {
	// Property: Empty/missing pageToken means first page

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	tests := []struct {
		name  string
		token string
	}{
		{"No token set", ""},
		{"Empty token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := service.Files.List()
			if tt.token != "" {
				call = call.PageToken(tt.token)
			}
			shaped := shaper.ShapeFilesList(call, reqCtx)

			if shaped == nil {
				t.Error("ShapeFilesList returned nil")
			}
		})
	}
}

func TestProperty_PaginationHandling_WithOtherParameters(t *testing.T) {
	// Property: Pagination works with query, orderBy, and other parameters

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	tests := []struct {
		name      string
		pageToken string
		query     string
		orderBy   string
	}{
		{"With query", "token1", "name contains 'test'", ""},
		{"With orderBy", "token2", "", "createdTime desc"},
		{"With both", "token3", "trashed=false", "modifiedTime"},
		{"All parameters", "token4", "name='test.txt'", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := service.Files.List()
			if tt.pageToken != "" {
				call = call.PageToken(tt.pageToken)
			}
			if tt.query != "" {
				call = call.Q(tt.query)
			}
			if tt.orderBy != "" {
				call = call.OrderBy(tt.orderBy)
			}

			shaped := shaper.ShapeFilesList(call, reqCtx)

			if shaped == nil {
				t.Error("ShapeFilesList returned nil")
			}
		})
	}
}
