package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/milcgroup/gdrv/internal/cache"
	"github.com/milcgroup/gdrv/internal/config"
	"github.com/milcgroup/gdrv/internal/utils"
)

// CacheCmd provides cache management commands
type CacheCmd struct {
	Clear  CacheClearCmd  `cmd:"" help:"Clear all cached data"`
	Status CacheStatusCmd `cmd:"" help:"Show cache statistics"`
}

// CacheClearCmd clears the cache
type CacheClearCmd struct{}

// CacheStatusCmd shows cache statistics
type CacheStatusCmd struct{}

func (cmd *CacheClearCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Load config to get cache settings
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if !cfg.CacheEnabled {
		out.Log("Cache is disabled. Nothing to clear.")
		return out.WriteSuccess("cache.clear", map[string]interface{}{
			"cleared": false,
			"reason":  "cache disabled",
		})
	}

	// Confirm if not forced
	if !flags.Force {
		fmt.Print("Are you sure you want to clear the cache? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return out.WriteSuccess("cache.clear", map[string]interface{}{
				"cleared": false,
				"reason":  "user cancelled",
			})
		}
	}

	// Create cache manager
	opts := cache.ManagerOptions{
		Enabled:    cfg.CacheEnabled,
		CacheType:  cfg.CacheType,
		DefaultTTL: cfg.GetCacheTTL(),
		MaxSize:    cfg.MaxCacheSize,
	}

	// Set cache path for SQLite
	if cfg.CacheType == "sqlite" {
		configDir, err := config.GetConfigDir()
		if err == nil {
			opts.CachePath = filepath.Join(configDir, "cache.db")
		}
	}

	mgr, err := cache.NewManager(opts)
	if err != nil {
		return out.WriteError("cache.clear", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to initialize cache: %v", err)).Build())
	}
	defer func() { _ = mgr.Close() }()

	// Clear the cache
	if err := mgr.InvalidateAll(); err != nil {
		return out.WriteError("cache.clear", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to clear cache: %v", err)).Build())
	}

	out.Log("Cache cleared successfully")
	return out.WriteSuccess("cache.clear", map[string]interface{}{
		"cleared": true,
	})
}

func (cmd *CacheStatusCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Load config to get cache settings
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Create cache manager
	opts := cache.ManagerOptions{
		Enabled:    cfg.CacheEnabled,
		CacheType:  cfg.CacheType,
		DefaultTTL: cfg.GetCacheTTL(),
		MaxSize:    cfg.MaxCacheSize,
	}

	// Set cache path for SQLite
	if cfg.CacheType == "sqlite" {
		configDir, err := config.GetConfigDir()
		if err == nil {
			opts.CachePath = filepath.Join(configDir, "cache.db")
		}
	}

	mgr, err := cache.NewManager(opts)
	if err != nil {
		return out.WriteError("cache.status", utils.NewCLIError(utils.ErrCodeUnknown,
			fmt.Sprintf("Failed to initialize cache: %v", err)).Build())
	}
	defer func() { _ = mgr.Close() }()

	// Get stats
	stats := mgr.Stats()

	// Add config info to response
	result := map[string]interface{}{
		"enabled":    cfg.CacheEnabled,
		"type":       cfg.CacheType,
		"ttl":        cfg.CacheTTL,
		"maxSize":    cfg.MaxCacheSize,
		"hits":       stats.Hits,
		"misses":     stats.Misses,
		"hitRate":    stats.HitRate(),
		"entries":    stats.Entries,
		"persistent": stats.Persistent,
	}

	// For SQLite, check if the file exists and get its size
	if cfg.CacheType == "sqlite" && opts.CachePath != "" {
		if info, err := os.Stat(opts.CachePath); err == nil {
			result["fileSize"] = info.Size()
			result["cachePath"] = opts.CachePath
		}
	}

	return out.WriteSuccess("cache.status", result)
}
