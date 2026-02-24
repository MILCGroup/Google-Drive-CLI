## Summary

Add Google Admin SDK Directory API support for user and group management in Google Workspace.

## Current State

`gog groups` uses Cloud Identity API for listing groups and members. This proposal adds Admin Directory API for full user provisioning and group management with domain-wide admin capabilities.

## Proposed Commands

```bash
gog admin users list --domain example.com
gog admin users get user@example.com
gog admin users create user@example.com --given "John" --family "Doe" --password "..."
gog admin users suspend user@example.com
gog admin groups list --domain example.com
gog admin groups members list group@example.com
gog admin groups members add group@example.com user@example.com --role MEMBER
```

## Use Cases

- User provisioning and deprovisioning
- Onboarding/offboarding automation
- Bulk group membership management
- Admin audits and compliance

## API Notes

- API: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users
- Scopes: `admin.directory.user`, `admin.directory.group`
- Requires service account with domain-wide delegation (matches existing SA support)

## Related

- #179 — GAM feature parity for Workspace admin (stale PR)
- Existing `gog groups` command (Cloud Identity API)

---

I've implemented this in another Drive CLI and can share implementation details if helpful.
