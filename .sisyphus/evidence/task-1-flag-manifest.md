# gdrv Pre-Migration Flag Manifest

## Summary
- **Total Commands**: 140 cobra.Command definitions
- **Total Flags**: 261 flag registrations
- **Total Required Flags**: 17 (via MarkFlagRequired)
- **Persistent Flags**: 14 (inherited by all commands)
- **Domains**: 16 (files, auth, admin, permissions, sheets, docs, slides, folders, drives, changes, labels, chat, sync, config, activity, about)

---

## Persistent Flags (Root — Inherited by All Commands)

These 14 flags are registered on `rootCmd.PersistentFlags()` and inherited by all subcommands.

| Flag | Type | Default | Short | Required | Notes |
|------|------|---------|-------|----------|-------|
| `--profile` | string | "default" | | | Authentication profile to use |
| `--drive-id` | string | "" | | | Shared Drive ID to operate in |
| `--output` | string | "json" | | | Output format (json, table) |
| `--quiet` | bool | false | `-q` | | Suppress non-essential output |
| `--verbose` | bool | false | `-v` | | Enable verbose logging |
| `--debug` | bool | false | | | Enable debug output |
| `--strict` | bool | false | | | Convert warnings to errors |
| `--no-cache` | bool | false | | | Bypass path resolution cache |
| `--cache-ttl` | int | 300 | | | Path cache TTL in seconds |
| `--include-shared-with-me` | bool | false | | | Include shared-with-me items |
| `--config` | string | "" | | | Path to configuration file |
| `--log-file` | string | "" | | | Path to log file |
| `--dry-run` | bool | false | | | Show what would be done without making changes |
| `--force` | bool | false | `-f` | | Force operation without confirmation |
| `--yes` | bool | false | `-y` | | Answer yes to all prompts |
| `--json` | bool | false | | | Output in JSON format (alias for --output json) |

---

## Domain: files (14 commands)

### files list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum files to return per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--include-trashed` | bool | false | | | Include trashed files |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### files get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### files upload
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |
| `--name` | string | "" | | | File name |
| `--mime-type` | string | "" | | | MIME type |

### files download
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--output` | string | "" | | | Output path |
| `--mime-type` | string | "" | | | Export MIME type |
| `--doc` / `--doc-text` | bool | false | | | Export Google Docs as plain text |

### files delete
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--permanent` | bool | false | | | Permanently delete |
| `--force` | bool | false | | | Skip confirmation |

### files copy
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--name` | string | "" | | | New file name |
| `--parent` | string | "" | | | Destination folder ID |

### files move
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | ✓ New parent folder ID |

### files trash
(No local flags)

### files restore
(No local flags)

### files revisions
(No local flags)

### files revisions download
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--output` | string | "" | | | ✓ Output path for revision download |

### files revisions restore
(No local flags)

### files list-trashed
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum files to return per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### files export-formats
(No local flags)

---

## Domain: auth (8 commands)

### auth login
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--preset` | string | "" | | | Scope preset (workspace-basic, workspace-full, admin, workspace-with-admin, workspace-activity, workspace-labels, workspace-sync, workspace-complete) |
| `--scopes` | string | "" | | | Custom OAuth scopes (comma-separated) |
| `--no-browser` | bool | false | | | Disable browser-based auth (use manual code entry) |
| `--client-id` | string | "" | | | OAuth client ID |
| `--client-secret` | string | "" | | | OAuth client secret |
| `--wide` | bool | false | | | Use wide terminal output |
| `--profile` | string | "default" | | | Authentication profile name |

### auth device
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--preset` | string | "" | | | Scope preset |
| `--scopes` | string | "" | | | Custom OAuth scopes |
| `--client-id` | string | "" | | | OAuth client ID |
| `--client-secret` | string | "" | | | OAuth client secret |
| `--wide` | bool | false | | | Use wide terminal output |
| `--profile` | string | "default" | | | Authentication profile name |

### auth service-account
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--key-file` | string | "" | | | ✓ Path to service account JSON key file |
| `--preset` | string | "" | | | Scope preset |
| `--scopes` | string | "" | | | Custom OAuth scopes |
| `--impersonate-user` | string | "" | | | User email to impersonate (for domain-wide delegation) |
| `--profile` | string | "default" | | | Authentication profile name |

### auth status
(No local flags)

### auth logout
(No local flags)

### auth profiles
(No local flags)

### auth diagnose
(No local flags)

---

## Domain: admin (15 commands)

### admin users list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--domain` | string | "" | | | Domain name (or use --customer) |
| `--customer` | string | "" | | | Customer ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### admin users get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### admin users create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--given-name` | string | "" | | | ✓ User's given name |
| `--family-name` | string | "" | | | ✓ User's family name |
| `--password` | string | "" | | | ✓ Initial password |

### admin users update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--given-name` | string | "" | | | User's given name |
| `--family-name` | string | "" | | | User's family name |
| `--suspended` | bool | false | | | Suspend user |
| `--org-unit-path` | string | "" | | | Organization unit path |

### admin users delete
(No local flags)

### admin users suspend
(No local flags)

### admin users unsuspend
(No local flags)

