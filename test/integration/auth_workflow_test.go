// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// Integration Test: Auth Workflow Validation
// These tests require actual Google OAuth credentials and should be run manually
// Run with: go test -tags=integration ./test/integration/...

func TestIntegration_AuthWorkflow_OAuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires manual intervention (browser OAuth flow)
	t.Skip("Requires manual OAuth flow - run manually when needed")

	profile := "workflow-test-" + time.Now().Format("20060102150405")

	manager := auth.NewManager("")

	// Test authentication
	ctx := context.Background()
	err := manager.Authenticate(ctx, profile, types.AuthOptions{
		Scopes:    []string{"https://www.googleapis.com/auth/drive"},
		NoBrowser: false,
	})

	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Test token retrieval
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token == nil {
		t.Fatal("Token is nil")
	}

	if !token.Valid() {
		t.Error("Token is not valid")
	}

	// Clean up
	err = manager.RemoveProfile(profile)
	if err != nil {
		t.Errorf("Failed to remove test profile: %v", err)
	}
}

func TestIntegration_AuthWorkflow_DeviceCodeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires manual device code entry
	t.Skip("Requires manual device code flow - run manually when needed")

	profile := "device-workflow-" + time.Now().Format("20060102150405")

	manager := auth.NewManager("")

	ctx := context.Background()
	err := manager.Authenticate(ctx, profile, types.AuthOptions{
		Scopes:    []string{"https://www.googleapis.com/auth/drive"},
		NoBrowser: true, // Force device code flow
	})

	if err != nil {
		t.Fatalf("Device code authentication failed: %v", err)
	}

	// Verify token
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if !token.Valid() {
		t.Error("Token is not valid")
	}

	// Clean up
	manager.RemoveProfile(profile)
}

func TestIntegration_AuthWorkflow_TokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set - skipping token refresh test")
	}

	manager := auth.NewManager("")

	// Get current token
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	originalExpiry := token.Expiry

	// Force token refresh (this would happen automatically on API call)
	// In real usage, the OAuth2 library handles this

	// Verify we can still get a valid token
	newToken, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken after potential refresh failed: %v", err)
	}

	if !newToken.Valid() {
		t.Error("Refreshed token is not valid")
	}

	t.Logf("Original expiry: %v, New expiry: %v", originalExpiry, newToken.Expiry)
}

func TestIntegration_AuthWorkflow_ProfileManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	manager := auth.NewManager("")
	ctx := context.Background()

	profiles := []string{
		"workflow-profile-1-" + time.Now().Format("20060102150405"),
		"workflow-profile-2-" + time.Now().Format("20060102150405"),
	}

	// Authenticate multiple profiles
	for _, profile := range profiles {
		err := manager.Authenticate(ctx, profile, types.AuthOptions{
			Scopes: []string{"https://www.googleapis.com/auth/drive"},
		})
		if err != nil {
			t.Fatalf("Failed to authenticate profile %s: %v", profile, err)
		}
	}

	// Verify both profiles have valid tokens
	for _, profile := range profiles {
		token, err := manager.GetToken(profile)
		if err != nil {
			t.Errorf("Failed to get token for %s: %v", profile, err)
			continue
		}
		if !token.Valid() {
			t.Errorf("Token for %s is not valid", profile)
		}
	}

	// Clean up
	for _, profile := range profiles {
		manager.RemoveProfile(profile)
	}
}

func TestIntegration_AuthWorkflow_SecureStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	profile := "storage-workflow-" + time.Now().Format("20060102150405")

	manager := auth.NewManager("")
	ctx := context.Background()

	// Authenticate
	err := manager.Authenticate(ctx, profile, types.AuthOptions{
		Scopes: []string{"https://www.googleapis.com/auth/drive"},
	})
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Create new manager instance to test persistence
	manager2 := auth.NewManager("")

	// Should be able to retrieve stored credentials
	token, err := manager2.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to retrieve stored credentials: %v", err)
	}

	if token == nil || !token.Valid() {
		t.Error("Stored credentials are not valid")
	}

	// Clean up
	manager.RemoveProfile(profile)
}</content>
<parameter name="filePath">test/integration/auth_workflow_test.go