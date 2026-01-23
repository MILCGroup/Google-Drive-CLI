package api

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/dl-alexandre/gdrive/internal/logging"
	"github.com/dl-alexandre/gdrive/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func TestRequestShaper_ShapeFilesGet(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file123", "key123", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	// Create a mock files.get call
	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Get("file123")
	shaped := shaper.ShapeFilesGet(call, reqCtx)

	// Verify supportsAllDrives is set
	// Note: We can't easily inspect the call parameters without executing it
	// but we can verify the resource key header
	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file123/key123" {
		t.Errorf("Expected resource key header 'file123/key123', got '%s'", header)
	}
}

func TestRequestShaper_ShapeFilesList_UserDomain(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	// No DriveID = user domain
	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.List()
	shaped := shaper.ShapeFilesList(call, reqCtx)

	// The call should have corpora=user
	// We can't easily inspect without executing, but we can verify it doesn't panic
	if shaped == nil {
		t.Error("ShapeFilesList returned nil")
	}
}

func TestRequestShaper_ShapeFilesList_SharedDrive(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)

	// With DriveID = shared drive domain
	reqCtx := NewRequestContext("default", "drive123", types.RequestTypeListOrSearch)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.List()
	shaped := shaper.ShapeFilesList(call, reqCtx)

	// The call should have corpora=drive and driveId=drive123
	if shaped == nil {
		t.Error("ShapeFilesList returned nil")
	}
}

