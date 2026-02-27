//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/folders"
	"github.com/milcgroup/gdrv/internal/types"
)

// TestIntegration_FolderOperations_CreationDeletion tests folder creation and deletion workflow
func TestIntegration_FolderOperations_CreationDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	folderManager := folders.NewManager(client)

	// Create test folder
	reqCtx := &types.RequestContext{}
	folderName := "test-folder-" + time.Now().Format("20060102150405")

	folder, err := folderManager.Create(ctx, reqCtx, folderName, "")
	if err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	if folder.Name != folderName {
		t.Errorf("Expected folder name %s, got %s", folderName, folder.Name)
	}

	// Verify folder exists by listing
	results, err := folderManager.List(ctx, reqCtx, "root", 100, "")
	if err != nil {
		t.Fatalf("Failed to list folders: %v", err)
	}

	found := false
	for _, item := range results.Files {
		if item.Name == folderName && item.MimeType == "application/vnd.google-apps.folder" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created folder not found in list")
	}

	// Delete the folder
	err = folderManager.Delete(ctx, reqCtx, folder.ID, false)
	if err != nil {
		t.Fatalf("Failed to delete folder: %v", err)
	}

	// Verify deletion
	results, err = folderManager.List(ctx, reqCtx, "root", 100, "")
	if err != nil {
		t.Fatalf("Failed to list after deletion: %v", err)
	}

	stillFound := false
	for _, item := range results.Files {
		if item.Name == folderName && item.MimeType == "application/vnd.google-apps.folder" {
			stillFound = true
			break
		}
	}
	if stillFound {
		t.Error("Folder still exists after deletion")
	}
}

// TestIntegration_FolderOperations_Hierarchies tests nested folder creation
func TestIntegration_FolderOperations_Hierarchies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	folderManager := folders.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create parent folder
	parentName := "parent-folder-" + time.Now().Format("20060102150405")
	parent, err := folderManager.Create(ctx, reqCtx, parentName, "")
	if err != nil {
		t.Fatalf("Failed to create parent folder: %v", err)
	}

	// Create child folder
	childName := "child-folder-" + time.Now().Format("20060102150405")
	child, err := folderManager.Create(ctx, reqCtx, childName, parent.ID)
	if err != nil {
		t.Fatalf("Failed to create child folder: %v", err)
	}

	// Verify hierarchy
	if len(child.Parents) == 0 || child.Parents[0] != parent.ID {
		t.Error("Child folder does not have correct parent")
	}

	// List children of parent
	results, err := folderManager.List(ctx, reqCtx, parent.ID, 100, "")
	if err != nil {
		t.Fatalf("Failed to list child folders: %v", err)
	}

	foundChild := false
	for _, item := range results.Files {
		if item.ID == child.ID {
			foundChild = true
			break
		}
	}
	if !foundChild {
		t.Error("Child folder not found in parent's listing")
	}

	// Clean up
	err = folderManager.Delete(ctx, reqCtx, child.ID, false)
	if err != nil {
		t.Errorf("Failed to delete child folder: %v", err)
	}

	err = folderManager.Delete(ctx, reqCtx, parent.ID, false)
	if err != nil {
		t.Errorf("Failed to delete parent folder: %v", err)
	}
}

// TestIntegration_FolderOperations_Move tests moving folders
func TestIntegration_FolderOperations_Move(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	folderManager := folders.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create source and destination folders
	srcName := "source-folder-" + time.Now().Format("20060102150405")
	destName := "dest-folder-" + time.Now().Format("20060102150405")

	src, err := folderManager.Create(ctx, reqCtx, srcName, "")
	if err != nil {
		t.Fatalf("Failed to create source folder: %v", err)
	}

	dest, err := folderManager.Create(ctx, reqCtx, destName, "")
	if err != nil {
		t.Fatalf("Failed to create destination folder: %v", err)
	}

	// Move src into dest
	_, err = folderManager.Move(ctx, reqCtx, src.ID, dest.ID)
	if err != nil {
		t.Fatalf("Failed to move folder: %v", err)
	}

	// Verify move
	moved, err := folderManager.Get(ctx, reqCtx, src.ID, "id,parents")
	if err != nil {
		t.Fatalf("Failed to get moved folder: %v", err)
	}

	if len(moved.Parents) == 0 || moved.Parents[0] != dest.ID {
		t.Error("Folder was not moved to correct parent")
	}

	// Clean up
	err = folderManager.Delete(ctx, reqCtx, src.ID, false)
	if err != nil {
		t.Errorf("Failed to delete moved folder: %v", err)
	}

	err = folderManager.Delete(ctx, reqCtx, dest.ID, false)
	if err != nil {
		t.Errorf("Failed to delete destination folder: %v", err)
	}
}

// TestIntegration_FolderOperations_RecursiveDeletion tests recursive deletion
func TestIntegration_FolderOperations_RecursiveDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	folderManager := folders.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create parent folder
	parentName := "recursive-test-" + time.Now().Format("20060102150405")
	parent, err := folderManager.Create(ctx, reqCtx, parentName, "")
	if err != nil {
		t.Fatalf("Failed to create parent folder: %v", err)
	}

	// Create child folder
	child, err := folderManager.Create(ctx, reqCtx, "child", parent.ID)
	if err != nil {
		t.Fatalf("Failed to create child folder: %v", err)
	}

	// Recursive delete
	err = folderManager.Delete(ctx, reqCtx, parent.ID, true)
	if err != nil {
		t.Fatalf("Failed to recursively delete folder: %v", err)
	}

	// Verify both are deleted
	_, err = folderManager.Get(ctx, reqCtx, parent.ID, "id")
	if err == nil {
		t.Error("Parent folder still exists after recursive deletion")
	}

	_, err = folderManager.Get(ctx, reqCtx, child.ID, "id")
	if err == nil {
		t.Error("Child folder still exists after recursive deletion")
	}
}
