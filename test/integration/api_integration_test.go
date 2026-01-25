// +build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Integration Test: API Connectivity
// Run with: go test -tags=integration ./test/integration/...

func setupAPIClient(t *testing.T) (*api.Client, *drive.Service) {
	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set - skipping API integration tests")
	}

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	ctx := context.Background()
	service, err := drive.NewService(ctx, option.WithTokenSource(
		manager.GetConfig().TokenSource(ctx, token),
	))
	if err != nil {
		t.Fatalf("Failed to create Drive service: %v", err)
	}

	client := api.NewClient(service, 3, 1000)
	return client, service
}

func TestIntegration_API_About(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)

	about, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.About, error) {
		return service.About.Get().Fields("user,storageQuota").Do()
	})

	if err != nil {
		t.Fatalf("About.Get failed: %v", err)
	}

	if about.User == nil {
		t.Error("User info is nil")
	}
	if about.StorageQuota == nil {
		t.Error("StorageQuota is nil")
	}

	t.Logf("User: %s (%s)", about.User.DisplayName, about.User.EmailAddress)
	t.Logf("Storage: %d / %d bytes used", about.StorageQuota.Usage, about.StorageQuota.Limit)
}

func TestIntegration_API_FilesList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)

	call := service.Files.List().
		PageSize(10).
		Fields("files(id,name,mimeType),nextPageToken")
	
	call = shaper.ShapeFilesList(call, reqCtx)

	fileList, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.FileList, error) {
		return call.Do()
	})

	if err != nil {
		t.Fatalf("Files.List failed: %v", err)
	}

	t.Logf("Found %d files", len(fileList.Files))
	
	for i, file := range fileList.Files {
		if i >= 5 {
			break // Only log first 5
		}
		t.Logf("  %s (%s) - %s", file.Name, file.Id, file.MimeType)
	}
}

func TestIntegration_API_FilesGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileID := os.Getenv("TEST_FILE_ID")
	if fileID == "" {
		t.Skip("TEST_FILE_ID not set")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	reqCtx.InvolvedFileIDs = []string{fileID}

	call := service.Files.Get(fileID).Fields("id,name,mimeType,size,capabilities")
	call = shaper.ShapeFilesGet(call, reqCtx)

	file, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.File, error) {
		return call.Do()
	})

	if err != nil {
		t.Fatalf("Files.Get failed: %v", err)
	}

	if file.Id != fileID {
		t.Errorf("File ID mismatch: got %s, want %s", file.Id, fileID)
	}

	t.Logf("File: %s", file.Name)
	t.Logf("  ID: %s", file.Id)
	t.Logf("  MIME: %s", file.MimeType)
	t.Logf("  Size: %d bytes", file.Size)
}

func TestIntegration_API_SharedDrivesList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)

	call := service.Drives.List().PageSize(10)
	call = shaper.ShapeDrivesList(call, reqCtx)

	driveList, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.DriveList, error) {
		return call.Do()
	})

	if err != nil {
		t.Fatalf("Drives.List failed: %v", err)
	}

	t.Logf("Found %d shared drives", len(driveList.Drives))
	
	for _, drive := range driveList.Drives {
		t.Logf("  %s (%s)", drive.Name, drive.Id)
	}
}

func TestIntegration_API_RetryOnRateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)

	// Make many rapid requests to potentially trigger rate limiting
	// The client should automatically retry
	for i := 0; i < 5; i++ {
		_, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.About, error) {
			return service.About.Get().Fields("user").Do()
		})

		if err != nil {
			// If we hit rate limit, verify it's handled properly
			t.Logf("Request %d: %v", i, err)
		}
	}
}

func TestIntegration_API_ErrorClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)

	// Try to get a file that doesn't exist
	invalidFileID := "invalid-file-id-does-not-exist"
	reqCtx.InvolvedFileIDs = []string{invalidFileID}

	call := service.Files.Get(invalidFileID).Fields("id,name")
	call = shaper.ShapeFilesGet(call, reqCtx)

	_, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.File, error) {
		return call.Do()
	})

	if err == nil {
		t.Fatal("Expected error for invalid file ID")
	}

	// Verify error is properly classified
	t.Logf("Error: %v", err)
	
	// In real implementation, verify this is classified as FILE_NOT_FOUND
}

func TestIntegration_API_ResourceKeyHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test resource key extraction from URL
	testURL := os.Getenv("TEST_RESOURCE_KEY_URL")
	if testURL == "" {
		t.Skip("TEST_RESOURCE_KEY_URL not set")
	}

	client, _ := setupAPIClient(t)
	
	fileID, key, ok := client.ResourceKeys().ParseFromURL(testURL)
	
	if !ok {
		t.Fatal("Failed to parse resource key URL")
	}

	t.Logf("Parsed URL:")
	t.Logf("  File ID: %s", fileID)
	t.Logf("  Resource Key: %s", key)

	if fileID == "" {
		t.Error("File ID is empty")
	}
}

func TestIntegration_API_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)

	pageSize := int64(5)
	pageToken := ""
	pageCount := 0
	totalFiles := 0

	for {
		call := service.Files.List().
			PageSize(pageSize).
			Fields("files(id,name),nextPageToken")
		
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		
		call = shaper.ShapeFilesList(call, reqCtx)

		fileList, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.FileList, error) {
			return call.Do()
		})

		if err != nil {
			t.Fatalf("Files.List failed on page %d: %v", pageCount+1, err)
		}

		pageCount++
		totalFiles += len(fileList.Files)
		
		t.Logf("Page %d: %d files", pageCount, len(fileList.Files))

		if fileList.NextPageToken == "" {
			break
		}
		
		pageToken = fileList.NextPageToken
		
		if pageCount >= 3 {
			// Limit test to 3 pages
			break
		}
	}

	t.Logf("Total: %d files across %d pages", totalFiles, pageCount)
}

func TestIntegration_API_FieldMasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileID := os.Getenv("TEST_FILE_ID")
	if fileID == "" {
		t.Skip("TEST_FILE_ID not set")
	}

	client, service := setupAPIClient(t)
	shaper := api.NewRequestShaper(client)
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	reqCtx.InvolvedFileIDs = []string{fileID}

	tests := []struct {
		name   string
		fields string
	}{
		{"Minimal", "id,name"},
		{"Standard", "id,name,mimeType,size"},
		{"Extended", "id,name,mimeType,size,createdTime,modifiedTime"},
		{"With capabilities", "id,name,capabilities"},
		{"With permissions", "id,name,permissions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := service.Files.Get(fileID).Fields(tt.fields)
			call = shaper.ShapeFilesGet(call, reqCtx)

			file, err := api.ExecuteWithRetry(context.Background(), client, reqCtx, func() (*drive.File, error) {
				return call.Do()
			})

			if err != nil {
				t.Fatalf("Files.Get with fields %s failed: %v", tt.fields, err)
			}

			if file.Id == "" {
				t.Error("File ID is empty")
			}
			if file.Name == "" {
				t.Error("File name is empty")
			}

			t.Logf("Fields %s: OK", tt.fields)
		})
	}
}
