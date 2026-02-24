# Vendor gogcli APIs — Add 7 Google API Modules to gdrv

## TL;DR

> **Quick Summary**: Add 7 new Google API modules (Gmail, Calendar, People, Tasks, Forms, Apps Script, Groups) to gdrv, implementing ~59 commands at "core + power user" depth. Each module follows gdrv's established Manager/Types/CLI/Test pattern. All CLI commands target kong (not cobra) since the kong migration is underway.
>
> **Deliverables**:
> - 7 new Manager packages: `internal/gmail/`, `internal/calendar/`, `internal/people/`, `internal/tasks/`, `internal/forms/`, `internal/appscript/`, `internal/groups/`
> - 7 new type files: `internal/types/gmail.go`, `calendar.go`, `people.go`, `tasks.go`, `forms.go`, `appscript.go`, `groups.go`
> - 7 new CLI files: `internal/cli/gmail.go`, `calendar.go`, `people.go`, `tasks.go`, `forms.go`, `appscript.go`, `groups.go`
> - 7 test files (TDD): `internal/<module>/manager_test.go`
> - Updated foundation files: `internal/utils/constants.go` (scopes + presets), `internal/auth/service_factory.go` (7 new service types), `internal/auth/manager.go` (7 new service getters)
> - Updated `go.mod` with 7 new Google API client imports
>
> **Estimated Effort**: XL (5-7 days with parallel execution)
> **Parallel Execution**: YES — 4 waves, up to 7 parallel tasks
> **Critical Path**: Task 1 → Task 2 → Tasks 3-9 (parallel) → Tasks 10-16 (parallel) → Tasks 17-23 (parallel) → F1-F4

---

## Context

### Original Request
"Vendor the best of gogcli" — implement the 7 Google APIs that gogcli has but gdrv lacks (Gmail, Calendar, People, Tasks, Forms, Apps Script, Groups), adapted to gdrv's architecture and patterns.

### Interview Summary
**Key Discussions**:
- Feature depth: "Core + power user" — essential commands plus power-user features per API. Skip niche features (email tracking, delegation, watch/Pub-Sub, propose-time, team calendars).
- Naming: Keep `gdrv` binary name despite expanded scope.
- Test strategy: TDD (RED → GREEN → REFACTOR) — write failing tests first, then implement.
- CLI framework: Kong migration is underway (separate plan). New modules must use kong struct-tag patterns, NOT cobra.
- Module pattern: Follow gdrv's established 8-step checklist exactly.

**Research Findings**:
- All 7 Go API client libraries are under `google.golang.org/api` — already a dependency in go.mod. Just need specific package imports.
- Gmail: Raw field must be `base64.URLEncoding`, List returns message stubs (must Get for full content), replies need ThreadId + In-Reply-To/References headers, `BatchDelete` requires `mail.google.com` restricted scope.
- Calendar: EventDateTime all-day vs timed are mutually exclusive, SendUpdates param controls notification emails, SyncToken enables incremental sync.
- People: PersonFields MANDATORY on every read (400 error without it), Etag required for updates, `people/me` vs `people/c{id}` resource name format.
- Tasks: Due date time portion is ignored (date-only), ShowCompleted defaults false, no batch API, 50K requests/day limit (most restrictive), subtask ordering via Move with Parent/Previous.
- Forms: Immutable after creation (questions only via BatchUpdate), WriteControl with RevisionId needed for concurrent edits, forms created after March 2026 are unpublished by default.
- Apps Script: Script must be deployed for Run, params primitives only, caller needs all scopes the script uses, 6-min timeout, script and caller must share GCP project.
- Cloud Identity Groups: Groups.Create returns async Operation, Parent must be `customers/{customerId}`, Labels determine group type (not Admin SDK groups).

### Metis Review
**Identified Gaps** (all addressed in plan):
- **Kong dependency**: New CLI files must target kong struct patterns. The kong migration plan (Tasks 1-20 in separate plan) must have completed at least the root struct + pattern establishment (Tasks 1-6) before CLI tasks in THIS plan execute. Added as explicit dependency.
- **Groups redundancy with Admin SDK**: gdrv already has `admin groups` (Admin Directory API). New `groups` module uses Cloud Identity API — must have distinct command path and documentation. Decision: use `gdrv groups` (Cloud Identity) vs existing `gdrv admin groups` (Admin SDK).
- **Scope preset explosion**: 7 new APIs means many new scope constants and potentially many new presets. Decision: add per-API presets + one "suite-complete" mega-preset.
- **Tasks API 50K/day rate limit**: Most restrictive of all 7 APIs. Must document and consider rate-limit awareness.
- **Forms API March 2026 breaking change**: Forms created after March 31, 2026 are unpublished by default. The `create` command must handle this (auto-publish or separate `publish` command).
- **Apps Script GCP project constraint**: The `run` command only works when script and OAuth client share GCP project. This dramatically limits utility — marked as experimental.
- **Gmail `batch-delete` requires `mail.google.com`**: Restricted scope, must be a separate scope tier.
- **People API has 3 scope families**: contacts, contacts.other, directory — all need coverage in presets.

---

## Work Objectives

### Core Objective
Add 7 new Google API modules to gdrv following the established Manager/Types/CLI/Test pattern, with TDD methodology, targeting kong CLI framework.

### Concrete Deliverables
- 7 working API modules with ~59 total CLI commands
- Full test coverage via TDD (each manager method has tests)
- Scope constants, presets, service factory entries for all 7 APIs
- kong-compatible CLI command structs for all 7 modules

### Definition of Done
- [ ] `go build -o bin/gdrv ./cmd/gdrv` succeeds with all 7 modules
- [ ] `go test -v -race ./internal/gmail/... ./internal/calendar/... ./internal/people/... ./internal/tasks/... ./internal/forms/... ./internal/appscript/... ./internal/groups/...` all pass
- [ ] `go vet ./...` reports no issues
- [ ] `bin/gdrv gmail --help` shows all Gmail subcommands
- [ ] `bin/gdrv calendar --help` shows all Calendar subcommands
- [ ] `bin/gdrv people --help` shows all People subcommands
- [ ] `bin/gdrv tasks --help` shows all Tasks subcommands
- [ ] `bin/gdrv forms --help` shows all Forms subcommands
- [ ] `bin/gdrv appscript --help` shows all Apps Script subcommands
- [ ] `bin/gdrv groups --help` shows all Groups subcommands

### Must Have
- All ~59 commands accessible via kong CLI struct dispatch
- All types implement `Tabular` interface (Headers(), Rows(), EmptyMessage())
- All API calls wrapped in `api.ExecuteWithRetry[T]()`
- All managers follow `NewManager(client *api.Client, service *<api>.Service)` pattern
- JSON output support (`--json` flag) for every command
- TDD: tests written before implementation for each manager method
- Scope constants for all 7 APIs in `internal/utils/constants.go`
- Service types + factory methods + getters for all 7 APIs
- Pagination support where APIs support it (Gmail, Calendar, People, Tasks, Forms responses)
- `--paginate` flag for auto-pagination where applicable

### Must NOT Have (Guardrails)
- Do NOT modify existing modules (files, folders, sheets, docs, slides, admin, chat, etc.)
- Do NOT use cobra for new CLI commands — kong only
- Do NOT implement niche features: email tracking, delegation, watch/Pub-Sub, propose-time, team calendars, focus-time/OOO/working-location, calendar ACL/colors
- Do NOT implement history sync (Gmail), autoforward, forwarding
- Do NOT add excessive error handling beyond what existing modules do
- Do NOT add excessive comments or JSDoc — match existing code style
- Do NOT over-abstract — each module is standalone, no shared "workspace base manager"
- Do NOT implement Apps Script `run` as production-ready — mark experimental with prominent GCP project constraint warning
- Do NOT rename the binary or restructure the project layout
- Do NOT add watch/webhook/Pub-Sub for any API (can be added later)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (`go test` with race detection)
- **Automated tests**: YES (TDD — RED → GREEN → REFACTOR)
- **Framework**: `go test` (standard library)
- **TDD Flow**: Each manager task writes tests FIRST (RED), then implements until tests pass (GREEN), then refactors

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Manager methods**: Use Bash — `go test -v -race ./internal/<module>/...`
- **CLI commands**: Use Bash — `bin/gdrv <module> --help`, `bin/gdrv <module> <cmd> --help`
- **Build verification**: Use Bash — `go build`, `go test -race`, `go vet`
- **Type compliance**: Use Bash — `go vet ./internal/types/...` (ensures Tabular interface satisfaction)

---

## Execution Strategy

### Parallel Execution Waves

> **IMPORTANT DEPENDENCY**: This plan assumes the kong migration plan Tasks 1-6 are complete (root CLI struct, main.go rewrite, files.go pattern established). If not yet complete, Wave 3 (CLI tasks) must wait until kong pattern tasks finish.

