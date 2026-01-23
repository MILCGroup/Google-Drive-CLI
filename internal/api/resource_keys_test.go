package api

import (
	"testing"
)

func TestResourceKeyManager_AddAndGet(t *testing.T) {
	mgr := NewResourceKeyManager()

	mgr.AddKey("file1", "key1", "url")
	mgr.AddKey("file2", "key2", "api")

	key, ok := mgr.GetKey("file1")
	if !ok || key != "key1" {
		t.Errorf("GetKey(file1) = %s, %v; want key1, true", key, ok)
	}

	key, ok = mgr.GetKey("file2")
	if !ok || key != "key2" {
		t.Errorf("GetKey(file2) = %s, %v; want key2, true", key, ok)
	}

	_, ok = mgr.GetKey("nonexistent")
	if ok {
		t.Error("GetKey(nonexistent) should return false")
	}
}

func TestResourceKeyManager_BuildHeader(t *testing.T) {
	mgr := NewResourceKeyManager()

	mgr.AddKey("file1", "key1", "url")
	mgr.AddKey("file2", "key2", "api")

	// Empty list
	header := mgr.BuildHeader([]string{})
	if header != "" {
		t.Errorf("BuildHeader([]) = %s, want empty", header)
	}

	// Single file
	header = mgr.BuildHeader([]string{"file1"})
	if header != "file1/key1" {
		t.Errorf("BuildHeader([file1]) = %s, want file1/key1", header)
	}

	// Multiple files
	header = mgr.BuildHeader([]string{"file1", "file2"})
	if header != "file1/key1,file2/key2" {
		t.Errorf("BuildHeader([file1,file2]) = %s, want file1/key1,file2/key2", header)
	}

	// File without key
	header = mgr.BuildHeader([]string{"file1", "unknown"})
	if header != "file1/key1" {
		t.Errorf("BuildHeader([file1,unknown]) = %s, want file1/key1", header)
	}
}

func TestResourceKeyManager_ParseFromURL(t *testing.T) {
	mgr := NewResourceKeyManager()

	tests := []struct {
		url        string
		wantFileID string
		wantKey    string
		wantOK     bool
	}{
		{
			"https://drive.google.com/file/d/ABC123/view?resourcekey=KEY456",
			"ABC123", "KEY456", true,
		},
		{
			"https://drive.google.com/open?id=ABC123&resourcekey=KEY456",
			"ABC123", "KEY456", true,
		},
		{
			"https://drive.google.com/file/d/ABC123/view",
			"ABC123", "", true,
		},
		{
			"https://example.com/other",
			"", "", false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			fileID, key, ok := mgr.ParseFromURL(tt.url)
			if ok != tt.wantOK {
				t.Errorf("ParseFromURL ok = %v, want %v", ok, tt.wantOK)
			}
			if fileID != tt.wantFileID {
				t.Errorf("ParseFromURL fileID = %s, want %s", fileID, tt.wantFileID)
			}
			if key != tt.wantKey {
				t.Errorf("ParseFromURL key = %s, want %s", key, tt.wantKey)
			}
		})
	}
}

func TestResourceKeyManager_Invalidate(t *testing.T) {
	mgr := NewResourceKeyManager()

	mgr.AddKey("file1", "key1", "url")
	mgr.Invalidate("file1")

	_, ok := mgr.GetKey("file1")
	if ok {
		t.Error("GetKey should return false after Invalidate")
	}
}

func TestResourceKeyManager_Clear(t *testing.T) {
	mgr := NewResourceKeyManager()

	mgr.AddKey("file1", "key1", "url")
	mgr.AddKey("file2", "key2", "api")
	mgr.Clear()

	_, ok1 := mgr.GetKey("file1")
	_, ok2 := mgr.GetKey("file2")
	if ok1 || ok2 {
		t.Error("Clear should remove all keys")
	}
}
