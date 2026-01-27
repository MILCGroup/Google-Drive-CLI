package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataFilePath(t *testing.T) {
	dir := t.TempDir()
	path := metadataFilePath(dir, "key123")
	expected := filepath.Join(dir, "credentials", "key123"+metadataSuffix)
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestWriteReadMetadata(t *testing.T) {
	dir := t.TempDir()
	meta := &AuthMetadata{
		Profile:        "default",
		ClientIDHash:   "hash",
		ClientIDLast4:  "1234",
		Scopes:         []string{"a", "b"},
		ExpiryDate:     "2030-01-01T00:00:00Z",
		RefreshToken:   true,
		CredentialType: "user",
		StorageBackend: "file",
		UpdatedAt:      "2026-01-01T00:00:00Z",
	}
	if err := writeMetadata(dir, "key123", meta); err != nil {
		t.Fatalf("writeMetadata error: %v", err)
	}
	path := metadataFilePath(dir, "key123")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected metadata file, got error: %v", err)
	}
	read, err := readMetadata(path)
	if err != nil {
		t.Fatalf("readMetadata error: %v", err)
	}
	if read.Profile != meta.Profile || read.ClientIDHash != meta.ClientIDHash || read.ClientIDLast4 != meta.ClientIDLast4 {
		t.Fatalf("read metadata mismatch: %+v", read)
	}
	if read.UpdatedAt != meta.UpdatedAt || read.StorageBackend != meta.StorageBackend || read.CredentialType != meta.CredentialType {
		t.Fatalf("read metadata mismatch: %+v", read)
	}
	if len(read.Scopes) != len(meta.Scopes) {
		t.Fatalf("expected %d scopes, got %d", len(meta.Scopes), len(read.Scopes))
	}
}

func TestMetadataTimestamp(t *testing.T) {
	ts := metadataTimestamp()
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %q", ts)
	}
}
