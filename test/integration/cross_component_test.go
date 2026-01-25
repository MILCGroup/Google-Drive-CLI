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
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/folders"
	"github.com/dl-alexandre/gdrv/internal/permissions"
	"github.com/dl-alexandre/gdrv/internal/resolver"
	"github.com/dl-alexandre/gdrv/internal/revisions"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_CrossComponent_PathResolverWithSharedDrives tests path resolver with Shared Drives
func TestIntegration_CrossComponent_PathResolverWithSharedDrives(t *testing.T) {
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

	// Test resolving Shared Drive root
	path := fmt.Sprintf("SharedDrives/%s", driveID)
	result, err := resolver.Resolve(ctx, reqCtx, path, "")
	if err != nil {
		t.Fatalf("Failed to resolve Shared Drive path: %v", err)
	}

	if result.Type != types.ResolveTypeSharedDrive {
		t.Errorf("Expected resolve type SharedDrive, got %s", result.Type)
	}

	t.Logf("Resolved Shared Drive path: %s -> %s", path, result.ID)
}

// TestIntegration_CrossComponent_FileManagerWithSafety tests file manager with safety controls
func TestIntegration_CrossComponent_FileManagerWithSafety(t *testing.T) {
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
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create test files
	fileNames := []string{
		"safety-test-1-" + time.Now().Format("20060102150405") + ".txt",
		"safety-test-2-" + time.Now().Format("20060102150405") + ".txt",
	}

	var fileIDs []string
	for _, name := range fileNames {
		file, err := fileManager.Create(ctx, reqCtx, name, "", "text/plain", nil)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
		fileIDs = append(fileIDs, file.ID)
	}

	// Test safety controls - dry run delete
	reqCtx.DryRun = true
	err = fileManager.Delete(ctx, reqCtx, fileIDs[0], false)
	if err != nil {
		t.Errorf("Dry run delete should not fail: %v", err)
	}

	// Verify file still exists
	_, err = fileManager.Get(ctx, reqCtx, fileIDs[0])
	if err != nil {
		t.Errorf("File should still exist after dry run: %v", err)
	}

	// Real delete
	reqCtx.DryRun = false
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
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Enable logging
	reqCtx.LogLevel = types.LogLevelVerbose

	// Perform operations with logging
	fileName := "logging-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create file with logging: %v", err)
	}

	// List with logging
	listReq := &types.FileListRequest{
		Query: fmt.Sprintf("name='%s'", fileName),
	}
	_, err = fileManager.List(ctx, reqCtx, listReq)
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
	file, err := fileManager.Create(ctx, reqCtx, fileName, folder.ID, "text/plain", []byte("initial content"))
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// 3. Update file content (create new revision)
	updatedContent := []byte("updated content")
	_, err = fileManager.Update(ctx, reqCtx, file.ID, updatedContent, "text/plain")
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// 4. List revisions
	revisions, err := revManager.List(ctx, reqCtx, file.ID)
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
	perms, err := permManager.List(ctx, reqCtx, file.ID)
	if err != nil {
		t.Fatalf("Failed to list permissions: %v", err)
	}

	if len(perms.Permissions) < 2 { // At least owner + our permission
		t.Errorf("Expected at least 2 permissions, got %d", len(perms.Permissions))
	}

	// 7. Clean up
	err = permManager.Delete(ctx, reqCtx, file.ID, "anyone")
	if err != nil {
		t.Errorf("Failed to delete permission: %v", err)
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

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Set environment variables for configuration
	os.Setenv("GDRV_MAX_RETRIES", "5")
	os.Setenv("GDRV_REQUEST_TIMEOUT", "60")
	defer func() {
		os.Unsetenv("GDRV_MAX_RETRIES")
		os.Unsetenv("GDRV_REQUEST_TIMEOUT")
	}()

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Configuration should be loaded and applied to managers
	fileName := "config-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create file with config: %v", err)
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	t.Log("Configuration propagation test completed")
}</content>
<parameter name="filePath">test/integration/cross_component_test.go