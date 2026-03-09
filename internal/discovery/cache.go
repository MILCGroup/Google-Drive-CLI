package discovery

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dl-alexandre/gdrv/internal/cache"
)

const (
	// DiscoveryCacheTTL is the default TTL for discovery documents (24 hours)
	DiscoveryCacheTTL = 24 * time.Hour

	// DiscoveryCacheKeyPrefix is used for cache key namespacing
	DiscoveryCacheKeyPrefix = "discovery:"
)

// Cache provides caching for discovery documents
type Cache struct {
	backend cache.Cache
}

// NewCache creates a new discovery cache backed by the provided cache
func NewCache(backend cache.Cache) *Cache {
	return &Cache{backend: backend}
}

// GetDocument retrieves a cached discovery document
func (c *Cache) GetDocument(serviceName, version string) (*CachedDiscoveryDoc, bool) {
	key := c.docKey(serviceName, version)
	data, found := c.backend.Get(key)
	if !found {
		return nil, false
	}

	doc, err := CachedDiscoveryDocFromJSON(data)
	if err != nil {
		return nil, false
	}

	// Check if expired (cache might have TTL but we double-check)
	if time.Since(doc.FetchedAt) > DiscoveryCacheTTL {
		return nil, false
	}

	return doc, true
}

// SetDocument stores a discovery document in the cache
func (c *Cache) SetDocument(serviceName, version string, doc *CachedDiscoveryDoc) error {
	key := c.docKey(serviceName, version)
	data, err := doc.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize discovery document: %w", err)
	}

	return c.backend.Set(key, data, DiscoveryCacheTTL)
}

// GetDirectory retrieves a cached API directory
func (c *Cache) GetDirectory() (*APIDirectoryList, bool) {
	key := DiscoveryCacheKeyPrefix + "directory"
	data, found := c.backend.Get(key)
	if !found {
		return nil, false
	}

	var dir APIDirectoryList
	if err := json.Unmarshal(data, &dir); err != nil {
		return nil, false
	}

	return &dir, true
}

// SetDirectory stores the API directory in the cache
func (c *Cache) SetDirectory(dir *APIDirectoryList) error {
	key := DiscoveryCacheKeyPrefix + "directory"
	data, err := json.Marshal(dir)
	if err != nil {
		return fmt.Errorf("failed to serialize API directory: %w", err)
	}

	return c.backend.Set(key, data, DiscoveryCacheTTL)
}

// Invalidate removes a cached discovery document
func (c *Cache) Invalidate(serviceName, version string) error {
	key := c.docKey(serviceName, version)
	// Use Delete if available, otherwise set with immediate expiration
	// The cache interface may vary, so we try different approaches
	if sqliteCache, ok := c.backend.(*cache.SQLiteCache); ok {
		return sqliteCache.Delete(key)
	}
	// For other cache types, we can't easily delete, so just let it expire
	return nil
}

// InvalidateAll clears all discovery-related cached data
func (c *Cache) InvalidateAll() error {
	// Since we don't have a wildcard delete, this is a no-op for now
	// The caller can clear the entire cache if needed
	return nil
}

// docKey generates a cache key for a discovery document
func (c *Cache) docKey(serviceName, version string) string {
	if version == "" {
		version = "preferred"
	}
	return fmt.Sprintf("%s%s:%s", DiscoveryCacheKeyPrefix, serviceName, version)
}
