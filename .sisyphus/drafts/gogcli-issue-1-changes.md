## Summary

Add support for Google Drive Changes API v3 to enable sync tools, incremental backups, and real-time change monitoring.

## Current State

`gog drive` supports file operations (ls, search, upload, download, etc.) but has no mechanism to track changes over time. Users must poll or maintain external state to detect modifications.

## Proposed Commands

```bash
gog drive changes start-token                    # Get starting page token
gog drive changes list --token <token>           # List changes since token
gog drive changes list --token <token> --max 50  # With pagination
gog drive changes watch --token <token> --webhook-url <url>  # Set up webhook
gog drive changes stop <channelId> <resourceId>  # Stop webhook
```

## Use Cases

- Build sync tools that mirror Drive to local storage
- Incremental backup systems that only fetch changed files
- Automation workflows triggered by file changes
- Real-time monitoring for compliance/security

## API Notes

- API: https://developers.google.com/drive/api/v3/reference/changes
- Scopes: Standard Drive scope (no additional scopes required)
- The Changes API provides page tokens for resumable change tracking

## Related

None — this is a new capability.

---

I've implemented this in another Drive CLI and can share implementation details if helpful.
