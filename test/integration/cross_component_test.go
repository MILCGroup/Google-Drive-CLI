//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/folders"
	"github.com/dl-alexandre/gdrv/internal/permissions"
	"github.com/dl-alexandre/gdrv/internal/resolver"
	"github.com/dl-alexandre/gdrv/internal/revisions"
	"github.com/dl-alexandre/gdrv/internal/safety"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_CrossComponent_PathResolverWithSharedDrives tests path resolver with Shared Drives
func TestIntegration_CrossComponent_PathResolverWithSharedDrives(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	pathResolver := resolver.NewPathResolver(client, time.Minute)
	folderManager := folders.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a folder in the Shared Drive to resolve
	folderName := "resolver-test-" + time.Now().Format("20060102150405")
	folder, err := folderManager.Create(ctx, reqCtx, folderName, driveID)
	if err != nil {
		t.Fatalf("Failed to create Shared Drive folder: %v", err)
	}

	path := folderName
	result, err := pathResolver.Resolve(ctx, reqCtx, path, resolver.ResolveOptions{
		DriveID:  driveID,
		UseCache: false,
	})
	if err != nil {
		t.Fatalf("Failed to resolve Shared Drive path: %v", err)
	}

	if result.FileID != folder.ID {
		t.Errorf("Expected resolved ID %s, got %s", folder.ID, result.FileID)
	}

	t.Logf("Resolved Shared Drive path: %s -> %s", path, result.FileID)

	_ = folderManager.Delete(ctx, reqCtx, folder.ID, false)
}

// TestIntegration_CrossComponent_FileManagerWithSafety tests file manager with safety controls
func TestIntegration_CrossComponent_FileManagerWithSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create test files
	fileNames := []string{
		"safety-test-1-" + time.Now().Format("20060102150405") + ".txt",
		"safety-test-2-" + time.Now().Format("20060102150405") + ".txt",
	}

	var fileIDs []string
	for _, name := range fileNames {
		file := uploadTempFile(t, ctx, fileManager, reqCtx, name, "", "text/plain", nil)
		fileIDs = append(fileIDs, file.ID)
	}

	// Test safety controls - dry run delete
	recorder := safety.NewDryRunRecorder()
	err := fileManager.DeleteWithSafety(ctx, reqCtx, fileIDs[0], false, safety.SafetyOptions{
		DryRun: true,
	}, recorder)
	if err != nil {
		t.Errorf("Dry run delete should not fail: %v", err)
	}

	// Verify file still exists
	_, err = fileManager.Get(ctx, reqCtx, fileIDs[0], "id")
	if err != nil {
		t.Errorf("File should still exist after dry run: %v", err)
	}

	// Real delete
	err = fileManager.Delete(ctx, reqCtx, fileIDs[0], false)
	if err != nil {
		t.Errorf("Real delete failed: %v", err)
	}

	// Clean up second file
	err = fileManager.Delete(ctx, reqCtx, fileIDs[1], false)
	if err != nil {
		t.Errorf("Failed to clean up second file: %v", err)
	}
}

// TestIntegration_CrossComponent_LoggingIntegration tests logging integration
func TestIntegration_CrossComponent_LoggingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Perform operations with logging
	fileName := "logging-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", nil)

	// List with logging
	listReq := files.ListOptions{
		Query: fmt.Sprintf("name='%s'", fileName),
	}
	_, err := fileManager.List(ctx, reqCtx, listReq)
	if err != nil {
		t.Errorf("Failed to list with logging: %v", err)
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete with logging: %v", err)
	}

	t.Log("Logging integration test completed - check logs for detailed output")
}

// TestIntegration_CrossComponent_ComplexWorkflow tests complex workflow with all managers
func TestIntegration_CrossComponent_ComplexWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	folderManager := folders.NewManager(client)
	permManager := permissions.NewManager(client)
	revManager := revisions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// 1. Create folder
	folderName := "complex-workflow-" + time.Now().Format("20060102150405")
	folder, err := folderManager.Create(ctx, reqCtx, folderName, "")
	if err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	// 2. Create file in folder
	fileName := "workflow-file.txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, folder.ID, "text/plain", []byte("initial content"))

	// 3. Update file content (create new revision)
	updatedContent := []byte("updated content")
	updatePath := filepath.Join(t.TempDir(), "workflow-update.txt")
	if writeErr := os.WriteFile(updatePath, updatedContent, 0644); writeErr != nil {
		t.Fatalf("Failed to create update file: %v", writeErr)
	}
	_, err = fileManager.UpdateContent(ctx, reqCtx, file.ID, updatePath, files.UpdateContentOptions{
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// 4. List revisions
	revisions, err := revManager.List(ctx, reqCtx, file.ID, revisions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list revisions: %v", err)
	}

	if len(revisions.Revisions) < 2 {
		t.Errorf("Expected at least 2 revisions, got %d", len(revisions.Revisions))
	}

	// 5. Create permission
	permOpts := permissions.CreateOptions{
		Type: "anyone",
		Role: "reader",
	}
	_, err = permManager.Create(ctx, reqCtx, file.ID, permOpts)
	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}

	// 6. List permissions
	perms, err := permManager.List(ctx, reqCtx, file.ID, permissions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list permissions: %v", err)
	}

	if len(perms) < 2 { // At least owner + our permission
		t.Errorf("Expected at least 2 permissions, got %d", len(perms))
	}

	// 7. Clean up
	permID := ""
	for _, perm := range perms {
		if perm.Type == "anyone" && perm.Role == "reader" {
			permID = perm.ID
			break
		}
	}
	if permID == "" {
		t.Errorf("Expected to find anyone permission for cleanup")
	} else {
		err = permManager.Delete(ctx, reqCtx, file.ID, permID, permissions.DeleteOptions{})
		if err != nil {
			t.Errorf("Failed to delete permission: %v", err)
		}
	}

	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	err = folderManager.Delete(ctx, reqCtx, folder.ID, false)
	if err != nil {
		t.Errorf("Failed to delete folder: %v", err)
	}

	t.Log("Complex workflow completed successfully")
}

// TestIntegration_CrossComponent_ConfigurationPropagation tests configuration propagation
func TestIntegration_CrossComponent_ConfigurationPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set environment variables for configuration
	os.Setenv("GDRV_MAX_RETRIES", "5")
	os.Setenv("GDRV_REQUEST_TIMEOUT", "60")
	defer func() {
		os.Unsetenv("GDRV_MAX_RETRIES")
		os.Unsetenv("GDRV_REQUEST_TIMEOUT")
	}()

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Configuration should be loaded and applied to managers
	fileName := "config-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", nil)

	// Clean up
	err := fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	t.Log("Configuration propagation test completed")
}
