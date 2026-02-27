package conflict

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/sync/diff"
	"github.com/milcgroup/gdrv/internal/sync/index"
	"github.com/milcgroup/gdrv/internal/sync/scanner"
)

func TestPolicyConstants(t *testing.T) {
	tests := []struct {
		policy Policy
		want   string
	}{
		{PolicyLocalWins, "local-wins"},
		{PolicyRemoteWins, "remote-wins"},
		{PolicyRenameBoth, "rename-both"},
	}

	for _, tc := range tests {
		t.Run(string(tc.policy), func(t *testing.T) {
			if string(tc.policy) != tc.want {
				t.Errorf("Policy = %q, want %q", tc.policy, tc.want)
			}
		})
	}
}

func TestResolveEmpty(t *testing.T) {
	t.Run("empty conflicts with local-wins", func(t *testing.T) {
		conflicts := []diff.Conflict{}
		actions, remaining := Resolve(conflicts, PolicyLocalWins)
		if len(actions) != 0 {
			t.Errorf("Expected 0 actions, got %d", len(actions))
		}
		if len(remaining) != 0 {
			t.Errorf("Expected 0 remaining, got %d", len(remaining))
		}
	})

	t.Run("empty conflicts with remote-wins", func(t *testing.T) {
		conflicts := []diff.Conflict{}
		actions, remaining := Resolve(conflicts, PolicyRemoteWins)
		if len(actions) != 0 {
			t.Errorf("Expected 0 actions, got %d", len(actions))
		}
		if len(remaining) != 0 {
			t.Errorf("Expected 0 remaining, got %d", len(remaining))
		}
	})
}

func TestResolveLocalWins(t *testing.T) {
	localEntry := &scanner.LocalEntry{
		RelativePath: "file.txt",
		AbsPath:      "/tmp/file.txt",
		Size:         100,
		ModTime:      1234567890,
		Hash:         "abc123",
	}

	remoteEntry := &scanner.RemoteEntry{
		RelativePath: "file.txt",
		ID:           "drive123",
		ParentID:     "parent456",
		Size:         200,
		ModifiedTime: "2024-01-15T10:00:00Z",
		MD5Checksum:  "def456",
		MimeType:     "text/plain",
	}

	prevEntry := &index.SyncEntry{
		ConfigID:     "config1",
		RelativePath: "file.txt",
		DriveFileID:  "drive123",
		ContentHash:  "oldhash",
		RemoteMD5:    "oldmd5",
	}

	tests := []struct {
		name       string
		conflict   diff.Conflict
		wantAction diff.ActionType
		wantPath   string
	}{
		{
			name: "both modified - should update",
			conflict: diff.Conflict{
				Path:   "file.txt",
				Kind:   diff.ConflictBothModified,
				Local:  localEntry,
				Remote: remoteEntry,
				Prev:   prevEntry,
			},
			wantAction: diff.ActionUpdate,
			wantPath:   "file.txt",
		},
		{
			name: "local deleted remote modified - should delete remote",
			conflict: diff.Conflict{
				Path:   "file.txt",
				Kind:   diff.ConflictLocalDeletedRemoteModified,
				Local:  nil,
				Remote: remoteEntry,
				Prev:   prevEntry,
			},
			wantAction: diff.ActionDeleteRemote,
			wantPath:   "file.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conflicts := []diff.Conflict{tc.conflict}
			actions, remaining := Resolve(conflicts, PolicyLocalWins)

			if len(remaining) != 0 {
				t.Errorf("Expected 0 remaining conflicts, got %d", len(remaining))
			}
			if len(actions) == 0 {
				t.Fatal("Expected at least 1 action, got 0")
			}
			if actions[0].Type != tc.wantAction {
				t.Errorf("Expected action type %v, got %v", tc.wantAction, actions[0].Type)
			}
			if actions[0].Path != tc.wantPath {
				t.Errorf("Expected action path %q, got %q", tc.wantPath, actions[0].Path)
			}
		})
	}
}

func TestResolveRemoteWins(t *testing.T) {
	localEntry := &scanner.LocalEntry{
		RelativePath: "file.txt",
		AbsPath:      "/tmp/file.txt",
		Size:         100,
		ModTime:      1234567890,
		Hash:         "abc123",
	}

	remoteEntry := &scanner.RemoteEntry{
		RelativePath: "file.txt",
		ID:           "drive123",
		ParentID:     "parent456",
		Size:         200,
		ModifiedTime: "2024-01-15T10:00:00Z",
		MD5Checksum:  "def456",
		MimeType:     "text/plain",
	}

	prevEntry := &index.SyncEntry{
		ConfigID:     "config1",
		RelativePath: "file.txt",
		DriveFileID:  "drive123",
		ContentHash:  "oldhash",
		RemoteMD5:    "oldmd5",
	}

	tests := []struct {
		name       string
		conflict   diff.Conflict
		wantAction diff.ActionType
		wantPath   string
	}{
		{
			name: "both modified - should download",
			conflict: diff.Conflict{
				Path:   "file.txt",
				Kind:   diff.ConflictBothModified,
				Local:  localEntry,
				Remote: remoteEntry,
				Prev:   prevEntry,
			},
			wantAction: diff.ActionDownload,
			wantPath:   "file.txt",
		},
		{
			name: "remote deleted local modified - should delete local",
			conflict: diff.Conflict{
				Path:   "file.txt",
				Kind:   diff.ConflictRemoteDeletedLocalModified,
				Local:  localEntry,
				Remote: nil,
				Prev:   prevEntry,
			},
			wantAction: diff.ActionDeleteLocal,
			wantPath:   "file.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conflicts := []diff.Conflict{tc.conflict}
			actions, remaining := Resolve(conflicts, PolicyRemoteWins)

			if len(remaining) != 0 {
				t.Errorf("Expected 0 remaining conflicts, got %d", len(remaining))
			}
			if len(actions) == 0 {
				t.Fatal("Expected at least 1 action, got 0")
			}
			if actions[0].Type != tc.wantAction {
				t.Errorf("Expected action type %v, got %v", tc.wantAction, actions[0].Type)
			}
		})
	}
}

func TestAddSuffix(t *testing.T) {
	tests := []struct {
		path   string
		suffix string
		want   string
	}{
		{"file.txt", ".local", "file.local.txt"},
		{"document.pdf", ".remote", "document.remote.pdf"},
		{"noext", ".bak", "noext.bak"},
		// Note: ".hidden" becomes ".tmp.hidden" because path.Ext(".hidden") = ".hidden"
		{".hidden", ".tmp", ".tmp.hidden"},
		{"deep/nested/file.go", ".local", "deep/nested/file.local.go"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := addSuffix(tc.path, tc.suffix)
			if got != tc.want {
				t.Errorf("addSuffix(%q, %q) = %q, want %q", tc.path, tc.suffix, got, tc.want)
			}
		})
	}
}
