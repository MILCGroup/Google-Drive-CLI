// +build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/drives"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Integration Test: Shared Drives Operations
// Run with: go test -tags=integration ./test/integration/...

func setupDrivesManager(t *testing.T) (*drives.Manager, *api.Client, context.Context) {
	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
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
	drivesMgr := drives.NewManager(client)

	return drivesMgr, client, ctx
}

func TestIntegration_Drives_ListSharedDrives(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	drivesMgr, _, ctx := setupDrivesManager(t)

	// List all accessible Shared Drives
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
	result, err := drivesMgr.List(ctx, reqCtx, 10, "")
	if err != nil {
		t.Fatalf("List drives failed: %v", err)
	}

	t.Logf("Found %d Shared Drives", len(result.Drives))

	for i, drive := range result.Drives {
		t.Logf("  Drive %d: %s (ID: %s)", i+1, drive.Name, drive.ID)
	}

	// Verify structure
	if len(result.Drives) > 0 {
		firstDrive := result.Drives[0]
		if firstDrive.ID == "" {
			t.Error("Drive ID is empty")
		}
		if firstDrive.Name == "" {
			t.Error("Drive Name is empty")
		}
	}
}

func TestIntegration_Drives_ListPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	drivesMgr, _, ctx := setupDrivesManager(t)

	// List with small page size to test pagination
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
	pageSize := 2
	allDrives := []*drives.SharedDrive{}
	pageToken := ""
	pageCount := 0

	for {
		result, err := drivesMgr.List(ctx, reqCtx, pageSize, pageToken)
		if err != nil {
			t.Fatalf("List drives page %d failed: %v", pageCount+1, err)
		}

		pageCount++
		t.Logf("Page %d: %d drives", pageCount, len(result.Drives))
		allDrives = append(allDrives, result.Drives...)

		if result.NextPageToken == "" {
			break
		}

		// Safety limit to prevent infinite loops
		if pageCount > 20 {
			t.Log("Stopping after 20 pages for safety")
			break
		}

		pageToken = result.NextPageToken
	}

	t.Logf("Total drives across %d pages: %d", pageCount, len(allDrives))
}

func TestIntegration_Drives_GetDriveMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	drivesMgr, _, ctx := setupDrivesManager(t)

	// Get drive with basic fields
	reqCtx := api.NewRequestContext("default", driveID, types.RequestTypeGetByID)
	drive, err := drivesMgr.Get(ctx, reqCtx, driveID, "id,name,kind")
	if err != nil {
		t.Fatalf("Get drive failed: %v", err)
	}

	if drive.ID != driveID {
		t.Errorf("Drive ID mismatch: got %s, want %s", drive.ID, driveID)
	}
	if drive.Name == "" {
		t.Error("Drive name is empty")
	}

	t.Logf("Drive: %s (ID: %s)", drive.Name, drive.ID)
	t.Logf("Kind: %s", drive.Kind)
}

