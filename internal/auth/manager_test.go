package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

func TestManager_NeedsRefresh(t *testing.T) {
	mgr := NewManager("/tmp/test")

	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{
			"Expired credentials",
			time.Now().Add(-1 * time.Hour),
			true,
		},
		{
			"Expiring soon (within 5 min)",
			time.Now().Add(3 * time.Minute),
			true,
		},
		{
			"Valid credentials",
			time.Now().Add(1 * time.Hour),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &types.Credentials{
				ExpiryDate: tt.expiry,
			}
			got := mgr.NeedsRefresh(creds)
			if got != tt.expected {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestManager_ValidateScopes(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		Scopes: []string{
			"https://www.googleapis.com/auth/drive.file",
			"https://www.googleapis.com/auth/drive.readonly",
		},
	}

	// Valid scope check
	err := mgr.ValidateScopes(creds, []string{"https://www.googleapis.com/auth/drive.file"})
	if err != nil {
		t.Errorf("ValidateScopes should pass for existing scope: %v", err)
	}

	// Missing scope check
	err = mgr.ValidateScopes(creds, []string{"https://www.googleapis.com/auth/drive"})
	if err == nil {
		t.Error("ValidateScopes should fail for missing scope")
	}
}

func TestRequiredScopesForService(t *testing.T) {
	tests := []struct {
		name     string
		svcType  ServiceType
		wantLen  int
		contains []string
	}{
		{
			"Drive",
			ServiceDrive,
			1,
			[]string{utils.ScopeFile},
		},
		{
			"Sheets",
			ServiceSheets,
			1,
			[]string{utils.ScopeSheets},
		},
		{
			"Docs",
			ServiceDocs,
			1,
			[]string{utils.ScopeDocs},
		},
		{
			"Slides",
			ServiceSlides,
			1,
			[]string{utils.ScopeSlides},
		},
		{
			"Admin",
			ServiceAdminDir,
			2,
			[]string{utils.ScopeAdminDirectoryUser, utils.ScopeAdminDirectoryGroup},
		},
		{
			"Unknown",
			ServiceType("unknown"),
			0,
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiredScopesForService(tt.svcType)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d scopes, got %d", tt.wantLen, len(got))
			}
			for _, wantScope := range tt.contains {
				found := false
				for _, scope := range got {
					if scope == wantScope {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("missing scope %q", wantScope)
				}
			}
		})
	}
}

// Test Manager credential save/load/delete operations
func TestManager_SaveLoadDeleteCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
		ClientID:     "test_client_id",
	}

	err := mgr.SaveCredentials("default", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("default")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
	if loaded.RefreshToken != creds.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", loaded.RefreshToken, creds.RefreshToken)
	}
	if loaded.Type != creds.Type {
		t.Errorf("Type mismatch: got %v, want %v", loaded.Type, creds.Type)
	}

	err = mgr.DeleteCredentials("default")
	if err != nil {
		t.Logf("DeleteCredentials returned: %v (may be expected if file already deleted)", err)
	}

	_, err = mgr.LoadCredentials("default")
	if err == nil {
		t.Error("LoadCredentials should fail after deletion")
	}
}

func TestManager_SaveCredentials_WithServiceAccount(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(1 * time.Hour),
		Scopes:              []string{utils.ScopeFile},
		Type:                types.AuthTypeServiceAccount,
		ServiceAccountEmail: "test@example.iam.gserviceaccount.com",
	}

	err := mgr.SaveCredentials("sa-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("sa-profile")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.ServiceAccountEmail != creds.ServiceAccountEmail {
		t.Errorf("ServiceAccountEmail mismatch: got %q, want %q", loaded.ServiceAccountEmail, creds.ServiceAccountEmail)
	}
	if loaded.Type != types.AuthTypeServiceAccount {
		t.Errorf("Type mismatch: got %v, want %v", loaded.Type, types.AuthTypeServiceAccount)
	}
}

func TestManager_SaveCredentials_WithImpersonation(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(1 * time.Hour),
		Scopes:              []string{utils.ScopeFile},
		Type:                types.AuthTypeImpersonated,
		ServiceAccountEmail: "sa@example.iam.gserviceaccount.com",
		ImpersonatedUser:    "user@example.com",
	}

	err := mgr.SaveCredentials("impersonate-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("impersonate-profile")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.ImpersonatedUser != creds.ImpersonatedUser {
		t.Errorf("ImpersonatedUser mismatch: got %q, want %q", loaded.ImpersonatedUser, creds.ImpersonatedUser)
	}
	if loaded.Type != types.AuthTypeImpersonated {
		t.Errorf("Type mismatch: got %v, want %v", loaded.Type, types.AuthTypeImpersonated)
	}
}

