// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/permissions"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_PermissionOperations_CreationListingDeletion tests permission creation, listing, and deletion
func TestIntegration_PermissionOperations_CreationListingDeletion(t *testing.T) {
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
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "test-file-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// List initial permissions
	perms, err := permManager.List(ctx, reqCtx, file.ID)
	if err != nil {
		t.Fatalf("Failed to list initial permissions: %v", err)
	}

	initialCount := len(perms.Permissions)
	t.Logf("Initial permission count: %d", initialCount)

	// Create a new permission (reader role)
	opts := permissions.CreateOptions{
		Type:  "anyone",
		Role:  "reader",
		Email: "",
	}
	perm, err := permManager.Create(ctx, reqCtx, file.ID, opts)
	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}

	if perm.Role != "reader" {
		t.Errorf("Expected role 'reader', got '%s'", perm.Role)
	}

	// List permissions again
	perms, err = permManager.List(ctx, reqCtx, file.ID)
	if err != nil {
		t.Fatalf("Failed to list permissions after creation: %v", err)
	}

	if len(perms.Permissions) != initialCount+1 {
		t.Errorf("Expected %d permissions, got %d", initialCount+1, len(perms.Permissions))
	}

	// Delete the permission
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID)
	if err != nil {
		t.Fatalf("Failed to delete permission: %v", err)
	}

	// Verify deletion
	perms, err = permManager.List(ctx, reqCtx, file.ID)
	if err != nil {
		t.Fatalf("Failed to list permissions after deletion: %v", err)
	}

	if len(perms.Permissions) != initialCount {
		t.Errorf("Expected %d permissions after deletion, got %d", initialCount, len(perms.Permissions))
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
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "public-link-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

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
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID)
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
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a test file
	fileName := "update-perm-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

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
	err = permManager.Delete(ctx, reqCtx, file.ID, perm.ID)
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

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// This test requires a Shared Drive to be available
	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set - skipping Shared Drive permission test")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	permManager := permissions.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create a file in the Shared Drive
	fileName := "shared-drive-perm-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, driveID, "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create file in Shared Drive: %v", err)
	}

	// List permissions (should include Shared Drive specific ones)
	perms, err := permManager.List(ctx, reqCtx, file.ID)
	if err != nil {
		t.Fatalf("Failed to list Shared Drive permissions: %v", err)
	}

	t.Logf("Shared Drive file has %d permissions", len(perms.Permissions))

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete Shared Drive test file: %v", err)
	}
}</content>
<parameter name="filePath">test/integration/permission_operations_test.go