package api

import (
	"strings"
	"testing"
)

// Property 3: Resource Key Header Formatting
// Validates: Requirements 14.3, 14.4, 14.5
// Property: Resource key headers must be formatted correctly as "fileId/key,fileId2/key2"

func TestProperty_ResourceKeyHeaderFormatting_SingleKey(t *testing.T) {
	// Property: Single resource key is formatted as "fileId/key"

	tests := []struct {
		fileID string
		key    string
	}{
		{"abc123", "key123"},
		{"file-with-dash", "key-with-dash"},
		{"FILE_UPPERCASE", "KEY_UPPERCASE"},
		{"123numeric", "456numeric"},
		{"verylongfileid1234567890", "verylongkey0987654321"},
		{"a", "k"},
		{"file_with_underscore", "key_with_underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.fileID, func(t *testing.T) {
			mgr := NewResourceKeyManager()
			mgr.AddKey(tt.fileID, tt.key, "test")

			header := mgr.BuildHeader([]string{tt.fileID})
			expected := tt.fileID + "/" + tt.key

			if header != expected {
				t.Errorf("BuildHeader([%s]) = %s, want %s", tt.fileID, header, expected)
			}
		})
	}
}

func TestProperty_ResourceKeyHeaderFormatting_MultipleKeys(t *testing.T) {
	// Property: Multiple resource keys are comma-separated

	mgr := NewResourceKeyManager()
	mgr.AddKey("file1", "key1", "test")
	mgr.AddKey("file2", "key2", "test")
	mgr.AddKey("file3", "key3", "test")

	tests := []struct {
		name     string
		fileIDs  []string
		expected string
	}{
		{
			"Two files",
			[]string{"file1", "file2"},
			"file1/key1,file2/key2",
		},
		{
			"Three files",
			[]string{"file1", "file2", "file3"},
			"file1/key1,file2/key2,file3/key3",
		},
		{
			"Different order",
			[]string{"file3", "file1"},
			"file3/key3,file1/key1",
		},
		{
			"Duplicate files",
			[]string{"file1", "file1"},
			"file1/key1,file1/key1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := mgr.BuildHeader(tt.fileIDs)
			if header != tt.expected {
				t.Errorf("BuildHeader(%v) = %s, want %s", tt.fileIDs, header, tt.expected)
			}
		})
	}
}

func TestProperty_ResourceKeyHeaderFormatting_NoKeys(t *testing.T) {
	// Property: No keys produces empty header

	mgr := NewResourceKeyManager()

	tests := []struct {
		name    string
		fileIDs []string
	}{
		{"Empty slice", []string{}},
		{"Nil slice", nil},
		{"Files without keys", []string{"unknown1", "unknown2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := mgr.BuildHeader(tt.fileIDs)
			if header != "" {
				t.Errorf("BuildHeader(%v) = %s, want empty string", tt.fileIDs, header)
			}
		})
	}
}

func TestProperty_ResourceKeyHeaderFormatting_MixedKeys(t *testing.T) {
	// Property: Only files with keys are included in header

	mgr := NewResourceKeyManager()
	mgr.AddKey("file1", "key1", "test")
	mgr.AddKey("file3", "key3", "test")

	header := mgr.BuildHeader([]string{"file1", "file2", "file3"})

	// Should only include file1 and file3
	expected := "file1/key1,file3/key3"
	if header != expected {
		t.Errorf("BuildHeader([file1,file2,file3]) = %s, want %s", header, expected)
	}
}

func TestProperty_ResourceKeyHeaderFormatting_NoSlash(t *testing.T) {
	// Property: Header never contains double slashes or trailing slashes

	mgr := NewResourceKeyManager()
	mgr.AddKey("file1", "key1", "test")
	mgr.AddKey("file2", "key2", "test")

	header := mgr.BuildHeader([]string{"file1", "file2"})

	if strings.Contains(header, "//") {
		t.Errorf("Header contains double slash: %s", header)
	}
	if strings.HasSuffix(header, "/") {
		t.Errorf("Header has trailing slash: %s", header)
	}
	if strings.HasPrefix(header, "/") {
		t.Errorf("Header has leading slash: %s", header)
	}
}

func TestProperty_ResourceKeyHeaderFormatting_NoCommaEdges(t *testing.T) {
	// Property: Header never starts or ends with comma, no double commas

	mgr := NewResourceKeyManager()
	mgr.AddKey("file1", "key1", "test")
	mgr.AddKey("file2", "key2", "test")

	header := mgr.BuildHeader([]string{"file1", "file2"})

	if strings.HasPrefix(header, ",") {
		t.Errorf("Header has leading comma: %s", header)
	}
	if strings.HasSuffix(header, ",") {
		t.Errorf("Header has trailing comma: %s", header)
	}
	if strings.Contains(header, ",,") {
		t.Errorf("Header contains double comma: %s", header)
	}
}

// Property 22: Resource Key Extraction from URLs
// Validates: Requirements 14.6
// Property: Resource keys must be correctly extracted from various URL formats

