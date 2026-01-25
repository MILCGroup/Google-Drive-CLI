// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/drives"
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/folders"
	"github.com/dl-alexandre/gdrv/internal/resolver"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_SharedDrives_Enumeration tests listing Shared Drives
func TestIntegration_SharedDrives_Enumeration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	driveManager := drives.NewManager(client)
	reqCtx := &types.RequestContext{}

	// List Shared Drives
	result, err := driveManager.List(ctx, reqCtx, 10, "")
	if err != nil {
		t.Fatalf("Failed to list Shared Drives: %v", err)
	}

	t.Logf("Found %d Shared Drives", len(result.Drives))

	// Check pagination if there are many drives
	if result.NextPageToken != "" {
		t.Logf("More drives available with page token: %s", result.NextPageToken)
	}

	// Verify drive structure
	for _, drive := range result.Drives {
		if drive.ID == "" {
			t.Error("Drive missing ID")
		}
		if drive.Name == "" {
			t.Error("Drive missing name")
		}
		t.Logf("Drive: %s (%s)", drive.Name, drive.ID)
	}
}

// TestIntegration_SharedDrives_GetDrive tests getting a specific Shared Drive
func TestIntegration_SharedDrives_GetDrive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Need a test Shared Drive ID
	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	driveManager := drives.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Get specific drive
	drive, err := driveManager.Get(ctx, reqCtx, driveID)
	if err != nil {
		t.Fatalf("Failed to get Shared Drive: %v", err)
	}

	if drive.ID != driveID {
		t.Errorf("Expected drive ID %s, got %s", driveID, drive.ID)
	}

	if drive.Name == "" {
		t.Error("Drive missing name")
	}

	t.Logf("Retrieved drive: %s", drive.Name)
}

// TestIntegration_SharedDrives_PathResolution tests path resolution within Shared Drives
func TestIntegration_SharedDrives_PathResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	resolver := resolver.NewResolver(client)
	reqCtx := &types.RequestContext{}

	// Test path resolution in Shared Drive context
	path := "SharedDrives/" + driveID + "/test-path"
	resolved, err := resolver.Resolve(ctx, reqCtx, path, "")
	if err != nil {
		// Path might not exist, that's okay for this test
		t.Logf("Path resolution failed (expected if path doesn't exist): %v", err)
	} else {
		t.Logf("Resolved path %s to ID: %s", path, resolved.ID)
	}
}

// TestIntegration_SharedDrives_SpecificOperations tests drive-specific operations
func TestIntegration_SharedDrives_SpecificOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	folderManager := folders.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a folder in the Shared Drive
	folderName := "shared-drive-test-" + time.Now().Format("20060102150405")
	folder, err := folderManager.Create(ctx, reqCtx, folderName, driveID)
	if err != nil {
		t.Fatalf("Failed to create folder in Shared Drive: %v", err)
	}

	// Create a file in the Shared Drive folder
	fileName := "shared-drive-file-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, folder.ID, "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create file in Shared Drive: %v", err)
	}

	// List contents of the Shared Drive folder
	listReq := &types.FileListRequest{
		Query: fmt.Sprintf("'%s' in parents and trashed=false", folder.ID),
	}
	results, err := fileManager.List(ctx, reqCtx, listReq)
	if err != nil {
		t.Fatalf("Failed to list Shared Drive folder contents: %v", err)
	}

	if len(results.Files) != 1 {
		t.Errorf("Expected 1 file in folder, got %d", len(results.Files))
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete Shared Drive file: %v", err)
	}

	err = folderManager.Delete(ctx, reqCtx, folder.ID, false)
	if err != nil {
		t.Errorf("Failed to delete Shared Drive folder: %v", err)
	}
}</content>
<parameter name="filePath">test/integration/shared_drives_test.go