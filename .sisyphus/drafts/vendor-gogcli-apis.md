# Draft: Vendor gogcli APIs into gdrv

## Requirements (confirmed)
- Add 7 new Google API modules to gdrv: Gmail, Calendar, People, Tasks, Forms, Apps Script, Groups
- Feature depth: Core + power user (skip niche: email tracking, delegation, watch/pub-sub, propose-time)
- Keep gdrv binary name (no rename)
- TDD approach (test-first)
- Follow existing Manager pattern exactly

## Technical Decisions
- Scope depth: Core + power user commands per API
- Brand: Keep gdrv
- Test strategy: TDD (RED → GREEN → REFACTOR)
- All 7 Go client libraries are in existing google.golang.org/api module — no new module deps
- Module pattern: Manager + Types + CLI + Scopes + ServiceFactory + Tests (8-step checklist from explore agent)

## Scope Per API (Core + Power User)

### Gmail (~18 commands)
CORE: search, thread get, message get, send (with reply/quote/HTML)
ADVANCED: drafts CRUD+send, labels CRUD, filters CRUD, vacation get/set, sendas list, batch delete/modify, attachments
SKIP: delegation, watch/Pub-Sub, email tracking, history, autoforward, forwarding

### Calendar (~13 commands)
CORE: events list (--today/--week/--days), event get, search, create (with attendees/location), update, delete, respond
ADVANCED: freebusy, conflicts, recurrence + reminders
SKIP: team calendars, focus-time/OOO/working-location shortcuts, propose-time, calendars metadata, acl, colors, users

### People/Contacts (~9 commands)
CORE: contacts list, search, get, create, update, delete
ADVANCED: other contacts list/search, workspace directory list/search
SKIP: JSON pipe update

### Tasks (~10 commands)
CORE: lists list/create, tasks list/get/add/update/done/delete
ADVANCED: undo, clear, repeat schedules

### Forms (~3 commands)
CORE: get, responses
ADVANCED: create

### Apps Script (~4 commands)
CORE: get, content
ADVANCED: create, run

### Groups (~2 commands)
CORE: list, members

## Key API Gotchas (from librarian)
- Gmail: Raw field must be base64.URLEncoding, List returns stubs only, replies need ThreadId + headers
- Calendar: EventDateTime all-day vs timed mutually exclusive, SendUpdates param, SyncToken for incremental
- People: PersonFields MANDATORY on every read, Etag required for updates
- Tasks: Due date time portion ignored, ShowCompleted defaults false, no batch API
- Forms: Immutable after creation (BatchUpdate only), WriteControl needed
- Apps Script: Script must be deployed, params primitives only, caller needs all script scopes
- Cloud Identity: Groups.Create returns Operation (async), Parent format critical, Labels determine type
