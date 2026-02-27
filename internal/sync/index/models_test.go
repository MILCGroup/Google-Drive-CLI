package index

import (
	"testing"
)

func TestSyncConfigStruct(t *testing.T) {
	cfg := SyncConfig{
		ID:              "test-config",
		LocalRoot:       "/tmp/test",
		RemoteRootID:    "drive123",
		ExcludePatterns: []string{"*.tmp", "*.log"},
		ConflictPolicy:  "local-wins",
		Direction:       "bidirectional",
		LastSyncTime:    1704067200,
		LastChangeToken: "token123",
	}

	if cfg.ID != "test-config" {
		t.Errorf("ID = %q, want %q", cfg.ID, "test-config")
	}
	if cfg.LocalRoot != "/tmp/test" {
		t.Errorf("LocalRoot = %q, want %q", cfg.LocalRoot, "/tmp/test")
	}
	if cfg.RemoteRootID != "drive123" {
		t.Errorf("RemoteRootID = %q, want %q", cfg.RemoteRootID, "drive123")
	}
	if len(cfg.ExcludePatterns) != 2 {
		t.Errorf("len(ExcludePatterns) = %d, want 2", len(cfg.ExcludePatterns))
	}
	if cfg.ConflictPolicy != "local-wins" {
		t.Errorf("ConflictPolicy = %q, want %q", cfg.ConflictPolicy, "local-wins")
	}
	if cfg.Direction != "bidirectional" {
		t.Errorf("Direction = %q, want %q", cfg.Direction, "bidirectional")
	}
}

func TestSyncEntryStruct(t *testing.T) {
	entry := SyncEntry{
		ConfigID:       "config1",
		RelativePath:   "file.txt",
		DriveFileID:    "drive123",
		DriveParentID:  "parent456",
		IsDir:          false,
		LocalMTime:     1234567890,
		LocalSize:      100,
		ContentHash:    "abc123",
		RemoteMTime:    "2024-01-15T10:00:00Z",
		RemoteSize:     100,
		RemoteMD5:      "def456",
		RemoteMimeType: "text/plain",
		SyncState:      "synced",
		LastSync:       1234567890,
	}

	if entry.ConfigID != "config1" {
		t.Errorf("ConfigID = %q, want %q", entry.ConfigID, "config1")
	}
	if entry.RelativePath != "file.txt" {
		t.Errorf("RelativePath = %q, want %q", entry.RelativePath, "file.txt")
	}
	if entry.DriveFileID != "drive123" {
		t.Errorf("DriveFileID = %q, want %q", entry.DriveFileID, "drive123")
	}
	if entry.IsDir {
		t.Error("IsDir should be false for a file")
	}
	if entry.LocalMTime != 1234567890 {
		t.Errorf("LocalMTime = %d, want %d", entry.LocalMTime, 1234567890)
	}
	if entry.LocalSize != 100 {
		t.Errorf("LocalSize = %d, want %d", entry.LocalSize, 100)
	}
	if entry.ContentHash != "abc123" {
		t.Errorf("ContentHash = %q, want %q", entry.ContentHash, "abc123")
	}
}

func TestSyncEntryDirectory(t *testing.T) {
	entry := SyncEntry{
		ConfigID:     "config1",
		RelativePath: "folder/",
		DriveFileID:  "drive789",
		IsDir:        true,
		LocalMTime:   1234567890,
		RemoteMTime:  "2024-01-15T10:00:00Z",
		SyncState:    "synced",
	}

	if !entry.IsDir {
		t.Error("IsDir should be true for a directory")
	}
	if entry.RelativePath != "folder/" {
		t.Errorf("RelativePath = %q, want %q", entry.RelativePath, "folder/")
	}
}

func TestSyncConfigDefaultValues(t *testing.T) {
	// Test with empty values
	cfg := SyncConfig{
		ID:           "minimal",
		LocalRoot:    "/tmp",
		RemoteRootID: "drive1",
	}

	if cfg.ConflictPolicy != "" {
		t.Errorf("ConflictPolicy should be empty by default, got %q", cfg.ConflictPolicy)
	}
	if cfg.Direction != "" {
		t.Errorf("Direction should be empty by default, got %q", cfg.Direction)
	}
	if cfg.LastSyncTime != 0 {
		t.Errorf("LastSyncTime should be 0 by default, got %d", cfg.LastSyncTime)
	}
	if cfg.LastChangeToken != "" {
		t.Errorf("LastChangeToken should be empty by default, got %q", cfg.LastChangeToken)
	}
}