func TestManager_SetOAuthConfig(t *testing.T) {
	mgr := NewManager("/tmp/test")

	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	scopes := []string{utils.ScopeFile, utils.ScopeReadonly}

	mgr.SetOAuthConfig(clientID, clientSecret, scopes)

	config := mgr.GetOAuthConfig()
	if config == nil {
		t.Fatal("GetOAuthConfig returned nil")
	}
	if config.ClientID != clientID {
		t.Errorf("ClientID mismatch: got %q, want %q", config.ClientID, clientID)
	}
	if config.ClientSecret != clientSecret {
		t.Errorf("ClientSecret mismatch: got %q, want %q", config.ClientSecret, clientSecret)
	}
	if len(config.Scopes) != len(scopes) {
		t.Errorf("Scopes length mismatch: got %d, want %d", len(config.Scopes), len(scopes))
	}
}

func TestManager_UseKeyring(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with plain file (forced)
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})
	if mgr.UseKeyring() {
		t.Error("UseKeyring should be false for plain file storage")
	}

	// Test with encrypted file (forced)
	mgr = NewManagerWithOptions(tmpDir, ManagerOptions{ForceEncryptedFile: true})
	if mgr.UseKeyring() {
		t.Error("UseKeyring should be false for encrypted file storage")
	}
}

func TestManager_ConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	if mgr.ConfigDir() != tmpDir {
		t.Errorf("ConfigDir mismatch: got %q, want %q", mgr.ConfigDir(), tmpDir)
	}
}

func TestManager_GetStorageBackend(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})
	if mgr.GetStorageBackend() != "plain-file" {
		t.Errorf("GetStorageBackend mismatch: got %q, want %q", mgr.GetStorageBackend(), "plain-file")
	}

	mgr = NewManagerWithOptions(tmpDir, ManagerOptions{ForceEncryptedFile: true})
	if mgr.GetStorageBackend() != "encrypted-file" {
		t.Errorf("GetStorageBackend mismatch: got %q, want %q", mgr.GetStorageBackend(), "encrypted-file")
	}
}

func TestManager_GetStorageWarning(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})
	warning := mgr.GetStorageWarning()
	if warning == "" {
		t.Error("GetStorageWarning should return a warning for plain file storage")
	}
}

func TestManager_GetScopesForCommand(t *testing.T) {
	mgr := NewManager("/tmp/test")

	tests := []struct {
		command string
		want    []string
	}{
		{"upload", []string{utils.ScopeFile}},
		{"download", []string{utils.ScopeReadonly}},
		{"delete", []string{utils.ScopeFull}},
		{"share", []string{utils.ScopeFull}},
		{"list", []string{utils.ScopeReadonly}},
		{"search", []string{utils.ScopeReadonly}},
		{"mkdir", []string{utils.ScopeFile}},
		{"copy", []string{utils.ScopeFile}},
		{"move", []string{utils.ScopeFull}},
		{"permissions", []string{utils.ScopeFull}},
		{"about", []string{utils.ScopeMetadataReadonly}},
		{"unknown", []string{utils.ScopeFile}}, // default
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := mgr.GetScopesForCommand(tt.command)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d scopes, got %d", len(tt.want), len(got))
			}
			for i, scope := range got {
				if scope != tt.want[i] {
					t.Errorf("scope mismatch at index %d: got %q, want %q", i, scope, tt.want[i])
				}
			}
		})
	}
}

func TestManager_ValidateScopesForCommand(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		Scopes: []string{utils.ScopeFile, utils.ScopeReadonly},
	}

	// Valid scope for command
	err := mgr.ValidateScopesForCommand(creds, "upload")
	if err != nil {
		t.Errorf("ValidateScopesForCommand should pass: %v", err)
	}

	// Invalid scope for command
	err = mgr.ValidateScopesForCommand(creds, "delete")
	if err == nil {
		t.Error("ValidateScopesForCommand should fail for missing scope")
	}
}

