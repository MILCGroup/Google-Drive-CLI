//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/auth"
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

	manager := setupAuthManager(t)

	// Test authentication
	ctx := context.Background()
	_, err := manager.Authenticate(ctx, profile, func(string) error { return nil }, auth.OAuthAuthOptions{
		NoBrowser: false,
	})

	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Test token retrieval
	creds, err := manager.GetValidCredentials(ctx, profile)
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	if creds == nil {
		t.Fatal("Credentials are nil")
	}

	if creds.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if time.Now().After(creds.ExpiryDate) {
		t.Error("Credentials are expired")
	}

	// Clean up
	err = manager.DeleteCredentials(profile)
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

	manager := setupAuthManager(t)

	ctx := context.Background()
	_, err := manager.Authenticate(ctx, profile, func(string) error { return nil }, auth.OAuthAuthOptions{
		NoBrowser: true, // Force device code flow
	})

	if err != nil {
		t.Fatalf("Device code authentication failed: %v", err)
	}

	// Verify token
	creds, err := manager.GetValidCredentials(ctx, profile)
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	if creds.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if time.Now().After(creds.ExpiryDate) {
		t.Error("Credentials are expired")
	}

	// Clean up
	_ = manager.DeleteCredentials(profile)
}

func TestIntegration_AuthWorkflow_TokenRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set - skipping token refresh test")
	}

	manager := setupAuthManager(t)

	// Get current token
	creds, err := manager.GetValidCredentials(context.Background(), profile)
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	originalExpiry := creds.ExpiryDate

	// Force token refresh (this would happen automatically on API call)
	// In real usage, the OAuth2 library handles this

	// Verify we can still get a valid token
	newCreds, err := manager.GetValidCredentials(context.Background(), profile)
	if err != nil {
		t.Fatalf("GetValidCredentials after potential refresh failed: %v", err)
	}

	if newCreds.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if time.Now().After(newCreds.ExpiryDate) {
		t.Error("Credentials are expired")
	}

	t.Logf("Original expiry: %v, New expiry: %v", originalExpiry, newCreds.ExpiryDate)
}

func TestIntegration_AuthWorkflow_ProfileManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	manager := setupAuthManager(t)
	ctx := context.Background()

	profiles := []string{
		"workflow-profile-1-" + time.Now().Format("20060102150405"),
		"workflow-profile-2-" + time.Now().Format("20060102150405"),
	}

	// Authenticate multiple profiles
	for _, profile := range profiles {
		_, err := manager.Authenticate(ctx, profile, func(string) error { return nil }, auth.OAuthAuthOptions{})
		if err != nil {
			t.Fatalf("Failed to authenticate profile %s: %v", profile, err)
		}
	}

	// Verify both profiles have valid tokens
	for _, profile := range profiles {
		creds, err := manager.GetValidCredentials(ctx, profile)
		if err != nil {
			t.Errorf("Failed to get credentials for %s: %v", profile, err)
			continue
		}
		if creds.AccessToken == "" || time.Now().After(creds.ExpiryDate) {
			t.Errorf("Credentials for %s are not valid", profile)
		}
	}

	// Clean up
	for _, profile := range profiles {
		_ = manager.DeleteCredentials(profile)
	}
}

func TestIntegration_AuthWorkflow_SecureStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	profile := "storage-workflow-" + time.Now().Format("20060102150405")

	manager := setupAuthManager(t)
	ctx := context.Background()

	// Authenticate
	_, err := manager.Authenticate(ctx, profile, func(string) error { return nil }, auth.OAuthAuthOptions{})
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Create new manager instance to test persistence
	manager2 := setupAuthManager(t)

	// Should be able to retrieve stored credentials
	creds, err := manager2.GetValidCredentials(ctx, profile)
	if err != nil {
		t.Fatalf("Failed to retrieve stored credentials: %v", err)
	}

	if creds == nil || creds.AccessToken == "" || time.Now().After(creds.ExpiryDate) {
		t.Error("Stored credentials are not valid")
	}

	// Clean up
	_ = manager.DeleteCredentials(profile)
}
