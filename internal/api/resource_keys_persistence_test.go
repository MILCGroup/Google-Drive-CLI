package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResourceKeyManager_SetCachePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	cachePath := filepath.Join(tmpDir, "resource_keys.json")

	// Test setting cache path (should not error)
	err = mgr.SetCachePath(cachePath)
	if err != nil {
		t.Errorf("SetCachePath failed: %v", err)
	}

	// Verify path was set
	if mgr.path != cachePath {
		t.Errorf("Expected path to be %s, got %s", cachePath, mgr.path)
	}
}

func TestResourceKeyManager_SetCachePath_WithExistingData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cachePath := filepath.Join(tmpDir, "resource_keys.json")

	// Create a cache file with existing data
	existingData := `{
		"file1": {"resourceKey": "key1", "timestamp": 1234567890, "source": "test"},
		"file2": {"resourceKey": "key2", "timestamp": 1234567891, "source": "test"}
	}`
	if err := os.WriteFile(cachePath, []byte(existingData), 0600); err != nil {
		t.Fatal(err)
	}

	mgr := NewResourceKeyManager()
	err = mgr.SetCachePath(cachePath)
	if err != nil {
		t.Errorf("SetCachePath failed: %v", err)
	}

	// Verify keys were loaded
	key1, ok := mgr.GetKey("file1")
	if !ok || key1 != "key1" {
		t.Errorf("Expected key1 to be loaded, got %s, %v", key1, ok)
	}

	key2, ok := mgr.GetKey("file2")
	if !ok || key2 != "key2" {
		t.Errorf("Expected key2 to be loaded, got %s, %v", key2, ok)
	}
}

func TestResourceKeyManager_SetCachePath_NonExistentFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	cachePath := filepath.Join(tmpDir, "non_existent", "resource_keys.json")

	// Should not error when file doesn't exist
	err = mgr.SetCachePath(cachePath)
	if err != nil {
		t.Errorf("SetCachePath should not error for non-existent file: %v", err)
	}
}

func TestResourceKeyManager_UpdateFromAPIResponse(t *testing.T) {
	mgr := NewResourceKeyManager()

	// Test adding a key
	mgr.UpdateFromAPIResponse("file123", "apiKey123")

	key, ok := mgr.GetKey("file123")
	if !ok || key != "apiKey123" {
		t.Errorf("Expected key 'apiKey123' for file123, got %s, %v", key, ok)
	}

	// Test with empty key (should not add)
	mgr.UpdateFromAPIResponse("file456", "")

	_, ok = mgr.GetKey("file456")
	if ok {
		t.Error("Should not add empty resource key")
	}
}

func TestResourceKeyManager_Load(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a cache file
	cachePath := filepath.Join(tmpDir, "resource_keys.json")
	data := `{
		"file1": {"resourceKey": "key1", "timestamp": 1234567890, "source": "url"}
	}`
	if err := os.WriteFile(cachePath, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}

	mgr := NewResourceKeyManager()
	mgr.path = cachePath

	// Test loading
	err = mgr.load()
	if err != nil {
		t.Errorf("load() failed: %v", err)
	}

	key, ok := mgr.GetKey("file1")
	if !ok || key != "key1" {
		t.Errorf("Expected key1 after loading, got %s, %v", key, ok)
	}
}

func TestResourceKeyManager_Load_NoPath(t *testing.T) {
	mgr := NewResourceKeyManager()

	// Should return nil when no path is set
	err := mgr.load()
	if err != nil {
		t.Errorf("load() should return nil when no path set, got %v", err)
	}
}

func TestResourceKeyManager_Load_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cachePath := filepath.Join(tmpDir, "resource_keys.json")
	// Write invalid JSON
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	mgr := NewResourceKeyManager()
	mgr.path = cachePath

	// Should return error for invalid JSON
	err = mgr.load()
	if err == nil {
		t.Error("load() should return error for invalid JSON")
	}
}

func TestResourceKeyManager_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	mgr.path = filepath.Join(tmpDir, "subdir", "resource_keys.json")

	mgr.AddKey("file1", "key1", "url")

	// Verify file was created
	_, err = os.Stat(mgr.path)
	if err != nil {
		t.Errorf("save() did not create file: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(mgr.path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("save() wrote empty file")
	}

	// Check if key1 is in the saved data
	content := string(data)
	if !contains(content, "file1") || !contains(content, "key1") {
		t.Errorf("Saved data doesn't contain expected content: %s", content)
	}
}

func TestResourceKeyManager_Save_NoPath(t *testing.T) {
	mgr := NewResourceKeyManager()

	// Should return nil when no path is set
	err := mgr.save()
	if err != nil {
		t.Errorf("save() should return nil when no path set, got %v", err)
	}
}

func TestResourceKeyManager_AddKey_PersistsToDisk(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	_ = mgr.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	mgr.AddKey("persisted_file", "persisted_key", "api")

	// Create new manager and load from same path
	mgr2 := NewResourceKeyManager()
	_ = mgr2.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	key, ok := mgr2.GetKey("persisted_file")
	if !ok || key != "persisted_key" {
		t.Errorf("Expected persisted key to be loaded, got %s, %v", key, ok)
	}
}

func TestResourceKeyManager_Invalidate_Persists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	_ = mgr.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	mgr.AddKey("to_remove", "key123", "url")

	// Verify it was added
	_, ok := mgr.GetKey("to_remove")
	if !ok {
		t.Fatal("Key should exist before invalidation")
	}

	// Invalidate
	mgr.Invalidate("to_remove")

	// Verify it was removed from memory
	_, ok = mgr.GetKey("to_remove")
	if ok {
		t.Error("Key should be removed from memory after invalidate")
	}

	// Create new manager and verify it persists
	mgr2 := NewResourceKeyManager()
	_ = mgr2.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	_, ok = mgr2.GetKey("to_remove")
	if ok {
		t.Error("Invalidate should persist to disk")
	}
}

func TestResourceKeyManager_Clear_Persists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gdrv-resource-keys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr := NewResourceKeyManager()
	_ = mgr.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	mgr.AddKey("file1", "key1", "url")
	mgr.AddKey("file2", "key2", "api")

	// Clear all
	mgr.Clear()

	// Verify all cleared from memory by trying to get a key
	_, ok1 := mgr.GetKey("file1")
	_, ok2 := mgr.GetKey("file2")
	if ok1 || ok2 {
		t.Error("Expected all keys to be cleared from memory")
	}

	// Create new manager and verify it persists
	mgr2 := NewResourceKeyManager()
	_ = mgr2.SetCachePath(filepath.Join(tmpDir, "keys.json"))

	_, ok := mgr2.GetKey("file1")
	if ok {
		t.Error("Clear should persist to disk - file1 still exists")
	}

	_, ok = mgr2.GetKey("file2")
	if ok {
		t.Error("Clear should persist to disk - file2 still exists")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			return containsAt(s, substr, start+1)
		}
	}
	return true
}
