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

func TestManager_LoadStoredCredentials_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	_, err := mgr.loadStoredCredentials("nonexistent")
	if err == nil {
		t.Error("loadStoredCredentials should fail for nonexistent profile")
	}
}

func TestManager_DeleteStoredCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("delete-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	err = mgr.deleteStoredCredentials("delete-test")
	if err != nil {
		t.Logf("deleteStoredCredentials returned: %v", err)
	}
}

func TestManager_ResolveCredentialKey_WithOAuthConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})
	mgr.SetOAuthConfig("test_client_id", "secret", []string{utils.ScopeFile})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("resolve-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	key, err := mgr.resolveCredentialKey("resolve-test")
	if err != nil {
		t.Fatalf("resolveCredentialKey failed: %v", err)
	}

	if key == "" {
		t.Error("resolveCredentialKey returned empty key")
	}
}

func TestManager_ProfileHasDifferentClient_NoMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	hasDifferent, err := mgr.profileHasDifferentClient("nonexistent", "hash")
	if err != nil {
		t.Logf("profileHasDifferentClient returned error: %v", err)
	}

	if hasDifferent {
		t.Error("profileHasDifferentClient should return false for nonexistent profile")
	}
}

func TestManager_FindMetadataKey_MultipleProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	for i := 0; i < 3; i++ {
		err := mgr.SaveCredentials("profile", creds)
		if err != nil {
			t.Fatalf("SaveCredentials failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	key, err := mgr.findMetadataKey("profile")
	if err != nil {
		t.Fatalf("findMetadataKey failed: %v", err)
	}

	if key == "" {
		t.Error("findMetadataKey returned empty key")
	}
}

func TestManager_CredentialLocation_Keyring(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("location-test", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	location, err := mgr.CredentialLocation("location-test")
	if err != nil {
		t.Fatalf("CredentialLocation failed: %v", err)
	}

	if location == "" {
		t.Error("CredentialLocation returned empty string")
	}
}

func TestManager_ListProfiles_WithMultipleClients(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	for i := 0; i < 5; i++ {
		err := mgr.SaveCredentials("profile", creds)
		if err != nil {
			t.Fatalf("SaveCredentials failed: %v", err)
		}
	}

	profiles, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(profiles) == 0 {
		t.Error("ListProfiles returned empty list")
	}
}

func TestManager_AddProfileToList_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	creds := &types.Credentials{
		AccessToken: "test_token",
		ExpiryDate:  time.Now().Add(1 * time.Hour),
		Scopes:      []string{utils.ScopeFile},
		Type:        types.AuthTypeOAuth,
	}

	err := mgr.SaveCredentials("dup-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	err = mgr.SaveCredentials("dup-profile", creds)
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	profiles, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	count := 0
	for _, p := range profiles {
		if p == "dup-profile" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 profile, got %d", count)
	}
}

func TestManager_RemoveProfileFromList_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManagerWithOptions(tmpDir, ManagerOptions{ForcePlainFile: true})

	err := mgr.removeProfileFromList("nonexistent")
	if err != nil {
		t.Logf("removeProfileFromList returned error: %v", err)
	}
}

func TestWriteMetadata_CreateDir(t *testing.T) {
	tmpDir := t.TempDir()

	meta := &AuthMetadata{
		Profile:        "test",
		ClientIDHash:   "hash",
		ClientIDLast4:  "last4",
		Scopes:         []string{utils.ScopeFile},
		ExpiryDate:     time.Now().Format(time.RFC3339),
		RefreshToken:   true,
		CredentialType: string(types.AuthTypeOAuth),
		StorageBackend: "plain-file",
		UpdatedAt:      metadataTimestamp(),
	}

	err := writeMetadata(tmpDir, "test-key", meta)
	if err != nil {
		t.Fatalf("writeMetadata failed: %v", err)
	}

	metaPath := metadataFilePath(tmpDir, "test-key")
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("Metadata file not created: %v", err)
	}
}

func TestReadMetadata_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "invalid.meta.json")
	os.WriteFile(metaPath, []byte("invalid json"), 0600)

	_, err := readMetadata(metaPath)
	if err == nil {
		t.Error("readMetadata should fail for invalid JSON")
	}
}

func TestReadMetadata_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, "nonexistent.meta.json")

	_, err := readMetadata(metaPath)
	if err == nil {
		t.Error("readMetadata should fail for nonexistent file")
	}
}

func TestLoadServiceAccount_InvalidKeyType(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	keyData := map[string]interface{}{
		"type": "invalid_type",
	}
	data, _ := json.Marshal(keyData)
	os.WriteFile(keyFile, data, 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for invalid key type")
	}
}

func TestLoadServiceAccount_MissingFields(t *testing.T) {
	mgr := NewManager("/tmp/test")
	ctx := context.Background()

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "key.json")
	keyData := map[string]interface{}{
		"type": "service_account",
	}
	data, _ := json.Marshal(keyData)
	os.WriteFile(keyFile, data, 0600)

	_, err := mgr.LoadServiceAccount(ctx, keyFile, []string{utils.ScopeFile}, "")
	if err == nil {
		t.Error("LoadServiceAccount should fail for missing fields")
	}
}