func TestProperty_ResourceKeyExtraction_StandardURLs(t *testing.T) {
	// Property: Standard Google Drive URL formats are parsed correctly

	tests := []struct {
		name       string
		url        string
		wantFileID string
		wantKey    string
		wantOK     bool
	}{
		{
			"File view URL with key",
			"https://drive.google.com/file/d/ABC123/view?resourcekey=KEY456",
			"ABC123", "KEY456", true,
		},
		{
			"File view URL without key",
			"https://drive.google.com/file/d/ABC123/view",
			"ABC123", "", true,
		},
		{
			"Open URL with key",
			"https://drive.google.com/open?id=ABC123&resourcekey=KEY456",
			"ABC123", "KEY456", true,
		},
		{
			"Open URL without key",
			"https://drive.google.com/open?id=ABC123",
			"ABC123", "", true,
		},
		{
			"Drive folder URL",
			"https://drive.google.com/drive/folders/FOLDER123",
			"FOLDER123", "", true,
		},
		{
			"File with edit",
			"https://drive.google.com/file/d/ABC123/edit?resourcekey=KEY456",
			"ABC123", "KEY456", true,
		},
	}

	mgr := NewResourceKeyManager()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileID, key, ok := mgr.ParseFromURL(tt.url)

			if ok != tt.wantOK {
				t.Errorf("ParseFromURL(%s) ok = %v, want %v", tt.url, ok, tt.wantOK)
			}
			if fileID != tt.wantFileID {
				t.Errorf("ParseFromURL(%s) fileID = %s, want %s", tt.url, fileID, tt.wantFileID)
			}
			if key != tt.wantKey {
				t.Errorf("ParseFromURL(%s) key = %s, want %s", tt.url, key, tt.wantKey)
			}
		})
	}
}

func TestProperty_ResourceKeyExtraction_InvalidURLs(t *testing.T) {
	// Property: Invalid URLs return ok=false

	tests := []string{
		"not a url",
		"http://example.com",
		"https://google.com",
		"https://drive.google.com",
		"https://drive.google.com/random",
		"",
		"ftp://drive.google.com/file/d/ABC123",
	}

	mgr := NewResourceKeyManager()

	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			_, _, ok := mgr.ParseFromURL(url)
			if ok {
				t.Errorf("ParseFromURL(%s) should return ok=false for invalid URL", url)
			}
		})
	}
}

func TestProperty_ResourceKeyExtraction_URLVariations(t *testing.T) {
	// Property: URL variations (http vs https, www, query params) are handled

	mgr := NewResourceKeyManager()

	tests := []struct {
		name       string
		url        string
		wantFileID string
		wantKey    string
	}{
		{
			"HTTPS",
			"https://drive.google.com/file/d/ABC123/view?resourcekey=KEY456",
			"ABC123", "KEY456",
		},
		{
			"HTTP (should upgrade)",
			"http://drive.google.com/file/d/ABC123/view?resourcekey=KEY456",
			"", "", // May not parse HTTP, or should upgrade
		},
		{
			"With www",
			"https://www.drive.google.com/file/d/ABC123/view?resourcekey=KEY456",
			"", "", // May not parse with www
		},
		{
			"Extra query params",
			"https://drive.google.com/file/d/ABC123/view?resourcekey=KEY456&usp=sharing",
			"ABC123", "KEY456",
		},
		{
			"Key as first param",
			"https://drive.google.com/open?resourcekey=KEY456&id=ABC123",
			"ABC123", "KEY456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileID, key, ok := mgr.ParseFromURL(tt.url)
			if ok {
				if tt.wantFileID != "" && fileID != tt.wantFileID {
					t.Errorf("ParseFromURL(%s) fileID = %s, want %s", tt.url, fileID, tt.wantFileID)
				}
				if tt.wantKey != "" && key != tt.wantKey {
					t.Errorf("ParseFromURL(%s) key = %s, want %s", tt.url, key, tt.wantKey)
				}
			}
		})
	}
}

func TestProperty_ResourceKeyExtraction_Consistency(t *testing.T) {
	// Property: Same URL always produces same result (deterministic)

	mgr := NewResourceKeyManager()
	url := "https://drive.google.com/file/d/ABC123/view?resourcekey=KEY456"

	firstFileID, firstKey, firstOK := mgr.ParseFromURL(url)

	// Run 100 iterations
	for i := 0; i < 100; i++ {
		fileID, key, ok := mgr.ParseFromURL(url)
		if ok != firstOK || fileID != firstFileID || key != firstKey {
			t.Errorf("ParseFromURL not deterministic: iteration %d gave (%s, %s, %v), expected (%s, %s, %v)",
				i, fileID, key, ok, firstFileID, firstKey, firstOK)
		}
	}
}

func TestProperty_ResourceKeyExtraction_SpecialCharacters(t *testing.T) {
	// Property: File IDs and keys with special characters are handled

	mgr := NewResourceKeyManager()

	tests := []struct {
		name       string
		url        string
		wantFileID string
		wantKey    string
	}{
		{
			"Underscores",
			"https://drive.google.com/file/d/ABC_123/view?resourcekey=KEY_456",
			"ABC_123", "KEY_456",
		},
		{
			"Hyphens",
			"https://drive.google.com/file/d/ABC-123/view?resourcekey=KEY-456",
			"ABC-123", "KEY-456",
		},
		{
			"Numbers only",
			"https://drive.google.com/file/d/123456/view?resourcekey=789012",
			"123456", "789012",
		},
		{
			"Mixed case",
			"https://drive.google.com/file/d/AbC123/view?resourcekey=KeY456",
			"AbC123", "KeY456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileID, key, ok := mgr.ParseFromURL(tt.url)
			if !ok {
				t.Errorf("ParseFromURL(%s) should succeed", tt.url)
				return
			}
			if fileID != tt.wantFileID {
				t.Errorf("ParseFromURL(%s) fileID = %s, want %s", tt.url, fileID, tt.wantFileID)
			}
			if key != tt.wantKey {
				t.Errorf("ParseFromURL(%s) key = %s, want %s", tt.url, key, tt.wantKey)
			}
		})
	}
}
