# Authentication Guide

The CLI supports multiple authentication methods and scope presets. OAuth uses Authorization Code + PKCE for installed/desktop apps (public client). The default client is ID-only (no embedded secret).

## OAuth Client Sources and Precedence

Credentials are resolved in this order:
1. CLI flags (`--client-id`, `--client-secret`)
2. Environment variables (`GDRV_CLIENT_ID`, `GDRV_CLIENT_SECRET`)
3. Config file (`oauthClientId`, `oauthClientSecret`)
4. Default public OAuth client (embedded ID, no secret)

No partial overrides: if any OAuth client variable is set, all required OAuth client fields must be set (client ID always; secret only if your client type requires it).

**Config file path (defaults):**
- macOS: `~/Library/Application Support/gdrv/config.json`
- Linux: `~/.config/gdrv/config.json`
- Windows: `%APPDATA%\\gdrv\\config.json`
- Override with `GDRV_CONFIG_DIR`

**Contributor/CI policy:** set `GDRV_REQUIRE_CUSTOM_OAUTH=1` to refuse default credentials.

Default public client IDs may rotate between releases. If you see `invalid_client` errors with the default client, upgrade or configure a custom client.

### Default OAuth Client Notes

- The default public client is ID-only; any secret provided is treated as public (PKCE is used).
- The shared client is hosted in a dedicated Google Cloud project with quota monitoring and a rotation plan.
- If the shared client is disabled or rotated, the CLI will instruct you to upgrade or configure a custom client.

## Token Storage

- Preferred: system keyring (Keychain / Secret Service / Credential Manager).
- Fallback: encrypted file storage at `.../credentials/<profile>.enc` with `0600` permissions and a local key file at `.../.keyfile`.
- Plain file storage is development-only and must be explicitly forced.
- `gdrv auth logout` removes local credentials only (does not revoke remote consent).

## Authentication Methods

### OAuth2 Flow (Recommended)

```bash
gdrv auth login
```

Opens a browser for authentication using a loopback redirect on `127.0.0.1` with an ephemeral port.
If the browser cannot be opened, gdrv will fall back to manual code entry.

**Manual fallback (headless):**
```bash
gdrv auth login --no-browser
```

Prompts for a manual code paste after you approve access in a browser. This fallback is used for headless environments, browser launch failures, or loopback binding issues. You can force it with `--no-browser` or `GDRV_NO_BROWSER=1`.

### Device Code Flow (Headless)

```bash
gdrv auth device
```

Displays a code to enter at https://www.google.com/device.

### Service Account Authentication

```bash
gdrv auth service-account --key-file ./service-account-key.json --preset workspace-basic
```

Loads credentials from a service account JSON key file. Use `--impersonate-user` for Admin SDK scopes.

## Scope Presets

| Preset | Description | Use Case |
|--------|-------------|----------|
| `workspace-basic` | Read-only Drive, Sheets, Docs, Slides, Labels | Viewing and downloading |
| `workspace-full` | Full Drive, Sheets, Docs, Slides, Labels | Editing and management |
| `admin` | Admin Directory users and groups, Admin Labels | User/group/label administration |
| `workspace-with-admin` | Workspace full + Admin Directory + Admin Labels | Full workspace + admin |
| `workspace-activity` | Workspace basic + Activity API | Read-only with activity auditing |
| `workspace-labels` | Workspace full + Labels API | Full access with label management |
| `workspace-sync` | Workspace full + Changes API | Full access with change tracking |
| `workspace-complete` | All Workspace APIs + Activity + Labels + Changes | Complete API access |

Use `workspace-basic` for least-privilege read-only access; use `workspace-full` only when write access is required. Use the specialized presets (`workspace-activity`, `workspace-labels`, `workspace-sync`, `workspace-complete`) when you need the advanced APIs.

```bash
# Basic presets
gdrv auth login --preset workspace-basic
gdrv auth login --preset workspace-full
gdrv auth login --preset admin
gdrv auth login --preset workspace-with-admin

# Advanced API presets
gdrv auth login --preset workspace-activity
gdrv auth login --preset workspace-labels
gdrv auth login --preset workspace-sync
gdrv auth login --preset workspace-complete

# Device code flow
gdrv auth device --preset workspace-basic

# Service account
gdrv auth service-account --key-file ./key.json --preset workspace-complete
```

## Custom Scopes

```bash
gdrv auth login --scopes "https://www.googleapis.com/auth/drive.file,https://www.googleapis.com/auth/spreadsheets.readonly"
gdrv auth service-account --key-file ./key.json --scopes "https://www.googleapis.com/auth/drive.file"
```

### Available Scopes

**Drive Scopes:**
- `https://www.googleapis.com/auth/drive` - Full Drive access
- `https://www.googleapis.com/auth/drive.file` - Per-file access
- `https://www.googleapis.com/auth/drive.readonly` - Read-only Drive access
- `https://www.googleapis.com/auth/drive.metadata.readonly` - Read-only metadata

**Workspace Scopes:**
- `https://www.googleapis.com/auth/spreadsheets` - Full Sheets access
- `https://www.googleapis.com/auth/spreadsheets.readonly` - Read-only Sheets
- `https://www.googleapis.com/auth/documents` - Full Docs access
- `https://www.googleapis.com/auth/documents.readonly` - Read-only Docs
- `https://www.googleapis.com/auth/presentations` - Full Slides access
- `https://www.googleapis.com/auth/presentations.readonly` - Read-only Slides

**Admin SDK Scopes:**
- `https://www.googleapis.com/auth/admin.directory.user` - User management
- `https://www.googleapis.com/auth/admin.directory.user.readonly` - Read-only users
- `https://www.googleapis.com/auth/admin.directory.group` - Group management
- `https://www.googleapis.com/auth/admin.directory.group.readonly` - Read-only groups

**Advanced API Scopes:**
- `https://www.googleapis.com/auth/drive.activity` - Full Activity API access
- `https://www.googleapis.com/auth/drive.activity.readonly` - Read-only Activity
- `https://www.googleapis.com/auth/drive.labels` - Full Labels access
- `https://www.googleapis.com/auth/drive.labels.readonly` - Read-only Labels
- `https://www.googleapis.com/auth/drive.admin.labels` - Admin label management
- `https://www.googleapis.com/auth/drive.admin.labels.readonly` - Read-only admin labels

## Multiple Profiles

```bash
# Create and switch profiles
gdrv auth login --profile work
gdrv auth login --profile personal

# Use specific profile
gdrv --profile work files list
```

## OAuth Testing-Mode Limits

If your OAuth consent screen is in testing mode, refresh tokens expire after 7 days and Google enforces a 100 refresh-token issuance cap per client. If you see repeated `invalid_grant` errors, re-authenticate and revoke unused tokens in Google Cloud Console or move the app to production to avoid the testing-mode limits.

## Custom OAuth Client Prerequisites

If you want to use your own OAuth client:
1. Create a project in Google Cloud Console
2. Enable the Google Drive API
3. Create OAuth 2.0 credentials (Desktop application)
4. Set credentials via environment variables or command flags:

```bash
export GDRV_CLIENT_ID="your-client-id"
export GDRV_CLIENT_SECRET="your-client-secret" # only if required by your client type

gdrv auth login --client-id "your-client-id" --client-secret "your-client-secret"
```