```
Wave 1 (Foundation — scopes, service factory, go.mod):
├── Task 1: Scope constants + presets for all 7 APIs [quick]
└── Task 2: Service factory + service getters for all 7 APIs [quick]

Wave 2 (Types — all 7 type files in parallel):
├── Task 3: Gmail types (internal/types/gmail.go) [quick]
├── Task 4: Calendar types (internal/types/calendar.go) [quick]
├── Task 5: People types (internal/types/people.go) [quick]
├── Task 6: Tasks types (internal/types/tasks.go) [quick]
├── Task 7: Forms types (internal/types/forms.go) [quick]
├── Task 8: Apps Script types (internal/types/appscript.go) [quick]
└── Task 9: Groups types (internal/types/groups.go) [quick]

Wave 3 (Managers + TDD Tests — 7 parallel, LARGEST wave):
├── Task 10: Gmail manager + tests (~18 methods) [deep]
├── Task 11: Calendar manager + tests (~13 methods) [deep]
├── Task 12: People manager + tests (~9 methods) [unspecified-high]
├── Task 13: Tasks manager + tests (~10 methods) [unspecified-high]
├── Task 14: Forms manager + tests (~3 methods) [quick]
├── Task 15: Apps Script manager + tests (~4 methods) [unspecified-high]
└── Task 16: Groups manager + tests (~2 methods) [quick]

Wave 4 (CLI commands — 7 parallel, depends on kong migration):
├── Task 17: Gmail CLI commands (kong structs, ~18 cmds) [unspecified-high]
├── Task 18: Calendar CLI commands (kong structs, ~13 cmds) [unspecified-high]
├── Task 19: People CLI commands (kong structs, ~9 cmds) [unspecified-high]
├── Task 20: Tasks CLI commands (kong structs, ~10 cmds) [unspecified-high]
├── Task 21: Forms CLI commands (kong structs, ~3 cmds) [quick]
├── Task 22: Apps Script CLI commands (kong structs, ~4 cmds) [quick]
└── Task 23: Groups CLI commands (kong structs, ~2 cmds) [quick]

Wave FINAL (After ALL — 4 parallel verification agents):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: Task 1 → Task 2 → Task 10 (Gmail manager) → Task 17 (Gmail CLI) → F1-F4
Parallel Speedup: ~70% faster than sequential
Max Concurrent: 7 (Waves 2, 3, and 4)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 2, 10-16, 17-23 | 1 |
| 2 | 1 | 10-16 | 1 |
| 3-9 | 1 | 10-16 (corresponding module) | 2 |
| 10 | 2, 3 | 17 | 3 |
| 11 | 2, 4 | 18 | 3 |
| 12 | 2, 5 | 19 | 3 |
| 13 | 2, 6 | 20 | 3 |
| 14 | 2, 7 | 21 | 3 |
| 15 | 2, 8 | 22 | 3 |
| 16 | 2, 9 | 23 | 3 |
| 17 | 10, kong migration T6+ | F1-F4 | 4 |
| 18 | 11, kong migration T6+ | F1-F4 | 4 |
| 19 | 12, kong migration T6+ | F1-F4 | 4 |
| 20 | 13, kong migration T6+ | F1-F4 | 4 |
| 21 | 14, kong migration T6+ | F1-F4 | 4 |
| 22 | 15, kong migration T6+ | F1-F4 | 4 |
| 23 | 16, kong migration T6+ | F1-F4 | 4 |
| F1-F4 | 17-23 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **2 tasks** — T1→`quick`, T2→`quick`
- **Wave 2**: **7 tasks** — T3-T9→`quick`
- **Wave 3**: **7 tasks** — T10→`deep`, T11→`deep`, T12-T13→`unspecified-high`, T14→`quick`, T15→`unspecified-high`, T16→`quick`
- **Wave 4**: **7 tasks** — T17-T20→`unspecified-high`, T21-T23→`quick`
- **FINAL**: **4 tasks** — F1→`oracle`, F2→`unspecified-high`, F3→`unspecified-high`, F4→`deep`

---

## TODOs

> Implementation + Test = ONE Task. Never separate.
> EVERY task MUST have: Recommended Agent Profile + Parallelization info + QA Scenarios.
> TDD: Manager tasks write tests FIRST, then implement.

- [ ] 1. Foundation — Scope Constants + Presets for All 7 APIs

  **What to do**:
  - Add scope constants to `internal/utils/constants.go` for all 7 APIs:
    - Gmail: `ScopeGmailReadonly`, `ScopeGmailSend`, `ScopeGmailCompose`, `ScopeGmailModify`, `ScopeGmailLabels`, `ScopeGmailSettingsBasic`, `ScopeGmailFull` (mail.google.com — for batch-delete only)
    - Calendar: `ScopeCalendar`, `ScopeCalendarReadonly`
    - People: `ScopeContacts`, `ScopeContactsReadonly`, `ScopeContactsOtherReadonly`, `ScopeDirectoryReadonly`
    - Tasks: `ScopeTasks` (single scope — no readonly variant)
    - Forms: `ScopeFormsBody`, `ScopeFormsBodyReadonly`, `ScopeFormsResponsesReadonly`
    - Apps Script: `ScopeScriptProjects`, `ScopeScriptProjectsReadonly`
    - Cloud Identity Groups: `ScopeCloudIdentityGroupsReadonly`, `ScopeCloudIdentityGroups`
  - Add scope presets:
    - `ScopesGmail` — Gmail send + compose + modify + labels + settings (core read/write)
    - `ScopesGmailReadonly` — Gmail readonly
    - `ScopesCalendar` — Calendar full
    - `ScopesCalendarReadonly` — Calendar readonly
    - `ScopesPeople` — Contacts + contacts.other.readonly + directory.readonly
    - `ScopesTasks` — Tasks (single scope)
    - `ScopesForms` — Forms body + responses
    - `ScopesAppScript` — Script projects
    - `ScopesGroups` — Cloud Identity groups
    - `ScopesSuiteComplete` — ALL existing scopes + ALL new API scopes (mega-preset)
  - Register new presets in the scope preset map used by `internal/cli/auth.go` (preset name → scope slice)

  **Must NOT do**:
  - Modify existing scope presets (workspace-basic, workspace-full, admin, etc.)
  - Add the Gmail full-access scope (`mail.google.com`) to any default preset — it must be explicitly requested

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Adding constants and array literals — no complex logic
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 2)
  - **Blocks**: Tasks 2, 3-9, 10-16, 17-23
  - **Blocked By**: None

  **References**:
  **Pattern References**:
  - `internal/utils/constants.go:14-37` — Existing scope constant pattern (ScopeFull, ScopeSheets, etc.)
  - `internal/utils/constants.go:40-107` — Existing preset pattern (ScopesWorkspaceBasic, ScopesWorkspaceFull, etc.)

  **API/Type References**:
  - `internal/cli/auth.go:375-396` — Preset name-to-scope mapping used in CLI

  **External References**:
  - Gmail scopes: `google.golang.org/api/gmail/v1` — GmailReadonlyScope, GmailSendScope, GmailComposeScope, GmailModifyScope, GmailLabelsScope, GmailSettingsBasicScope
  - Calendar scopes: `google.golang.org/api/calendar/v3` — CalendarScope, CalendarReadonlyScope
  - People scopes: `google.golang.org/api/people/v1` — ContactsScope, ContactsReadonlyScope, ContactsOtherReadonlyScope, DirectoryReadonlyScope
  - Tasks scopes: `google.golang.org/api/tasks/v1` — TasksScope
  - Forms scopes: `google.golang.org/api/forms/v1` — FormsBodyScope, FormsBodyReadonlyScope, FormsResponsesReadonlyScope
  - Apps Script scopes: `google.golang.org/api/script/v1` — ScriptProjectsScope, ScriptProjectsReadonlyScope
  - Cloud Identity scopes: `google.golang.org/api/cloudidentity/v1` — CloudIdentityGroupsScope, CloudIdentityGroupsReadonlyScope

  **WHY Each Reference Matters**:
  - `constants.go:14-37` — Follow exact naming pattern (Scope prefix + CamelCase domain + qualifier)
  - `constants.go:40-107` — Follow exact preset pattern (Scopes prefix + name, slice of scope constants)
  - `auth.go:375-396` — The preset map MUST include new preset names or they won't be usable via `--preset`

  **Acceptance Criteria**:
  - [ ] All scope constants compile: `go build ./internal/utils/...`
  - [ ] All presets defined as `[]string` slices
  - [ ] `ScopesSuiteComplete` contains ALL scopes (existing + new)
  - [ ] No existing presets modified

  **QA Scenarios**:

  ```
  Scenario: Scope constants compile
    Tool: Bash
    Preconditions: None
    Steps:
      1. go build ./internal/utils/...
      2. go vet ./internal/utils/...
    Expected Result: Build and vet pass with 0 errors
    Evidence: .sisyphus/evidence/task-1-scopes-build.txt

  Scenario: All new scope constants are defined
    Tool: Bash
    Steps:
      1. grep -c "ScopeGmail\|ScopeCalendar\|ScopeContacts\|ScopeDirectory\|ScopeTasks\|ScopeForms\|ScopeScript\|ScopeCloudIdentity" internal/utils/constants.go
    Expected Result: At least 18 matches (18+ new scope constants)
    Evidence: .sisyphus/evidence/task-1-scope-count.txt

  Scenario: Suite-complete preset includes all APIs
    Tool: Bash
    Steps:
      1. grep -A 50 "ScopesSuiteComplete" internal/utils/constants.go
      2. Verify it contains Gmail, Calendar, People, Tasks, Forms, Script, CloudIdentity scope references
    Expected Result: All 7 API families represented in the mega-preset
    Evidence: .sisyphus/evidence/task-1-suite-complete.txt
  ```

  **Commit**: YES (group with Task 2)
  - Message: `feat(auth): add OAuth scope constants and presets for 7 new APIs`
  - Files: `internal/utils/constants.go`
  - Pre-commit: `go build ./internal/utils/...`

---

- [ ] 2. Foundation — Service Factory + Service Getters for All 7 APIs

  **What to do**:
  - Add 7 new ServiceType constants to `internal/auth/service_factory.go`:
    - `ServiceGmail ServiceType = "gmail"`
    - `ServiceCalendar ServiceType = "calendar"`
    - `ServicePeople ServiceType = "people"`
    - `ServiceTasks ServiceType = "tasks"`
    - `ServiceForms ServiceType = "forms"`
    - `ServiceAppScript ServiceType = "appscript"`
    - `ServiceCloudIdentity ServiceType = "cloudidentity"`
  - Add 7 new `Create*Service` methods on `ServiceFactory`:
    - `CreateGmailService(ctx, creds) (*gmail.Service, error)` — pattern: `gmail.NewService(ctx, option.WithHTTPClient(client))`
    - `CreateCalendarService(ctx, creds) (*calendar.Service, error)`
    - `CreatePeopleService(ctx, creds) (*people.Service, error)`
    - `CreateTasksService(ctx, creds) (*tasks.Service, error)`
    - `CreateFormsService(ctx, creds) (*forms.Service, error)`
    - `CreateAppScriptService(ctx, creds) (*script.Service, error)`
    - `CreateCloudIdentityService(ctx, creds) (*cloudidentity.Service, error)`
  - Add 7 switch cases to `CreateService()` method
  - Add 7 new service getter methods to `internal/auth/manager.go`:
    - `GetGmailService(ctx, creds) (*gmail.Service, error)` — follows pattern of existing GetSheetsService
    - `GetCalendarService(ctx, creds) (*calendar.Service, error)`
    - `GetPeopleService(ctx, creds) (*people.Service, error)`
    - `GetTasksService(ctx, creds) (*tasks.Service, error)`
    - `GetFormsService(ctx, creds) (*forms.Service, error)`
    - `GetAppScriptService(ctx, creds) (*script.Service, error)`
    - `GetCloudIdentityService(ctx, creds) (*cloudidentity.Service, error)`
  - Add required imports to both files:
    - `google.golang.org/api/gmail/v1`
    - `google.golang.org/api/calendar/v3`
    - `google.golang.org/api/people/v1`
    - `google.golang.org/api/tasks/v1`
    - `google.golang.org/api/forms/v1`
    - `google.golang.org/api/script/v1`
    - `google.golang.org/api/cloudidentity/v1`
  - Run `go mod tidy` to pull in any missing sub-packages (all under google.golang.org/api which is already a dependency)

  **Must NOT do**:
  - Modify existing Create*Service methods
  - Change the ServiceFactory struct definition
  - Change the Manager struct definition

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Repetitive pattern — copy existing Create*Service 7 times with different types
  - **Skills**: [`golang-pro`]
    - `golang-pro`: Needs correct Go import paths and type assertions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Task 1)
  - **Blocks**: Tasks 10-16 (managers need service factory)
  - **Blocked By**: Task 1 (scope constants must exist for validation)

  **References**:
  **Pattern References**:
  - `internal/auth/service_factory.go:17-53` — Existing ServiceType constants and CreateService switch (EXACT pattern to extend)
  - `internal/auth/service_factory.go:55-83` — Existing Create*Service methods (EXACT pattern to copy)
  - `internal/auth/manager.go:574-598` — Existing Get*Service methods (EXACT pattern to copy)

  **External References**:
  - `google.golang.org/api/gmail/v1` — `gmail.NewService(ctx, option.WithHTTPClient(client))`
  - `google.golang.org/api/calendar/v3` — `calendar.NewService(ctx, ...)`
  - `google.golang.org/api/people/v1` — `people.NewService(ctx, ...)`
  - `google.golang.org/api/tasks/v1` — `tasks.NewService(ctx, ...)`
  - `google.golang.org/api/forms/v1` — `forms.NewService(ctx, ...)`
  - `google.golang.org/api/script/v1` — `script.NewService(ctx, ...)`
  - `google.golang.org/api/cloudidentity/v1` — `cloudidentity.NewService(ctx, ...)`

  **WHY Each Reference Matters**:
  - `service_factory.go:55-83` — Each new method is a copy of this pattern with different Google API package. The pattern is: get HTTP client from manager, call `<pkg>.NewService(ctx, option.WithHTTPClient(client))`.
  - `manager.go:574-598` — Each new getter delegates to the factory. Pattern: `f.CreateXxxService(ctx, creds)`.

  **Acceptance Criteria**:
  - [ ] 7 new ServiceType constants defined
  - [ ] 7 new Create*Service methods on ServiceFactory
  - [ ] 7 new switch cases in CreateService
  - [ ] 7 new Get*Service methods on Manager
  - [ ] `go build ./internal/auth/...` succeeds
  - [ ] `go mod tidy` exits cleanly

  **QA Scenarios**:

  ```
  Scenario: Service factory compiles with all 7 new services
    Tool: Bash
    Steps:
      1. go build ./internal/auth/...
      2. go vet ./internal/auth/...
    Expected Result: Build and vet pass
    Evidence: .sisyphus/evidence/task-2-factory-build.txt

  Scenario: All service types defined
    Tool: Bash
    Steps:
      1. grep -c "ServiceType = " internal/auth/service_factory.go
    Expected Result: 13 matches (6 existing + 7 new)
    Evidence: .sisyphus/evidence/task-2-service-types.txt

  Scenario: Go mod tidy succeeds
    Tool: Bash
    Steps:
      1. go mod tidy
      2. go build ./...
    Expected Result: Both succeed without error
    Evidence: .sisyphus/evidence/task-2-mod-tidy.txt
  ```

  **Commit**: YES (group with Task 1)
  - Message: `feat(auth): add service factory and getters for 7 new Google APIs`
  - Files: `internal/auth/service_factory.go`, `internal/auth/manager.go`, `go.mod`, `go.sum`
  - Pre-commit: `go build ./internal/auth/...`

- [ ] 3. Types — Gmail Domain Types (`internal/types/gmail.go`)

  **What to do**:
  - Create `internal/types/gmail.go` with types implementing the `Tabular` interface:
    - `GmailMessage` — ID, ThreadID, From, To, Subject, Date, Snippet, LabelIDs, SizeEstimate. Headers: ID, From, To, Subject, Date. Used by message get.
    - `GmailMessageList` — Messages []GmailMessage, ResultSizeEstimate. Headers: ID, From, Subject, Date, Labels. Used by search/list.
    - `GmailThread` — ID, Snippet, Messages []GmailMessage. Headers: ID, Messages, Snippet. Used by thread get.
    - `GmailDraft` — ID, Message GmailMessage. Used by drafts list/get.
    - `GmailDraftList` — Drafts []GmailDraft. Used by drafts list.
    - `GmailLabel` — ID, Name, Type, MessagesTotal, MessagesUnread, ThreadsTotal, ThreadsUnread. Used by labels list/get.
    - `GmailLabelList` — Labels []GmailLabel. Used by labels list.
    - `GmailFilter` — ID, Criteria (From, To, Subject, Query, HasAttachment), Action (AddLabelIDs, RemoveLabelIDs, Forward). Used by filters list/get.
    - `GmailFilterList` — Filters []GmailFilter.
    - `GmailVacationSettings` — EnableAutoReply, ResponseSubject, ResponseBodyPlainText, ResponseBodyHtml, StartTime, EndTime. Used by vacation get/set.
    - `GmailSendAs` — SendAsEmail, DisplayName, ReplyToAddress, IsPrimary, IsDefault. Used by sendas list.
    - `GmailSendAsList` — SendAs []GmailSendAs.
    - `GmailSendResult` — ID, ThreadID, LabelIDs. Used by send/reply result.
    - `GmailAttachment` — MessageID, AttachmentID, Filename, MimeType, Size. Used by attachments get.
    - `GmailBatchResult` — Count int, Action string. Used by batch-delete/batch-modify results.
  - All list types must implement Tabular: Headers(), Rows(), EmptyMessage()
  - Single-item types must also implement Tabular for table output

  **Must NOT do**:
  - Import Gmail API package in types — types are pure data structs
  - Add methods beyond Tabular interface

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4-9)
  - **Blocks**: Task 10 (Gmail manager)
  - **Blocked By**: Task 1 (scope constants for reference)

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — CANONICAL type pattern. Study: struct fields with JSON tags, Headers/Rows/EmptyMessage methods, truncateID helper usage.
  - `internal/types/chat.go` — Another type pattern with nested structs.
  - `internal/types/admin.go` — Complex type pattern with user/group types.

  **WHY Each Reference Matters**:
  - `sheets.go` shows the exact Tabular interface implementation pattern: Headers returns column names, Rows returns string matrices, EmptyMessage returns a context-specific message.

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] All types have JSON tags
  - [ ] All list types implement Tabular (Headers, Rows, EmptyMessage)
  - [ ] `go vet ./internal/types/...` clean

  **QA Scenarios**:

  ```
  Scenario: Gmail types compile and satisfy Tabular
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
      2. go vet ./internal/types/...
    Expected Result: Build and vet pass
    Evidence: .sisyphus/evidence/task-3-gmail-types.txt

  Scenario: Type count verification
    Tool: Bash
    Steps:
      1. grep -c "^type Gmail" internal/types/gmail.go
    Expected Result: At least 12 types defined
    Evidence: .sisyphus/evidence/task-3-gmail-type-count.txt
  ```

  **Commit**: YES (group with Tasks 4-9 — one commit for all types)
  - Message: `feat(types): add Gmail domain types with Tabular interface`
  - Files: `internal/types/gmail.go`

---

- [ ] 4. Types — Calendar Domain Types (`internal/types/calendar.go`)

  **What to do**:
  - Create `internal/types/calendar.go` with Tabular types:
    - `CalendarEvent` — ID, Summary, Description, Location, Start (DateTime string), End (DateTime string), Status, Creator, Organizer, Attendees []CalendarAttendee, HangoutLink, HtmlLink, Recurrence []string, Reminders. Headers: ID, Summary, Start, End, Location, Status.
    - `CalendarEventList` — Events []CalendarEvent, Summary (calendar name). Used by events list/search.
    - `CalendarAttendee` — Email, DisplayName, ResponseStatus, Self bool, Optional bool. Nested type for event attendees.
    - `CalendarFreeBusy` — Calendars map[string][]CalendarBusyPeriod. Used by freebusy command.
    - `CalendarBusyPeriod` — Start, End string.
    - `CalendarConflict` — Event CalendarEvent, ConflictsWith []CalendarEvent. Used by conflicts command.
    - `CalendarEventResult` — ID, Summary, HtmlLink. Used by create/update/delete results.
  - Handle all-day vs timed events: Start/End fields should be strings that can represent either "2026-01-15" (all-day) or "2026-01-15T10:00:00-05:00" (timed).

  **Must NOT do**:
  - Import Calendar API package in types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 3, 5-9)
  - **Blocks**: Task 11 (Calendar manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Canonical Tabular pattern
  - `internal/types/admin.go` — Complex struct with nested types (User/Group/Member)

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] CalendarEvent and CalendarEventList implement Tabular
  - [ ] All-day vs timed events use single string field (no separate date/dateTime)

  **QA Scenarios**:

  ```
  Scenario: Calendar types compile
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
      2. go vet ./internal/types/...
    Expected Result: Build and vet pass
    Evidence: .sisyphus/evidence/task-4-calendar-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add Calendar domain types with Tabular interface`
  - Files: `internal/types/calendar.go`

---

- [ ] 5. Types — People Domain Types (`internal/types/people.go`)

  **What to do**:
  - Create `internal/types/people.go` with Tabular types:
    - `Contact` — ResourceName, Etag, DisplayName, GivenName, FamilyName, Emails []ContactEmail, Phones []ContactPhone, Organizations []ContactOrg, Addresses []ContactAddress. Headers: Name, Email, Phone, Organization.
    - `ContactList` — Contacts []Contact, TotalPeople int, TotalItems int. Used by contacts list/search.
    - `ContactEmail` — Value, Type (home/work/other), Primary bool.
    - `ContactPhone` — Value, Type, Primary bool.
    - `ContactOrg` — Name, Title, Department.
    - `ContactAddress` — FormattedValue, Type, City, Region, Country.
    - `OtherContact` — ResourceName, DisplayName, Emails []ContactEmail, Phones []ContactPhone. Simplified version for "other contacts".
    - `OtherContactList` — Contacts []OtherContact.
    - `DirectoryPerson` — ResourceName, DisplayName, Emails []ContactEmail, Phones []ContactPhone, Organizations []ContactOrg. For Workspace directory.
    - `DirectoryPersonList` — People []DirectoryPerson.
    - `ContactResult` — ResourceName, DisplayName. Used by create/update/delete results.
  - Note: Etag field is critical — People API requires it for updates

  **Must NOT do**:
  - Import People API package in types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 12 (People manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Tabular pattern
  - `internal/types/admin.go:User` — Complex type with nested structs for people-like data

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] Contact and ContactList implement Tabular
  - [ ] Etag field present on Contact type

  **QA Scenarios**:

  ```
  Scenario: People types compile
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
    Expected Result: Build passes
    Evidence: .sisyphus/evidence/task-5-people-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add People/Contacts domain types with Tabular interface`
  - Files: `internal/types/people.go`

- [ ] 6. Types — Tasks Domain Types (`internal/types/tasks.go`)

  **What to do**:
  - Create `internal/types/tasks.go` with Tabular types:
    - `TaskList` — ID, Title, Updated. Headers: ID, Title, Updated. Used by task lists list.
    - `TaskListResult` — TaskLists []TaskList. Implements Tabular.
    - `Task` — ID, Title, Notes, Status (needsAction/completed), Due (date string, time ignored), Completed (datetime), Parent (parent task ID), Position, Links []TaskLink. Headers: ID, Title, Status, Due.
    - `TaskResult` — Tasks []Task. Implements Tabular.
    - `TaskLink` — Type, Description, Link.
    - `TaskMutationResult` — ID, Title, Status. Used by add/update/done/delete results.
  - Note: Due field is DATE ONLY (time portion ignored by API). Store as string "YYYY-MM-DD".
  - Note: Status is one of "needsAction" or "completed"

  **Must NOT do**:
  - Import Tasks API package in types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 13 (Tasks manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Tabular pattern

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] Task and TaskResult implement Tabular
  - [ ] Due field is string type (not time.Time)

  **QA Scenarios**:

  ```
  Scenario: Tasks types compile
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
    Expected Result: Build passes
    Evidence: .sisyphus/evidence/task-6-tasks-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add Tasks domain types with Tabular interface`
  - Files: `internal/types/tasks.go`

---

- [ ] 7. Types — Forms Domain Types (`internal/types/forms.go`)

  **What to do**:
  - Create `internal/types/forms.go` with Tabular types:
    - `Form` — FormID, Info (Title, Description, DocumentTitle), RevisionID, ResponderURI, LinkedSheetID, Items []FormItem. Headers: ID, Title, Items, Responses URL.
    - `FormItem` — ItemID, Title, Description, QuestionItem (QuestionID, Required bool, ChoiceQuestion/TextQuestion/etc). Simplified — show title + type.
    - `FormResponse` — ResponseID, CreateTime, LastSubmittedTime, RespondentEmail, Answers map[string]FormAnswer. Headers: Response ID, Email, Submitted.
    - `FormResponseList` — Responses []FormResponse. Implements Tabular.
    - `FormAnswer` — QuestionID, TextAnswers []string, FileUploadAnswers []string, ChoiceAnswers []string.
    - `FormCreateResult` — FormID, Title, ResponderURI. Used by create result.
  - Note: Answers are keyed by questionId, NOT question title

  **Must NOT do**:
  - Import Forms API package in types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 14 (Forms manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Tabular pattern

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] Form and FormResponseList implement Tabular

  **QA Scenarios**:

  ```
  Scenario: Forms types compile
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
    Expected Result: Build passes
    Evidence: .sisyphus/evidence/task-7-forms-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add Forms domain types with Tabular interface`
  - Files: `internal/types/forms.go`

---

- [ ] 8. Types — Apps Script Domain Types (`internal/types/appscript.go`)

  **What to do**:
  - Create `internal/types/appscript.go` with Tabular types:
    - `ScriptProject` — ScriptID, Title, ParentID, CreateTime, UpdateTime. Headers: ID, Title, Created, Updated.
    - `ScriptContent` — ScriptID, Files []ScriptFile. Used by content command.
    - `ScriptFile` — Name, Type (SERVER_JS, HTML, JSON), Source string, CreateTime, UpdateTime, FunctionSet []string. Headers: Name, Type, Functions.
    - `ScriptRunResult` — Done bool, Error *ScriptError, Response map[string]interface{}. Used by run result.
    - `ScriptError` — Code int, Message string, Details []string.
    - `ScriptCreateResult` — ScriptID, Title. Used by create result.
  - Note: Script ID is NOT the same as the Drive file ID

  **Must NOT do**:
  - Import Apps Script API package in types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 15 (Apps Script manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Tabular pattern
  - `internal/types/docs.go` — Simpler type pattern

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] ScriptProject and ScriptContent implement Tabular

  **QA Scenarios**:

  ```
  Scenario: AppScript types compile
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
    Expected Result: Build passes
    Evidence: .sisyphus/evidence/task-8-appscript-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add Apps Script domain types with Tabular interface`
  - Files: `internal/types/appscript.go`

---

- [ ] 9. Types — Groups Domain Types (`internal/types/groups.go`)

  **What to do**:
  - Create `internal/types/groups.go` with Tabular types:
    - `CloudIdentityGroup` — Name (resource name), GroupKey (ID/Namespace), DisplayName, Description, CreateTime, UpdateTime, Labels map[string]string. Headers: Name, Display Name, Email, Created.
    - `CloudIdentityGroupList` — Groups []CloudIdentityGroup. Implements Tabular.
    - `CloudIdentityMember` — Name, PreferredMemberKey (ID/Namespace), Roles []MemberRole, CreateTime. Headers: Member, Email, Role, Joined.
    - `CloudIdentityMemberList` — Members []CloudIdentityMember. Implements Tabular.
    - `MemberRole` — Name string (OWNER, MANAGER, MEMBER).
  - Note: Cloud Identity groups use resource names like `groups/{groupId}`, members use `groups/{groupId}/memberships/{membershipId}`
  - Note: This is Cloud Identity API, NOT Admin SDK. `gdrv admin groups` already exists for Admin SDK. These types must NOT conflict with `internal/types/admin.go` types.

  **Must NOT do**:
  - Import Cloud Identity API package in types
  - Use type names that conflict with existing admin.go types (AdminGroup, AdminGroupMember)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 16 (Groups manager)
  - **Blocked By**: Task 1

  **References**:
  **Pattern References**:
  - `internal/types/sheets.go` — Tabular pattern
  - `internal/types/admin.go` — AdminGroup and AdminGroupMember types — MUST NOT conflict with these names

  **WHY Each Reference Matters**:
  - `admin.go` — Existing group types use `AdminGroup`, `AdminGroupMember` names. New types use `CloudIdentityGroup`, `CloudIdentityMember` to avoid name collision.

  **Acceptance Criteria**:
  - [ ] `go build ./internal/types/...` succeeds
  - [ ] No name conflicts with existing admin.go types
  - [ ] CloudIdentityGroupList and CloudIdentityMemberList implement Tabular

  **QA Scenarios**:

  ```
  Scenario: Groups types compile without conflicts
    Tool: Bash
    Steps:
      1. go build ./internal/types/...
      2. grep "AdminGroup\|CloudIdentityGroup" internal/types/groups.go internal/types/admin.go
    Expected Result: Build passes. groups.go uses CloudIdentity prefix, admin.go uses Admin prefix — no collision.
    Evidence: .sisyphus/evidence/task-9-groups-types.txt
  ```

  **Commit**: YES (group with Wave 2 types commit)
  - Message: `feat(types): add Cloud Identity Groups domain types with Tabular interface`
  - Files: `internal/types/groups.go`

- [ ] 10. Manager + TDD Tests — Gmail (`internal/gmail/`)

  **What to do**:
  TDD approach: Write tests FIRST in `manager_test.go`, then implement in `manager.go`.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *gmail.Service
  }
  func NewManager(client *api.Client, service *gmail.Service) *Manager
  ```

  **Methods to implement** (~18, matching command scope):
  1. `Search(ctx, reqCtx, query, maxResults, pageToken) (*types.GmailMessageList, string, error)` — gmail.Users.Messages.List with Q param. Returns stubs only.
  2. `GetMessage(ctx, reqCtx, messageID, format) (*types.GmailMessage, error)` — gmail.Users.Messages.Get. Format: "full"/"metadata"/"minimal". Parse headers (From, To, Subject, Date) from payload.Headers.
  3. `GetThread(ctx, reqCtx, threadID, format) (*types.GmailThread, error)` — gmail.Users.Threads.Get.
  4. `Send(ctx, reqCtx, to, subject, body, cc, bcc, htmlBody, inReplyTo, threadID) (*types.GmailSendResult, error)` — Compose RFC 2822 message, base64.URLEncoding.EncodeToString the raw bytes, set Raw field on gmail.Message. If inReplyTo set, add In-Reply-To + References headers and ThreadId.
  5. `ListDrafts(ctx, reqCtx, maxResults, pageToken) (*types.GmailDraftList, string, error)`
  6. `GetDraft(ctx, reqCtx, draftID) (*types.GmailDraft, error)`
  7. `CreateDraft(ctx, reqCtx, to, subject, body) (*types.GmailDraft, error)` — Same RFC 2822 + base64 encoding as Send.
  8. `UpdateDraft(ctx, reqCtx, draftID, to, subject, body) (*types.GmailDraft, error)`
  9. `DeleteDraft(ctx, reqCtx, draftID) error`
  10. `SendDraft(ctx, reqCtx, draftID) (*types.GmailSendResult, error)`
  11. `ListLabels(ctx, reqCtx) (*types.GmailLabelList, error)`
  12. `CreateLabel(ctx, reqCtx, name) (*types.GmailLabel, error)`
  13. `DeleteLabel(ctx, reqCtx, labelID) error`
  14. `ListFilters(ctx, reqCtx) (*types.GmailFilterList, error)`
  15. `CreateFilter(ctx, reqCtx, criteria, action) (*types.GmailFilter, error)` — criteria: from/to/subject/query/hasAttachment; action: addLabelIDs/removeLabelIDs/forward
  16. `DeleteFilter(ctx, reqCtx, filterID) error`
  17. `GetVacation(ctx, reqCtx) (*types.GmailVacationSettings, error)`
  18. `SetVacation(ctx, reqCtx, settings) (*types.GmailVacationSettings, error)`
  19. `ListSendAs(ctx, reqCtx) (*types.GmailSendAsList, error)`
  20. `BatchDelete(ctx, reqCtx, messageIDs) (*types.GmailBatchResult, error)` — Uses gmail.BatchDeleteMessagesRequest. NOTE: requires mail.google.com scope.
  21. `BatchModify(ctx, reqCtx, messageIDs, addLabelIDs, removeLabelIDs) (*types.GmailBatchResult, error)` — Uses gmail.BatchModifyMessagesRequest.
  22. `GetAttachment(ctx, reqCtx, messageID, attachmentID) ([]byte, *types.GmailAttachment, error)` — Returns raw attachment bytes + metadata.

  **Critical implementation details**:
  - All userID parameters use `"me"` (current authenticated user)
  - RFC 2822 message composition: use `net/mail` or manual header construction. Must include MIME headers for HTML emails.
  - base64.URLEncoding (NOT StdEncoding) for Raw field
  - Message.List returns stubs — only ID and ThreadID. Must call Get for full content.
  - Replies: set ThreadId on message, add `In-Reply-To: <messageId>` and `References: <messageId>` headers
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Test approach (TDD)**:
  - Mock the gmail.Service using interface or test helpers
  - Table-driven subtests for each method
  - Test RFC 2822 composition separately (unit test the message builder)
  - Test base64 encoding is URLEncoding not StdEncoding
  - Test header parsing from API response

  **Must NOT do**:
  - Implement watch/Pub-Sub for real-time notifications
  - Implement email tracking
  - Implement history sync
  - Implement delegation/forwarding

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Largest module (22 methods), RFC 2822 composition, base64 encoding edge case, header parsing
  - **Skills**: [`golang-pro`]
    - `golang-pro`: Complex Go patterns — RFC 2822 construction, base64 encoding, Google API pagination, interface mocking

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 11-16)
  - **Blocks**: Task 17 (Gmail CLI)
  - **Blocked By**: Tasks 2 (service factory), 3 (Gmail types)

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — CANONICAL Manager pattern. Study: struct definition, NewManager constructor, ExecuteWithRetry usage, type conversion from API types to domain types.
  - `internal/sheets/manager_test.go` — Test pattern with table-driven subtests.
  - `internal/docs/manager.go` — Alternative pattern with simpler API calls.

  **API/Type References**:
  - `internal/types/gmail.go` (from Task 3) — Domain types to return
  - `internal/api/client.go:67` — `ExecuteWithRetry[T]()` generic function signature

  **External References**:
  - `google.golang.org/api/gmail/v1` — Gmail Go client: `gmail.Service`, `gmail.UsersMessagesService`, `gmail.Message`, `gmail.MessagePartHeader`
  - Gmail API: Messages.List returns `gmail.ListMessagesResponse` with `[]gmail.Message` (stubs — only ID/ThreadID). Must call Messages.Get for full content.
  - Gmail API: Message.Raw is base64url-encoded RFC 2822 message for sending
  - Gmail API: `BatchDeleteMessagesRequest` and `BatchModifyMessagesRequest` are first-class Go methods

  **WHY Each Reference Matters**:
  - `sheets/manager.go` — The EXACT pattern to follow for struct, constructor, and ExecuteWithRetry wrapping
  - `sheets/manager_test.go` — The EXACT test pattern (table-driven subtests, assertions)
  - Gmail API docs on Raw field — MUST use base64.URLEncoding, not StdEncoding, or messages will fail to send

  **Acceptance Criteria**:
  - [ ] `internal/gmail/manager_test.go` exists with tests for all 22 methods
  - [ ] `internal/gmail/manager.go` exists with all 22 methods implemented
  - [ ] `go test -v -race ./internal/gmail/...` passes
  - [ ] RFC 2822 message builder tested (correct headers, MIME, encoding)
  - [ ] base64.URLEncoding used (not StdEncoding) — verified by test

  **QA Scenarios**:

  ```
  Scenario: Gmail manager tests pass
    Tool: Bash
    Preconditions: Types from Task 3 exist
    Steps:
      1. go test -v -race ./internal/gmail/...
    Expected Result: All tests pass, 0 failures
    Failure Indicators: Any FAIL line, compilation errors
    Evidence: .sisyphus/evidence/task-10-gmail-tests.txt

  Scenario: RFC 2822 message composition test
    Tool: Bash
    Steps:
      1. go test -v -run "TestSend\|TestRFC\|TestMessage" ./internal/gmail/...
    Expected Result: Tests verify correct From/To/Subject headers, MIME type, base64url encoding
    Evidence: .sisyphus/evidence/task-10-gmail-rfc2822.txt

  Scenario: Manager struct follows pattern
    Tool: Bash
    Steps:
      1. grep "type Manager struct" internal/gmail/manager.go
      2. grep "func NewManager" internal/gmail/manager.go
      3. grep "ExecuteWithRetry" internal/gmail/manager.go
    Expected Result: Manager struct with client + service fields, NewManager constructor, ExecuteWithRetry in every API method
    Evidence: .sisyphus/evidence/task-10-gmail-pattern.txt
  ```

  **Commit**: YES
  - Message: `feat(gmail): add Gmail manager with TDD tests (22 methods)`
  - Files: `internal/gmail/manager.go`, `internal/gmail/manager_test.go`
  - Pre-commit: `go test -race ./internal/gmail/...`

