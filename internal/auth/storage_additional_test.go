package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyringStorage_Save_Load_Delete(t *testing.T) {
	storage := NewKeyringStorage("test-service")

	err := storage.Save("test-key", []byte("test-data"))
	if err != nil {
		t.Logf("Save returned error (may be expected if keyring unavailable): %v", err)
	}

	data, err := storage.Load("test-key")
	if err != nil {
		t.Logf("Load returned error (may be expected if keyring unavailable): %v", err)
	}

	if data != nil && string(data) != "test-data" {
		t.Errorf("Data mismatch: got %q, want %q", string(data), "test-data")
	}

	err = storage.Delete("test-key")
	if err != nil {
		t.Logf("Delete returned error (may be expected if keyring unavailable): %v", err)
	}
}

func TestKeyringStorage_Name(t *testing.T) {
	storage := NewKeyringStorage("test-service")
	if storage.Name() != "system-keyring" {
		t.Errorf("Expected 'system-keyring', got %q", storage.Name())
	}
}

func TestEncryptedFileStorage_Save_Load_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewEncryptedFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptedFileStorage failed: %v", err)
	}

	err = storage.Save("test-profile", []byte("test-data"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := storage.Load("test-profile")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if string(data) != "test-data" {
		t.Errorf("Data mismatch: got %q, want %q", string(data), "test-data")
	}

	err = storage.Delete("test-profile")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Load("test-profile")
	if err == nil {
		t.Error("Load should fail after deletion")
	}
}

func TestEncryptedFileStorage_Name(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewEncryptedFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptedFileStorage failed: %v", err)
	}

	if storage.Name() != "encrypted-file" {
		t.Errorf("Expected 'encrypted-file', got %q", storage.Name())
	}
}

func TestPlainFileStorage_Save_Load_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPlainFileStorage(tmpDir)

	err := storage.Save("test-profile", []byte("test-data"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, err := storage.Load("test-profile")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if string(data) != "test-data" {
		t.Errorf("Data mismatch: got %q, want %q", string(data), "test-data")
	}

	err = storage.Delete("test-profile")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Load("test-profile")
	if err == nil {
		t.Error("Load should fail after deletion")
	}
}

func TestPlainFileStorage_Name(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPlainFileStorage(tmpDir)

	if storage.Name() != "plain-file" {
		t.Errorf("Expected 'plain-file', got %q", storage.Name())
	}
}

func TestPlainFileStorage_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPlainFileStorage(tmpDir)

	_, err := storage.Load("nonexistent")
	if err == nil {
		t.Error("Load should fail for nonexistent profile")
	}
}

func TestEncryptedFileStorage_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewEncryptedFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptedFileStorage failed: %v", err)
	}

	_, err = storage.Load("nonexistent")
	if err == nil {
		t.Error("Load should fail for nonexistent profile")
	}
}

func TestEncryptedFileStorage_CorruptedData(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := NewEncryptedFileStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewEncryptedFileStorage failed: %v", err)
	}

	credDir := filepath.Join(tmpDir, "credentials")
	os.MkdirAll(credDir, 0700)

	credFile := filepath.Join(credDir, "corrupted.enc")
	os.WriteFile(credFile, []byte("corrupted-data"), 0600)

	_, err = storage.Load("corrupted")
	if err == nil {
		t.Error("Load should fail for corrupted data")
	}
}

func TestGetOrCreateEncryptionKey_Existing(t *testing.T) {
	tmpDir := t.TempDir()

	key1, err := getOrCreateEncryptionKey(tmpDir)
	if err != nil {
		t.Fatalf("getOrCreateEncryptionKey failed: %v", err)
	}

	key2, err := getOrCreateEncryptionKey(tmpDir)
	if err != nil {
		t.Fatalf("getOrCreateEncryptionKey failed: %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("Keys should be identical for same directory")
	}
}

func TestGetOrCreateEncryptionKey_Length(t *testing.T) {
	tmpDir := t.TempDir()

	key, err := getOrCreateEncryptionKey(tmpDir)
	if err != nil {
		t.Fatalf("getOrCreateEncryptionKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}
}