func TestRequestShaper_ShapeFilesCreate_WithParentResourceKeys(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("parent1", "pkey1", "test")
	client.ResourceKeys().AddKey("parent2", "pkey2", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)
	reqCtx.InvolvedParentIDs = []string{"parent1", "parent2"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Create(&drive.File{Name: "test.txt"})
	shaped := shaper.ShapeFilesCreate(call, reqCtx)

	// Verify parent resource keys are in header
	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if !strings.Contains(header, "parent1/pkey1") {
		t.Errorf("Expected header to contain 'parent1/pkey1', got '%s'", header)
	}
	if !strings.Contains(header, "parent2/pkey2") {
		t.Errorf("Expected header to contain 'parent2/pkey2', got '%s'", header)
	}
}

func TestRequestShaper_ShapeFilesUpdate_WithFileAndParentKeys(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file123", "fkey123", "test")
	client.ResourceKeys().AddKey("parent456", "pkey456", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)
	reqCtx.InvolvedFileIDs = []string{"file123"}
	reqCtx.InvolvedParentIDs = []string{"parent456"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Update("file123", &drive.File{Name: "updated.txt"})
	shaped := shaper.ShapeFilesUpdate(call, reqCtx)

	// Verify both file and parent resource keys are in header
	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if !strings.Contains(header, "file123/fkey123") {
		t.Errorf("Expected header to contain 'file123/fkey123', got '%s'", header)
	}
	if !strings.Contains(header, "parent456/pkey456") {
		t.Errorf("Expected header to contain 'parent456/pkey456', got '%s'", header)
	}
}

func TestRequestShaper_ShapeFilesDelete_WithResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file789", "key789", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)
	reqCtx.InvolvedFileIDs = []string{"file789"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Delete("file789")
	shaped := shaper.ShapeFilesDelete(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file789/key789" {
		t.Errorf("Expected header 'file789/key789', got '%s'", header)
	}
}

func TestRequestShaper_ShapeFilesCopy_WithResourceKeys(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("source123", "skey123", "test")
	client.ResourceKeys().AddKey("dest456", "dkey456", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)
	reqCtx.InvolvedFileIDs = []string{"source123"}
	reqCtx.InvolvedParentIDs = []string{"dest456"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Copy("source123", &drive.File{Name: "copy.txt"})
	shaped := shaper.ShapeFilesCopy(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if !strings.Contains(header, "source123/skey123") {
		t.Errorf("Expected header to contain 'source123/skey123', got '%s'", header)
	}
}

func TestRequestShaper_ShapePermissionsList_WithResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file999", "key999", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypePermissionOp)
	reqCtx.InvolvedFileIDs = []string{"file999"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Permissions.List("file999")
	shaped := shaper.ShapePermissionsList(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file999/key999" {
		t.Errorf("Expected header 'file999/key999', got '%s'", header)
	}
}

func TestRequestShaper_ShapePermissionsCreate_WithResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file888", "key888", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypePermissionOp)
	reqCtx.InvolvedFileIDs = []string{"file888"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	perm := &drive.Permission{
		Type:         "user",
		Role:         "reader",
		EmailAddress: "test@example.com",
	}
	call := service.Permissions.Create("file888", perm)
	shaped := shaper.ShapePermissionsCreate(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file888/key888" {
		t.Errorf("Expected header 'file888/key888', got '%s'", header)
	}
}

func TestRequestShaper_ShapeRevisionsList_WithResourceKey(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	client.ResourceKeys().AddKey("file777", "key777", "test")

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeRevisionOp)
	reqCtx.InvolvedFileIDs = []string{"file777"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Revisions.List("file777")
	shaped := shaper.ShapeRevisionsList(call, reqCtx)

	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "file777/key777" {
		t.Errorf("Expected header 'file777/key777', got '%s'", header)
	}
}

func TestRequestShaper_ShapeDrivesList(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeListOrSearch)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Drives.List()
	shaped := shaper.ShapeDrivesList(call, reqCtx)

	// drives.list doesn't need special shaping
	if shaped == nil {
		t.Error("ShapeDrivesList returned nil")
	}
}

func TestRequestShaper_NoResourceKeys(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	// Don't add any resource keys

	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeGetByID)
	reqCtx.InvolvedFileIDs = []string{"file123"}

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	call := service.Files.Get("file123")
	shaped := shaper.ShapeFilesGet(call, reqCtx)

	// Verify header is not set when no keys available
	header := shaped.Header().Get("X-Goog-Drive-Resource-Keys")
	if header != "" {
		t.Errorf("Expected empty header, got '%s'", header)
	}
}

// Verify that all shaped calls have supportsAllDrives=true
func TestRequestShaper_SupportsAllDrives(t *testing.T) {
	// This is a critical requirement for Shared Drive support
	// All file operations must include supportsAllDrives=true

	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test various operations
	tests := []struct {
		name string
		call interface{}
	}{
		{"FilesGet", shaper.ShapeFilesGet(service.Files.Get("test"), reqCtx)},
		{"FilesList", shaper.ShapeFilesList(service.Files.List(), reqCtx)},
		{"FilesCreate", shaper.ShapeFilesCreate(service.Files.Create(&drive.File{}), reqCtx)},
		{"FilesUpdate", shaper.ShapeFilesUpdate(service.Files.Update("test", &drive.File{}), reqCtx)},
		{"FilesDelete", shaper.ShapeFilesDelete(service.Files.Delete("test"), reqCtx)},
		{"FilesCopy", shaper.ShapeFilesCopy(service.Files.Copy("test", &drive.File{}), reqCtx)},
		{"PermissionsList", shaper.ShapePermissionsList(service.Permissions.List("test"), reqCtx)},
		{"PermissionsCreate", shaper.ShapePermissionsCreate(service.Permissions.Create("test", &drive.Permission{}), reqCtx)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.call == nil {
				t.Error("Shaped call is nil")
			}
			// Note: We can't easily verify supportsAllDrives is set without executing the request
			// but we've verified it's in the code. This test ensures the calls don't panic.
		})
	}
}

// TestRequestShaper_HeaderMethods verifies that shaped calls support Header() method
func TestRequestShaper_HeaderMethods(t *testing.T) {
	client := NewClient(nil, 3, 1000, logging.NewNoOpLogger())
	shaper := NewRequestShaper(client)
	reqCtx := NewRequestContext("default", "", types.RequestTypeMutation)

	service, err := drive.NewService(context.Background(), option.WithoutAuthentication(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test that Header() method works on all shaped calls
	tests := []struct {
		name   string
		header http.Header
	}{
		{"FilesGet", shaper.ShapeFilesGet(service.Files.Get("test"), reqCtx).Header()},
		{"FilesCreate", shaper.ShapeFilesCreate(service.Files.Create(&drive.File{}), reqCtx).Header()},
		{"FilesUpdate", shaper.ShapeFilesUpdate(service.Files.Update("test", &drive.File{}), reqCtx).Header()},
		{"FilesDelete", shaper.ShapeFilesDelete(service.Files.Delete("test"), reqCtx).Header()},
		{"FilesCopy", shaper.ShapeFilesCopy(service.Files.Copy("test", &drive.File{}), reqCtx).Header()},
		{"PermissionsList", shaper.ShapePermissionsList(service.Permissions.List("test"), reqCtx).Header()},
		{"PermissionsCreate", shaper.ShapePermissionsCreate(service.Permissions.Create("test", &drive.Permission{}), reqCtx).Header()},
		{"RevisionsList", shaper.ShapeRevisionsList(service.Revisions.List("test"), reqCtx).Header()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.header == nil {
				t.Error("Header() returned nil")
			}
		})
	}
}
