# AGENTS.md

This file provides guidance to AI agents (Claude Code, Claude CLI, Claude Desktop, and other AI coding assistants) when working with code in this repository.

## About This Repository

**This is a company-specific fork of `gdrv`** - a production-grade Google Drive CLI written in Go. 

### Use Case: Company Deployment with Pre-Bundled OAuth

This fork exists to solve a specific problem: **simplified deployment for companies and AI agents**. Instead of requiring every user or agent to configure OAuth credentials separately, this fork embeds company OAuth credentials directly into the binary.

**Why this fork?**
- **Simplified onboarding**: New team members run `gdrv auth login` without OAuth setup
- **AI agent ready**: Agents can use the tool immediately without credential configuration
- **Single binary deployment**: No external configuration files or environment variables needed
- **Consistent authentication**: All users authenticate through the same OAuth app

### Build Types: Official vs Source

This fork supports two build modes:

**Official Build** (default for downloads)
- OAuth credentials bundled in binary
- Works immediately without env vars
- Built via CI with `make build-official`
- Marked as "official" build type

**Source Build** (for development)
- Requires `GDRV_CLIENT_ID` and `GDRV_CLIENT_SECRET` env vars
- Use `make build` or standard `go build`
- Runtime check enforces official builds for bundled creds

### What's Different from Upstream

| Aspect | Upstream gdrv | This Fork |
|--------|---------------|-----------|
| OAuth Setup | Requires `GDRV_CLIENT_ID` and `GDRV_CLIENT_SECRET` env vars | Pre-bundled in official binaries |
| Installation Steps | 3+ steps (install + OAuth config + auth) | 2 steps (download + auth) |
| AI Agent Usage | Requires manual credential setup | Works out of the box |
| Deployment | Complex with secrets management | Official binary = zero config |

## Quick Reference

### For AI Agents (Zero-Config Operation)

```bash
# Just works - no env vars needed
./bin/gdrv files list --json
./bin/gdrv files upload document.pdf --json
./bin/gdrv folders create "New Project" --json
```

### Authentication (One-Time Setup)

```bash
# Login with bundled company credentials
./bin/gdrv auth login --preset workspace-basic

# For headless/agent environments
./bin/gdrv auth device --preset workspace-basic
```

## Common Commands

### Build and Run

```bash
# Build for current platform (includes bundled credentials)
make build
# Output: bin/gdrv

# Build for all platforms
make build-all

# Build and run with arguments
make run ARGS='version'
make run ARGS='files list --json'
```

### Testing

```bash
# Run all unit tests with race detection
go test -v -race -cover ./...

# Run tests with coverage report
make test-coverage

# Run integration tests (requires credentials)
go test -tags=integration ./test/integration/...
```

### Linting

```bash
# Run golangci-lint
make lint

# Run go vet
go vet ./...
```

## High-Level Architecture

### Core Design Patterns

1. **Manager Pattern**: Each domain has a Manager type in `internal/<domain>/manager.go`
2. **API Client Layer**: `internal/api/client.go` with retry logic and rate limiting
3. **Request Context**: All operations use `types.RequestContext` with TraceID
4. **CLI Command Structure**: Commands in `internal/cli/` using spf13/cobra

### Critical Components

#### Authentication (`internal/auth/`)

- **Manager**: Handles OAuth2, device code flow, service accounts
- **Bundled Credentials**: Company OAuth credentials in `oauth_client_defaults.go`
- **Storage Backends**: Three-tier storage (keyring → encrypted file → plain file)
- **Scope Presets**: Pre-configured scope sets (workspace-basic, workspace-full, admin)
- **Profile Support**: Multiple authentication profiles

#### OAuth Credentials Location

OAuth credentials are injected at **build time** via ldflags. The source contains no credentials.

**For Official CI Builds:**
- Credentials stored in GitHub Secrets: `GDRV_CLIENT_ID`, `GDRV_CLIENT_SECRET`
- Build command: `make build-official` (sets `OFFICIAL_BUILD=true`)
- ldflags inject credentials + mark build as "official"

**For Source Builds:**
- Requires `GDRV_CLIENT_ID` and `GDRV_CLIENT_SECRET` env vars
- Or config file OAuth client settings

The `GetBundledOAuthClient()` function checks:
1. `BundledOAuthClientID` / `BundledOAuthClientSecret` vars (set by ldflags)
2. Returns credentials only if build is marked "official"

### API Flow Pattern

```
CLI Command → Manager Method → API Client → ExecuteWithRetry → Google API
                                                      ↓
                                             Logging + Error Classification
```

### Exit Codes