### admin groups list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--domain` | string | "" | | | Domain name |
| `--customer` | string | "" | | | Customer ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### admin groups get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### admin groups create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--description` | string | "" | | | Group description |

### admin groups update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--name` | string | "" | | | Group name |
| `--description` | string | "" | | | Group description |

### admin groups delete
(No local flags)

### admin groups members list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--roles` | string | "" | | | Filter by roles (OWNER, MANAGER, MEMBER) |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### admin groups members add
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--role` | string | "MEMBER" | | | Member role (OWNER, MANAGER, MEMBER) |

### admin groups members remove
(No local flags)

---

## Domain: permissions (14 commands)

### permissions list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### permissions create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--type` | string | "" | | | ✓ Permission type (user, group, domain, anyone) |
| `--role` | string | "" | | | ✓ Permission role (reader, commenter, writer) |
| `--email` | string | "" | | | Email address (for user/group types) |
| `--domain` | string | "" | | | Domain (for domain type) |

### permissions update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--role` | string | "" | | | ✓ New permission role |

### permissions delete
(No local flags)

### permissions create-link
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--role` | string | "reader" | | | Link role (reader, commenter, writer) |

### permissions audit public
(No local flags)

### permissions audit external
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--internal-domain` | string | "" | | | ✓ Internal domain for external detection |

### permissions audit anyone-with-link
(No local flags)

### permissions audit user
(No local flags)

### permissions analyze
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--recursive` | bool | false | | | Include descendants |

### permissions report
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--internal-domain` | string | "" | | | Internal domain for reporting |

### permissions bulk remove-public
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--folder-id` | string | "" | | | ✓ Folder ID to process |

### permissions bulk update-role
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--folder-id` | string | "" | | | ✓ Folder ID to process |
| `--from-role` | string | "" | | | ✓ Source role |
| `--to-role` | string | "" | | | ✓ Target role |

### permissions search
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--email` | string | "" | | | Email address to search |
| `--role` | string | "" | | | Role to filter by |
| `--type` | string | "" | | | Permission type to filter by |

---

## Domain: sheets (10 commands)

### sheets list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### sheets create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |

### sheets get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### sheets batch-update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--requests` | string | "" | | | JSON requests (inline) |
| `--requests-file` | string | "" | | | Path to JSON requests file |

### sheets values get
(No local flags)

### sheets values update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--values` | string | "" | | | JSON values (inline) |
| `--values-file` | string | "" | | | Path to JSON values file |
| `--value-input-option` | string | "RAW" | | | Value input option (RAW, USER_ENTERED) |

### sheets values append
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--values` | string | "" | | | JSON values (inline) |
| `--values-file` | string | "" | | | Path to JSON values file |
| `--value-input-option` | string | "RAW" | | | Value input option (RAW, USER_ENTERED) |

### sheets values clear
(No local flags)

---

## Domain: docs (5 commands)

### docs list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### docs create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |

### docs get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### docs read
(No local flags)

### docs update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--requests` | string | "" | | | JSON requests (inline) |
| `--requests-file` | string | "" | | | Path to JSON requests file |

---

## Domain: slides (6 commands)

### slides list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |
| `--query` | string | "" | | | Search query |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--order-by` | string | "" | | | Sort order |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### slides create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--parent` | string | "" | | | Parent folder ID |

### slides get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### slides read
(No local flags)

### slides update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--requests` | string | "" | | | JSON requests (inline) |
| `--requests-file` | string | "" | | | Path to JSON requests file |

### slides replace
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--data` | string | "" | | | JSON replacement data (inline) |
| `--file` | string | "" | | | Path to JSON replacement file |

---

## Domain: folders (5 commands)

### folders create
(No local flags)

### folders list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### folders get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

### folders delete
(No local flags)

### folders move
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--recursive` | bool | false | | | Include descendants |

---

## Domain: drives (2 commands)

### drives list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--fields` | string | "" | | | Fields to return |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### drives get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Fields to return |

---

## Domain: changes (4 commands)

### changes start-page-token
(No local flags)

### changes list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--page-token` | string | "" | | | ✓ Page token to list changes from |
| `--drive-id` | string | "" | | | Shared Drive ID |
| `--include-removed` | bool | false | | | Include removed items |
| `--include-items-from-all-drives` | bool | false | | | Include items from all drives |
| `--include-permissions-for-view` | string | "" | | | Include permissions with published view |
| `--restrict-to-my-drive` | bool | false | | | Restrict to My Drive only |
| `--limit` | int | 100 | | | Maximum results per page |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### changes watch
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--page-token` | string | "" | | | ✓ Page token to watch from |
| `--webhook-url` | string | "" | | | ✓ Webhook URL for notifications |
| `--expiration` | string | "" | | | Webhook expiration time |

### changes stop
(No local flags)

---

## Domain: labels (8 commands)

### labels list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--view` | string | "" | | | Label view mode |
| `--customer` | string | "" | | | Customer ID |
| `--fields` | string | "" | | | Fields to return |

