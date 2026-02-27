//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/files"
	"github.com/milcgroup/gdrv/internal/permissions"
	"github.com/milcgroup/gdrv/internal/types"
)

// TestIntegration_PermissionOperations_CreationListingDeletion tests permission creation, listing, and deletion
func TestIntegration_PermissionOperations_CreationListingDeletion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "test-file-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", nil)

	// List initial permissions
	perms, err := permManager.List(ctx, reqCtx, file.ID, permissions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list initial permissions: %v", err)
	}

	initialCount := len(perms)
	t.Logf("Initial permission count: %d", initialCount)

	// Create a new permission (reader role)
	opts := permissions.CreateOptions{
		Type: "anyone",
		Role: "reader",
	}
	perm, err := permManager.Create(ctx, reqCtx, file.ID, opts)
	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}

	if perm.Role != "reader" {
		t.Errorf("Expected role 'reader', got '%s'", perm.Role)
	}

	// List permissions again
	perms, err = permManager.List(ctx, reqCtx, file.ID, permissions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list permissions after creation: %v", err)
	}

	if len(perms) != initialCount+1 {
		t.Errorf("Expected %d permissions, got %d", initialCount+1, len(perms))
	}

	// Delete the permission
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID, permissions.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete permission: %v", err)
	}

	// Verify deletion
	perms, err = permManager.List(ctx, reqCtx, file.ID, permissions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list permissions after deletion: %v", err)
	}

	if len(perms) != initialCount {
		t.Errorf("Expected %d permissions after deletion, got %d", initialCount, len(perms))
	}

	// Clean up file
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_PermissionOperations_PublicLink tests public link creation
func TestIntegration_PermissionOperations_PublicLink(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "public-link-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", nil)

	// Create public link
	perm, err := permManager.CreatePublicLink(ctx, reqCtx, file.ID, "reader", true)
	if err != nil {
		t.Fatalf("Failed to create public link: %v", err)
	}

	if perm.Role != "reader" {
		t.Errorf("Expected role 'reader', got '%s'", perm.Role)
	}

	if perm.Type != "anyone" {
		t.Errorf("Expected type 'anyone', got '%s'", perm.Type)
	}

	// Verify link is accessible (this is limited without browser automation)
	t.Logf("Created public link permission with ID: %s", perm.ID)

	// Clean up
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID, permissions.DeleteOptions{})
	if err != nil {
		t.Errorf("Failed to delete public link permission: %v", err)
	}

	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_PermissionOperations_Update tests permission updates
func TestIntegration_PermissionOperations_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "update-perm-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", nil)

	// Create initial permission
	opts := permissions.CreateOptions{
		Type: "anyone",
		Role: "reader",
	}
	perm, err := permManager.Create(ctx, reqCtx, file.ID, opts)
	if err != nil {
		t.Fatalf("Failed to create initial permission: %v", err)
	}

	// Update permission role
	updateOpts := permissions.UpdateOptions{
		Role: "writer",
	}
	updatedPerm, err := permManager.Update(ctx, reqCtx, file.ID, perm.ID, updateOpts)
	if err != nil {
		t.Fatalf("Failed to update permission: %v", err)
	}

	if updatedPerm.Role != "writer" {
		t.Errorf("Expected updated role 'writer', got '%s'", updatedPerm.Role)
	}

	// Clean up
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID, permissions.DeleteOptions{})
	if err != nil {
		t.Errorf("Failed to delete permission: %v", err)
	}

	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_PermissionOperations_SharedDrive tests Shared Drive permission scenarios
func TestIntegration_PermissionOperations_SharedDrive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a Shared Drive to be available
	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set - skipping Shared Drive permission test")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a file in the Shared Drive
	fileName := "shared-drive-perm-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, driveID, "text/plain", nil)

	// List permissions (should include Shared Drive specific ones)
	perms, err := permManager.List(ctx, reqCtx, file.ID, permissions.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list Shared Drive permissions: %v", err)
	}

	t.Logf("Shared Drive file has %d permissions", len(perms))

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete Shared Drive test file: %v", err)
	}
}