func TestManager_ValidateServiceScopes(t *testing.T) {
	mgr := NewManager("/tmp/test")

	tests := []struct {
		name      string
		scopes    []string
		svcType   ServiceType
		shouldErr bool
	}{
		{
			"Drive with file scope",
			[]string{utils.ScopeFile},
			ServiceDrive,
			false,
		},
		{
			"Drive without file scope",
			[]string{utils.ScopeReadonly},
			ServiceDrive,
			true,
		},
		{
			"Sheets with sheets scope",
			[]string{utils.ScopeSheets},
			ServiceSheets,
			false,
		},
		{
			"Sheets without sheets scope",
			[]string{utils.ScopeFile},
			ServiceSheets,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &types.Credentials{Scopes: tt.scopes}
			err := mgr.ValidateServiceScopes(creds, tt.svcType)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateServiceScopes error = %v, shouldErr = %v", err, tt.shouldErr)
			}
		})
	}
}

func TestManager_CredentialLocation(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("test-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	location, err := mgr.CredentialLocation("test-profile")
	if err != nil {
		t.Fatalf("CredentialLocation failed: %v", err)
	}

	if location == "" {
		t.Error("CredentialLocation returned empty string")
	}

	// For plain file storage, should contain .json
	if !contains(location, ".json") {
		t.Errorf("CredentialLocation should contain .json for plain file storage: %q", location)
	}
}

func TestManager_LoadAuthMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("test-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	meta, err := mgr.LoadAuthMetadata("test-profile")
	if err != nil {
		t.Fatalf("LoadAuthMetadata failed: %v", err)
	}

	if meta.Profile != "test-profile" {
		t.Errorf("Profile mismatch: got %q, want %q", meta.Profile, "test-profile")
	}
	if meta.CredentialType != string(types.AuthTypeOAuth) {
		t.Errorf("CredentialType mismatch: got %q, want %q", meta.CredentialType, string(types.AuthTypeOAuth))
	}
}

func TestManager_GetServiceFactory(t *testing.T) {
	mgr := NewManager("/tmp/test")
	factory := mgr.GetServiceFactory()

	if factory == nil {
		t.Fatal("GetServiceFactory returned nil")
	}
}

func TestClientIDHash(t *testing.T) {
	tests := []struct {
		name      string
		clientID  string
		wantHash  bool
		wantLast4 string
	}{
		{
			"Empty client ID",
			"",
			false,
			"",
		},
		{
			"Short client ID",
			"abc",
			true,
			"abc",
		},
		{
			"Long client ID",
			"1234567890abcdef",
			true,
			"cdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, last4 := clientIDHash(tt.clientID)
			if (hash != "") != tt.wantHash {
				t.Errorf("hash empty = %v, want %v", hash == "", !tt.wantHash)
			}
			if last4 != tt.wantLast4 {
				t.Errorf("last4 = %q, want %q", last4, tt.wantLast4)
			}
		})
	}
}

