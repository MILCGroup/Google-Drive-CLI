## Summary

Add permission auditing and bulk operations to `gog drive` for security compliance and bulk permission management.

## Current State

`gog drive` has `permissions`, `share`, and `unshare` commands for individual file operations. No capabilities exist for auditing access patterns or bulk permission changes across folders.

## Proposed Commands

```bash
gog drive audit public                           # Find files with public access
gog drive audit external --domain example.com    # Find externally shared files
gog drive audit user user@example.com            # Audit a specific user's access
gog drive bulk remove-public --parent <folderId> --dry-run  # Bulk remove public links
gog drive bulk update-role --parent <folderId> --from writer --to reader --dry-run
```

## Use Cases

- Security compliance audits (find public/external shares)
- Bulk permission cleanup when employees leave
- Risk assessment: identify overshared files
- Downgrade permissions en masse (e.g., writers → readers)

## API Notes

- Uses existing Drive Permissions API (no additional scopes)
- Bulk operations respect existing `--dry-run` and `confirmDestructive()` patterns
- Could benefit from idempotency tracking for safe resume of interrupted operations

## Related

- #291 — confirmDestructive guards for dangerous operations

---

I've implemented this in another Drive CLI and can share implementation details if helpful.