- 0: Success
- 1: Unknown error
- 2: Authentication required
- 3: Invalid argument
- 4: Resource not found
- 5: Permission denied
- 6: Rate limited

## Key Implementation Notes

### When Adding New Commands

1. Define command in `internal/cli/<domain>.go`
2. Create/extend Manager in `internal/<domain>/manager.go`
3. Add types to `internal/types/<domain>.go` if needed
4. Use `api.ExecuteWithRetry()` for all API calls
5. Add `--json` flag support for all commands
6. Return proper exit codes using `internal/utils/` constants

### Authentication Context (Fork-Specific)

- **OAuth credentials are pre-bundled** - no env vars needed
- For Admin SDK: Use `--key-file` and `--impersonate-user`
- Profiles stored in `~/.config/gdrv/` (or `GDRV_CONFIG_DIR`)
- Token refresh happens automatically when within 5 minutes of expiration

### Working with Shared Drives

- Always pass `--drive-id` flag or set it globally
- Resource keys automatically managed by `ResourceKeyManager`
- Use `supportsAllDrives=true` for all Drive API calls

### Google Workspace APIs

- **Sheets**: `internal/sheets/manager.go`
- **Docs**: `internal/docs/manager.go`
- **Slides**: `internal/slides/manager.go`
- **Admin SDK**: `internal/admin/manager.go`

### Error Handling

Use `internal/errors/errors.go`:
- `ClassifyGoogleAPIError()`: Converts Google API errors to CLI errors
- `NewCLIError()`: Creates structured CLI errors
- `AppError`: Wraps errors with additional context

### Version Information

Injected at build time via ldflags (see Makefile):
- `pkg/version.Version`: Git tag or "dev"
- `pkg/version.GitCommit`: Git commit SHA
- `pkg/version.BuildTime`: Build timestamp

## Company-Specific Fork

This is a company-specific fork of [gdrv](https://github.com/dl-alexandre/Google-Drive-CLI) with pre-bundled OAuth credentials for seamless deployment.

## Upstream Merge Workflow

When updates from the upstream `gdrv` repository need to be merged into this fork, follow this workflow to preserve fork-specific changes while incorporating upstream improvements.

### Fork-Specific Files (Do Not Overwrite)

These files contain company-specific modifications and **must never be replaced** with upstream versions:

| File | Description |
|------|-------------|
| `internal/auth/oauth_client_defaults.go` | Pre-bundled OAuth credentials |
| `AGENTS.md` | This documentation file |
| `README.md` | Fork-specific README (maintain fork branding) |

### Merge Procedure

```bash
# 1. Add upstream remote (one-time setup)
git remote add upstream https://github.com/dl-alexandre/Google-Drive-CLI.git

# 2. Fetch upstream changes
git fetch upstream

# 3. Create merge branch
git checkout -b merge-upstream-$(date +%Y%m%d)

# 4. Merge upstream (expect conflicts in fork-specific files)
git merge upstream/main

# 5. For each conflict in fork-specific files, keep OUR version:
#    - internal/auth/oauth_client_defaults.go
#    - AGENTS.md
#    - README.md

# Resolve conflicts: Keep company OAuth credentials
git checkout --ours internal/auth/oauth_client_defaults.go
git add internal/auth/oauth_client_defaults.go

# Resolve conflicts: Keep fork documentation
git checkout --ours AGENTS.md README.md
git add AGENTS.md README.md

# 6. Continue merge for remaining conflicts
git merge --continue

# 7. Build and test
go build -o bin/gdrv ./cmd/gdrv
make test

# 8. Tag with fork version
# Format: upstream-tag-fork-N (e.g., v1.2.3-fork-1)
git tag v$(git describe upstream/main --tags)-fork-$(git tag -l "v*-fork-*" | wc -l | xargs -I {} echo $(({} + 1)))
```

### Post-Merge Checklist

After merging upstream changes:

- [ ] Build succeeds: `make build`
- [ ] All tests pass: `make test`
- [ ] OAuth credentials still work: `gdrv auth status`
- [ ] Fork-specific functionality intact
- [ ] Update version string in `pkg/version/version.go` if needed
- [ ] Update CHANGELOG.md with upstream changes
- [ ] Tag release with fork version suffix

### Handling Breaking Changes

If upstream introduces breaking API changes or restructures authentication:

1. **Stop the merge** - Do not proceed with automatic resolution
2. **Review the diff** carefully, especially `internal/auth/` changes
3. **Update fork code** to maintain bundled OAuth functionality
4. **Test thoroughly** - The bundled OAuth feature is the primary purpose of this fork
5. **Update this section** if merge procedure changes due to upstream restructuring