func TestIntegration_Drives_GetDriveFullMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	drivesMgr, _, ctx := setupDrivesManager(t)

	// Get drive with all fields
	reqCtx := api.NewRequestContext("default", driveID, types.RequestTypeGetByID)
	fields := "id,name,kind,colorRgb,backgroundImageLink,themeId,createdTime,hidden,capabilities,restrictions,orgUnitId"
	drive, err := drivesMgr.Get(ctx, reqCtx, driveID, fields)
	if err != nil {
		t.Fatalf("Get drive with full metadata failed: %v", err)
	}

	// Log all metadata
	t.Logf("Drive Full Metadata:")
	t.Logf("  ID: %s", drive.ID)
	t.Logf("  Name: %s", drive.Name)
	t.Logf("  Kind: %s", drive.Kind)
	t.Logf("  Color: %s", drive.ColorRgb)
	t.Logf("  Theme ID: %s", drive.ThemeID)
	t.Logf("  Created: %s", drive.CreatedTime)
	t.Logf("  Hidden: %v", drive.Hidden)
	t.Logf("  Org Unit ID: %s", drive.OrgUnitID)

	if drive.Capabilities != nil {
		t.Logf("  Capabilities:")
		t.Logf("    CanAddChildren: %v", drive.Capabilities.CanAddChildren)
		t.Logf("    CanManageMembers: %v", drive.Capabilities.CanManageMembers)
		t.Logf("    CanEdit: %v", drive.Capabilities.CanEdit)
		t.Logf("    CanShare: %v", drive.Capabilities.CanShare)
		t.Logf("    CanDeleteDrive: %v", drive.Capabilities.CanDeleteDrive)
		t.Logf("    CanRenameDrive: %v", drive.Capabilities.CanRenameDrive)
	}

	if drive.Restrictions != nil {
		t.Logf("  Restrictions:")
		t.Logf("    DomainUsersOnly: %v", drive.Restrictions.DomainUsersOnly)
		t.Logf("    DriveMembersOnly: %v", drive.Restrictions.DriveMembersOnly)
		t.Logf("    AdminManagedRestrictions: %v", drive.Restrictions.AdminManagedRestrictions)
		t.Logf("    CopyRequiresWriterPermission: %v", drive.Restrictions.CopyRequiresWriterPermission)
	}
}

func TestIntegration_Drives_RequestContextIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	driveID := os.Getenv("TEST_SHARED_DRIVE_ID")
	if driveID == "" {
		t.Skip("TEST_SHARED_DRIVE_ID not set")
	}

	drivesMgr, client, ctx := setupDrivesManager(t)

	// Test 1: Create RequestContext with driveID
	reqCtx := api.NewRequestContext("default", driveID, types.RequestTypeListOrSearch)
	if reqCtx.DriveID != driveID {
		t.Errorf("RequestContext DriveID not set correctly: got %s, want %s", reqCtx.DriveID, driveID)
	}

	// Test 2: Verify DriveID is properly maintained
	drive, err := drivesMgr.Get(ctx, reqCtx, driveID, "id,name")
	if err != nil {
		t.Fatalf("Get drive with RequestContext failed: %v", err)
	}

	t.Logf("Successfully retrieved drive %s with RequestContext.DriveID=%s", drive.Name, reqCtx.DriveID)

	// Test 3: Verify TraceID is generated
	if reqCtx.TraceID == "" {
		t.Error("RequestContext TraceID was not generated")
	}
	t.Logf("Request TraceID: %s", reqCtx.TraceID)

	// Test 4: Test shaper integration
	shaper := api.NewRequestShaper(client)
	listCall := client.Service().Files.List()
	shapedCall := shaper.ShapeFilesList(listCall, reqCtx)
	if shapedCall == nil {
		t.Error("ShapeFilesList returned nil")
	}
}

func TestIntegration_Drives_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	drivesMgr, _, ctx := setupDrivesManager(t)

	// Test getting non-existent drive
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	invalidDriveID := "invalid-drive-id-12345"

	_, err := drivesMgr.Get(ctx, reqCtx, invalidDriveID, "id,name")
	if err == nil {
		t.Error("Expected error for invalid drive ID, got nil")
	}

	t.Logf("Got expected error for invalid drive: %v", err)
}

func TestIntegration_Drives_ListEmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test may pass if the user has no Shared Drives
	// It validates that empty results are handled correctly

	drivesMgr, _, ctx := setupDrivesManager(t)

	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
	result, err := drivesMgr.List(ctx, reqCtx, 100, "")
	if err != nil {
		t.Fatalf("List drives failed: %v", err)
	}

	// Should succeed even with no drives
	if result == nil {
		t.Error("Result is nil")
	}

	if result.Drives == nil {
		t.Error("Drives array is nil")
	}

	t.Logf("Drive list returned successfully with %d drives", len(result.Drives))

	// No NextPageToken expected if all drives fit in one page
	if len(result.Drives) < 100 && result.NextPageToken != "" {
		t.Errorf("Unexpected NextPageToken when all drives fit in page: %s", result.NextPageToken)
	}
}