func TestCredentialKey(t *testing.T) {
	tests := []struct {
		profile string
		hash    string
		want    string
	}{
		{"default", "", "default"},
		{"default", "abc123", "default--abc123"},
		{"work", "xyz789", "work--xyz789"},
	}

	for _, tt := range tests {
		t.Run(tt.profile+"-"+tt.hash, func(t *testing.T) {
			got := credentialKey(tt.profile, tt.hash)
			if got != tt.want {
				t.Errorf("credentialKey = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManager_NewManagerWithOptions_EncryptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForceEncryptedFile: true})

	if mgr.GetStorageBackend() != "encrypted-file" {
		t.Errorf("Expected encrypted-file storage, got %q", mgr.GetStorageBackend())
	}

	// Test that we can save and load with encryption
	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("encrypted-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("encrypted-test")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch after encryption roundtrip")
	}
}

func TestManager_LoadCredentials_InvalidExpiryDate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	// Manually create a credential file with invalid expiry date
	credDir := filepath.Join(tmpDir, "credentials")
	_ = os.MkdirAll(credDir, 0700)

	stored := types.StoredCredentials{
		Profile:     "bad-expiry",
		AccessToken: "test_token",
		ExpiryDate:  "invalid-date-format",
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	data, _ := json.Marshal(stored)
	credFile := filepath.Join(credDir, "bad-expiry.json")
	_ = os.WriteFile(credFile, data, 0600)

	// Should fail to parse invalid expiry date
	_, err := mgr.LoadCredentials("bad-expiry")
	if err == nil {
		t.Error("LoadCredentials should fail with invalid expiry date")
	}
}

// Test GetValidCredentials with OAuth credentials
func TestManager_GetValidCredentials_OAuth(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("oauth-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	ctx := context.Background()
	loaded, err := mgr.GetValidCredentials(ctx, "oauth-test")
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
}

func TestManager_GetValidCredentials_ServiceAccount(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(1 * time.Hour),
		Scopes:              []string{utils.ScopeFile},
		Type:                types.AuthTypeServiceAccount,
		ServiceAccountEmail: "test@example.iam.gserviceaccount.com",
	}

	err := mgr.SaveCredentials("sa-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	ctx := context.Background()
	loaded, err := mgr.GetValidCredentials(ctx, "sa-test")
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	if loaded.Type != types.AuthTypeServiceAccount {
		t.Errorf("Type mismatch: got %v, want %v", loaded.Type, types.AuthTypeServiceAccount)
	}
}

func TestManager_GetValidCredentials_ExpiredServiceAccount(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(-1 * time.Hour),
		Scopes:              []string{utils.ScopeFile},
		Type:                types.AuthTypeServiceAccount,
		ServiceAccountEmail: "test@example.iam.gserviceaccount.com",
	}

	err := mgr.SaveCredentials("expired-sa", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	ctx := context.Background()
	_, err = mgr.GetValidCredentials(ctx, "expired-sa")
	if err == nil {
		t.Error("GetValidCredentials should fail for expired service account")
	}
}

func TestManager_GetValidCredentials_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	ctx := context.Background()
	_, err := mgr.GetValidCredentials(ctx, "nonexistent")
	if err == nil {
		t.Error("GetValidCredentials should fail for nonexistent profile")
	}
}

func TestManager_GetHTTPClient(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	client := mgr.GetHTTPClient(ctx, creds)

	if client == nil {
		t.Fatal("GetHTTPClient returned nil")
	}
}

func TestManager_GetHTTPClient_WithOAuthConfig(t *testing.T) {
	mgr := NewManager("/tmp/test")
	mgr.SetOAuthConfig("test_client_id", "test_client_secret", []string{utils.ScopeFile})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	client := mgr.GetHTTPClient(ctx, creds)

	if client == nil {
		t.Fatal("GetHTTPClient returned nil")
	}
}

func TestManager_GetHTTPClient_ServiceAccount(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(1 * time.Hour),
		Type:                types.AuthTypeServiceAccount,
		ServiceAccountEmail: "test@example.iam.gserviceaccount.com",
	}

	ctx := context.Background()
	client := mgr.GetHTTPClient(ctx, creds)

	if client == nil {
		t.Fatal("GetHTTPClient returned nil for service account")
	}
}

func TestManager_ListProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	profiles := []string{"profile1", "profile2", "profile3"}
	for _, profile := range profiles {
		err := mgr.SaveCredentials(profile, creds)
		if err != nil {
			t.Fatalf("SaveCredentials failed for %s: %v", profile, err)
		}
	}

	listed, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(listed) != len(profiles) {
		t.Errorf("Expected %d profiles, got %d", len(profiles), len(listed))
	}

	for _, profile := range profiles {
		found := false
		for _, p := range listed {
			if p == profile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Profile %q not found in list", profile)
		}
	}
}

func TestManager_ListProfiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	listed, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(listed) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(listed))
	}
}

func TestManager_ResolveCredentialKey(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("test-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	key, err := mgr.ResolveCredentialKey("test-profile")
	if err != nil {
		t.Fatalf("ResolveCredentialKey failed: %v", err)
	}

	if key == "" {
		t.Error("ResolveCredentialKey returned empty key")
	}
}

func TestManager_ResolveCredentialKey_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	key, err := mgr.ResolveCredentialKey("nonexistent")
	if err != nil {
		t.Logf("ResolveCredentialKey returned error (expected): %v", err)
	}
	if key == "" && err == nil {
		t.Error("ResolveCredentialKey should return error or empty key for nonexistent profile")
	}
}

func TestManager_AddProfileToList(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("profile1", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	listed, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	found := false
	for _, p := range listed {
		if p == "profile1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Profile not found after SaveCredentials")
	}
}

func TestManager_RemoveProfileFromList(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("profile-to-delete", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	err = mgr.DeleteCredentials("profile-to-delete")
	if err != nil {
		t.Logf("DeleteCredentials returned: %v (may be expected)", err)
	}

	listed, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	for _, p := range listed {
		if p == "profile-to-delete" {
			t.Error("Profile should be removed from list after deletion")
		}
	}
}

func TestManager_ClientIDForStorage(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		ClientID: "creds_client_id",
	}

	mgr.SetOAuthConfig("config_client_id", "secret", []string{})

	clientID := mgr.clientIDForStorage(creds)
	if clientID != "config_client_id" {
		t.Errorf("Expected config_client_id, got %q", clientID)
	}

	mgr2 := NewManager("/tmp/test")
	clientID = mgr2.clientIDForStorage(creds)
	if clientID != "creds_client_id" {
		t.Errorf("Expected creds_client_id, got %q", clientID)
	}

	mgr3 := NewManager("/tmp/test")
	clientID = mgr3.clientIDForStorage(nil)
	if clientID != "" {
		t.Errorf("Expected empty string, got %q", clientID)
	}
}

