# Google Drive CLI (gdrv) - Company Fork

A **fast**, **lightweight**, and **AI-agent friendly** CLI for Google Drive.

## About This Fork

**This is a company-specific fork of the upstream `gdrv` project.** It solves a specific deployment problem: **pre-bundled OAuth credentials for zero-configuration usage**.

### Why This Fork Exists

| Problem | Upstream Solution | This Fork's Solution |
|---------|-------------------|---------------------|
| OAuth setup complexity | Each user configures `GDRV_CLIENT_ID` env var | Credentials pre-bundled in official binary |
| AI agent friction | Agents need credential setup | Works immediately, zero config |
| Deployment overhead | Multiple steps + secrets management | Single binary, install and go |
| Onboarding friction | Technical OAuth configuration required | Just run `gdrv auth login` |

**Use case**: Companies and teams who want their members and AI agents to use Google Drive CLI without dealing with OAuth client setup.

### Build Types

| Build Type | How to Build | OAuth Credentials |
|------------|--------------|-------------------|
| **Official** | Download from GitHub Releases or `make build-official` | Bundled in binary |
| **Source** | `make build` or `go build` | Requires `GDRV_CLIENT_ID` env var |

Official builds are recommended for end users and AI agents. Source builds are for developers who want to customize the tool.

## Quick Start

### 1. Install

```bash
# Download the official binary for your platform from GitHub Releases
# https://github.com/MILCGroup/Google-Drive-CLI/releases

# Or on macOS with Homebrew
brew install milcgroup/gdrv/gdrv

# Or build from source (requires env vars - see below)
git clone https://github.com/MILCGroup/Google-Drive-CLI.git
cd Google-Drive-CLI
make build
sudo cp bin/gdrv /usr/local/bin/
```

### 2. Authenticate

```bash
# Login using the bundled company credentials
gdrv auth login --preset workspace-basic
```

That's it! No OAuth client setup required (official builds only).

## For AI Agents (Zero-Config Operation)

This CLI is designed for seamless AI agent integration:

```bash
# Just works - no env vars needed
gdrv files list --json
gdrv files upload report.pdf --json
gdrv folders create "Project X" --json
```

### Agent Best Practices

```bash
# Always use --json for machine-readable output
gdrv files list --json

# Use --paginate for complete results
gdrv files list --paginate --json

# Check exit codes programmatically
# 0 = Success, 2 = Auth required, 4 = Not found, etc.
```

## Authentication

### OAuth Login (Default)

Opens browser for authentication:

```bash
gdrv auth login --preset workspace-basic
```

**Headless/Server environments:**

```bash
# Device code flow (no browser needed)
gdrv auth device --preset workspace-basic

# Or manual mode
gdrv auth login --no-browser --preset workspace-basic
```

### Scope Presets

| Preset | Access Level |
|--------|--------------|
| `workspace-basic` | Read-only (recommended for agents) |
| `workspace-full` | Full read/write |
| `admin` | Admin SDK (users/groups) |
| `workspace-complete` | All APIs |

```bash
# Read-only access (safest for automation)
gdrv auth login --preset workspace-basic

# Full access when needed
gdrv auth login --preset workspace-full
```

### Service Account (For Admin Operations)

For Admin SDK operations, use a service account:

```bash
gdrv auth service-account \
  --key-file ./service-account-key.json \
  --impersonate-user admin@company.com \
  --preset admin
```

## Common Commands

### Files

```bash
# List files
gdrv files list --json

# Upload
gdrv files upload document.pdf --json

# Download
gdrv files download <file-id> --output doc.pdf

# Search
gdrv files list --query "name contains 'Report'" --json
```

### Folders

```bash
# Create folder
gdrv folders create "Q1 Reports" --json

# List folder contents
gdrv folders list <folder-id> --json
```

### Shared Drives

```bash
# List Shared Drives
gdrv drives list --json

# Access Shared Drive content
gdrv files list --drive-id <drive-id> --json
```

### Google Workspace

```bash
# Sheets
gdrv sheets list --json
gdrv sheets get <spreadsheet-id> --json
gdrv sheets values get <id> "Sheet1!A1:B10" --json

# Docs
gdrv docs list --json
gdrv docs read <doc-id> --json

# Slides
gdrv slides list --json
gdrv slides read <presentation-id> --json
```

### Admin SDK

```bash
# List users
gdrv admin users list --domain company.com --json

# List groups
gdrv admin groups list --domain company.com --json
```

## Advanced Features

### Drive Activity API

```bash
# Query recent activity
gdrv activity query --json

# Activity for specific file
gdrv activity query --file-id <id> --json
```

### Drive Labels API

```bash
# List labels
gdrv labels list --json

# Apply label to file
gdrv labels file apply <file-id> <label-id> --json
```

### Permission Auditing

```bash
# Find publicly shared files
gdrv permissions audit public --json

# External sharing audit
gdrv permissions audit external --internal-domain company.com --json
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication required |
| 3 | Invalid argument |
| 4 | Resource not found |
| 5 | Permission denied |
| 6 | Rate limited |

## Configuration

Optional configuration file locations:
- macOS: `~/Library/Application Support/gdrv/config.json`
- Linux: `~/.config/gdrv/config.json`
- Windows: `%APPDATA%\gdrv\config.json`

**Note:** OAuth credentials are pre-configured in official builds. No need to set `GDRV_CLIENT_ID`.

## Building from Source

### Source Build (Requires Env Vars)

```bash
# Set your OAuth credentials first
export GDRV_CLIENT_ID="your-client-id"
export GDRV_CLIENT_SECRET="your-secret"

# Build
make build
# Output: bin/gdrv
```

**Source builds do NOT include bundled OAuth credentials.** You'll need to provide your own via environment variables or use the official release binaries.

### Official Build (Bundled Credentials)

For CI/CD to create official builds with bundled OAuth credentials:

```bash
# Requires GDRV_CLIENT_ID and GDRV_CLIENT_SECRET in CI secrets
make build-official
```

The official build injects credentials via ldflags and marks the binary as "official" at build time.

### Build for All Platforms

```bash
make build-all
# Outputs: bin/gdrv-<os>-<arch>
```

### Custom OAuth Override (Optional)

To override bundled credentials in any build:

```bash
export GDRV_CLIENT_ID="your-client-id"
export GDRV_CLIENT_SECRET="your-secret"
make build
```

## Troubleshooting

### "official release build" error

This build does not have bundled OAuth credentials. 

**Solutions:**
1. Download official release from GitHub Releases
2. Or set env vars and build:
   ```bash
   export GDRV_CLIENT_ID="your-id"
   export GDRV_CLIENT_SECRET="your-secret"
   make build
   ```

### "Authentication required"

```bash
# Re-authenticate
gdrv auth logout
gdrv auth login --preset workspace-basic
```

### "Browser not opening"

```bash
# Use device code flow instead
gdrv auth device --preset workspace-basic
```

### Permission Denied

```bash
# Check auth status
gdrv auth status --json

# May need higher scope preset
gdrv auth login --preset workspace-full
```

## Security Notes

- Credentials are stored in system keyring (preferred) or encrypted file
- Tokens refresh automatically
- `gdrv auth logout` removes local credentials only
- All API communication is over HTTPS

## Original Project

This is a company-specific fork of [gdrv](https://github.com/dl-alexandre/Google-Drive-CLI) with pre-configured OAuth credentials for seamless deployment.

## License

MIT License - see LICENSE file for details.
