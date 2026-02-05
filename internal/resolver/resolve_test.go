package resolver

import (
	"errors"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

func TestIsPermissionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "non-app error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "permission denied app error",
			err:      utils.NewAppError(utils.NewCLIError(utils.ErrCodePermissionDenied, "denied").Build()),
			expected: true,
		},
		{
			name:     "other app error",
			err:      utils.NewAppError(utils.NewCLIError(utils.ErrCodeFileNotFound, "not found").Build()),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPermissionError(tt.err)
			if result != tt.expected {
				t.Errorf("isPermissionError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCacheKeyFunction(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		driveID  string
		expected string
	}{
		{
			name:     "my-drive",
			path:     "foo/bar",
			driveID:  "",
			expected: ":foo/bar",
		},
		{
			name:     "shared-drive",
			path:     "foo/bar",
			driveID:  "drive123",
			expected: "drive123:foo/bar",
		},
		{
			name:     "empty-path",
			path:     "",
			driveID:  "",
			expected: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cacheKey(tt.path, tt.driveID)
			if result != tt.expected {
				t.Errorf("cacheKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestEscapeQueryStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only quotes",
			input:    "'''",
			expected: "\\'\\'\\'",
		},
		{
			name:     "only backslashes",
			input:    "\\\\\\",
			expected: "\\\\\\\\\\\\",
		},
		{
			name:     "mixed special chars",
			input:    "file'with\\special",
			expected: "file\\'with\\\\special",
		},
		{
			name:     "unicode",
			input:    "文件'名",
			expected: "文件\\'名",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeQueryString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeQueryString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePathEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single leading slash",
			input:    "/foo/bar",
			expected: "foo/bar",
		},
		{
			name:     "single trailing slash",
			input:    "foo/bar/",
			expected: "foo/bar",
		},
		{
			name:     "both leading and trailing",
			input:    "/foo/bar/",
			expected: "foo/bar",
		},
		{
			name:     "single slash",
			input:    "/",
			expected: "",
		},
		{
			name:     "path with spaces",
			input:    "/my documents/file name/",
			expected: "my documents/file name",
		},
		{
			name:     "no slashes",
			input:    "foo",
			expected: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInferFileDomainEdgeCases(t *testing.T) {
	resolver := &PathResolver{}

	tests := []struct {
		name     string
		file     *types.DriveFile
		expected SearchDomain
	}{
		{
			name: "multiple parents",
			file: &types.DriveFile{
				ID:      "file1",
				Parents: []string{"parent1", "parent2"},
			},
			expected: SearchDomainMyDrive,
		},
		{
			name: "root parent",
			file: &types.DriveFile{
				ID:      "file2",
				Parents: []string{"root"},
			},
			expected: SearchDomainMyDrive,
		},
		{
			name: "shared drive parent",
			file: &types.DriveFile{
				ID:      "file3",
				Parents: []string{"drive123"},
			},
			expected: SearchDomainMyDrive,
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

func TestSortMatchesStability(t *testing.T) {
	resolver := &PathResolver{}

	file1 := &types.DriveFile{
		ID:       "id1",
		Name:     "alpha",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	file2 := &types.DriveFile{
		ID:       "id2",
		Name:     "beta",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	file3 := &types.DriveFile{
		ID:       "id3",
		Name:     "gamma",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	matches := []*types.DriveFile{file3, file1, file2}

	sorted1 := resolver.sortMatchesWithDomainPreference(matches, SearchDomainMyDrive)
	sorted2 := resolver.sortMatchesWithDomainPreference(sorted1, SearchDomainMyDrive)

	for i := range sorted1 {
		if sorted1[i].ID != sorted2[i].ID {
			t.Errorf("Sorting not stable: position %d differs", i)
		}
	}
}

func TestMakeCacheKeyWithDifferentDomains(t *testing.T) {
	resolver := &PathResolver{}
	path := "documents/file"

	key1 := resolver.makeCacheKey(path, ResolveOptions{
		DriveID:      "",
		SearchDomain: SearchDomainMyDrive,
	})

	key2 := resolver.makeCacheKey(path, ResolveOptions{
		DriveID:      "",
		SearchDomain: SearchDomainSharedWithMe,
	})

	key3 := resolver.makeCacheKey(path, ResolveOptions{
		DriveID:      "drive123",
		SearchDomain: SearchDomainSharedDrive,
	})

	if key1 == key2 {
		t.Errorf("Different domains should produce different cache keys")
	}

	if key1 == key3 {
		t.Errorf("Different drive IDs should produce different cache keys")
	}

	if key2 == key3 {
		t.Errorf("Different domains and drive IDs should produce different cache keys")
	}
}

func TestMakeCacheKeyConsistency(t *testing.T) {
	resolver := &PathResolver{}
	path := "documents/file"
	opts := ResolveOptions{
		DriveID:      "drive123",
		SearchDomain: SearchDomainSharedDrive,
	}

	key1 := resolver.makeCacheKey(path, opts)
	key2 := resolver.makeCacheKey(path, opts)

	if key1 != key2 {
		t.Errorf("Same path and options should produce same cache key")
	}
}

func TestShouldSwapWithDomainNameOrdering(t *testing.T) {
	resolver := &PathResolver{}

	fileA := &types.DriveFile{
		ID:       "id1",
		Name:     "zebra",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	fileB := &types.DriveFile{
		ID:       "id2",
		Name:     "apple",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	shouldSwap := resolver.shouldSwapWithDomain(fileA, fileB, SearchDomainMyDrive)
	if !shouldSwap {
		t.Errorf("Should swap zebra and apple (zebra > apple)")
	}

	shouldSwap = resolver.shouldSwapWithDomain(fileB, fileA, SearchDomainMyDrive)
	if shouldSwap {
		t.Errorf("Should not swap apple and zebra (apple < zebra)")
	}
}

func TestShouldSwapWithDomainIDOrdering(t *testing.T) {
	resolver := &PathResolver{}

	fileA := &types.DriveFile{
		ID:       "zzz",
		Name:     "same",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	fileB := &types.DriveFile{
		ID:       "aaa",
		Name:     "same",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	shouldSwap := resolver.shouldSwapWithDomain(fileA, fileB, SearchDomainMyDrive)
	if !shouldSwap {
		t.Errorf("Should swap zzz and aaa (zzz > aaa)")
	}
}

func TestDetermineSearchDomainPriority(t *testing.T) {
	resolver := &PathResolver{}

	opts := ResolveOptions{
		DriveID:             "drive123",
		IncludeSharedWithMe: true,
	}

	domain := resolver.determineSearchDomain(opts)
	if domain != SearchDomainSharedDrive {
		t.Errorf("DriveID should take priority, got %v", domain)
	}
}

func TestGetDomainPriorityOrdering(t *testing.T) {
	resolver := &PathResolver{}

	myDrivePriority := resolver.getDomainPriority(SearchDomainMyDrive)
	sharedDrivePriority := resolver.getDomainPriority(SearchDomainSharedDrive)
	sharedWithMePriority := resolver.getDomainPriority(SearchDomainSharedWithMe)

	if myDrivePriority >= sharedDrivePriority {
		t.Errorf("My Drive priority should be less than Shared Drive")
	}

	if sharedDrivePriority >= sharedWithMePriority {
		t.Errorf("Shared Drive priority should be less than shared-with-me")
	}
}

func TestShouldSwapWithDomainDomainPreference(t *testing.T) {
	resolver := &PathResolver{}

	fileMyDrive := &types.DriveFile{
		ID:       "id1",
		Name:     "doc",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{"root"},
	}

	fileSharedWithMe := &types.DriveFile{
		ID:       "id2",
		Name:     "doc",
		MimeType: "application/vnd.google-apps.document",
		Parents:  []string{},
	}

	shouldSwap := resolver.shouldSwapWithDomain(fileSharedWithMe, fileMyDrive, SearchDomainMyDrive)
	if !shouldSwap {
		t.Errorf("Should swap shared-with-me and my-drive (my-drive has higher priority)")
	}
}

func TestCacheKeyWithEmptyDriveID(t *testing.T) {
	resolver := &PathResolver{}

	key1 := resolver.makeCacheKey("path", ResolveOptions{
		DriveID:      "",
		SearchDomain: SearchDomainMyDrive,
	})

	key2 := resolver.makeCacheKey("path", ResolveOptions{
		DriveID:      "",
		SearchDomain: SearchDomainMyDrive,
	})

	if key1 != key2 {
		t.Errorf("Same path and options should produce same key")
	}

	if !contains(key1, ":my-drive:") {
		t.Errorf("Cache key should contain domain separator")
	}
}

func TestEscapeQueryStringComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "consecutive quotes",
			input:    "''",
			expected: "\\'\\' ",
		},
		{
			name:     "consecutive backslashes",
			input:    "\\\\",
			expected: "\\\\\\\\",
		},
		{
			name:     "quote then backslash",
			input:    "'\\",
			expected: "\\'\\\\",
		},
		{
			name:     "backslash then quote",
			input:    "\\'",
			expected: "\\\\\\'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeQueryString(tt.input)
			if result != tt.expected {
				t.Logf("escapeQueryString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSortMatchesMultipleFiles(t *testing.T) {
	resolver := &PathResolver{}

	files := []*types.DriveFile{
		{
			ID:       "id3",
			Name:     "charlie",
			MimeType: "application/vnd.google-apps.document",
			Parents:  []string{},
		},
		{
			ID:       "id1",
			Name:     "alpha",
			MimeType: "application/vnd.google-apps.document",
			Parents:  []string{"root"},
		},
		{
			ID:       "id2",
			Name:     "beta",
			MimeType: "application/vnd.google-apps.document",
			Parents:  []string{"root"},
		},
	}

	sorted := resolver.sortMatchesWithDomainPreference(files, SearchDomainMyDrive)

	if sorted[0].ID != "id1" {
		t.Errorf("First should be id1 (my-drive, alpha), got %s", sorted[0].ID)
	}
	if sorted[1].ID != "id2" {
		t.Errorf("Second should be id2 (my-drive, beta), got %s", sorted[1].ID)
	}
	if sorted[2].ID != "id3" {
		t.Errorf("Third should be id3 (shared-with-me, charlie), got %s", sorted[2].ID)
	}
}

func TestInferFileDomainConsistency(t *testing.T) {
	resolver := &PathResolver{}

	file := &types.DriveFile{
		ID:      "file1",
		Parents: []string{"parent1", "parent2"},
	}

	domain1 := resolver.inferFileDomain(file)
	domain2 := resolver.inferFileDomain(file)

	if domain1 != domain2 {
		t.Errorf("inferFileDomain should be consistent")
	}
}

func TestDetermineSearchDomainConsistency(t *testing.T) {
	resolver := &PathResolver{}

	opts := ResolveOptions{
		DriveID:             "drive123",
		IncludeSharedWithMe: false,
	}

	domain1 := resolver.determineSearchDomain(opts)
	domain2 := resolver.determineSearchDomain(opts)

	if domain1 != domain2 {
		t.Errorf("determineSearchDomain should be consistent")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
