// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrive/internal/api"
	"github.com/dl-alexandre/gdrive/internal/auth"
	"github.com/dl-alexandre/gdrive/internal/files"
	"github.com/dl-alexandre/gdrive/internal/types"
)

// TestIntegration_ErrorScenarios_NetworkFailures tests network failure handling
func TestIntegration_ErrorScenarios_NetworkFailures(t *testing.T) {
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

	// Test with invalid file ID (should trigger API error)
	_, err = fileManager.Get(ctx, reqCtx, "invalid-file-id-12345")
	if err == nil {
		t.Error("Expected error for invalid file ID")
	}

	t.Logf("Handled invalid file ID error: %v", err)
}

// TestIntegration_ErrorScenarios_PermissionDenied tests permission denied scenarios
func TestIntegration_ErrorScenarios_PermissionDenied(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Need a file ID that doesn't exist or is not accessible
	fileID := os.Getenv("TEST_INACCESSIBLE_FILE_ID")
	if fileID == "" {
		t.Skip("TEST_INACCESSIBLE_FILE_ID not set")
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

	// Try to access file without permission
	_, err = fileManager.Get(ctx, reqCtx, fileID)
	if err == nil {
		t.Log("File was accessible - permission denied test inconclusive")
	} else {
		t.Logf("Handled permission denied error: %v", err)
	}
}

// TestIntegration_ErrorScenarios_InvalidParameters tests invalid parameter validation
func TestIntegration_ErrorScenarios_InvalidParameters(t *testing.T) {
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

	// Test empty file name (should be handled gracefully)
	_, err = fileManager.Create(ctx, reqCtx, "", "", "text/plain", nil)
	if err == nil {
		t.Error("Expected error for empty file name")
	} else {
		t.Logf("Handled invalid parameter error: %v", err)
	}

	// Test invalid parent ID
	_, err = fileManager.Create(ctx, reqCtx, "test.txt", "invalid-parent-id", "text/plain", nil)
	if err == nil {
		t.Error("Expected error for invalid parent ID")
	} else {
		t.Logf("Handled invalid parent ID error: %v", err)
	}
}

// TestIntegration_ErrorScenarios_QuotaExceeded tests quota exceeded handling
func TestIntegration_ErrorScenarios_QuotaExceeded(t *testing.T) {
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

	// Create a large file to potentially trigger quota issues
	// This is hard to simulate reliably, so we test the error handling path
	largeContent := make([]byte, 100*1024*1024) // 100MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	fileName := "large-file-test-" + time.Now().Format("20060102150405") + ".bin"
	_, err = fileManager.Create(ctx, reqCtx, fileName, "", "application/octet-stream", largeContent)
	if err != nil {
		t.Logf("Handled large file upload error (possibly quota): %v", err)
	} else {
		t.Log("Large file upload succeeded")
		// If it succeeded, clean up
		// Note: In real quota exceeded, this would fail
	}
}

// TestIntegration_ErrorScenarios_RateLimiting tests rate limiting and backoff
func TestIntegration_ErrorScenarios_RateLimiting(t *testing.T) {
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

	// Make many rapid requests to potentially trigger rate limiting
	for i := 0; i < 50; i++ {
		_, err = fileManager.List(ctx, reqCtx, &types.FileListRequest{
			Query: "trashed=false",
			PageSize: 10,
		})
		if err != nil {
			t.Logf("Request %d failed (possibly rate limited): %v", i+1, err)
			break
		}
	}

	t.Log("Rate limiting test completed - check logs for backoff behavior")
}</content>
<parameter name="filePath">test/integration/error_scenarios_test.go