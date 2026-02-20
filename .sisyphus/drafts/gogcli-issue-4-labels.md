## Summary

Add Google Drive Labels API v2 support for structured file metadata and enterprise content classification.

## Current State

`gog drive` has no label/metadata taxonomy support. This is the Drive Labels API (not Gmail labels, which `gog gmail labels` already handles).

## Proposed Commands

```bash
gog drive labels list                             # List available label schemas
gog drive labels get <labelId>                    # Get label schema details
gog drive labels file list <fileId>               # Labels applied to a file
gog drive labels file apply <fileId> <labelId> --fields '{"key":"value"}'
gog drive labels file remove <fileId> <labelId>
```

## Use Cases

- Enterprise content classification (document types, retention policies)
- Project tagging with structured metadata
- Workflow automation based on label values
- Compliance and records management

## API Notes

- API: https://developers.google.com/drive/labels/overview
- Additional scope: `https://www.googleapis.com/auth/drive.labels`
- Labels are structured metadata schemas defined at the organization level

## Related

None — this is Drive Labels (structured metadata), distinct from Gmail labels.

---

I've implemented this in another Drive CLI and can share implementation details if helpful.
