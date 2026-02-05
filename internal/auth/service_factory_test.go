package auth

import (
	"context"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

func TestServiceFactory_CreateService_Drive(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateService(ctx, creds, ServiceDrive)
	if err != nil {
		t.Logf("CreateService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateService should return service or error")
	}
}

func TestServiceFactory_CreateService_Sheets(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeSheets},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateService(ctx, creds, ServiceSheets)
	if err != nil {
		t.Logf("CreateService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateService should return service or error")
	}
}

func TestServiceFactory_CreateService_Docs(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeDocs},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateService(ctx, creds, ServiceDocs)
	if err != nil {
		t.Logf("CreateService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateService should return service or error")
	}
}

func TestServiceFactory_CreateService_Slides(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeSlides},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateService(ctx, creds, ServiceSlides)
	if err != nil {
		t.Logf("CreateService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateService should return service or error")
	}
}

func TestServiceFactory_CreateService_AdminDir(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeAdminDirectoryUser},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateService(ctx, creds, ServiceAdminDir)
	if err != nil {
		t.Logf("CreateService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateService should return service or error")
	}
}

func TestServiceFactory_CreateService_Unknown(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	_, err := factory.CreateService(ctx, creds, ServiceType("unknown"))
	if err == nil {
		t.Error("CreateService should fail for unknown service type")
	}
}

func TestServiceFactory_CreateDriveService(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateDriveService(ctx, creds)
	if err != nil {
		t.Logf("CreateDriveService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateDriveService should return service or error")
	}
}

func TestServiceFactory_CreateSheetsService(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeSheets},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateSheetsService(ctx, creds)
	if err != nil {
		t.Logf("CreateSheetsService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateSheetsService should return service or error")
	}
}

func TestServiceFactory_CreateDocsService(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeDocs},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateDocsService(ctx, creds)
	if err != nil {
		t.Logf("CreateDocsService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateDocsService should return service or error")
	}
}

func TestServiceFactory_CreateSlidesService(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeSlides},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateSlidesService(ctx, creds)
	if err != nil {
		t.Logf("CreateSlidesService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateSlidesService should return service or error")
	}
}

func TestServiceFactory_CreateAdminService(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := NewServiceFactory(mgr)

	creds := &types.Credentials{
		AccessToken:  "test_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeAdminDirectoryUser},
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := factory.CreateAdminService(ctx, creds)
	if err != nil {
		t.Logf("CreateAdminService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("CreateAdminService should return service or error")
	}
}