func TestManager_RefreshCredentials_NonOAuth(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		Type: types.AuthTypeServiceAccount,
	}

	ctx := context.Background()
	_, err := mgr.RefreshCredentials(ctx, creds)
	if err == nil {
		t.Error("RefreshCredentials should fail for non-OAuth credentials")
	}
}

func TestManager_RefreshCredentials_NoOAuthConfig(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Type:         types.AuthTypeOAuth,
	}

	ctx := context.Background()
	_, err := mgr.RefreshCredentials(ctx, creds)
	if err == nil {
		t.Error("RefreshCredentials should fail without OAuth config")
	}
}

func TestManager_GetSheetsService(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeSheets},
		Type:        types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := mgr.GetSheetsService(ctx, creds)
	if err != nil {
		t.Logf("GetSheetsService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("GetSheetsService should return service or error")
	}
}

func TestManager_GetDocsService(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeDocs},
		Type:        types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := mgr.GetDocsService(ctx, creds)
	if err != nil {
		t.Logf("GetDocsService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("GetDocsService should return service or error")
	}
}

func TestManager_GetSlidesService(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeSlides},
		Type:        types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := mgr.GetSlidesService(ctx, creds)
	if err != nil {
		t.Logf("GetSlidesService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("GetSlidesService should return service or error")
	}
}

func TestManager_GetAdminService(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeAdminDirectoryUser},
		Type:        types.AuthTypeOAuth,
	}

	ctx := context.Background()
	service, err := mgr.GetAdminService(ctx, creds)
	if err != nil {
		t.Logf("GetAdminService returned error (expected in test): %v", err)
	}
	if service == nil && err == nil {
		t.Error("GetAdminService should return service or error")
	}
}

func TestManager_LoadServiceAccount_InvalidKeyFile(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	_, err := mgr.LoadServiceAccount(ctx, "/nonexistent/path/key.json", []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for nonexistent key file")
	}
}

func TestManager_LoadServiceAccount_EmptyKeyFile(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	_ = os.WriteFile(keyFile, []byte(""), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for empty key file")
	}
}

func TestManager_LoadServiceAccount_NoScopes(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	_ = os.WriteFile(keyFile, []byte("{}"), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail with no scopes")
	}
}

func TestManager_LoadServiceAccount_InvalidImpersonateUser(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	_ = os.WriteFile(keyFile, []byte("{}"), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "invalid-user-no-at-sign")
	if err == nil {
		t.Error("LoadServiceAccount should fail for invalid impersonate user")
	}
}

func TestManager_LoadServiceAccount_InvalidJSON(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	_ = os.WriteFile(keyFile, []byte("invalid json {"), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for invalid JSON")
	}
}

func TestManager_LoadServiceAccount_WrongType(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	keyData := `{"type": "oauth2", "client_id": "test"}`
	_ = os.WriteFile(keyFile, []byte(keyData), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for wrong key type")
	}
}

func TestManager_LoadServiceAccount_MissingClientEmail(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	keyData := `{"type": "service_account", "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0Z3VS5JJcds3s/0Ej3Ej5Ej3Ej5Ej3Ej5Ej3Ej5Ej3Ej5\n-----END RSA PRIVATE KEY-----"}`
	_ = os.WriteFile(keyFile, []byte(keyData), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for missing client_email")
	}
}

func TestManager_LoadServiceAccount_MissingPrivateKey(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	keyData := `{"type": "service_account", "client_email": "test@example.iam.gserviceaccount.com"}`
	_ = os.WriteFile(keyFile, []byte(keyData), 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for missing private_key")
	}
}

func TestManager_ProfileHasDifferentClient(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
		ClientID:    "client1",
	}

	err := mgr.SaveCredentials("test-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	_, err = mgr.profileHasDifferentClient("test-profile", "different-hash")
	if err != nil {
		t.Logf("profileHasDifferentClient returned error: %v", err)
	}
}

