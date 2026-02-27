# AI Agent Best Practices

This CLI is designed to be used by AI agents and automation scripts.

## Quick Reference

```bash
# Always use --json for machine-readable output
gdrv files list --json

# Use --paginate to get all results automatically
gdrv files list --paginate --json

# Check exit codes for error handling
# 0 = Success, 1 = General error, 2 = Auth required, etc.
```

## JSON Output

Always use `--json` for machine-readable output:

```bash
# List files as JSON
gdrv files list --json

# Get file metadata
gdrv files get 1abc123... --json

# Upload returns the created file object
gdrv files upload report.pdf --json
```

## Pagination Control

Use `--paginate` to automatically fetch all pages:

```bash
# Get ALL files (auto-pagination)
gdrv files list --paginate --json

# Get all trashed files
gdrv files list-trashed --paginate --json

# Get all Shared Drives
gdrv drives list --paginate --json
```

Or control pagination manually:

```bash
# Get first page
gdrv files list --limit 100 --json

# Use nextPageToken from response for next page
gdrv files list --limit 100 --page-token "TOKEN_FROM_PREVIOUS" --json
```

## Sorting and Filtering

```bash
# Sort by modified time (newest first)
gdrv files list --order-by "modifiedTime desc" --json

# Search by name
gdrv files list --query "name contains 'report'" --json

# Combined: recent PDFs
gdrv files list --query "mimeType = 'application/pdf'" --order-by "modifiedTime desc" --json
```

## Non-Interactive Mode

Destructive commands run without prompts by default. Use `--dry-run` to preview:

```bash
# Preview what would be deleted
gdrv files delete 1abc123... --dry-run

# Actually delete (no prompt)
gdrv files delete 1abc123...

# Permanently delete (bypasses trash)
gdrv files delete 1abc123... --permanent
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

## Best Practices Checklist

1. **Always use `--json`** - Parse structured output, not human-readable tables
2. **Use `--paginate`** - Don't miss items due to pagination limits
3. **Check exit codes** - Handle errors programmatically
4. **Use file IDs** - More reliable than paths for Shared Drives
5. **Use `--dry-run`** - Preview destructive operations before executing

## Environment Variables

```bash
# Use a specific profile
export GDRV_PROFILE=work

# Custom config directory
export GDRV_CONFIG_DIR=/path/to/config

# Force custom OAuth (CI/contributor policy)
export GDRV_REQUIRE_CUSTOM_OAUTH=1

# Disable browser auto-open
export GDRV_NO_BROWSER=1
```
