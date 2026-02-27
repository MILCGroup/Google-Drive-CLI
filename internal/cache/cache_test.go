package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	opts := MemoryCacheOptions{
		MaxSize:    100,
		DefaultTTL: time.Minute,
	}
	cache := NewMemoryCache(opts)
	defer func() { _ = cache.Close() }()

	// Test Set and Get
	t.Run("SetAndGet", func(t *testing.T) {
		key := "test:key"
		value := []byte("test value")

		if err := cache.Set(key, value, 0); err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		got, found := cache.Get(key)
		if !found {
			t.Fatal("Expected to find key")
		}
		if string(got) != string(value) {
			t.Errorf("Expected %q, got %q", value, got)
		}
	})

	// Test cache miss
	t.Run("CacheMiss", func(t *testing.T) {
		_, found := cache.Get("nonexistent:key")
		if found {
			t.Error("Expected cache miss for nonexistent key")
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete:key"
		value := []byte("to be deleted")

		_ = cache.Set(key, value, 0)
		_ = cache.Delete(key)

		_, found := cache.Get(key)
		if found {
			t.Error("Expected key to be deleted")
		}
	})

	// Test Clear
	t.Run("Clear", func(t *testing.T) {
		_ = cache.Set("key1", []byte("value1"), 0)
		_ = cache.Set("key2", []byte("value2"), 0)

		if err := cache.Clear(); err != nil {
			t.Fatalf("Clear failed: %v", err)
		}

		_, found1 := cache.Get("key1")
		_, found2 := cache.Get("key2")
		if found1 || found2 {
			t.Error("Expected all keys to be cleared")
		}
	})

	// Test TTL expiration
	t.Run("TTLExpiration", func(t *testing.T) {
		key := "ttl:key"
		value := []byte("expires quickly")

		_ = cache.Set(key, value, 1*time.Millisecond)
		time.Sleep(50 * time.Millisecond) // Wait for expiration

		_, found := cache.Get(key)
		if found {
			t.Error("Expected key to expire")
		}
	})

	// Test Stats
	t.Run("Stats", func(t *testing.T) {
		_ = cache.Clear()

		_ = cache.Set("stats:key1", []byte("value1"), 0)
		_ = cache.Set("stats:key2", []byte("value2"), 0)
		cache.Get("stats:key1") // Hit
		cache.Get("stats:miss") // Miss

		stats := cache.Stats()
		if stats.Entries != 2 {
			t.Errorf("Expected 2 entries, got %d", stats.Entries)
		}
		if stats.Hits != 1 {
			t.Errorf("Expected 1 hit, got %d", stats.Hits)
		}
		if stats.Misses < 1 {
			t.Errorf("Expected at least 1 miss, got %d", stats.Misses)
		}
	})
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		prefix   string
		id       string
		expected string
	}{
		{"file:", "123", "file:123"},
		{"folder:", "abc", "folder:abc"},
		{"drive:", "xyz", "drive:xyz"},
	}

	for _, tt := range tests {
		got := BuildKey(tt.prefix, tt.id)
		if got != tt.expected {
			t.Errorf("BuildKey(%q, %q) = %q; want %q", tt.prefix, tt.id, got, tt.expected)
		}
	}
}

func TestEntryIsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entry    Entry
		expected bool
	}{
		{
			name: "not expired",
			entry: Entry{
				ExpiresAt: now.Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			entry: Entry{
				ExpiresAt: now.Add(-time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStatsHitRate(t *testing.T) {
	tests := []struct {
		name     string
		stats    Stats
		expected float64
	}{
		{
			name:     "empty stats",
			stats:    Stats{Hits: 0, Misses: 0},
			expected: 0,
		},
		{
			name:     "all hits",
			stats:    Stats{Hits: 100, Misses: 0},
			expected: 100,
		},
		{
			name:     "all misses",
			stats:    Stats{Hits: 0, Misses: 100},
			expected: 0,
		},
		{
			name:     "50/50",
			stats:    Stats{Hits: 50, Misses: 50},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stats.HitRate(); got != tt.expected {
				t.Errorf("HitRate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
