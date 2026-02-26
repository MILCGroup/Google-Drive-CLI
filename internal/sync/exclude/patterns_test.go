package exclude

import (
	"testing"
)

func TestDefaultPatterns(t *testing.T) {
	patterns := DefaultPatterns()

	if len(patterns) == 0 {
		t.Fatal("DefaultPatterns should return non-empty slice")
	}

	expectedPatterns := []string{
		".git/",
		".DS_Store",
		"._*",
		"node_modules/",
		"vendor/",
		"*.log",
		"*.tmp",
		".env",
		".env.*",
		"*.key",
		"*.pem",
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultPatterns should include %q", expected)
		}
	}
}

func TestNew(t *testing.T) {
	t.Run("nil patterns", func(t *testing.T) {
		m := New(nil)
		if m == nil {
			t.Fatal("New(nil) should return non-nil Matcher")
		}
		if len(m.patterns) == 0 {
			t.Error("New(nil) should still include default patterns")
		}
	})

	t.Run("empty patterns", func(t *testing.T) {
		m := New([]string{})
		if m == nil {
			t.Fatal("New([]) should return non-nil Matcher")
		}
		if len(m.patterns) == 0 {
			t.Error("New([]) should still include default patterns")
		}
	})

	t.Run("custom patterns", func(t *testing.T) {
		custom := []string{"*.custom", "custom_dir/"}
		m := New(custom)
		if m == nil {
			t.Fatal("New(custom) should return non-nil Matcher")
		}
		// Should include defaults + custom
		if len(m.patterns) < len(custom) {
			t.Errorf("Matcher should have at least %d patterns, got %d", len(custom), len(m.patterns))
		}
	})

	t.Run("patterns with empty strings", func(t *testing.T) {
		custom := []string{"", "*.custom", "", "custom_dir/", ""}
		m := New(custom)
		if m == nil {
			t.Fatal("New(custom) should return non-nil Matcher")
		}
		// Empty strings should be filtered out
		for _, p := range m.patterns {
			if p == "" {
				t.Error("Empty patterns should be filtered out")
			}
		}
	})
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		isDir   bool
		want    bool
	}{
		{"git directory", ".git/", true, true},
		{"DS_Store file", ".DS_Store", false, true},
		{"DS_Store in subdir", "subdir/.DS_Store", false, true},
		{"node_modules directory", "node_modules/", true, true},
		{"vendor directory", "vendor/", true, true},
		{"log file", "app.log", false, true},
		{"tmp file", "temp.tmp", false, true},
		{"dot env file", ".env", false, true},
		{"dot env with suffix", ".env.local", false, true},
		{"key file", "server.key", false, true},
		{"pem file", "cert.pem", false, true},
		{"normal file", "main.go", false, false},
		{"normal directory", "src/", true, false},
		{"file in git", ".git/config", false, true},
		{"file in node_modules", "node_modules/lodash/index.js", false, true},
	}

	m := New(nil)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := m.IsExcluded(tc.relPath, tc.isDir)
			if got != tc.want {
				t.Errorf("IsExcluded(%q, %v) = %v, want %v", tc.relPath, tc.isDir, got, tc.want)
			}
		})
	}
}

func TestIsExcludedNilMatcher(t *testing.T) {
	var m *Matcher
	got := m.IsExcluded(".git/", true)
	if got {
		t.Error("IsExcluded on nil Matcher should return false")
	}
}

func TestIsExcludedGlobPatterns(t *testing.T) {
	m := New(nil)

	tests := []struct {
		name    string
		relPath string
		want    bool
	}{
		{"apple double file", "._file.txt", true},
		{"log wildcard", "debug.log", true},
		{"tmp wildcard", "backup.tmp", true},
		{"key wildcard", "private.key", true},
		{"pem wildcard", "ca.pem", true},
		{"non-matching", "main.go", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := m.IsExcluded(tc.relPath, false)
			if got != tc.want {
				t.Errorf("IsExcluded(%q) = %v, want %v", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestIsExcludedPrefix(t *testing.T) {
	m := New(nil)

	tests := []struct {
		name    string
		relPath string
		isDir   bool
		want    bool
	}{
		{"git prefix - exact", ".git", true, true},
		{"git prefix - file", ".git/config", false, true},
		{"vendor prefix - exact", "vendor", true, true},
		{"vendor prefix - file", "vendor/lib/main.go", false, true},
		{"non-matching", "myvendor/", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := m.IsExcluded(tc.relPath, tc.isDir)
			if got != tc.want {
				t.Errorf("IsExcluded(%q, %v) = %v, want %v", tc.relPath, tc.isDir, got, tc.want)
			}
		})
	}
}
