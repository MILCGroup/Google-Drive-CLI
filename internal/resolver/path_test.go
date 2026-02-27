package resolver

import (
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/types"
)

// TestPathNormalization tests that paths are properly normalized
func TestPathNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/foo/bar", "foo/bar"},
		{"foo/bar/", "foo/bar"},
		{"/foo/bar/", "foo/bar"},
		{"foo", "foo"},
		{"/foo", "foo"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestQueryStringEscaping tests that query strings are properly escaped
func TestQueryStringEscaping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"name with spaces", "name with spaces"},
		{"name'with'quotes", "name\\'with\\'quotes"},
		{"name\\with\\backslash", "name\\\\with\\\\backslash"},
		{"name'with\\both", "name\\'with\\\\both"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeQueryString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeQueryString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCacheKeyGeneration tests cache key generation with domain awareness
func TestCacheKeyGeneration(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		path     string
		opts     ResolveOptions
		expected string
	}{
		{
			name: "my-drive",
			path: "foo/bar",
			opts: ResolveOptions{
				DriveID:      "",
				SearchDomain: SearchDomainMyDrive,
			},
			expected: ":my-drive:foo/bar",
		},
		{
			name: "shared-drive",
			path: "foo/bar",
			opts: ResolveOptions{
				DriveID:      "drive123",
				SearchDomain: SearchDomainSharedDrive,
			},
			expected: "drive123:shared-drive:foo/bar",
		},
		{
			name: "shared-with-me",
			path: "document",
			opts: ResolveOptions{
				DriveID:      "",
				SearchDomain: SearchDomainSharedWithMe,
			},
			expected: ":shared-with-me:document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.makeCacheKey(tt.path, tt.opts)
			if result != tt.expected {
				t.Errorf("makeCacheKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDomainPriority tests that domain priorities are correct
func TestDomainPriority(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		domain   SearchDomain
		expected int
	}{
		{SearchDomainMyDrive, 1},
		{SearchDomainSharedDrive, 2},
		{SearchDomainSharedWithMe, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.domain), func(t *testing.T) {
			result := resolver.getDomainPriority(tt.domain)
			if result != tt.expected {
				t.Errorf("getDomainPriority(%v) = %d, want %d", tt.domain, result, tt.expected)
			}
		})
	}

	// Verify ordering: My Drive < Shared Drive < shared-with-me
	myDrivePriority := resolver.getDomainPriority(SearchDomainMyDrive)
	sharedDrivePriority := resolver.getDomainPriority(SearchDomainSharedDrive)
	sharedWithMePriority := resolver.getDomainPriority(SearchDomainSharedWithMe)

	if myDrivePriority >= sharedDrivePriority {
		t.Errorf("My Drive priority (%d) should be less than Shared Drive priority (%d)",
			myDrivePriority, sharedDrivePriority)
	}

	if sharedDrivePriority >= sharedWithMePriority {
		t.Errorf("Shared Drive priority (%d) should be less than shared-with-me priority (%d)",
			sharedDrivePriority, sharedWithMePriority)
	}
}

// TestDisambiguationOrdering tests stable ordering for disambiguation
func TestDisambiguationOrdering(t *testing.T) {
	resolver := &PathResolver{}

	// Create test files
	file1 := &types.DriveFile{
		ID:       "id1",
		Name:     "document",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	file2 := &types.DriveFile{
		ID:       "id2",
		Name:     "document",
		MimeType: "application/vnd.google-apps.shortcut",
		Parents:  []string{"root"},
	}

	file3 := &types.DriveFile{
		ID:       "id3",
		Name:     "document",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{},
	}

	tests := []struct {
		name     string
		matches  []*types.DriveFile
		domain   SearchDomain
		expected []string // Expected order of IDs
	}{
		{
			name:     "prefer non-shortcut",
			matches:  []*types.DriveFile{file2, file1},
			domain:   SearchDomainMyDrive,
			expected: []string{"id1", "id2"}, // file1 (non-shortcut) before file2 (shortcut)
		},
		{
			name:     "prefer my-drive over shared-with-me",
			matches:  []*types.DriveFile{file3, file1},
			domain:   SearchDomainMyDrive,
			expected: []string{"id1", "id3"}, // file1 (My Drive) before file3 (shared-with-me)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := resolver.sortMatchesWithDomainPreference(tt.matches, tt.domain)

			if len(sorted) != len(tt.expected) {
				t.Fatalf("Expected %d results, got %d", len(tt.expected), len(sorted))
			}

			for i, expectedID := range tt.expected {
				if sorted[i].ID != expectedID {
					t.Errorf("Position %d: expected ID %s, got %s", i, expectedID, sorted[i].ID)
				}
			}
		})
	}
}

// TestCacheTTL tests that cache entries expire correctly
func TestCacheTTL(t *testing.T) {
	cacheTTL := 100 * time.Millisecond
	resolver := NewPathResolver(nil, cacheTTL)

	// Add a cache entry
	key := "test:my-drive:foo/bar"
	fileID := "file123"
	resolver.updateCacheByKey(key, fileID)

	// Should be cached immediately
	if cached, ok := resolver.checkCacheByKey(key); !ok || cached != fileID {
		t.Errorf("Expected cached value %s, got ok=%v, cached=%s", fileID, ok, cached)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should no longer be cached
	if _, ok := resolver.checkCacheByKey(key); ok {
		t.Errorf("Expected cache entry to be expired, but it was still present")
	}
}

// TestCacheInvalidation tests cache invalidation
func TestCacheInvalidation(t *testing.T) {
	resolver := NewPathResolver(nil, 1*time.Hour)

	// Add a cache entry
	path := "foo/bar"
	driveID := ""
	fileID := "file123"
	resolver.updateCache(path, driveID, fileID)

	// Verify it's cached
	if cached, ok := resolver.checkCache(path, driveID); !ok || cached != fileID {
		t.Errorf("Expected cached value %s, got ok=%v, cached=%s", fileID, ok, cached)
	}

	// Invalidate
	resolver.InvalidateCache(path, driveID)

	// Should no longer be cached
	if _, ok := resolver.checkCache(path, driveID); ok {
		t.Errorf("Expected cache entry to be invalidated, but it was still present")
	}
}

// TestClearCache tests clearing all cache entries
func TestClearCache(t *testing.T) {
	resolver := NewPathResolver(nil, 1*time.Hour)

	// Add multiple cache entries
	entries := map[string]string{
		"test1:my-drive:foo":     "file1",
		"test2:shared-drive:bar": "file2",
		":shared-with-me:doc":    "file3",
	}

	for key, fileID := range entries {
		resolver.updateCacheByKey(key, fileID)
	}

	// Verify all are cached
	for key, expectedFileID := range entries {
		if cached, ok := resolver.checkCacheByKey(key); !ok || cached != expectedFileID {
			t.Errorf("Key %s: expected cached value %s, got ok=%v, cached=%s",
				key, expectedFileID, ok, cached)
		}
	}

	// Clear all
	resolver.ClearCache()

	// Verify all are gone
	for key := range entries {
		if _, ok := resolver.checkCacheByKey(key); ok {
			t.Errorf("Expected cache entry %s to be cleared, but it was still present", key)
		}
	}
}

// TestSearchDomainDetermination tests automatic search domain determination
func TestSearchDomainDetermination(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		opts     ResolveOptions
		expected SearchDomain
	}{
		{
			name: "my-drive-default",
			opts: ResolveOptions{
				DriveID:             "",
				IncludeSharedWithMe: false,
			},
			expected: SearchDomainMyDrive,
		},
		{
			name: "shared-drive-with-id",
			opts: ResolveOptions{
				DriveID:             "drive123",
				IncludeSharedWithMe: false,
			},
			expected: SearchDomainSharedDrive,
		},
		{
			name: "all-drives-with-shared-with-me",
			opts: ResolveOptions{
				DriveID:             "",
				IncludeSharedWithMe: true,
			},
			expected: SearchDomainAllDrives,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.determineSearchDomain(tt.opts)
			if result != tt.expected {
				t.Errorf("determineSearchDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFileDomainInference tests inferring which domain a file belongs to
func TestFileDomainInference(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		file     *types.DriveFile
		expected SearchDomain
	}{
		{
			name: "file-with-parents",
			file: &types.DriveFile{
				ID:      "file1",
				Parents: []string{"parent1"},
			},
			expected: SearchDomainMyDrive,
		},
		{
			name: "file-without-parents",
			file: &types.DriveFile{
				ID:      "file2",
				Parents: []string{},
			},
			expected: SearchDomainSharedWithMe,
		},
		{
			name: "file-with-nil-parents",
			file: &types.DriveFile{
				ID:      "file3",
				Parents: nil,
			},
			expected: SearchDomainSharedWithMe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.inferFileDomain(tt.file)
			if result != tt.expected {
				t.Errorf("inferFileDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestInvalidateCacheByKey tests cache invalidation by key
func TestInvalidateCacheByKey(t *testing.T) {
	resolver := NewPathResolver(nil, 1*time.Hour)

	// Add a cache entry
	key := "test:my-drive:foo/bar"
	fileID := "file123"
	resolver.updateCacheByKey(key, fileID)

	// Verify it's cached
	if cached, ok := resolver.checkCacheByKey(key); !ok || cached != fileID {
		t.Errorf("Expected cached value %s, got ok=%v, cached=%s", fileID, ok, cached)
	}

	// Invalidate by key
	resolver.InvalidateCacheByKey(key)

	// Should no longer be cached
	if _, ok := resolver.checkCacheByKey(key); ok {
		t.Errorf("Expected cache entry to be invalidated, but it was still present")
	}
}

// TestGetDomainPriorityEdgeCases tests domain priority for edge cases
func TestGetDomainPriorityEdgeCases(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		domain   SearchDomain
		expected int
	}{
		{name: "my-drive", domain: SearchDomainMyDrive, expected: 1},
		{name: "shared-drive", domain: SearchDomainSharedDrive, expected: 2},
		{name: "shared-with-me", domain: SearchDomainSharedWithMe, expected: 3},
		{name: "all-drives", domain: SearchDomainAllDrives, expected: 4},
		{name: "domain", domain: SearchDomainDomain, expected: 4},
		{name: "unknown", domain: SearchDomain("unknown"), expected: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.getDomainPriority(tt.domain)
			if result != tt.expected {
				t.Errorf("getDomainPriority(%v) = %d, want %d", tt.domain, result, tt.expected)
			}
		})
	}
}

// TestShouldSwapWithDomainShortcutPreference tests shortcut preference in disambiguation
func TestShouldSwapWithDomainShortcutPreference(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		fileA    *types.DriveFile
		fileB    *types.DriveFile
		domain   SearchDomain
		expected bool // true if should swap (A before B)
	}{
		{
			name: "prefer non-shortcut",
			fileA: &types.DriveFile{
				ID:       "id1",
				Name:     "doc",
				MimeType: "application/vnd.google-apps.shortcut",
				Parents:  []string{"root"},
			},
			fileB: &types.DriveFile{
				ID:       "id2",
				Name:     "doc",
				MimeType: "application/vnd.google-apps.document",
				Parents:  []string{"root"},
			},
			domain:   SearchDomainMyDrive,
			expected: true, // A is shortcut, should swap to put B first
		},
		{
			name: "both non-shortcuts, prefer by name",
			fileA: &types.DriveFile{
				ID:       "id1",
				Name:     "zebra",
				MimeType: "application/vnd.google-apps.document",
				Parents:  []string{"root"},
			},
			fileB: &types.DriveFile{
				ID:       "id2",
				Name:     "apple",
				MimeType: "application/vnd.google-apps.document",
				Parents:  []string{"root"},
			},
			domain:   SearchDomainMyDrive,
			expected: true, // zebra > apple, should swap
		},
		{
			name: "same name and type, prefer by ID",
			fileA: &types.DriveFile{
				ID:       "zzz",
				Name:     "doc",
				MimeType: "application/vnd.google-apps.document",
				Parents:  []string{"root"},
			},
			fileB: &types.DriveFile{
				ID:       "aaa",
				Name:     "doc",
				MimeType: "application/vnd.google-apps.document",
				Parents:  []string{"root"},
			},
			domain:   SearchDomainMyDrive,
			expected: true, // zzz > aaa, should swap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.shouldSwapWithDomain(tt.fileA, tt.fileB, tt.domain)
			if result != tt.expected {
				t.Errorf("shouldSwapWithDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}