---

- [ ] 11. Manager + TDD Tests — Calendar (`internal/calendar/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *calendar.Service
  }
  func NewManager(client *api.Client, service *calendar.Service) *Manager
  ```

  **Methods to implement** (~13):
  1. `ListEvents(ctx, reqCtx, calendarID, timeMin, timeMax, maxResults, pageToken, query, singleEvents, orderBy) (*types.CalendarEventList, string, error)` — calendar.Events.List. Support --today/--week/--days via timeMin/timeMax calculation at CLI layer. singleEvents=true expands recurring events.
  2. `GetEvent(ctx, reqCtx, calendarID, eventID) (*types.CalendarEvent, error)` — calendar.Events.Get.
  3. `SearchEvents(ctx, reqCtx, calendarID, query, timeMin, timeMax) (*types.CalendarEventList, string, error)` — Same as ListEvents with Q param. May be merged with ListEvents.
  4. `CreateEvent(ctx, reqCtx, calendarID, summary, description, location, startTime, endTime, attendees, allDay, recurrence, reminders, sendUpdates) (*types.CalendarEventResult, error)` — Handle all-day (Date field) vs timed (DateTime field) via allDay bool. Set EventDateTime.Date for all-day, EventDateTime.DateTime for timed. CRITICAL: these are mutually exclusive.
  5. `UpdateEvent(ctx, reqCtx, calendarID, eventID, updates, sendUpdates) (*types.CalendarEventResult, error)` — Patch event fields. sendUpdates controls notification emails ("all"/"externalOnly"/"none").
  6. `DeleteEvent(ctx, reqCtx, calendarID, eventID, sendUpdates) error` — calendar.Events.Delete.
  7. `RespondToEvent(ctx, reqCtx, calendarID, eventID, response, sendUpdates) (*types.CalendarEventResult, error)` — Update self-attendee's ResponseStatus to "accepted"/"declined"/"tentative". Get event first, find self in attendees, patch ResponseStatus.
  8. `FreeBusy(ctx, reqCtx, calendarIDs, timeMin, timeMax) (*types.CalendarFreeBusy, error)` — calendar.Freebusy.Query.
  9. `FindConflicts(ctx, reqCtx, calendarID, startTime, endTime) ([]types.CalendarConflict, error)` — List events in time range, check for overlaps.

  **Critical implementation details**:
  - Default calendarID is `"primary"` (current user's primary calendar)
  - EventDateTime: all-day uses `.Date = "2026-01-15"`, timed uses `.DateTime = "2026-01-15T10:00:00-05:00"`. NEVER set both.
  - SendUpdates parameter: "all", "externalOnly", "none" — controls whether attendees get email notifications
  - SyncToken for incremental sync (optional, for future use) — not in scope but leave room
  - Recurrence uses RFC 5545 RRULE format (e.g., `RRULE:FREQ=WEEKLY;COUNT=10`)
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Must NOT do**:
  - Implement team calendar management
  - Implement focus-time/OOO/working-location shortcuts
  - Implement propose-time
  - Implement calendar ACL/colors/settings management

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Second largest module, EventDateTime mutual exclusivity, attendee response flow, recurrence handling
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 18 (Calendar CLI)
  - **Blocked By**: Tasks 2, 4

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern
  - `internal/sheets/manager_test.go` — Test pattern

  **External References**:
  - `google.golang.org/api/calendar/v3` — Calendar Go client: `calendar.Service`, `calendar.Event`, `calendar.EventDateTime`, `calendar.FreeBusyRequest`
  - Calendar API: EventDateTime.Date (all-day) vs EventDateTime.DateTime (timed) — mutually exclusive
  - Calendar API: Events.List supports Q (text search), TimeMin, TimeMax, SingleEvents, OrderBy

  **Acceptance Criteria**:
  - [ ] `internal/calendar/manager_test.go` exists with tests for all 9 methods
  - [ ] `internal/calendar/manager.go` exists with all 9 methods
  - [ ] `go test -v -race ./internal/calendar/...` passes
  - [ ] All-day vs timed event creation tested (mutually exclusive fields)

  **QA Scenarios**:

  ```
  Scenario: Calendar manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/calendar/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-11-calendar-tests.txt

  Scenario: EventDateTime mutual exclusivity test
    Tool: Bash
    Steps:
      1. go test -v -run "TestCreate\|TestAllDay\|TestDateTime" ./internal/calendar/...
    Expected Result: Tests verify all-day sets Date (not DateTime), timed sets DateTime (not Date)
    Evidence: .sisyphus/evidence/task-11-calendar-datetime.txt
  ```

  **Commit**: YES
  - Message: `feat(calendar): add Calendar manager with TDD tests (9 methods)`
  - Files: `internal/calendar/manager.go`, `internal/calendar/manager_test.go`
  - Pre-commit: `go test -race ./internal/calendar/...`

- [ ] 12. Manager + TDD Tests — People (`internal/people/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *people.Service
  }
  func NewManager(client *api.Client, service *people.Service) *Manager
  ```

  **Methods to implement** (~9):
  1. `ListContacts(ctx, reqCtx, pageSize, pageToken, sortOrder) (*types.ContactList, string, error)` — people.People.Connections.List on "people/me". CRITICAL: must set PersonFields ("names,emailAddresses,phoneNumbers,organizations,addresses") or API returns 400.
  2. `SearchContacts(ctx, reqCtx, query, pageSize) (*types.ContactList, error)` — people.People.SearchContacts. Also requires readMask (equivalent of PersonFields for search).
  3. `GetContact(ctx, reqCtx, resourceName) (*types.Contact, error)` — people.People.Get. Resource name format: "people/c{id}".
  4. `CreateContact(ctx, reqCtx, givenName, familyName, emails, phones) (*types.ContactResult, error)` — people.People.CreateContact.
  5. `UpdateContact(ctx, reqCtx, resourceName, etag, updates) (*types.ContactResult, error)` — people.People.UpdateContact. CRITICAL: Etag required — API rejects updates without matching etag. updatePersonFields must specify which fields are being updated.
  6. `DeleteContact(ctx, reqCtx, resourceName) error` — people.People.DeleteContact.
  7. `ListOtherContacts(ctx, reqCtx, pageSize, pageToken) (*types.OtherContactList, string, error)` — people.OtherContacts.List. readMask limited to names,emailAddresses,phoneNumbers.
  8. `SearchOtherContacts(ctx, reqCtx, query, pageSize) (*types.OtherContactList, error)` — people.OtherContacts.Search.
  9. `ListDirectory(ctx, reqCtx, pageSize, pageToken, query) (*types.DirectoryPersonList, string, error)` — people.People.ListDirectoryPeople. Requires Workspace account. readMask for directory.
  10. `SearchDirectory(ctx, reqCtx, query, pageSize) (*types.DirectoryPersonList, error)` — people.People.SearchDirectoryPeople.

  **Critical implementation details**:
  - PersonFields / readMask is MANDATORY on every read call. Without it, API returns 400.
  - Default personFields: "names,emailAddresses,phoneNumbers,organizations,addresses"
  - Etag is required for updates — get current etag first, pass it in update
  - Resource name formats: `people/me` (self), `people/c{id}` (contacts by ID)
  - OtherContacts has limited readMask (only names, emailAddresses, phoneNumbers)
  - Directory requires Workspace account — document this limitation

  **Must NOT do**:
  - Implement JSON pipe update (complex batch update pattern)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: PersonFields mandatory parameter, Etag update flow, 3 different contact types
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 19 (People CLI)
  - **Blocked By**: Tasks 2, 5

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern
  - `internal/admin/manager.go` — Pattern for list/get/create/update/delete CRUD operations (similar to contacts)

  **External References**:
  - `google.golang.org/api/people/v1` — People Go client: `people.Service`, `people.Person`, `people.Name`, `people.EmailAddress`
  - People API: PersonFields parameter — comma-separated field names, MANDATORY
  - People API: UpdateContact requires etag + updatePersonFields

  **Acceptance Criteria**:
  - [ ] `internal/people/manager_test.go` exists with tests for all 10 methods
  - [ ] `internal/people/manager.go` exists with all 10 methods
  - [ ] `go test -v -race ./internal/people/...` passes
  - [ ] PersonFields parameter included in every read call (tested)
  - [ ] Etag update flow tested

  **QA Scenarios**:

  ```
  Scenario: People manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/people/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-12-people-tests.txt

  Scenario: PersonFields mandatory parameter tested
    Tool: Bash
    Steps:
      1. go test -v -run "TestList\|TestGet\|TestSearch" ./internal/people/...
    Expected Result: Tests verify PersonFields is set on every read call
    Evidence: .sisyphus/evidence/task-12-people-personfields.txt
  ```

  **Commit**: YES
  - Message: `feat(people): add People/Contacts manager with TDD tests (10 methods)`
  - Files: `internal/people/manager.go`, `internal/people/manager_test.go`
  - Pre-commit: `go test -race ./internal/people/...`

---

- [ ] 13. Manager + TDD Tests — Tasks (`internal/tasks/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *tasks.Service
  }
  func NewManager(client *api.Client, service *tasks.Service) *Manager
  ```

  **Methods to implement** (~10):
  1. `ListTaskLists(ctx, reqCtx, maxResults, pageToken) (*types.TaskListResult, string, error)` — tasks.Tasklists.List.
  2. `CreateTaskList(ctx, reqCtx, title) (*types.TaskList, error)` — tasks.Tasklists.Insert.
  3. `ListTasks(ctx, reqCtx, taskListID, maxResults, pageToken, showCompleted, showHidden, dueMin, dueMax) (*types.TaskResult, string, error)` — tasks.Tasks.List. NOTE: showCompleted defaults to false.
  4. `GetTask(ctx, reqCtx, taskListID, taskID) (*types.Task, error)` — tasks.Tasks.Get.
  5. `AddTask(ctx, reqCtx, taskListID, title, notes, due, parent, previous) (*types.TaskMutationResult, error)` — tasks.Tasks.Insert. Due is date-only string. Parent/Previous for subtask positioning.
  6. `UpdateTask(ctx, reqCtx, taskListID, taskID, updates) (*types.TaskMutationResult, error)` — tasks.Tasks.Update.
  7. `CompleteTask(ctx, reqCtx, taskListID, taskID) (*types.TaskMutationResult, error)` — Convenience: Update task status to "completed" + set completed timestamp.
  8. `UncompleteTask(ctx, reqCtx, taskListID, taskID) (*types.TaskMutationResult, error)` — Set status back to "needsAction", clear completed timestamp.
  9. `DeleteTask(ctx, reqCtx, taskListID, taskID) error` — tasks.Tasks.Delete.
  10. `ClearCompleted(ctx, reqCtx, taskListID) error` — tasks.Tasks.Clear. Removes all completed tasks from a list.

  **Critical implementation details**:
  - Due date: time portion is IGNORED by API. Store/send as "YYYY-MM-DDT00:00:00.000Z" or RFC 3339 with zero time.
  - ShowCompleted defaults to false — completed tasks hidden unless explicitly requested
  - No batch API — each operation is individual
  - 50K requests/day limit — most restrictive of all 7 APIs. Add a comment noting this.
  - Subtask ordering: Use tasks.Tasks.Move with parent/previous parameters
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Must NOT do**:
  - Implement repeat/recurring task schedules (API doesn't support this — it's a client-side feature)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Date handling quirk, show-completed default, subtask ordering
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 20 (Tasks CLI)
  - **Blocked By**: Tasks 2, 6

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern

  **External References**:
  - `google.golang.org/api/tasks/v1` — Tasks Go client: `tasks.Service`, `tasks.TaskList`, `tasks.Task`
  - Tasks API: Due is RFC 3339 but time portion is ignored (date-only)
  - Tasks API: 50K requests/day courtesy limit

  **Acceptance Criteria**:
  - [ ] `internal/tasks/manager_test.go` exists with tests for all 10 methods
  - [ ] `internal/tasks/manager.go` exists with all 10 methods
  - [ ] `go test -v -race ./internal/tasks/...` passes
  - [ ] Due date handling tested (time portion ignored)
  - [ ] Rate limit documented in comments

  **QA Scenarios**:

  ```
  Scenario: Tasks manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/tasks/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-13-tasks-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(tasks): add Tasks manager with TDD tests (10 methods)`
  - Files: `internal/tasks/manager.go`, `internal/tasks/manager_test.go`
  - Pre-commit: `go test -race ./internal/tasks/...`

---

- [ ] 14. Manager + TDD Tests — Forms (`internal/forms/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *forms.Service
  }
  func NewManager(client *api.Client, service *forms.Service) *Manager
  ```

  **Methods to implement** (~3):
  1. `GetForm(ctx, reqCtx, formID) (*types.Form, error)` — forms.Forms.Get. Returns form structure including items/questions.
  2. `ListResponses(ctx, reqCtx, formID, pageSize, pageToken) (*types.FormResponseList, string, error)` — forms.Forms.Responses.List. Answers keyed by questionId NOT title.
  3. `CreateForm(ctx, reqCtx, title) (*types.FormCreateResult, error)` — forms.Forms.Create. NOTE: Only sets title. Questions must be added via BatchUpdate (not in scope for create command). Forms created after March 2026 are unpublished by default — add a comment noting this.

  **Critical implementation details**:
  - Create only sets title — questions require BatchUpdate with WriteControl (out of scope for MVP)
  - Forms created after March 31, 2026 are unpublished by default. Document this in create method comments.
  - Answers in responses are keyed by questionId, not question title
  - WriteControl with RevisionId needed for concurrent edit safety (implement if doing BatchUpdate)
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Must NOT do**:
  - Implement BatchUpdate for adding questions (too complex for MVP, can be added later)
  - Implement form publishing/unpublishing (can be added later)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Only 3 methods, simplest module
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 21 (Forms CLI)
  - **Blocked By**: Tasks 2, 7

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern
  - `internal/docs/manager.go` — Simpler manager with fewer methods (similar scope)

  **External References**:
  - `google.golang.org/api/forms/v1` — Forms Go client: `forms.Service`, `forms.Form`, `forms.ListFormResponsesResponse`
  - Forms API: Create returns form with only title set, questions via BatchUpdate only
  - Forms API: March 2026 breaking change — forms unpublished by default

  **Acceptance Criteria**:
  - [ ] `internal/forms/manager_test.go` exists with tests for all 3 methods
  - [ ] `internal/forms/manager.go` exists with all 3 methods
  - [ ] `go test -v -race ./internal/forms/...` passes
  - [ ] March 2026 note documented in create method

  **QA Scenarios**:

  ```
  Scenario: Forms manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/forms/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-14-forms-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(forms): add Forms manager with TDD tests (3 methods)`
  - Files: `internal/forms/manager.go`, `internal/forms/manager_test.go`
  - Pre-commit: `go test -race ./internal/forms/...`

---

- [ ] 15. Manager + TDD Tests — Apps Script (`internal/appscript/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *script.Service
  }
  func NewManager(client *api.Client, service *script.Service) *Manager
  ```

  **Methods to implement** (~4):
  1. `GetProject(ctx, reqCtx, scriptID) (*types.ScriptProject, error)` — script.Projects.Get.
  2. `GetContent(ctx, reqCtx, scriptID) (*types.ScriptContent, error)` — script.Projects.GetContent. Returns all script files with source code.
  3. `CreateProject(ctx, reqCtx, title, parentID) (*types.ScriptCreateResult, error)` — script.Projects.Create. parentID is optional Drive folder.
  4. `RunFunction(ctx, reqCtx, scriptID, functionName, parameters) (*types.ScriptRunResult, error)` — script.Scripts.Run. EXPERIMENTAL: requires script deployed as executable, caller needs all scopes script uses, script and caller must share GCP project, 6-min timeout. Parameters must be primitives (string, number, bool) — no objects/arrays.

  **Critical implementation details**:
  - Script ID ≠ Drive file ID. Scripts have their own ID system.
  - Run requires: (a) script deployed as API executable, (b) shared GCP project, (c) caller has all scopes the script uses, (d) parameters are primitives only
  - Run has 6-minute timeout (set context timeout)
  - Mark RunFunction as experimental with prominent warnings in code comments
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Must NOT do**:
  - Present Run as a production-ready feature — mark experimental
  - Implement script deployment management
  - Implement script version management

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Run command has complex constraints (GCP project, scopes, deployment, timeout)
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 22 (Apps Script CLI)
  - **Blocked By**: Tasks 2, 8

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern

  **External References**:
  - `google.golang.org/api/script/v1` — Script Go client: `script.Service`, `script.Project`, `script.Content`, `script.ExecutionRequest`
  - Apps Script API: Run requires deployment, shared GCP project, all script scopes on caller
  - Apps Script API: 6-minute execution timeout

  **Acceptance Criteria**:
  - [ ] `internal/appscript/manager_test.go` exists with tests for all 4 methods
  - [ ] `internal/appscript/manager.go` exists with all 4 methods
  - [ ] `go test -v -race ./internal/appscript/...` passes
  - [ ] RunFunction marked as experimental in code comments
  - [ ] GCP project constraint documented

  **QA Scenarios**:

  ```
  Scenario: AppScript manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/appscript/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-15-appscript-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(appscript): add Apps Script manager with TDD tests (4 methods, run experimental)`
  - Files: `internal/appscript/manager.go`, `internal/appscript/manager_test.go`
  - Pre-commit: `go test -race ./internal/appscript/...`

---

- [ ] 16. Manager + TDD Tests — Groups (`internal/groups/`)

  **What to do**:
  TDD approach: Write tests FIRST, then implement.

  **Manager struct**:
  ```go
  type Manager struct {
      client  *api.Client
      service *cloudidentity.Service
  }
  func NewManager(client *api.Client, service *cloudidentity.Service) *Manager
  ```

  **Methods to implement** (~2):
  1. `ListGroups(ctx, reqCtx, parent, pageSize, pageToken) (*types.CloudIdentityGroupList, string, error)` — cloudidentity.Groups.List. Parent format: `customers/{customerId}`. Returns groups with resource names.
  2. `ListMembers(ctx, reqCtx, groupName, pageSize, pageToken) (*types.CloudIdentityMemberList, string, error)` — cloudidentity.Groups.Memberships.List. groupName format: `groups/{groupId}`.

  **Critical implementation details**:
  - This is Cloud Identity API, NOT Admin SDK Directory API
  - Existing `gdrv admin groups` uses Admin SDK — this `gdrv groups` uses Cloud Identity
  - Parent parameter for ListGroups must be `customers/{customerId}` format
  - Groups.Create returns an async Operation (LRO) — not implementing create/delete in this scope
  - Labels on groups determine type (e.g., `cloudidentity.googleapis.com/groups.discussion_forum`)
  - All calls wrapped in `api.ExecuteWithRetry[T]()`

  **Must NOT do**:
  - Implement group create/update/delete (async Operations are complex, out of MVP scope)
  - Conflict with admin groups commands

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Only 2 methods, simplest module
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 23 (Groups CLI)
  - **Blocked By**: Tasks 2, 9

  **References**:
  **Pattern References**:
  - `internal/sheets/manager.go` — Canonical Manager pattern
  - `internal/admin/manager.go` — Similar list/get pattern for Admin groups (reference for contrast)

  **External References**:
  - `google.golang.org/api/cloudidentity/v1` — Cloud Identity Go client: `cloudidentity.Service`, `cloudidentity.Group`, `cloudidentity.Membership`
  - Cloud Identity API: Parent format `customers/{customerId}`, group resource name `groups/{groupId}`

  **Acceptance Criteria**:
  - [ ] `internal/groups/manager_test.go` exists with tests for both methods
  - [ ] `internal/groups/manager.go` exists with both methods
  - [ ] `go test -v -race ./internal/groups/...` passes
  - [ ] No conflicts with admin groups types or commands

  **QA Scenarios**:

  ```
  Scenario: Groups manager tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/groups/...
    Expected Result: All tests pass
    Evidence: .sisyphus/evidence/task-16-groups-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(groups): add Cloud Identity Groups manager with TDD tests (2 methods)`
  - Files: `internal/groups/manager.go`, `internal/groups/manager_test.go`
  - Pre-commit: `go test -race ./internal/groups/...`

- [ ] 17. CLI Commands — Gmail (kong structs, ~18 commands)

  **What to do**:
  - Create `internal/cli/gmail.go` with kong struct-tag CLI commands
  - Define `GmailCmd` with nested subcommand groups:
    ```
    GmailCmd → Search, ThreadGet, MessageGet, Send
             → Drafts → List, Get, Create, Update, Delete, Send
             → Labels → List, Create, Delete
             → Filters → List, Create, Delete
             → Vacation → Get, Set
             → SendAs → List
             → BatchDelete, BatchModify
             → Attachments → Get
    ```
  - Each subcommand struct has:
    - Flag fields with kong tags (`help:""`, `name:""`, `required:""`, `default:""`)
    - Positional arg fields with `arg:""` tags
    - `Run(globals *Globals) error` method that:
      1. Calls `globals.ToGlobalFlags()` to get `types.GlobalFlags`
      2. Creates auth manager, gets credentials
      3. Gets Gmail service via `authManager.GetGmailService(ctx, creds)`
      4. Creates `gmail.NewManager(apiClient, gmailService)`
      5. Calls appropriate manager method
      6. Outputs result via OutputWriter

  **Key flags per command**:
  - Search: `--query` (required), `--limit`, `--page-token`, `--paginate`
  - Send: `--to` (required), `--subject` (required), `--body`, `--html`, `--cc`, `--bcc`, `--reply-to` (message ID for replies), `--thread`
  - Drafts create: `--to`, `--subject`, `--body`
  - Labels create: `--name` (required)
  - Filters create: `--from`, `--to`, `--subject`, `--query`, `--has-attachment`, `--add-labels`, `--remove-labels`, `--forward`
  - Vacation set: `--enable`, `--subject`, `--body`, `--html-body`, `--start`, `--end`
  - BatchDelete: `--ids` (required, comma-separated)
  - BatchModify: `--ids` (required), `--add-labels`, `--remove-labels`
  - Attachments get: message-id (arg), `--attachment-id` (required), `--output` (file path)

  - Wire `GmailCmd` into root CLI struct

  **Must NOT do**:
  - Use cobra — kong only
  - Add business logic — delegate entirely to manager methods

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: ~18 commands with various flag configurations, nested subgroups
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 18-23)
  - **Blocks**: F1-F4
  - **Blocked By**: Task 10 (Gmail manager), kong migration Task 6+ (pattern must exist)

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong-migration) — The kong CLI pattern to follow. If kong migration hasn't completed yet, reference `.sisyphus/plans/cobra-to-kong-migration.md` Task 6 for the pattern definition.
  - `internal/cli/sheets.go` (post-kong-migration) — Nested subcommand pattern (sheets values → get, update, append, clear)
  - `internal/cli/admin.go` (post-kong-migration) — 3-level nesting pattern (admin → users/groups → members)

  **API/Type References**:
  - `internal/gmail/manager.go` (from Task 10) — Manager methods to call from Run()
  - `internal/types/gmail.go` (from Task 3) — Types returned by manager

  **Acceptance Criteria**:
  - [ ] `bin/gdrv gmail --help` lists all subcommand groups
  - [ ] `bin/gdrv gmail search --help` shows --query, --limit, --page-token flags
  - [ ] `bin/gdrv gmail send --help` shows --to, --subject, --body, --html flags
  - [ ] `bin/gdrv gmail drafts --help` lists draft subcommands
  - [ ] `bin/gdrv gmail labels --help` lists label subcommands
  - [ ] `go build ./cmd/gdrv` succeeds
  - [ ] No cobra imports in gmail.go

  **QA Scenarios**:

  ```
  Scenario: Gmail command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv gmail --help
      3. bin/gdrv gmail search --help
      4. bin/gdrv gmail drafts --help
      5. bin/gdrv gmail labels --help
      6. bin/gdrv gmail filters --help
      7. bin/gdrv gmail vacation --help
    Expected Result: All subgroups and commands listed with correct flags
    Evidence: .sisyphus/evidence/task-17-gmail-cli.txt

  Scenario: Gmail CLI compiles with no cobra
    Tool: Bash
    Steps:
      1. grep "spf13/cobra" internal/cli/gmail.go || echo "CLEAN"
    Expected Result: "CLEAN" — no cobra imports
    Evidence: .sisyphus/evidence/task-17-gmail-no-cobra.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Gmail CLI commands (kong, 18 commands)`
  - Files: `internal/cli/gmail.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 18. CLI Commands — Calendar (kong structs, ~13 commands)

  **What to do**:
  - Create `internal/cli/calendar.go` with kong structs:
    ```
    CalendarCmd → Events → List, Get, Search, Create, Update, Delete, Respond
               → FreeBusy
               → Conflicts
    ```
  - Key flags:
    - Events list: `--calendar` (default "primary"), `--today` (bool, sets timeMin/Max to today), `--week` (bool), `--days` (int, number of days ahead), `--limit`, `--page-token`, `--paginate`, `--query`
    - Events create: `--calendar`, `--summary` (required), `--start` (required), `--end` (required), `--description`, `--location`, `--attendees` (comma-separated emails), `--all-day` (bool), `--recurrence`, `--reminder` (e.g., "email:10,popup:5"), `--send-updates` (all/externalOnly/none)
    - Events respond: event-id (arg), `--response` (required: accepted/declined/tentative), `--send-updates`
    - FreeBusy: `--calendars` (comma-separated, default "primary"), `--start` (required), `--end` (required)
    - Conflicts: `--calendar`, `--start` (required), `--end` (required)
  - CLI layer converts `--today`/`--week`/`--days` to timeMin/timeMax before calling manager
  - Wire `CalendarCmd` into root CLI struct

  **Must NOT do**:
  - Use cobra
  - Implement team calendar, ACL, colors, settings

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 11, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/calendar/manager.go` (from Task 11) — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv calendar --help` lists all subcommands
  - [ ] `bin/gdrv calendar events list --help` shows --today, --week, --days, --calendar flags
  - [ ] `bin/gdrv calendar events create --help` shows --summary, --start, --end as required
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Calendar command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv calendar --help
      3. bin/gdrv calendar events --help
      4. bin/gdrv calendar events create --help
    Expected Result: All commands and flags listed correctly
    Evidence: .sisyphus/evidence/task-18-calendar-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Calendar CLI commands (kong, 13 commands)`
  - Files: `internal/cli/calendar.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 19. CLI Commands — People (kong structs, ~9 commands)

  **What to do**:
  - Create `internal/cli/people.go` with kong structs:
    ```
    PeopleCmd → Contacts → List, Search, Get, Create, Update, Delete
             → OtherContacts → List, Search
             → Directory → List, Search
    ```
  - Key flags:
    - Contacts list: `--limit`, `--page-token`, `--paginate`, `--sort` (FIRST_NAME_ASCENDING/LAST_NAME_ASCENDING)
    - Contacts search: `--query` (required), `--limit`
    - Contacts get: resource-name (arg)
    - Contacts create: `--given-name` (required), `--family-name`, `--email`, `--phone`, `--organization`, `--title`
    - Contacts update: resource-name (arg), `--given-name`, `--family-name`, `--email`, `--phone` (at least one required)
    - Contacts delete: resource-name (arg)
    - Directory list/search: `--limit`, `--page-token`, `--query`
  - Wire `PeopleCmd` into root CLI struct

  **Must NOT do**:
  - Use cobra
  - Implement JSON pipe update

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 12, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/people/manager.go` (from Task 12) — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv people --help` lists contacts, other-contacts, directory
  - [ ] `bin/gdrv people contacts create --help` shows --given-name as required
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: People command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv people --help
      3. bin/gdrv people contacts --help
      4. bin/gdrv people directory --help
    Expected Result: All subgroups and commands listed
    Evidence: .sisyphus/evidence/task-19-people-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add People/Contacts CLI commands (kong, 10 commands)`
  - Files: `internal/cli/people.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 20. CLI Commands — Tasks (kong structs, ~10 commands)

  **What to do**:
  - Create `internal/cli/tasks.go` with kong structs:
    ```
    TasksCmd → Lists → List, Create
            → Tasks → List, Get, Add, Update, Done, Undo, Delete
            → Clear
    ```
  - Key flags:
    - Lists list: `--limit`, `--page-token`, `--paginate`
    - Lists create: `--title` (required)
    - Tasks list: `--list` (task list ID, default "@default"), `--show-completed`, `--show-hidden`, `--due-min`, `--due-max`, `--limit`, `--page-token`, `--paginate`
    - Tasks add: `--list`, `--title` (required), `--notes`, `--due` (date: YYYY-MM-DD), `--parent` (subtask), `--previous` (ordering)
    - Tasks done: task-id (arg), `--list`
    - Tasks undo: task-id (arg), `--list`
    - Clear: `--list` (task list ID, default "@default")
  - Wire `TasksCmd` into root CLI struct
  - NOTE: Use `TasksCmd` to avoid conflict with the `tasks` Go package name

  **Must NOT do**:
  - Use cobra
  - Implement repeat/recurring schedules (not supported by API)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 13, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/tasks/manager.go` (from Task 13) — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv tasks --help` lists all subcommands
  - [ ] `bin/gdrv tasks add --help` shows --title as required, --due accepts YYYY-MM-DD
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Tasks command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv tasks --help
      3. bin/gdrv tasks lists --help
      4. bin/gdrv tasks add --help
    Expected Result: All commands listed with correct flags
    Evidence: .sisyphus/evidence/task-20-tasks-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Tasks CLI commands (kong, 10 commands)`
  - Files: `internal/cli/tasks.go`
  - Pre-commit: `go build ./cmd/gdrv`

- [ ] 21. CLI Commands — Forms (kong structs, ~3 commands)

  **What to do**:
  - Create `internal/cli/forms.go` with kong structs:
    ```
    FormsCmd → Get, Responses, Create
    ```
  - Key flags:
    - Get: form-id (arg)
    - Responses: form-id (arg), `--limit`, `--page-token`, `--paginate`
    - Create: `--title` (required)
  - Wire `FormsCmd` into root CLI struct

  **Must NOT do**:
  - Use cobra
  - Add question management (BatchUpdate)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Only 3 commands, simple structure
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 14, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/forms/manager.go` (from Task 14) — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv forms --help` lists get, responses, create
  - [ ] `bin/gdrv forms create --help` shows --title as required
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Forms command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv forms --help
    Expected Result: Lists get, responses, create commands
    Evidence: .sisyphus/evidence/task-21-forms-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Forms CLI commands (kong, 3 commands)`
  - Files: `internal/cli/forms.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 22. CLI Commands — Apps Script (kong structs, ~4 commands)

  **What to do**:
  - Create `internal/cli/appscript.go` with kong structs:
    ```
    AppScriptCmd → Get, Content, Create, Run
    ```
  - Key flags:
    - Get: script-id (arg)
    - Content: script-id (arg)
    - Create: `--title` (required), `--parent` (Drive folder ID, optional)
    - Run: script-id (arg), `--function` (required), `--parameters` (JSON array string, optional). Help text MUST include prominent warning about GCP project constraint and deployment requirement.
  - Wire `AppScriptCmd` into root CLI struct (command name: `appscript`)

  **Must NOT do**:
  - Use cobra
  - Hide the Run command — it should be visible but clearly marked experimental in help text

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Only 4 commands
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 15, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/appscript/manager.go` (from Task 15) — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv appscript --help` lists get, content, create, run
  - [ ] `bin/gdrv appscript run --help` shows --function as required AND includes GCP project warning
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: AppScript command tree
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv appscript --help
      3. bin/gdrv appscript run --help
    Expected Result: All commands listed. Run help includes experimental/GCP project warning text.
    Evidence: .sisyphus/evidence/task-22-appscript-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Apps Script CLI commands (kong, 4 commands, run experimental)`
  - Files: `internal/cli/appscript.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 23. CLI Commands — Groups (kong structs, ~2 commands)

  **What to do**:
  - Create `internal/cli/groups.go` with kong structs:
    ```
    GroupsCmd → List, Members
    ```
  - Key flags:
    - List: `--customer` (required — customer ID), `--limit`, `--page-token`, `--paginate`
    - Members: group-name (arg — resource name like "groups/{groupId}"), `--limit`, `--page-token`, `--paginate`
  - Help text must clarify this is Cloud Identity Groups (not Admin SDK groups)
  - Wire `GroupsCmd` into root CLI struct (command name: `groups`)

  **Must NOT do**:
  - Use cobra
  - Conflict with existing `admin groups` command — this is a SEPARATE command path
  - Implement group create/delete (async operations out of scope)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Only 2 commands, simplest CLI module
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: Task 16, kong migration Task 6+

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` (post-kong) — Kong CLI pattern
  - `internal/groups/manager.go` (from Task 16) — Manager methods
  - `internal/cli/admin.go` — Existing admin groups commands (for contrast — different API, different command path)

  **WHY Each Reference Matters**:
  - `admin.go` — Shows existing `admin groups` structure. New `groups` command must NOT overlap in command path or confuse users. Help text should clarify the distinction.

  **Acceptance Criteria**:
  - [ ] `bin/gdrv groups --help` lists list and members
  - [ ] `bin/gdrv groups list --help` shows --customer as required
  - [ ] Help text mentions Cloud Identity (not Admin SDK)
  - [ ] `bin/gdrv admin groups --help` still works separately
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Groups command tree and distinction from admin groups
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv groups --help
      3. bin/gdrv admin groups --help
    Expected Result: Both work independently. groups help mentions "Cloud Identity", admin groups help mentions "Admin SDK" or "Directory API".
    Evidence: .sisyphus/evidence/task-23-groups-cli.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add Cloud Identity Groups CLI commands (kong, 2 commands)`
  - Files: `internal/cli/groups.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.sisyphus/evidence/`. Compare deliverables against plan. Verify all 7 modules exist and build. Verify scope presets are registered. Verify service factory has all 7 entries.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go build ./...`, `go vet ./...`, `go test -race ./...`. Review all new files for: consistency with existing module patterns (Manager struct, NewManager constructor, ExecuteWithRetry usage), proper error handling, no dead code, no unused imports. Check AI slop: excessive comments, over-abstraction, generic names (data/result/item/temp). Verify all types implement Tabular interface.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Run `go build -o bin/gdrv ./cmd/gdrv`. Test every new top-level command (`gmail`, `calendar`, `people`, `tasks`, `forms`, `appscript`, `groups`). Verify `--help` output for every command and subcommand. Verify `--json` flag works. Test that auth errors return proper JSON envelope (not panics). Test that help text is consistent in style with existing commands. Save evidence.
  Output: `Commands [N/N pass] | Help [N/N] | JSON [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual implementation. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance across ALL files. Detect any modifications to existing modules (files, sheets, docs, etc.) — reject if found. Verify command count matches plan (~59 total). Verify no cobra imports in new files.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `feat(auth): add scope constants, presets, and service factory for 7 new APIs` — constants.go, service_factory.go, manager.go
- **Wave 2**: `feat(types): add domain types for gmail, calendar, people, tasks, forms, appscript, groups` — internal/types/*.go
- **Wave 3** (per module): `feat(<module>): add <module> manager with TDD tests` — internal/<module>/manager.go, manager_test.go
- **Wave 4** (per module): `feat(cli): add <module> CLI commands (kong)` — internal/cli/<module>.go
- **Final**: `chore: vendor-gogcli-apis complete — 7 new API modules`

---

## Success Criteria

### Verification Commands
```bash
go build -o bin/gdrv ./cmd/gdrv                    # Expected: successful build
go test -v -race ./internal/gmail/...               # Expected: all tests pass
go test -v -race ./internal/calendar/...            # Expected: all tests pass
go test -v -race ./internal/people/...              # Expected: all tests pass
go test -v -race ./internal/tasks/...               # Expected: all tests pass
go test -v -race ./internal/forms/...               # Expected: all tests pass
go test -v -race ./internal/appscript/...           # Expected: all tests pass
go test -v -race ./internal/groups/...              # Expected: all tests pass
go vet ./...                                        # Expected: no issues
bin/gdrv gmail --help                               # Expected: lists all Gmail subcommands
bin/gdrv calendar --help                            # Expected: lists all Calendar subcommands
bin/gdrv people --help                              # Expected: lists all People subcommands
bin/gdrv tasks --help                               # Expected: lists all Tasks subcommands
bin/gdrv forms --help                               # Expected: lists all Forms subcommands
bin/gdrv appscript --help                           # Expected: lists all Apps Script subcommands
bin/gdrv groups --help                              # Expected: lists all Groups subcommands
```

### Final Checklist
- [ ] All 7 modules build and test cleanly
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass (TDD — written before implementation)
- [ ] No modifications to existing modules
- [ ] All types implement Tabular interface
- [ ] All API calls use ExecuteWithRetry
- [ ] All CLI commands use kong struct patterns
- [ ] Scope presets registered and documented
- [ ] Service factory handles all 7 new service types
