package drives

import (
	"context"
	"net/http"
	"testing"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/logging"
	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func TestManager_New(t *testing.T) {
	service, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	logger := logging.NewNoOpLogger()
	client := api.NewClient(service, 3, 1000, logger)
	mgr := NewManager(client)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.client != client {
		t.Error("Manager client not set correctly")
	}

	if mgr.shaper == nil {
		t.Error("Manager shaper is nil")
	}
}

func TestSharedDrive_Structure(t *testing.T) {
	// Test that SharedDrive struct can be created and has expected fields
	sd := &SharedDrive{
		ID:                  "drive123",
		Name:                "Test Drive",
		Kind:                "drive#drive",
		ColorRgb:            "#ff0000",
		BackgroundImageLink: "https://example.com/bg.jpg",
		ThemeID:             "theme1",
		CreatedTime:         "2024-01-01T00:00:00.000Z",
		Hidden:              false,
		OrgUnitID:           "org123",
	}

	if sd.ID != "drive123" {
		t.Errorf("ID mismatch: got %s, want drive123", sd.ID)
	}
	if sd.Name != "Test Drive" {
		t.Errorf("Name mismatch: got %s, want Test Drive", sd.Name)
	}
	if sd.Kind != "drive#drive" {
		t.Errorf("Kind mismatch: got %s, want drive#drive", sd.Kind)
	}
	if sd.ColorRgb != "#ff0000" {
		t.Errorf("ColorRgb mismatch: got %s, want #ff0000", sd.ColorRgb)
	}
	if sd.Hidden != false {
		t.Error("Hidden should be false")
	}
}

func TestDriveCapabilities_Structure(t *testing.T) {
	caps := &DriveCapabilities{
		CanAddChildren:            true,
		CanManageMembers:          true,
		CanEdit:                   true,
		CanShare:                  false,
		CanDeleteDrive:            false,
		CanRenameDrive:            true,
		CanListChildren:           true,
		CanDownload:               true,
		CanCopy:                   true,
		CanComment:                true,
		CanReadRevisions:          true,
		CanRename:                 true,
		CanChangeDriveBackground:  true,
		CanResetDriveRestrictions: false,
	}

	if !caps.CanAddChildren {
		t.Error("CanAddChildren should be true")
	}
	if !caps.CanManageMembers {
		t.Error("CanManageMembers should be true")
	}
	if caps.CanShare {
		t.Error("CanShare should be false")
	}
}

func TestDriveRestrictions_Structure(t *testing.T) {
	restrictions := &DriveRestrictions{
		AdminManagedRestrictions:     true,
		CopyRequiresWriterPermission: false,
		DomainUsersOnly:              true,
		DriveMembersOnly:             false,
	}

	if !restrictions.AdminManagedRestrictions {
		t.Error("AdminManagedRestrictions should be true")
	}
	if restrictions.CopyRequiresWriterPermission {
		t.Error("CopyRequiresWriterPermission should be false")
	}
	if !restrictions.DomainUsersOnly {
		t.Error("DomainUsersOnly should be true")
	}
}

func TestListResult_Structure(t *testing.T) {
	result := &ListResult{
		Drives: []*SharedDrive{
			{ID: "drive1", Name: "Drive 1"},
			{ID: "drive2", Name: "Drive 2"},
		},
		NextPageToken: "token123",
	}

	if len(result.Drives) != 2 {
		t.Errorf("Expected 2 drives, got %d", len(result.Drives))
	}

	if result.NextPageToken != "token123" {
		t.Errorf("NextPageToken mismatch: got %s, want token123", result.NextPageToken)
	}
}

func TestMapDriveToSharedDrive_BasicFields(t *testing.T) {
	apiDrive := &drive.Drive{
		Id:                  "drive123",
		Name:                "Test Drive",
		Kind:                "drive#drive",
		ColorRgb:            "#00ff00",
		BackgroundImageLink: "https://example.com/image.png",
		ThemeId:             "theme456",
		CreatedTime:         "2024-01-15T12:00:00.000Z",
		Hidden:              true,
		OrgUnitId:           "org456",
	}

	sd := mapDriveToSharedDrive(apiDrive)

	if sd.ID != "drive123" {
		t.Errorf("ID mismatch: got %s, want drive123", sd.ID)
	}
	if sd.Name != "Test Drive" {
		t.Errorf("Name mismatch: got %s, want Test Drive", sd.Name)
	}
	if sd.Kind != "drive#drive" {
		t.Errorf("Kind mismatch: got %s, want drive#drive", sd.Kind)
	}
	if sd.ColorRgb != "#00ff00" {
		t.Errorf("ColorRgb mismatch: got %s, want #00ff00", sd.ColorRgb)
	}
	if sd.BackgroundImageLink != "https://example.com/image.png" {
		t.Errorf("BackgroundImageLink mismatch")
	}
	if sd.ThemeID != "theme456" {
		t.Errorf("ThemeID mismatch: got %s, want theme456", sd.ThemeID)
	}
	if sd.CreatedTime != "2024-01-15T12:00:00.000Z" {
		t.Errorf("CreatedTime mismatch")
	}
	if !sd.Hidden {
		t.Error("Hidden should be true")
	}
	if sd.OrgUnitID != "org456" {
		t.Errorf("OrgUnitID mismatch: got %s, want org456", sd.OrgUnitID)
	}
}

func TestMapDriveToSharedDrive_WithCapabilities(t *testing.T) {
	apiDrive := &drive.Drive{
		Id:   "drive123",
		Name: "Test Drive",
		Capabilities: &drive.DriveCapabilities{
			CanAddChildren:            true,
			CanManageMembers:          true,
			CanEdit:                   false,
			CanShare:                  true,
			CanDeleteDrive:            false,
			CanRenameDrive:            true,
			CanListChildren:           true,
			CanDownload:               true,
			CanCopy:                   false,
			CanComment:                true,
			CanReadRevisions:          true,
			CanRename:                 false,
			CanChangeDriveBackground:  true,
			CanResetDriveRestrictions: false,
			CanChangeCopyRequiresWriterPermissionRestriction: true,
			CanChangeDomainUsersOnlyRestriction:              false,
			CanChangeDriveMembersOnlyRestriction:             true,
			CanTrashChildren:                                 true,
		},
	}

	sd := mapDriveToSharedDrive(apiDrive)

	if sd.Capabilities == nil {
		t.Fatal("Capabilities is nil")
	}

	caps := sd.Capabilities
	if !caps.CanAddChildren {
		t.Error("CanAddChildren should be true")
	}
	if !caps.CanManageMembers {
		t.Error("CanManageMembers should be true")
	}
	if caps.CanEdit {
		t.Error("CanEdit should be false")
	}
	if !caps.CanShare {
		t.Error("CanShare should be true")
	}
	if caps.CanDeleteDrive {
		t.Error("CanDeleteDrive should be false")
	}
	if !caps.CanRenameDrive {
		t.Error("CanRenameDrive should be true")
	}
	if !caps.CanChangeCopyRequiresWriterPermissionRestriction {
		t.Error("CanChangeCopyRequiresWriterPermissionRestriction should be true")
	}
	if caps.CanChangeDomainUsersOnlyRestriction {
		t.Error("CanChangeDomainUsersOnlyRestriction should be false")
	}
	if !caps.CanChangeDriveMembersOnlyRestriction {
		t.Error("CanChangeDriveMembersOnlyRestriction should be true")
	}
	if !caps.CanTrashChildren {
		t.Error("CanTrashChildren should be true")
	}
}

func TestMapDriveToSharedDrive_WithRestrictions(t *testing.T) {
	apiDrive := &drive.Drive{
		Id:   "drive123",
		Name: "Test Drive",
		Restrictions: &drive.DriveRestrictions{
			AdminManagedRestrictions:     true,
			CopyRequiresWriterPermission: true,
			DomainUsersOnly:              false,
			DriveMembersOnly:             true,
		},
	}

	sd := mapDriveToSharedDrive(apiDrive)

	if sd.Restrictions == nil {
		t.Fatal("Restrictions is nil")
	}

	rest := sd.Restrictions
	if !rest.AdminManagedRestrictions {
		t.Error("AdminManagedRestrictions should be true")
	}
	if !rest.CopyRequiresWriterPermission {
		t.Error("CopyRequiresWriterPermission should be true")
	}
	if rest.DomainUsersOnly {
		t.Error("DomainUsersOnly should be false")
	}
	if !rest.DriveMembersOnly {
		t.Error("DriveMembersOnly should be true")
	}
}

func TestMapDriveToSharedDrive_NilCapabilitiesAndRestrictions(t *testing.T) {
	apiDrive := &drive.Drive{
		Id:           "drive123",
		Name:         "Test Drive",
		Capabilities: nil,
		Restrictions: nil,
	}

	sd := mapDriveToSharedDrive(apiDrive)

	if sd.Capabilities != nil {
		t.Error("Capabilities should be nil when not set in API response")
	}
	if sd.Restrictions != nil {
		t.Error("Restrictions should be nil when not set in API response")
	}
}

func TestMapDriveToSharedDrive_EmptyDrive(t *testing.T) {
	// Test with minimal drive data
	apiDrive := &drive.Drive{
		Id:   "drive123",
		Name: "",
	}

	sd := mapDriveToSharedDrive(apiDrive)

	if sd.ID != "drive123" {
		t.Errorf("ID mismatch: got %s, want drive123", sd.ID)
	}
	if sd.Name != "" {
		t.Errorf("Name should be empty string, got %s", sd.Name)
	}
	if sd.Capabilities != nil {
		t.Error("Capabilities should be nil")
	}
	if sd.Restrictions != nil {
		t.Error("Restrictions should be nil")
	}
}

func TestRequestContext_DriveID(t *testing.T) {
	// Test that RequestContext properly carries drive context
	driveID := "test-drive-123"
	reqCtx := api.NewRequestContext("default", driveID, types.RequestTypeListOrSearch)

	if reqCtx.DriveID != driveID {
		t.Errorf("DriveID mismatch: got %s, want %s", reqCtx.DriveID, driveID)
	}

	if reqCtx.Profile != "default" {
		t.Errorf("Profile mismatch: got %s, want default", reqCtx.Profile)
	}

	if reqCtx.RequestType != types.RequestTypeListOrSearch {
		t.Errorf("RequestType mismatch: got %s, want %s", reqCtx.RequestType, types.RequestTypeListOrSearch)
	}

	if reqCtx.TraceID == "" {
		t.Error("TraceID should be generated")
	}
}
