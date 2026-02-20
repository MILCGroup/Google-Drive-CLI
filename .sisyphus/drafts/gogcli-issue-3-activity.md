## Summary

Add Google Drive Activity API v2 support for audit trails, compliance monitoring, and security investigations.

## Current State

`gog drive` has no activity tracking. This is the Drive Activity API (distinct from Gmail History API / `gmail watch serve`).

## Proposed Commands

```bash
gog drive activity query                          # Recent activity across Drive
gog drive activity query --file <fileId>          # Activity for specific file
gog drive activity query --folder <folderId>      # Activity for folder tree
gog drive activity query --user user@example.com  # Activity by user
gog drive activity query --actions edit,share --from 2026-01-01T00:00:00Z
```

## Use Cases

- Compliance auditing: who accessed what and when
- Security monitoring: detect unauthorized access patterns
- Incident investigation: trace file lifecycle events
- Access tracking: monitor sensitive document interactions

## API Notes

- API: https://developers.google.com/drive/activity/v2
- Additional scope: `https://www.googleapis.com/auth/drive.activity.readonly`
- Supports filtering by action types (edit, comment, share, permission_change, etc.)

## Related

None — this is a new capability distinct from Gmail history.

---

I've implemented this in another Drive CLI and can share implementation details if helpful.
