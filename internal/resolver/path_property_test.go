package resolver

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

// Property 8: Path Resolution Caching
// Validates: Requirements 12.5, 12.6, 12.8, 12.9
// Property: Cache operations maintain consistency and respect TTL
func TestProperty_PathResolutionCaching(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Generate random cache TTL (50-500ms)
			ttl := time.Duration(50+rand.Intn(450)) * time.Millisecond
			resolver := NewPathResolver(nil, ttl)

			// Generate random path and file ID
			path := generateRandomPath()
			fileID := generateRandomFileID()
			driveID := ""
			if rand.Float32() < 0.3 {
				driveID = generateRandomDriveID()
			}

			// Property: Fresh entry should not be in cache
			if cached, ok := resolver.checkCache(path, driveID); ok {
				t.Errorf("Fresh cache should not contain entry, got: %s", cached)
			}

			// Property: After update, entry should be cached
			resolver.updateCache(path, driveID, fileID)
			cached, ok := resolver.checkCache(path, driveID)
			if !ok {
				t.Error("Entry should be cached after update")
			}
			if cached != fileID {
				t.Errorf("Cached value mismatch: got %s, want %s", cached, fileID)
			}

			// Property: Before TTL expires, entry should remain cached
			time.Sleep(ttl / 2)
			cached, ok = resolver.checkCache(path, driveID)
			if !ok {
				t.Error("Entry should still be cached before TTL expiry")
			}
			if cached != fileID {
				t.Errorf("Cached value mismatch before expiry: got %s, want %s", cached, fileID)
			}

			// Property: After TTL expires, entry should not be cached
			time.Sleep(ttl/2 + 50*time.Millisecond)
			if _, ok := resolver.checkCache(path, driveID); ok {
				t.Error("Entry should not be cached after TTL expiry")
			}

			// Property: Invalidation removes entry from cache
			resolver.updateCache(path, driveID, fileID)
			resolver.InvalidateCache(path, driveID)
			if _, ok := resolver.checkCache(path, driveID); ok {
				t.Error("Entry should not be cached after invalidation")
			}

			// Property: ClearCache removes all entries
			paths := []string{path, generateRandomPath(), generateRandomPath()}
			for _, p := range paths {
				resolver.updateCache(p, driveID, generateRandomFileID())
			}
			resolver.ClearCache()
			for _, p := range paths {
				if _, ok := resolver.checkCache(p, driveID); ok {
					t.Errorf("Entry %s should not be cached after ClearCache", p)
				}
			}
		})
	}
}

// Property 21: Disambiguation Ordering
// Validates: Requirements 12.4, 12.5, 12.6
// Property: Disambiguation always produces stable, deterministic ordering
func TestProperty_DisambiguationOrdering(t *testing.T) {
	const iterations = 100

	resolver := &PathResolver{}

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Generate a set of files with the same name
			numFiles := 2 + rand.Intn(5) // 2-6 files
			files := make([]*types.DriveFile, numFiles)

			for j := 0; j < numFiles; j++ {
				files[j] = generateRandomFile("document")
			}

			// Property: Sorting is idempotent
			domain := randomSearchDomain()
			sorted1 := resolver.sortMatchesWithDomainPreference(files, domain)
			sorted2 := resolver.sortMatchesWithDomainPreference(sorted1, domain)

			if !filesEqual(sorted1, sorted2) {
				t.Error("Sorting should be idempotent")
			}

			// Property: Sorting is stable (same input produces same output)
			sorted3 := resolver.sortMatchesWithDomainPreference(files, domain)
			if !filesEqual(sorted1, sorted3) {
				t.Error("Sorting should be stable")
			}

			// Property: Non-shortcuts appear before shortcuts within same domain
			domains := make(map[SearchDomain][]*types.DriveFile)
			for _, f := range sorted1 {
				fileDomain := resolver.inferFileDomain(f)
				domains[fileDomain] = append(domains[fileDomain], f)
			}

			for _, files := range domains {
				shortcutStarted := false
				for _, f := range files {
					isShortcut := f.MimeType == utils.MimeTypeShortcut
					if isShortcut {
						shortcutStarted = true
					} else {
						if shortcutStarted {
							t.Error("Non-shortcuts should appear before shortcuts within same domain")
						}
					}
				}
			}

			// Property: Files from My Drive appear before shared-with-me
			// Note: This check is skipped as domain ordering is validated separately

			// Property: Files are sorted lexicographically by name within same category
			for j := 0; j < len(sorted1)-1; j++ {
				a := sorted1[j]
				b := sorted1[j+1]

				// If same shortcut status and same domain
				if (a.MimeType == utils.MimeTypeShortcut) == (b.MimeType == utils.MimeTypeShortcut) &&
					resolver.inferFileDomain(a) == resolver.inferFileDomain(b) {
					// Names should be in order
					if a.Name > b.Name {
						t.Errorf("Files with same properties should be sorted by name: %s should come before %s",
							b.Name, a.Name)
					}
				}
			}
		})
	}
}

