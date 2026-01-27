# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`gdrv` is a production-grade Google Drive CLI written in Go. It provides fast, lightweight automation for Google Drive with emphasis on AI agent-friendliness (JSON output, explicit flags, clean exit codes). The tool supports complete Drive integration, Google Workspace APIs (Sheets, Docs, Slides), Admin SDK, and Shared Drives.

## Common Commands

### Build and Run
```bash
# Build for current platform
make build
# Output: bin/gdrv

# Build for all platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
make build-all

# Build and run with arguments
make run ARGS='version'
make run ARGS='files list --json'

# Quick build for development
go build -o bin/gdrv ./cmd/gdrv
```

### Testing
```bash
# Run all unit tests with race detection
go test -v -race -cover ./...

# Run tests with coverage report
make test-coverage
# Generates: coverage.out, coverage.html

# Run specific package tests
go test -v ./internal/files/...

# Run specific test
go test -v -run TestManagerUpload ./internal/files/...

# Run integration tests (requires credentials)
go test -tags=integration ./test/integration/...
```

### Linting
```bash
# Run golangci-lint (must be installed)
make lint

# Run go vet only
go vet ./...
```

### Dependencies
```bash
# Download dependencies
go mod download
# or
make deps

# Tidy modules
go mod tidy
# or
make tidy

# Verify dependencies
go mod verify
```

### Other
```bash
# Clean build artifacts
make clean

# Show version info
make version

# Generate checksums for binaries
make checksums
```

## High-Level Architecture

### Core Design Patterns

1. **Manager Pattern**: Each domain (files, folders, auth, sheets, docs, slides, admin, permissions, drives, revisions) has a Manager type in `internal/<domain>/manager.go` that encapsulates business logic and orchestrates API calls.

2. **API Client Layer**: `internal/api/client.go` provides a unified API client with built-in retry logic, exponential backoff, rate limit handling, and request shaping. All API calls flow through `ExecuteWithRetry[T]()` for consistent error handling.

3. **Request Context**: Every API operation uses `types.RequestContext` (with TraceID) for structured logging and request correlation across the call stack.

4. **CLI Command Structure**: Commands are organized in `internal/cli/` with separate files per domain (files.go, folders.go, sheets.go, docs.go, slides.go, admin.go, etc.). Uses spf13/cobra for command parsing.

### Critical Components

#### Authentication (`internal/auth/`)
- **Manager**: Handles OAuth2, device code flow, and service account authentication
- **Storage Backends**: Three-tier storage (keyring → encrypted file → plain file) for credential storage
- **Scope Presets**: Pre-configured scope sets (workspace-basic, workspace-full, admin, workspace-with-admin)
- **Profile Support**: Multiple authentication profiles stored with profile-specific keys
- **Service Account**: Supports domain-wide delegation with user impersonation for Admin SDK

#### API Client (`internal/api/`)
- **client.go**: Core API client with retry logic (configurable max retries, exponential backoff with jitter)
- **operations.go**: API operation wrappers for files, folders, permissions
- **request_shaper.go**: Configures API requests based on context (drive ID, fields, resource keys)
- **resource_keys.go**: Manages resource keys for Shared Drive items

#### Safety Layer (`internal/safety/`)
- **Dry-Run Mode**: Records planned operations without execution
- **Idempotency**: Tracks operation keys to prevent duplicate operations
- **Confirmation Prompts**: Interactive confirmation for destructive operations

#### Path Resolution (`internal/resolver/`)
- **PathResolver**: Resolves file paths to IDs with caching (configurable TTL)
- **Handles edge cases**: Shared Drives, multiple items with same name, "Shared with me" items

#### Logging (`internal/logging/`)
- **Multi-logger pattern**: Console + file logging with trace correlation
- **Structured logging**: Supports trace IDs, key-value fields, sensitive data redaction
- **Debug mode**: Detailed operation logging when `--debug` is enabled

#### Type System (`internal/types/`)
- **types.go**: Core domain types (File, Folder)
- **cli.go**: GlobalFlags and CLI configuration
- **api.go**: RequestContext and API types
- **output.go**: OutputFormat handling
- **sheets.go, docs.go, slides.go, admin.go**: Domain-specific types

### API Flow Pattern

All API operations follow this pattern:

```
CLI Command → Manager Method → API Client (with RequestContext) → ExecuteWithRetry → Google API
                                                                            ↓
                                                                   Logging + Error Classification
```

Example flow for file upload:
1. `internal/cli/files.go`: filesUploadCmd parses flags, creates Manager
2. `internal/files/manager.go`: Manager.Upload() prepares request
3. `internal/api/client.go`: ExecuteWithRetry() handles retries
4. Error classification via `internal/errors/errors.go`
5. Result returned up the stack with proper exit codes

### Testing Strategy

- **Unit tests**: `*_test.go` files alongside implementation (e.g., `manager_test.go`)
- **Property tests**: `*_property_test.go` for invariant testing
- **Integration tests**: `test/integration/*_test.go` with real API calls (requires auth)
- **Mock strategy**: Tests use interfaces and dependency injection

### Exit Codes

Defined in `internal/utils/constants.go`:
- 0: Success
- 1: Unknown error
- 2: Authentication required
- 3: Invalid argument
- 4: Resource not found
- 5: Permission denied
- 6: Rate limited

## Key Implementation Notes

### When Adding New Commands

1. Define command in `internal/cli/<domain>.go` using cobra.Command
2. Create or extend Manager in `internal/<domain>/manager.go`
3. Add types to `internal/types/<domain>.go` if needed
4. Use `api.ExecuteWithRetry()` for all API calls
5. Create `types.RequestContext` with appropriate RequestType
6. Add `--json` flag support for all commands
7. Return proper exit codes using `internal/utils/` constants
8. Add tests in `internal/<domain>/manager_test.go`

### Authentication Context

- OAuth credentials required: Set `GDRV_CLIENT_ID` and `GDRV_CLIENT_SECRET`
- Service accounts: Use `--key-file` and `--impersonate-user` for Admin SDK
- Profiles stored in `~/.config/gdrv/` (or `GDRV_CONFIG_DIR`)
- Token refresh happens automatically when within 5 minutes of expiration

### Working with Shared Drives

- Always pass `--drive-id` flag or set it globally
- Resource keys automatically managed by `ResourceKeyManager`
- Use `supportsAllDrives=true` for all Drive API calls (handled by RequestShaper)

### Logging Best Practices

- Always use the logger from `cli.GetLogger()`
- Attach trace IDs for request correlation: `logger.WithTraceID(reqCtx.TraceID)`
- Use structured fields: `logging.F("key", value)`
- Redact sensitive data (tokens, credentials) automatically

### Google Workspace APIs

- **Sheets**: `internal/sheets/manager.go` wraps Sheets API v4
- **Docs**: `internal/docs/manager.go` wraps Docs API v1
- **Slides**: `internal/slides/manager.go` wraps Slides API v1
- **Admin SDK**: `internal/admin/manager.go` wraps Admin Directory API v1

All Workspace APIs require appropriate OAuth scopes. Use scope presets for convenience.

### Error Handling

Use `internal/errors/errors.go` for consistent error classification:
- `ClassifyGoogleAPIError()`: Converts Google API errors to CLI errors with proper exit codes
- `NewCLIError()`: Creates structured CLI errors
- `AppError`: Wraps errors with additional context

### Version Information

Version metadata injected at build time via ldflags (see Makefile):
- `pkg/version.Version`: Git tag or "dev"
- `pkg/version.GitCommit`: Git commit SHA
- `pkg/version.BuildTime`: Build timestamp
