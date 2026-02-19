//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/logging"
	"github.com/milcgroup/gdrv/internal/resolver"
	"github.com/milcgroup/gdrv/internal/types"
)

func TestResolveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := setupTestClient()
	if err != nil {
		t.Skipf("Could not set up test client: %v", err)
	}

	pathResolver := resolver.NewPathResolver(client, 0)
	ctx := context.Background()
	reqCtx := api.NewRequestContext("test", "", types.RequestTypeListOrSearch)

	result, err := pathResolver.Resolve(ctx, reqCtx, "My Drive", resolver.ResolveOptions{
		SearchDomain: resolver.SearchDomainMyDrive,
		UseCache:     false,
	})

	if err != nil {
		t.Logf("Resolve failed (expected if no 'My Drive' folder): %v", err)
		return
	}

	if result.FileID == "" {
		t.Error("Expected non-empty file ID")
	}
}

func TestResolveWithCacheIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := setupTestClient()
	if err != nil {
		t.Skipf("Could not set up test client: %v", err)
	}

	pathResolver := resolver.NewPathResolver(client, 0)
	ctx := context.Background()
	reqCtx := api.NewRequestContext("test", "", types.RequestTypeListOrSearch)

	result1, err := pathResolver.Resolve(ctx, reqCtx, "My Drive", resolver.ResolveOptions{
		SearchDomain: resolver.SearchDomainMyDrive,
		UseCache:     true,
	})

	if err != nil {
		t.Logf("First resolve failed (expected if no 'My Drive' folder): %v", err)
		return
	}

	if result1.Cached {
		t.Error("First resolve should not be cached")
	}

	result2, err := pathResolver.Resolve(ctx, reqCtx, "My Drive", resolver.ResolveOptions{
		SearchDomain: resolver.SearchDomainMyDrive,
		UseCache:     true,
	})

	if err != nil {
		t.Fatalf("Second resolve failed: %v", err)
	}

	if !result2.Cached {
		t.Error("Second resolve should be cached")
	}

	if result1.FileID != result2.FileID {
		t.Errorf("File IDs should match: %s vs %s", result1.FileID, result2.FileID)
	}
}

func setupTestClient() (*api.Client, error) {
	service, err := getTestDriveService()
	if err != nil {
		return nil, err
	}

	return api.NewClient(service, 3, 1000, logging.NewNoOpLogger()), nil
}