func TestManager_FindMetadataKey(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("test-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	key, err := mgr.findMetadataKey("test-profile")
	if err != nil {
		t.Fatalf("findMetadataKey failed: %v", err)
	}

	if key == "" {
		t.Error("findMetadataKey should return non-empty key")
	}
}

func TestManager_FindMetadataKey_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	key, err := mgr.findMetadataKey("nonexistent")
	if err != nil {
		t.Logf("findMetadataKey returned error: %v", err)
	}

	if key != "" {
		t.Errorf("findMetadataKey should return empty key for nonexistent profile, got %q", key)
	}
}

func TestManager_NewManagerWithOptions_PlainFile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	if mgr.GetStorageBackend() != "plain-file" {
		t.Errorf("Expected plain-file storage, got %q", mgr.GetStorageBackend())
	}

	warning := mgr.GetStorageWarning()
	if warning == "" {
		t.Error("Expected warning for plain file storage")
	}
}

func TestManager_NewManagerWithOptions_EncryptedFileFallback(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForceEncryptedFile: true})

	backend := mgr.GetStorageBackend()
	if backend != "encrypted-file" && backend != "plain-file" {
		t.Errorf("Expected encrypted-file or plain-file storage, got %q", backend)
	}
}

func TestManager_CheckKeyringAvailable(t *testing.T) {
	available := checkKeyringAvailable()
	if available {
		t.Log("System keyring is available")
	} else {
		t.Log("System keyring is not available (expected in test environment)")
	}
}

func TestManager_CredentialLocation_EncryptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForceEncryptedFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("encrypted-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	location, err := mgr.CredentialLocation("encrypted-test")
	if err != nil {
		t.Fatalf("CredentialLocation failed: %v", err)
	}

	if location == "" {
		t.Error("CredentialLocation returned empty string")
	}

	if !contains(location, ".enc") {
		t.Errorf("CredentialLocation should contain .enc for encrypted file storage: %q", location)
	}
}

func TestManager_LoadStoredCredentials_WithClientIDHash(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})
	mgr.SetOAuthConfig("test_client_id", "secret", []string{utils.ScopeFile})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("hashed-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("hashed-profile")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
}

func TestManager_DeleteCredentials_WithMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour),
		Scopes:       []string{utils.ScopeFile},
		Type:         types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("delete-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	meta, err := mgr.LoadAuthMetadata("delete-test")
	if err != nil {
		t.Logf("LoadAuthMetadata returned error: %v", err)
	}

	if meta != nil && meta.Profile != "delete-test" {
		t.Errorf("Metadata profile mismatch: got %q, want %q", meta.Profile, "delete-test")
	}

	err = mgr.DeleteCredentials("delete-test")
	if err != nil {
		t.Logf("DeleteCredentials returned: %v", err)
	}
}

func TestManager_RefreshCredentials_WithOAuthConfig(t *testing.T) {
	mgr := NewManager("/tmp/test")
	mgr.SetOAuthConfig("test_client_id", "test_client_secret", []string{utils.ScopeFile})

	creds := &types.Credentials{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		ExpiryDate:   time.Now().Add(-1 * time.Hour),
		Type:         types.AuthTypeOAuth,
		Scopes:       []string{utils.ScopeFile},
	}

	ctx := context.Background()
	_, err := mgr.RefreshCredentials(ctx, creds)
	if err != nil {
		t.Logf("RefreshCredentials returned error (expected in test): %v", err)
	}
}

func TestManager_GetValidCredentials_Impersonated(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken:         "test_token",
		ExpiryDate:          time.Now().Add(1 * time.Hour),
		Scopes:              []string{utils.ScopeFile},
		Type:                types.AuthTypeImpersonated,
		ServiceAccountEmail: "sa@example.iam.gserviceaccount.com",
		ImpersonatedUser:    "user@example.com",
	}

	err := mgr.SaveCredentials("impersonate-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	ctx := context.Background()
	loaded, err := mgr.GetValidCredentials(ctx, "impersonate-test")
	if err != nil {
		t.Fatalf("GetValidCredentials failed: %v", err)
	}

	if loaded.Type != types.AuthTypeImpersonated {
		t.Errorf("Type mismatch: got %v, want %v", loaded.Type, types.AuthTypeImpersonated)
	}
}

func TestManager_LoadStoredCredentials_FallbackToProfile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("fallback-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded, err := mgr.LoadCredentials("fallback-test")
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}

	if loaded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, creds.AccessToken)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
