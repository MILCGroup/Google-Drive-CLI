package cache

import (
	"sync"
	"time"
)

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	mu          sync.RWMutex
	entries     map[string]*Entry
	maxSize     int
	defaultTTL  time.Duration
	stats       Stats
	stopCleanup chan struct{}
}

// MemoryCacheOptions contains configuration options for the memory cache
type MemoryCacheOptions struct {
	MaxSize    int
	DefaultTTL time.Duration
}

// DefaultMemoryCacheOptions returns default options for the memory cache
func DefaultMemoryCacheOptions() MemoryCacheOptions {
	return MemoryCacheOptions{
		MaxSize:    10000,           // Max 10,000 entries
		DefaultTTL: 5 * time.Minute, // Default 5 minute TTL
	}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(opts MemoryCacheOptions) *MemoryCache {
	if opts.MaxSize <= 0 {
		opts.MaxSize = DefaultMemoryCacheOptions().MaxSize
	}
	if opts.DefaultTTL <= 0 {
		opts.DefaultTTL = DefaultMemoryCacheOptions().DefaultTTL
	}

	cache := &MemoryCache{
		entries:    make(map[string]*Entry),
		maxSize:    opts.MaxSize,
		defaultTTL: opts.DefaultTTL,
		stats: Stats{
			MaxSize: opts.MaxSize,
			Enabled: true,
		},
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.stats.Misses++
		c.mu.Unlock()
		return nil, false
	}

	// Check if expired
	if entry.IsExpired() {
		c.mu.Lock()
		delete(c.entries, key)
		c.stats.Misses++
		c.stats.Entries = len(c.entries)
		c.mu.Unlock()
		return nil, false
	}

	// Update access metadata
	c.mu.Lock()
	entry.AccessedAt = time.Now()
	entry.AccessCount++
	c.stats.Hits++
	c.mu.Unlock()

	return entry.Value, true
}

// Set stores a value in the cache
func (c *MemoryCache) Set(key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxSize {
		c.evictOldest(1)
	}

	now := time.Now()
	c.entries[key] = &Entry{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		AccessedAt:  now,
		AccessCount: 0,
	}
	c.stats.Entries = len(c.entries)
	c.stats.Size = c.calculateSize()

	return nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.stats.Entries = len(c.entries)
	c.stats.Size = c.calculateSize()

	return nil
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
	c.stats.Entries = 0
	c.stats.Size = 0
	c.stats.Hits = 0
	c.stats.Misses = 0

	return nil
}

// Close closes the cache and stops the cleanup goroutine
func (c *MemoryCache) Close() error {
	close(c.stopCleanup)
	return c.Clear()
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Enabled = true
	stats.Persistent = false
	return stats
}

// evictOldest removes the oldest entries from the cache
func (c *MemoryCache) evictOldest(count int) {
	type keyTime struct {
		key  string
		time time.Time
	}

	// Collect entries with their access times
	entries := make([]keyTime, 0, len(c.entries))
	for k, v := range c.entries {
		entries = append(entries, keyTime{key: k, time: v.AccessedAt})
	}

	// Sort by access time (oldest first)
	// Simple bubble sort for small datasets
	for i := 0; i < len(entries)-1 && i < count; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].time.Before(entries[i].time) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		delete(c.entries, entries[i].key)
	}
}

// cleanupLoop runs periodically to remove expired entries
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired entries
func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
	c.stats.Entries = len(c.entries)
	c.stats.Size = c.calculateSize()
}

// calculateSize estimates the memory size of all entries
func (c *MemoryCache) calculateSize() int {
	size := 0
	for _, entry := range c.entries {
		size += len(entry.Key) + len(entry.Value) + 100 // 100 bytes overhead per entry
	}
	return size
}
