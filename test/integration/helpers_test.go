//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/logging"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"google.golang.org/api/drive/v3"
)

func setupAuthManager(t *testing.T) *auth.Manager {
	t.Helper()
	manager := auth.NewManager("")

	clientID := strings.TrimSpace(os.Getenv("GDRV_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GDRV_CLIENT_SECRET"))
	if clientID != "" {
		manager.SetOAuthConfig(clientID, clientSecret, []string{utils.ScopeFull})
		return manager
	}
	if bundledID, bundledSecret, ok := auth.GetBundledOAuthClient(); ok {
		manager.SetOAuthConfig(bundledID, bundledSecret, []string{utils.ScopeFull})
	}
	return manager
}

func loadTestCredentials(t *testing.T, profile string) *types.Credentials {
	t.Helper()
	manager := setupAuthManager(t)
	creds, err := manager.GetValidCredentials(context.Background(), profile)
	if err != nil {
		t.Fatalf("Failed to load credentials for profile %q: %v", profile, err)
	}
	return creds
}

func setupDriveClient(t *testing.T) (*api.Client, *drive.Service, string) {
	t.Helper()
	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	manager := setupAuthManager(t)
	ctx := context.Background()
	creds, err := manager.GetValidCredentials(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to get credentials: %v", err)
	}
	service, err := manager.GetDriveService(ctx, creds)
	if err != nil {
		t.Fatalf("Failed to create Drive service: %v", err)
	}
	client := api.NewClient(service, 3, 1000, logging.NewNoOpLogger())
	return client, service, profile
}

func uploadTempFile(t *testing.T, ctx context.Context, manager *files.Manager, reqCtx *types.RequestContext, name, parentID, mimeType string, contents []byte) *types.DriveFile {
	t.Helper()
	if contents == nil {
		contents = []byte("integration test")
	}

	tmpPath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(tmpPath, contents, 0644); err != nil {
		t.Fatalf("Failed to create temp file %q: %v", tmpPath, err)
	}

	opts := files.UploadOptions{
		ParentID: parentID,
		Name:     name,
		MimeType: mimeType,
	}
	file, err := manager.Upload(ctx, reqCtx, tmpPath, opts)
	if err != nil {
		t.Fatalf("Failed to upload file %q: %v", name, err)
	}
	return file
}
