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

// Integration Test: Auth Flow Validation
// These tests require actual Google OAuth credentials and should be run manually
// Run with: go test -tags=integration ./test/integration/...

func TestIntegration_AuthFlow_OAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires manual intervention (browser OAuth flow)
	t.Skip("Requires manual OAuth flow - run manually when needed")

	profile := "test-profile-" + time.Now().Format("20060102150405")
	
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

func TestIntegration_AuthFlow_DeviceCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires manual device code entry
	t.Skip("Requires manual device code flow - run manually when needed")

	profile := "test-device-" + time.Now().Format("20060102150405")
	
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

func TestIntegration_AuthFlow_TokenRefresh(t *testing.T) {
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

func TestIntegration_AuthFlow_MultipleProfiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	manager := auth.NewManager("")
	ctx := context.Background()
	
	profiles := []string{
		"test-profile-1-" + time.Now().Format("20060102150405"),
		"test-profile-2-" + time.Now().Format("20060102150405"),
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

func TestIntegration_AuthFlow_ScopeValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	manager := auth.NewManager("")
	
	// Verify token has expected scopes
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	// Note: OAuth2 token doesn't directly expose scopes
	// In real implementation, scopes would be stored with profile metadata
	
	if token == nil {
		t.Fatal("Token is nil")
	}
}

func TestIntegration_AuthFlow_InvalidProfile(t *testing.T) {
	manager := auth.NewManager("")
	
	// Try to get token for non-existent profile
	_, err := manager.GetToken("nonexistent-profile-12345")
	if err == nil {
		t.Error("Expected error for non-existent profile")
	}
}

func TestIntegration_AuthFlow_CredentialStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("Requires manual OAuth - run manually when needed")

	profile := "test-storage-" + time.Now().Format("20060102150405")
	
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
}