### labels get
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--view` | string | "" | | | Label view mode |
| `--fields` | string | "" | | | Fields to return |

### labels create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | Label field definitions |

### labels publish
(No local flags)

### labels disable
(No local flags)

### labels file list
(No local flags)

### labels file apply
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | JSON field values |

### labels file update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--fields` | string | "" | | | JSON field values |

### labels file remove
(No local flags)

---

## Domain: chat (14 commands)

### chat spaces list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### chat spaces get
(No local flags)

### chat spaces create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--display-name` | string | "" | | | Space display name |
| `--type` | string | "SPACE" | | | Space type (SPACE, GROUP_CHAT) |
| `--external-users` | bool | false | | | Allow external users |

### chat spaces delete
(No local flags)

### chat messages list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--filter` | string | "" | | | Message filter |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### chat messages get
(No local flags)

### chat messages create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--text` | string | "" | | | Message text |
| `--thread` | string | "" | | | Thread ID for replies |

### chat messages update
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--text` | string | "" | | | Updated message text |

### chat messages delete
(No local flags)

### chat members list
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--paginate` | bool | false | | | Automatically fetch all pages |

### chat members get
(No local flags)

### chat members create
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--email` | string | "" | | | Member email address |
| `--role` | string | "MEMBER" | | | Member role (MEMBER, MANAGER) |

### chat members delete
(No local flags)

---

## Domain: sync (6 commands)

### sync init
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--local-dir` | string | "" | | | Local directory path |
| `--remote-folder-id` | string | "" | | | Remote folder ID |
| `--sync-direction` | string | "bidirectional" | | | Sync direction (upload, download, bidirectional) |

### sync push
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--conflict-strategy` | string | "skip" | | | Conflict resolution (skip, overwrite, rename) |
| `--exclude-patterns` | string | "" | | | Patterns to exclude |
| `--include-patterns` | string | "" | | | Patterns to include |

### sync pull
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--conflict-strategy` | string | "skip" | | | Conflict resolution |
| `--exclude-patterns` | string | "" | | | Patterns to exclude |
| `--include-patterns` | string | "" | | | Patterns to include |

### sync status
(No local flags)

### sync list
(No local flags)

### sync remove
(No local flags)

---

## Domain: config (3 commands)

### config show
(No local flags)

### config set
(No local flags — takes key/value arguments)

### config reset
(No local flags)

---

## Domain: activity (1 command)

### activity query
| Flag | Type | Default | Short | Required |
|------|------|---------|-------|----------|
| `--file-id` | string | "" | | | File ID to query |
| `--folder-id` | string | "" | | | Folder ID to query |
| `--ancestor-name` | string | "" | | | Ancestor folder name |
| `--start-time` | string | "" | | | Start time (RFC3339) |
| `--end-time` | string | "" | | | End time (RFC3339) |
| `--action-types` | string | "" | | | Comma-separated action types |
| `--user` | string | "" | | | User email to filter |
| `--limit` | int | 100 | | | Maximum results per page |
| `--page-token` | string | "" | | | Page token for pagination |
| `--paginate` | bool | false | | | Automatically fetch all pages |

---

## Domain: about (1 command)

### about
(No local flags)

---

## Special Notes

### Dual Flag Names
- `--doc` and `--doc-text` are bound to the same boolean variable in `files download` command

### Shared Flag Variables
- `filesParentID` is used by: `files list`, `files upload`, `files copy`, `files move`
- `filesQuery` is used by: `files list`, `files list-trashed`
- `filesLimit` is used by: `files list`, `files list-trashed`
- `filesPageToken` is used by: `files list`, `files list-trashed`
- `filesOrderBy` is used by: `files list`, `files list-trashed`
- `filesFields` is used by: `files list`, `files list-trashed`
- `filesPaginate` is used by: `files list`, `files list-trashed`

### Command Aliases
- `perm` is an alias for `permissions` command

### Required Flags (17 total)
1. `files move --parent`
2. `files revisions download --output`
3. `auth service-account --key-file`
4. `admin users create --given-name`
5. `admin users create --family-name`
6. `admin users create --password`
7. `permissions create --type`
8. `permissions create --role`
9. `permissions update --role`
10. `permissions audit external --internal-domain`
11. `permissions bulk remove-public --folder-id`
12. `permissions bulk update-role --folder-id`
13. `permissions bulk update-role --from-role`
14. `permissions bulk update-role --to-role`
15. `changes list --page-token`
16. `changes watch --page-token`
17. `changes watch --webhook-url`

---

## JSON Output Envelope Structure

All commands with `--json` flag produce output with this structure:

```json
{
  "schemaVersion": "1.0",
  "traceId": "uuid-string",
  "command": "command.path",
  "data": { /* command-specific data */ },
  "warnings": [ /* optional warnings */ ],
  "errors": [ /* optional errors */ ]
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Unknown error |
| 2 | Authentication required |
| 3 | Invalid argument |
| 4 | Resource not found |
| 5 | Permission denied |
| 6 | Rate limited |

---

## Build Information

- **Build Command**: `go build -o bin/gdrv ./cmd/gdrv`
- **Version Injection**: Via ldflags at build time
- **Framework**: spf13/cobra (to be migrated to alecthomas/kong)
- **Go Version**: 1.21+ (inferred from project structure)