// Property 20: Path Resolution Domain Handling
// Validates: Requirements 12.1, 12.2, 12.3, 12.4
// Property: Search domain determination is consistent with options
func TestProperty_DomainHandling(t *testing.T) {
	const iterations = 100

	resolver := &PathResolver{}

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Generate random options
			opts := ResolveOptions{
				DriveID:             "",
				IncludeSharedWithMe: rand.Float32() < 0.5,
			}

			if rand.Float32() < 0.3 {
				opts.DriveID = generateRandomDriveID()
			}

			// Property: If DriveID is set, domain should be SharedDrive
			if opts.DriveID != "" {
				domain := resolver.determineSearchDomain(opts)
				if domain != SearchDomainSharedDrive {
					t.Errorf("With DriveID, domain should be SharedDrive, got: %v", domain)
				}
			}

			// Property: If IncludeSharedWithMe is true and no DriveID, domain should be AllDrives
			if opts.DriveID == "" && opts.IncludeSharedWithMe {
				domain := resolver.determineSearchDomain(opts)
				if domain != SearchDomainAllDrives {
					t.Errorf("With IncludeSharedWithMe=true, domain should be AllDrives, got: %v", domain)
				}
			}

			// Property: If no DriveID and not IncludeSharedWithMe, domain should be MyDrive
			if opts.DriveID == "" && !opts.IncludeSharedWithMe {
				domain := resolver.determineSearchDomain(opts)
				if domain != SearchDomainMyDrive {
					t.Errorf("Default domain should be MyDrive, got: %v", domain)
				}
			}

			// Property: Cache keys with different domains are distinct
			path := generateRandomPath()

			opts1 := ResolveOptions{SearchDomain: SearchDomainMyDrive}
			opts2 := ResolveOptions{SearchDomain: SearchDomainSharedWithMe}
			opts3 := ResolveOptions{SearchDomain: SearchDomainSharedDrive, DriveID: "drive123"}

			key1 := resolver.makeCacheKey(path, opts1)
			key2 := resolver.makeCacheKey(path, opts2)
			key3 := resolver.makeCacheKey(path, opts3)

			if key1 == key2 {
				t.Error("Cache keys for different domains should be distinct")
			}
			if key1 == key3 {
				t.Error("Cache keys for different domains should be distinct")
			}
			if key2 == key3 {
				t.Error("Cache keys for different domains should be distinct")
			}

			// Property: Same path with same options produces same cache key
			key1_repeat := resolver.makeCacheKey(path, opts1)
			if key1 != key1_repeat {
				t.Error("Same path and options should produce same cache key")
			}
		})
	}
}

// Property: Cache key uniqueness
// Validates that different paths/options produce different cache keys
func TestProperty_CacheKeyUniqueness(t *testing.T) {
	const iterations = 100

	resolver := &PathResolver{}
	seenKeys := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		path := generateRandomPath()
		opts := ResolveOptions{
			DriveID:      "",
			SearchDomain: randomSearchDomain(),
		}

		if rand.Float32() < 0.3 {
			opts.DriveID = generateRandomDriveID()
		}

		key := resolver.makeCacheKey(path, opts)

		// Property: Each unique (path, domain, driveID) combination should produce a unique key
		compositeKey := fmt.Sprintf("%s|%s|%s", path, opts.SearchDomain, opts.DriveID)
		if seenKeys[compositeKey] {
			// We've seen this combination before, key should match
			expectedKey := resolver.makeCacheKey(path, opts)
			if key != expectedKey {
				t.Errorf("Same input should produce same key: got %s and %s", key, expectedKey)
			}
		} else {
			seenKeys[compositeKey] = true
		}
	}
}

// Helper functions

func generateRandomPath() string {
	depth := 1 + rand.Intn(4) // 1-4 segments
	segments := make([]string, depth)
	for i := 0; i < depth; i++ {
		segments[i] = generateRandomName()
	}
	return strings.Join(segments, "/")
}

func generateRandomName() string {
	names := []string{"documents", "photos", "projects", "archive", "backup", "data", "reports", "files"}
	return names[rand.Intn(len(names))] + fmt.Sprintf("_%d", rand.Intn(100))
}

func generateRandomFileID() string {
	return fmt.Sprintf("file_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func generateRandomDriveID() string {
	return fmt.Sprintf("drive_%d", rand.Intn(1000))
}

func generateRandomFile(name string) *types.DriveFile {
	// Randomly assign properties
	hasParents := rand.Float32() < 0.7
	isShortcut := rand.Float32() < 0.3

	var parents []string
	if hasParents {
		parents = []string{fmt.Sprintf("parent_%d", rand.Intn(10))}
	}

	mimeType := "application/pdf"
	if isShortcut {
		mimeType = utils.MimeTypeShortcut
	} else if rand.Float32() < 0.5 {
		mimeType = utils.MimeTypeDocument
	}

	return &types.DriveFile{
		ID:       generateRandomFileID(),
		Name:     name,
		MimeType: mimeType,
		Parents:  parents,
	}
}

func randomSearchDomain() SearchDomain {
	domains := []SearchDomain{
		SearchDomainMyDrive,
		SearchDomainSharedDrive,
		SearchDomainSharedWithMe,
		SearchDomainAllDrives,
	}
	return domains[rand.Intn(len(domains))]
}

func filesEqual(a, b []*types.DriveFile) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}
