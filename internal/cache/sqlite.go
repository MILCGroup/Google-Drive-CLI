package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteCache is a persistent cache implementation using SQLite
type SQLiteCache struct {
	db         *sql.DB
	path       string
	defaultTTL time.Duration
	maxSize    int
	stats      Stats
}

// SQLiteCacheOptions contains configuration options for the SQLite cache
type SQLiteCacheOptions struct {
	Path       string
	DefaultTTL time.Duration
	MaxSize    int
}

// DefaultSQLiteCacheOptions returns default options for the SQLite cache
func DefaultSQLiteCacheOptions() SQLiteCacheOptions {
	configDir, _ := os.UserConfigDir()
	cachePath := filepath.Join(configDir, "gdrv", "cache.db")

	return SQLiteCacheOptions{
		Path:       cachePath,
		DefaultTTL: 5 * time.Minute,
		MaxSize:    100000, // Max 100,000 entries
	}
}

// NewSQLiteCache creates a new SQLite-backed cache
func NewSQLiteCache(opts SQLiteCacheOptions) (*SQLiteCache, error) {
	if opts.Path == "" {
		opts.Path = DefaultSQLiteCacheOptions().Path
	}
	if opts.DefaultTTL <= 0 {
		opts.DefaultTTL = DefaultSQLiteCacheOptions().DefaultTTL
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = DefaultSQLiteCacheOptions().MaxSize
	}

	// Ensure the directory exists
	dir := filepath.Dir(opts.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	cache := &SQLiteCache{
		db:         db,
		path:       opts.Path,
		defaultTTL: opts.DefaultTTL,
		maxSize:    opts.MaxSize,
		stats: Stats{
			MaxSize:    opts.MaxSize,
			Enabled:    true,
			Persistent: true,
		},
	}

	// Initialize the database schema
	if err := cache.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize cache schema: %w", err)
	}

	// Clean up expired entries on startup
	if err := cache.cleanupExpired(); err != nil {
		// Log but don't fail - this is just cleanup
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup expired cache entries: %v\n", err)
	}

	return cache, nil
}

// initSchema creates the necessary database tables and indexes
func (c *SQLiteCache) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cache_entries (
		key TEXT PRIMARY KEY,
		value BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		accessed_at INTEGER NOT NULL,
		access_count INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_expires_at ON cache_entries(expires_at);
	CREATE INDEX IF NOT EXISTS idx_accessed_at ON cache_entries(accessed_at);

	CREATE TABLE IF NOT EXISTS cache_stats (
		key TEXT PRIMARY KEY,
		value INTEGER NOT NULL
	);

	-- Insert default stats if not exists
	INSERT OR IGNORE INTO cache_stats (key, value) VALUES ('hits', 0);
	INSERT OR IGNORE INTO cache_stats (key, value) VALUES ('misses', 0);
	`

	_, err := c.db.Exec(schema)
	return err
}

// Get retrieves a value from the cache
func (c *SQLiteCache) Get(key string) ([]byte, bool) {
	var value []byte
	var expiresAt int64

	err := c.db.QueryRow(
		"SELECT value, expires_at FROM cache_entries WHERE key = ?",
		key,
	).Scan(&value, &expiresAt)

	if err == sql.ErrNoRows {
		c.incrementStat("misses")
		return nil, false
	}
	if err != nil {
		c.incrementStat("misses")
		return nil, false
	}

	// Check if expired
	if time.Now().Unix() > expiresAt {
		// Delete expired entry
		_, _ = c.db.Exec("DELETE FROM cache_entries WHERE key = ?", key)
		c.incrementStat("misses")
		return nil, false
	}

	// Update access metadata
	_, _ = c.db.Exec(
		"UPDATE cache_entries SET accessed_at = ?, access_count = access_count + 1 WHERE key = ?",
		time.Now().Unix(), key,
	)

	c.incrementStat("hits")
	return value, true
}

// Set stores a value in the cache
func (c *SQLiteCache) Set(key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	_, err := c.db.Exec(
		`INSERT INTO cache_entries (key, value, created_at, expires_at, accessed_at, access_count)
		VALUES (?, ?, ?, ?, ?, 0)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			created_at = excluded.created_at,
			expires_at = excluded.expires_at,
			accessed_at = excluded.accessed_at`,
		key, value, now.Unix(), expiresAt.Unix(), now.Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to set cache entry: %w", err)
	}

	// Check if we need to evict entries
	count, err := c.entryCount()
	if err == nil && count > c.maxSize {
		_ = c.evictOldest(count - c.maxSize)
	}

	return nil
}

// Delete removes a value from the cache
func (c *SQLiteCache) Delete(key string) error {
	_, err := c.db.Exec("DELETE FROM cache_entries WHERE key = ?", key)
	return err
}

// Clear removes all values from the cache
func (c *SQLiteCache) Clear() error {
	_, err := c.db.Exec("DELETE FROM cache_entries")
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	// Reset stats
	_, _ = c.db.Exec("UPDATE cache_stats SET value = 0 WHERE key IN ('hits', 'misses')")

	return nil
}

// Close closes the cache database connection
func (c *SQLiteCache) Close() error {
	return c.db.Close()
}

// Stats returns cache statistics
func (c *SQLiteCache) Stats() Stats {
	stats := Stats{
		Enabled:    true,
		Persistent: true,
		MaxSize:    c.maxSize,
	}

	// Get persisted stats
	_ = c.db.QueryRow("SELECT value FROM cache_stats WHERE key = 'hits'").Scan(&stats.Hits)
	_ = c.db.QueryRow("SELECT value FROM cache_stats WHERE key = 'misses'").Scan(&stats.Misses)

	// Get entry count
	count, _ := c.entryCount()
	stats.Entries = count

	return stats
}

// incrementStat increments a stat counter in the database
func (c *SQLiteCache) incrementStat(key string) {
	_, _ = c.db.Exec(
		"INSERT INTO cache_stats (key, value) VALUES (?, 1) ON CONFLICT(key) DO UPDATE SET value = value + 1",
		key,
	)
}

// entryCount returns the number of entries in the cache
func (c *SQLiteCache) entryCount() (int, error) {
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM cache_entries").Scan(&count)
	return count, err
}

// cleanupExpired removes all expired entries
func (c *SQLiteCache) cleanupExpired() error {
	now := time.Now().Unix()
	_, err := c.db.Exec("DELETE FROM cache_entries WHERE expires_at < ?", now)
	return err
}

// evictOldest removes the oldest entries from the cache
func (c *SQLiteCache) evictOldest(count int) error {
	_, err := c.db.Exec(
		"DELETE FROM cache_entries WHERE key IN (SELECT key FROM cache_entries ORDER BY accessed_at ASC LIMIT ?)",
		count,
	)
	return err
}
