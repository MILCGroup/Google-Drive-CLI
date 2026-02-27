# Changelog

For detailed release notes and version history, please see the [GitHub Releases](https://github.com/dl-alexandre/Google-Drive-CLI/releases) page.

## v1.0.3 (2026-02-27)

### Fixes
- **Style**: Applied `go fmt` formatting to 5 files to fix CI linting
  - `internal/api/client.go`
  - `internal/auth/oauth_client_defaults.go`
  - `internal/cli/output.go`
  - `internal/cli/sync.go`
  - `internal/config/output.go`

## v1.0.2 (2026-02-27)

### Fork-Specific Changes
- **CI/CD**: Fixed opencode workflow ProviderModelNotFoundError by using default model
- **CI/CD**: Improved sync-upstream workflow with conflict detection and OpenCode automation prompts
- **Docs**: Added fork divergence patterns documentation for future upstream syncs
- **Docs**: Documented recurring conflict types (import paths, OAuth requirements, fork dispatch)

### Upstream Sync (from dl-alexandre/gdrv)
- **Features**: Progress bars for uploads/downloads with rate limiting
- **Features**: Batch file operations (upload/download/delete)
- **Features**: Fuzzy file search with interactive selection
- **Features**: Metadata caching layer (memory + SQLite backend)
- **Features**: Shell completions for bash/zsh/fish/powershell
- **Docs**: Split documentation into focused guides (API, Auth, Troubleshooting)
- **CI/CD**: Updated GitHub Actions, added security scanning
- **Dependencies**: Updated tablewriter, google-api-go-client, sqlite

## Recent Highlights

### Advanced APIs
All four advanced APIs have been implemented:

- **Drive Activity API (v2)** - Query file and folder activity with comprehensive filtering
- **Drive Labels API (v2)** - Manage labels and apply structured metadata to files
- **Drive Changes API (v3)** - Track file changes for sync and automation workflows
- **Permissions Enhancements** - Audit, analyze, and bulk-manage permissions

### Available Scope Presets

All new scope presets are available for authentication:

- `workspace-activity`: Workspace + Activity API (read-only)
- `workspace-labels`: Workspace + Labels API
- `workspace-sync`: Workspace + Changes API
- `workspace-complete`: All Workspace APIs + Activity + Labels + Changes

## Version History

See [GitHub Releases](https://github.com/dl-alexandre/Google-Drive-CLI/releases) for:
- Detailed release notes
- Breaking changes
- New features
- Bug fixes
- Migration guides
