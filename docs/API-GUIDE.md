# API Guide

Complete reference for all gdrv commands organized by category.

## Table of Contents

- [File Operations](#file-operations)
- [Folder Operations](#folder-operations)
- [Permission Management](#permission-management)
- [Google Sheets](#google-sheets-operations)
- [Google Docs](#google-docs-operations)
- [Google Slides](#google-slides-operations)
- [Google Chat](#google-chat-api-operations)
- [Shared Drives](#shared-drives)
- [Admin SDK](#admin-sdk-operations)
- [Drive Activity API](#drive-activity-api-operations)
- [Drive Labels API](#drive-labels-api-operations)
- [Drive Changes API](#drive-changes-api-operations)
- [Permission Auditing](#permission-auditing-and-analysis)
- [Configuration](#configuration)

## File Operations

```bash
gdrv files upload <file>          # Upload file
gdrv files download <file-id>     # Download file
gdrv files list                   # List files
gdrv files delete <file-id>       # Delete file
gdrv files trash <file-id>        # Move to trash
gdrv files restore <file-id>      # Restore from trash
gdrv files revisions <file-id>    # List revisions
```

**Common Flags:**
- `--json`: Output as JSON
- `--quiet`: Suppress output
- `--dry-run`: Preview without executing
- `--output`: Specify download path
- `--doc`: Download Google Doc as text

## Folder Operations

```bash
gdrv folders create <name>        # Create folder
gdrv folders list <folder-id>     # List contents
gdrv folders delete <folder-id>   # Delete folder
gdrv folders move <id> <parent>   # Move folder
```

## Permission Management

```bash
gdrv permissions list <file-id>           # List permissions
gdrv permissions create <file-id> --type user --email user@example.com --role reader
gdrv permissions update <file-id> <perm-id> --role writer
gdrv permissions delete <file-id> <perm-id>
gdrv permissions public <file-id>         # Create public link
```

**Permission Roles:** `owner`, `writer`, `commenter`, `reader`
**Permission Types:** `user`, `group`, `domain`, `anyone`

## Google Sheets Operations

Manage Google Sheets spreadsheets with full read and write capabilities.

**Required OAuth Scopes:**
- Read operations: `https://www.googleapis.com/auth/spreadsheets.readonly` or `https://www.googleapis.com/auth/spreadsheets`
- Write operations: `https://www.googleapis.com/auth/spreadsheets`
- Use preset: `workspace-basic` (read-only) or `workspace-full` (read/write)

**API Documentation:** [Google Sheets API v4](https://developers.google.com/sheets/api)

### Commands

```bash
# List spreadsheets
gdrv sheets list                                # List all spreadsheets
gdrv sheets list --parent <folder-id>          # List spreadsheets in a folder
gdrv sheets list --query "name contains 'Report'" --json
gdrv sheets list --paginate --json            # Get all spreadsheets

# Create a spreadsheet
gdrv sheets create "My Spreadsheet"           # Create a new spreadsheet
gdrv sheets create "Budget 2026" --parent <folder-id> --json

# Get spreadsheet metadata
gdrv sheets get <spreadsheet-id>                # Get spreadsheet details
gdrv sheets get 1abc123... --json

# Read and write values
gdrv sheets values get <spreadsheet-id> <range> # Get values from a range
gdrv sheets values get 1abc123... "Sheet1!A1:B10" --json
gdrv sheets values update <spreadsheet-id> <range> # Update values in a range
gdrv sheets values update 1abc123... "Sheet1!A1:B2" --values '[[1,2],[3,4]]'
gdrv sheets values update 1abc123... "Sheet1!A1:B2" --values-file data.json
gdrv sheets values append <spreadsheet-id> <range> # Append values to a range
gdrv sheets values append 1abc123... "Sheet1!A1" --values '[[5,6]]' --value-input-option RAW
gdrv sheets values clear <spreadsheet-id> <range> # Clear values from a range

# Batch update spreadsheet
gdrv sheets batch-update <spreadsheet-id>       # Batch update spreadsheet
gdrv sheets batch-update 1abc123... --requests-file examples/sheets/batch-update.json
```

**Command Flags:**

- **List flags:** `--parent`, `--query`, `--limit`, `--page-token`, `--order-by`, `--fields`, `--paginate`
- **Create flags:** `--parent`
- **Batch-update flags:** `--requests`, `--requests-file`
- **Values update/append flags:** `--value-input-option` (RAW or USER_ENTERED), `--values`, `--values-file`

## Google Docs Operations

Manage Google Docs documents with content reading and batch update capabilities.

**Required OAuth Scopes:**
- Read operations: `https://www.googleapis.com/auth/documents.readonly` or `https://www.googleapis.com/auth/documents`
- Write operations: `https://www.googleapis.com/auth/documents`
- Use preset: `workspace-basic` (read-only) or `workspace-full` (read/write)

**API Documentation:** [Google Docs API v1](https://developers.google.com/docs/api)

### Commands

```bash
# List documents
gdrv docs list                                # List all documents
gdrv docs list --parent <folder-id>          # List documents in a folder
gdrv docs list --query "name contains 'Report'" --json
gdrv docs list --paginate --json            # Get all documents

# Create a document
gdrv docs create "My Document"               # Create a new document
gdrv docs create "Meeting Notes" --parent <folder-id> --json

# Get document metadata
gdrv docs get <document-id>                   # Get document details
gdrv docs get 1abc123... --json

# Read document content
gdrv docs read <document-id>                  # Extract plain text from document
gdrv docs read 1abc123...                     # Print plain text
gdrv docs read 1abc123... --json             # Get structured content

# Batch update document
gdrv docs update <document-id>                # Batch update document
gdrv docs update 1abc123... --requests-file updates.json
gdrv docs update 1abc123... --requests-file examples/docs/batch-update.json
```

**Command Flags:**

- **List flags:** `--parent`, `--query`, `--limit`, `--page-token`, `--order-by`, `--fields`, `--paginate`
- **Create flags:** `--parent`
- **Update flags:** `--requests`, `--requests-file`

## Google Slides Operations

Manage Google Slides presentations with content reading, batch updates, and text replacement (templating) capabilities.

**Required OAuth Scopes:**
- Read operations: `https://www.googleapis.com/auth/presentations.readonly` or `https://www.googleapis.com/auth/presentations`
- Write operations: `https://www.googleapis.com/auth/presentations`
- Use preset: `workspace-basic` (read-only) or `workspace-full` (read/write)

**API Documentation:** [Google Slides API v1](https://developers.google.com/slides/api)

### Commands

```bash
# List presentations
gdrv slides list                                # List all presentations
gdrv slides list --parent <folder-id>          # List presentations in a folder
gdrv slides list --query "name contains 'Deck'" --json
gdrv slides list --paginate --json            # Get all presentations

# Create a presentation
gdrv slides create "My Presentation"           # Create a new presentation
gdrv slides create "Q1 Review" --parent <folder-id> --json

# Get presentation metadata
gdrv slides get <presentation-id>               # Get presentation details
gdrv slides get 1abc123... --json

# Read presentation content
gdrv slides read <presentation-id>              # Extract text from all slides
gdrv slides read 1abc123...                     # Print text from all slides
gdrv slides read 1abc123... --json             # Get structured content

# Batch update presentation
gdrv slides update <presentation-id>            # Batch update presentation
gdrv slides update 1abc123... --requests-file updates.json
gdrv slides update 1abc123... --requests-file examples/slides/batch-update.json

# Replace text placeholders (templating)
gdrv slides replace <presentation-id>           # Replace text placeholders
gdrv slides replace 1abc123... --data '{"{{NAME}}":"Alice","{{DATE}}":"2026-01-24"}'
gdrv slides replace 1abc123... --file examples/slides/replacements.json
```

**Command Flags:**

- **List flags:** `--parent`, `--query`, `--limit`, `--page-token`, `--order-by`, `--fields`, `--paginate`
- **Create flags:** `--parent`
- **Update flags:** `--requests`, `--requests-file`
- **Replace flags:** `--data` (JSON string), `--file` (JSON file path)

## Google Chat API Operations

Manage Google Chat spaces, messages, and members through the Google Chat API.

**Required OAuth Scopes:**
- Read operations: `https://www.googleapis.com/auth/chat.readonly`
- Write operations: `https://www.googleapis.com/auth/chat`

**API Documentation:** [Google Chat API](https://developers.google.com/workspace/chat/api/reference/rest)

### Spaces Management

```bash
# List all spaces you have access to
gdrv chat spaces list --json
gdrv chat spaces list --paginate --json

# Get details about a specific space
gdrv chat spaces get <space-id> --json

# Create a new space
gdrv chat spaces create --display-name "Team Chat" --type SPACE --json
gdrv chat spaces create --display-name "Project Group" --type GROUP_CHAT --external-users --json

# Delete a space
gdrv chat spaces delete <space-id>
```

**Spaces Command Flags:**
- **List flags:** `--limit`, `--page-token`, `--paginate`
- **Create flags:** `--display-name` (required), `--type` (SPACE or GROUP_CHAT), `--external-users`

### Messages Management

```bash
# List messages in a space
gdrv chat messages list <space-id> --json
gdrv chat messages list <space-id> --limit 50 --paginate --json

# Get a specific message
gdrv chat messages get <space-id> <message-id> --json

# Create a message in a space
gdrv chat messages create <space-id> --text "Hello everyone!" --json
gdrv chat messages create <space-id> --text "Reply in thread" --thread <thread-id> --json

# Update a message
gdrv chat messages update <space-id> <message-id> --text "Updated message text" --json

# Delete a message
gdrv chat messages delete <space-id> <message-id>
```

**Messages Command Flags:**
- **List flags:** `--limit`, `--page-token`, `--filter`, `--paginate`
- **Create flags:** `--text` (required), `--thread`
- **Update flags:** `--text` (required)

### Members Management

```bash
# List members of a space
gdrv chat members list <space-id> --json
gdrv chat members list <space-id> --paginate --json

# Get member details
gdrv chat members get <space-id> <member-id> --json

# Add a member to a space
gdrv chat members create <space-id> --email user@example.com --role MEMBER --json
gdrv chat members create <space-id> --email manager@example.com --role MANAGER --json

# Remove a member from a space
gdrv chat members delete <space-id> <member-id>
```

**Members Command Flags:**
- **List flags:** `--limit`, `--page-token`, `--paginate`
- **Create flags:** `--email` (required), `--role` (MEMBER or MANAGER, default: MEMBER)

## Shared Drives

```bash
gdrv drives list                 # List Shared Drives
gdrv drives get <drive-id>       # Get drive details
```

## Admin SDK Operations

Manage Google Workspace users and groups through the Admin SDK Directory API.

**⚠️ Important Authentication Requirements:**

Admin SDK operations **require service account authentication** with domain-wide delegation enabled. This is different from regular OAuth authentication.

**Required Setup:**
1. Create a service account in Google Cloud Console
2. Enable domain-wide delegation for the service account
3. Authorize the required scopes in Google Workspace Admin Console
4. Download the service account JSON key file
5. Authenticate using the service account with user impersonation

**Required OAuth Scopes:**
- User management: `https://www.googleapis.com/auth/admin.directory.user`
- Group management: `https://www.googleapis.com/auth/admin.directory.group`
- Use preset: `admin` or `workspace-with-admin`

**API Documentation:** 
- [Admin SDK Directory API - Users](https://developers.google.com/admin-sdk/directory/reference/rest/v1/users)
- [Admin SDK Directory API - Groups](https://developers.google.com/admin-sdk/directory/reference/rest/v1/groups)

**Authentication Example:**

```bash
# Authenticate with service account and impersonate admin user
gdrv auth service-account \
  --key-file ./service-account-key.json \
  --impersonate-user admin@example.com \
  --preset admin
```

### User Management

```bash
# List users
gdrv admin users list --domain example.com
gdrv admin users list --domain example.com --json
gdrv admin users list --domain example.com --paginate --json
gdrv admin users list --domain example.com --query "name:John" --json

# Get user details
gdrv admin users get user@example.com
gdrv admin users get user@example.com --fields "id,name,email" --json

# Create a user
gdrv admin users create newuser@example.com \
  --given-name "John" \
  --family-name "Doe" \
  --password "TempPass123!"

# Update a user
gdrv admin users update user@example.com --given-name "Jane" --family-name "Smith"
gdrv admin users update user@example.com --suspended true
gdrv admin users update user@example.com --org-unit-path "/Engineering/Developers"

# Suspend/unsuspend a user
gdrv admin users suspend user@example.com
gdrv admin users unsuspend user@example.com

# Delete a user
gdrv admin users delete user@example.com
```

**User Command Flags:**
- **List flags:** `--domain` or `--customer`, `--query`, `--limit`, `--page-token`, `--order-by`, `--fields`, `--paginate`
- **Get flags:** `--fields`
- **Create flags:** `--given-name` (required), `--family-name` (required), `--password` (required)
- **Update flags:** `--given-name`, `--family-name`, `--suspended` (true/false), `--org-unit-path`

### Group Management

```bash
# List groups
gdrv admin groups list --domain example.com
gdrv admin groups list --domain example.com --json
gdrv admin groups list --domain example.com --paginate --json
gdrv admin groups list --domain example.com --query "name:Team" --json

# Get group details
gdrv admin groups get group@example.com
gdrv admin groups get group@example.com --fields "id,name,email" --json

# Create a group
gdrv admin groups create group@example.com "Team Group" \
  --description "Team access group"

# Update a group
gdrv admin groups update group@example.com --name "New Name"
gdrv admin groups update group@example.com --description "Updated description"

# Delete a group
gdrv admin groups delete group@example.com
```

**Group Command Flags:**
- **List flags:** `--domain` or `--customer`, `--query`, `--limit`, `--page-token`, `--order-by`, `--fields`, `--paginate`
- **Get flags:** `--fields`
- **Create flags:** `--description`
- **Update flags:** `--name`, `--description`

### Group Membership Management

```bash
# List group members
gdrv admin groups members list team@example.com
gdrv admin groups members list team@example.com --json
gdrv admin groups members list team@example.com --roles MANAGER --json
gdrv admin groups members list team@example.com --paginate --json

# Add member to group
gdrv admin groups members add team@example.com user@example.com --role MEMBER
gdrv admin groups members add team@example.com user@example.com --role MANAGER
gdrv admin groups members add team@example.com user@example.com --role OWNER

# Remove member from group
gdrv admin groups members remove team@example.com user@example.com
```

**Group Members Command Flags:**
- **List flags:** `--limit`, `--page-token`, `--roles` (OWNER, MANAGER, MEMBER), `--fields`, `--paginate`
- **Add flags:** `--role` (OWNER, MANAGER, or MEMBER, default: MEMBER)

## Drive Activity API Operations

Query and monitor file and folder activity across Google Drive with detailed activity streams.

**Required OAuth Scopes:**
- Read-only: `https://www.googleapis.com/auth/drive.activity.readonly`
- Full access: `https://www.googleapis.com/auth/drive.activity`
- Use preset: `workspace-activity` (read-only) or `workspace-complete`

**API Documentation:** [Drive Activity API v2](https://developers.google.com/drive/activity/v2)

```bash
# Query recent activity for all accessible files
gdrv activity query --json

# Query activity for a specific file
gdrv activity query --file-id 1abc123... --json

# Query activity within a time range
gdrv activity query --start-time "2026-01-01T00:00:00Z" --end-time "2026-01-31T23:59:59Z" --json

# Query activity for a folder (including descendants)
gdrv activity query --folder-id 0ABC123... --json

# Filter by activity types
gdrv activity query --action-types "edit,share,permission_change" --json

# Get activity for a specific user
gdrv activity query --user user@example.com --json

# Paginate through activity results
gdrv activity query --limit 100 --page-token "TOKEN" --json
```

## Drive Labels API Operations

Manage Drive labels and apply structured metadata to files and folders.

**Required OAuth Scopes:**
- Read-only: `https://www.googleapis.com/auth/drive.labels.readonly`
- Full access: `https://www.googleapis.com/auth/drive.labels`
- Admin: `https://www.googleapis.com/auth/drive.admin.labels`
- Use preset: `workspace-labels` or `workspace-complete`

**API Documentation:** [Drive Labels API v2](https://developers.google.com/drive/labels/overview)

```bash
# List available labels
gdrv labels list --json

# Get label schema
gdrv labels get <label-id> --json

# List labels applied to a file
gdrv labels file list <file-id> --json

# Apply a label to a file
gdrv labels file apply <file-id> <label-id> \
  --fields '{"field1":"value1","field2":"value2"}' --json

# Update label fields on a file
gdrv labels file update <file-id> <label-id> \
  --fields '{"field1":"new_value"}' --json

# Remove a label from a file
gdrv labels file remove <file-id> <label-id>

# Create a label (admin only)
gdrv labels create "Document Type" --json

# Publish a label (admin only)
gdrv labels publish <label-id>

# Disable a label (admin only)
gdrv labels disable <label-id>
```

## Drive Changes API Operations

Track changes to files and folders for real-time synchronization and automation.

**Required OAuth Scopes:**
- Uses standard Drive scopes (no additional scopes needed)
- Use preset: `workspace-sync` or `workspace-complete`

**API Documentation:** [Drive Changes API v3](https://developers.google.com/drive/api/v3/reference/changes)

```bash
# Get the starting page token
gdrv changes start-page-token --json

# List changes since a page token
gdrv changes list --page-token "12345" --json

# List changes with auto-pagination
gdrv changes list --page-token "12345" --paginate --json

# List changes for a specific Shared Drive
gdrv changes list --page-token "12345" --drive-id <drive-id> --json

# Watch for changes (webhook setup)
gdrv changes watch --page-token "12345" \
  --webhook-url "https://example.com/webhook" --json

# Stop watching for changes
gdrv changes stop <channel-id> <resource-id>

# List changes including removed files
gdrv changes list --page-token "12345" --include-removed --json
```

## Permission Auditing and Analysis

Enhanced permission auditing and access analysis tools for security and compliance.

**Required OAuth Scopes:**
- Uses standard Drive scopes (no additional scopes needed)

```bash
# Audit all files with public access
gdrv permissions audit public --json

# Audit all files shared with external domains
gdrv permissions audit external --internal-domain example.com --json

# Audit permissions for a specific user
gdrv permissions audit user user@example.com --json

# Find files with "anyone with link" access
gdrv permissions audit anyone-with-link --json

# Analyze permission inheritance for a folder
gdrv permissions analyze <folder-id> --recursive --json

# Generate permission report for a file/folder
gdrv permissions report <file-id> --internal-domain example.com --json

# Bulk remove public access (dry-run first)
gdrv permissions bulk remove-public --folder-id <folder-id> --dry-run --json

# Bulk change role (e.g., downgrade all "writer" to "reader")
gdrv permissions bulk update-role --folder-id <folder-id> \
  --from-role writer --to-role reader --dry-run --json

# Find files accessible by a specific email
gdrv permissions search --email user@example.com --json

# List all files with "commenter" access
gdrv permissions search --role commenter --json
```

## Configuration

```bash
gdrv config show                 # Show current config
gdrv config set <key> <value>    # Set config value
gdrv config reset                # Reset to defaults
```

Config file defaults:
- macOS: `~/Library/Application Support/gdrv/config.json`
- Linux: `~/.config/gdrv/config.json`
- Windows: `%APPDATA%\\gdrv\\config.json`
- Override with `GDRV_CONFIG_DIR`

OAuth client fields in config (optional):
- `oauthClientId`
- `oauthClientSecret` (only if required by your client type)

## Other Commands

```bash
gdrv auth login [--preset <preset>] [--wide] [--scopes <scopes>] [--no-browser] [--client-id <id>] [--client-secret <secret>] [--profile <name>]
gdrv auth device [--preset <preset>] [--wide] [--client-id <id>] [--client-secret <secret>] [--profile <name>]
gdrv auth service-account --key-file <file> [--preset <preset>] [--scopes <scopes>] [--impersonate-user <email>] [--profile <name>]
gdrv auth status                 # Show auth status
gdrv auth profiles               # Manage profiles
gdrv auth logout                 # Clear credentials
gdrv about                       # Show API capabilities
```

## Output Formats

### Table Format (Default)
```bash
gdrv files list
```

### JSON Format
```bash
gdrv files list --json
```

### Quiet Mode
```bash
gdrv files upload file.txt --quiet
```

## Safety Controls

### Dry Run (Preview)
Preview what would happen without executing:
```bash
gdrv files delete 123 --dry-run
```

### Default Behavior (Non-Interactive)
By default, commands execute without prompts for agent-friendliness:
```bash
# Executes immediately (no confirmation prompt)
gdrv files delete 123
```

### Interactive Mode
For interactive use, the CLI will prompt for confirmation when safety checks require it.
