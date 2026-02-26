package diff

import (
	"testing"

	"github.com/dl-alexandre/gdrv/internal/sync/index"
	"github.com/dl-alexandre/gdrv/internal/sync/scanner"
)

func TestConflictKinds(t *testing.T) {
	tests := []struct {
		kind ConflictKind
		want string
	}{
		{ConflictBothModified, "both_modified"},
		{ConflictLocalDeletedRemoteModified, "local_deleted_remote_modified"},
		{ConflictRemoteDeletedLocalModified, "remote_deleted_local_modified"},
		{ConflictTypeMismatch, "type_mismatch"},
	}

	for _, tc := range tests {
		t.Run(string(tc.kind), func(t *testing.T) {
			if string(tc.kind) != tc.want {
				t.Errorf("ConflictKind = %q, want %q", tc.kind, tc.want)
			}
		})
	}
}

func TestModes(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModePush, "push"},
		{ModePull, "pull"},
		{ModeBidirectional, "bidirectional"},
	}

	for _, tc := range tests {
		t.Run(string(tc.mode), func(t *testing.T) {
			if string(tc.mode) != tc.want {
				t.Errorf("Mode = %q, want %q", tc.mode, tc.want)
			}
		})
	}
}

func TestActionStruct(t *testing.T) {
	localEntry := &scanner.LocalEntry{
		RelativePath: "file.txt",
		AbsPath:      "/tmp/file.txt",
		IsDir:        false,
		Size:         100,
		ModTime:      1234567890,
		Hash:         "abc123",
	}

	remoteEntry := &scanner.RemoteEntry{
		RelativePath: "file.txt",
		ID:           "drive123",
		ParentID:     "parent456",
		IsDir:        false,
		Size:         100,
		ModifiedTime: "2024-01-15T10:00:00Z",
		MD5Checksum:  "abc123",
		MimeType:     "text/plain",
	}

	prevEntry := &index.SyncEntry{
		ConfigID:     "config1",
		RelativePath: "file.txt",
		DriveFileID:  "drive123",
		IsDir:        false,
		LocalMTime:   1234567890,
		LocalSize:    100,
		ContentHash:  "abc123",
		RemoteMTime:  "2024-01-15T10:00:00Z",
		RemoteSize:   100,
		RemoteMD5:    "abc123",
	}

	action := Action{
		Type:     ActionUpload,
		Path:     "file.txt",
		FromPath: "",
		ToPath:   "",
		Local:    localEntry,
		Remote:   remoteEntry,
		Prev:     prevEntry,
		Name:     "uploaded.txt",
	}

	if action.Type != ActionUpload {
		t.Errorf("Action.Type = %v, want %v", action.Type, ActionUpload)
	}
	if action.Path != "file.txt" {
		t.Errorf("Action.Path = %q, want %q", action.Path, "file.txt")
	}
	if action.Local == nil {
		t.Error("Action.Local should not be nil")
	}
	if action.Remote == nil {
		t.Error("Action.Remote should not be nil")
	}
	if action.Prev == nil {
		t.Error("Action.Prev should not be nil")
	}
}

func TestConflictStruct(t *testing.T) {
	localEntry := &scanner.LocalEntry{
		RelativePath: "file.txt",
		AbsPath:      "/tmp/file.txt",
		IsDir:        false,
		Size:         100,
		ModTime:      1234567890,
		Hash:         "abc123",
	}

	remoteEntry := &scanner.RemoteEntry{
		RelativePath: "file.txt",
		ID:           "drive123",
		ParentID:     "parent456",
		IsDir:        false,
		Size:         100,
		ModifiedTime: "2024-01-15T10:00:00Z",
		MD5Checksum:  "abc123",
		MimeType:     "text/plain",
	}

	prevEntry := &index.SyncEntry{
		ConfigID:     "config1",
		RelativePath: "file.txt",
		DriveFileID:  "drive123",
		IsDir:        false,
		LocalMTime:   1234567890,
		LocalSize:    100,
		ContentHash:  "abc123",
		RemoteMTime:  "2024-01-15T10:00:00Z",
		RemoteSize:   100,
		RemoteMD5:    "abc123",
	}

	conflict := Conflict{
		Path:   "file.txt",
		Kind:   ConflictBothModified,
		Local:  localEntry,
		Remote: remoteEntry,
		Prev:   prevEntry,
	}

	if conflict.Path != "file.txt" {
		t.Errorf("Conflict.Path = %q, want %q", conflict.Path, "file.txt")
	}
	if conflict.Kind != ConflictBothModified {
		t.Errorf("Conflict.Kind = %v, want %v", conflict.Kind, ConflictBothModified)
	}
	if conflict.Local == nil {
		t.Error("Conflict.Local should not be nil")
	}
	if conflict.Remote == nil {
		t.Error("Conflict.Remote should not be nil")
	}
	if conflict.Prev == nil {
		t.Error("Conflict.Prev should not be nil")
	}
}
