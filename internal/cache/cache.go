// Package cache provides caching functionality for Google Drive metadata.
// It supports both in-memory and persistent (SQLite) caching with TTL support.
package cache

import (
	"encoding/json"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache by key.
	// Returns nil if the key doesn't exist or has expired.
	Get(key string) ([]byte, bool)

	// Set stores a value in the cache with the given TTL.
	// If ttl is 0, the default TTL is used.
	Set(key string, value []byte, ttl time.Duration) error

	// Delete removes a value from the cache by key.
	Delete(key string) error

	// Clear removes all values from the cache.
	Clear() error

	// Close closes the cache and releases any resources.
	Close() error

	// Stats returns cache statistics.
	Stats() Stats
}

// Stats represents cache statistics
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Size       int   `json:"size"`
	MaxSize    int   `json:"maxSize,omitempty"`
	Entries    int   `json:"entries"`
	Enabled    bool  `json:"enabled"`
	Persistent bool  `json:"persistent"`
}

// Entry represents a cached entry
type Entry struct {
	Key         string    `json:"key"`
	Value       []byte    `json:"value"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	AccessedAt  time.Time `json:"accessedAt"`
	AccessCount int64     `json:"accessCount"`
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Key types for different cached data
const (
	// Key prefixes for different types of cached data
	KeyPrefixFile       = "file:"
	KeyPrefixFolder     = "folder:"
	KeyPrefixFolderList = "folder_list:"
	KeyPrefixDrive      = "drive:"
	KeyPrefixUser       = "user:"
	KeyPrefixPath       = "path:"
)

// BuildKey builds a cache key with the given prefix and identifier
func BuildKey(prefix, id string) string {
	return prefix + id
}

// BuildFolderListKey builds a cache key for folder listings
func BuildFolderListKey(folderID string, driveID string, pageToken string) string {
	key := KeyPrefixFolderList + folderID
	if driveID != "" {
		key += ":" + driveID
	}
	if pageToken != "" {
		key += ":" + pageToken
	}
	return key
}

// BuildPathKey builds a cache key for path resolution
func BuildPathKey(path string, driveID string) string {
	key := KeyPrefixPath + path
	if driveID != "" {
		key += ":" + driveID
	}
	return key
}

// MarshalJSON implements json.Marshaler for Stats
func (s Stats) MarshalJSON() ([]byte, error) {
	type Alias Stats
	return json.Marshal(&struct {
		HitRate float64 `json:"hitRate"`
		*Alias
	}{
		HitRate: s.HitRate(),
		Alias:   (*Alias)(&s),
	})
}

// HitRate calculates the cache hit rate as a percentage
func (s *Stats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total) * 100
}
