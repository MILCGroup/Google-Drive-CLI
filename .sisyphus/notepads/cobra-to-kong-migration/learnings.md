# Cobra-to-Kong Migration Learnings

## [2026-02-20] Task 4: Helper file cobra decoupling

### Summary
Verified framework decoupling for helper files that will be called from kong `Run()` methods.

### Files Analyzed
1. **internal/cli/output.go** - CLEAN ✅
   - No cobra imports
   - No pflag imports
   - Pure Go types: OutputWriter, OutputFormat, CLIOutput
   - Safe to call from kong Run() methods

2. **internal/cli/drive_helpers.go** - CLEAN ✅
   - No cobra imports
   - No pflag imports
   - Single helper function: convertDriveFile()
   - Converts google.golang.org/api/drive/v3 types to internal types
   - Safe to call from kong Run() methods

3. **internal/cli/root.go** - COBRA DEPENDENT (expected)
   - Contains cobra.Command definitions
   - Helper functions (ResolveFileID, GetPathResolver, GetResolveOptions) use plain Go types
   - These helpers are framework-agnostic and can be extracted if needed

### Key Findings
- **output.go**: Uses only standard library + internal types. No framework dependencies.
- **drive_helpers.go**: Minimal, pure conversion function. No framework dependencies.
- **root.go**: Cobra-dependent at command level, but helper functions are decoupled.

### Build Status
- `go build ./...` ✅ SUCCESS
- `go test -race ./internal/cli/...` ✅ PASS
- No cobra imports in helper files ✅

### Implications for Migration
These files are **safe to call from kong Run() methods** without any refactoring needed. They have zero framework coupling and can be used directly in the new kong-based CLI structure.

### Next Steps
- Tasks 5-16 can safely import and use these helpers
- No circular dependency risks
- No refactoring required for these files
## [2026-02-20] Task 2: Kong dependency added
- Kong version: v1.14.0
- go build ./... passes with both cobra and kong present
- No import conflicts
- All existing tests pass with -race flag
- Kong added as indirect dependency (not yet imported in code)

## [2026-02-20] Task 3: Kong root foundation
- Kong version in use: v1.14.0
- `Globals.AfterApply()` is the kong replacement for cobra `PersistentPreRunE`: it normalizes `--json`, validates output format, and initializes logger before command `Run()` executes.
- `Globals.Logger logging.Logger` with `kong:"-"` is internal-only (not a flag) and is available to command `Run()` methods via `ctx.Run(&c.Globals)` binding.
- `Globals.ToGlobalFlags()` keeps manager-layer compatibility by converting kong-parsed values to `types.GlobalFlags`.
- `internal/cli/stubs.go` allows incremental migration: root CLI references all top-level commands immediately, while each domain can replace its stub type without repeated edits to `root.go`.
- Issues encountered and resolved: missing `go.sum` entry for kong after first import (`go get github.com/alecthomas/kong@v1.14.0` fixed it), and local `gopls` not on PATH for diagnostics tooling.

## [2026-02-20] Task 6: Files domain migration
- Full files domain migrated from cobra command vars/functions to kong struct commands with `Run(globals *Globals) error` methods.
- `getFileManager()` stayed reusable and unchanged; all command methods now call `globals.ToGlobalFlags()` and existing manager logic.
- Kong v1.14 does not allow a command node to mix positional args and subcommands. Revisions list moved to `files revisions list <file-id>` using `default:"withargs"` to preserve `files revisions <file-id>` behavior.
- Kong v1.14 also rejects duplicate flag names across inherited global flags and command-local flags. Local `--output`/`--force` had to be renamed to avoid runtime parser panic.


## [2026-02-20] Task: Auth domain migration to kong
- Migrated `internal/cli/auth.go` from cobra command vars/init wiring to kong struct commands with `Run(globals *Globals) error` methods.
- Preserved command set under `auth`: login, logout, service-account, status, device, profiles, diagnose.
- Preserved client override edge-case semantics by switching `--client-id`/`--client-secret` to `*string` fields and checking nil in `resolveOAuthClient`.
- Converted scope resolution helper to explicit inputs (`preset`, `wide`, `scopes`) so command logic no longer relies on global mutable flag vars.
- Removed temporary `AuthCmd` stub in `internal/cli/stubs.go`; root now binds to real kong auth command tree.
- Verification passed: `go build -o bin/gdrv ./cmd/gdrv`, `bin/gdrv auth --help`, `bin/gdrv auth service-account --help`, and `go test -race ./internal/cli/...`.

## permissions.go migration (2026-02-20)

### Patterns confirmed
- `getPermissionManager()` helper signature changed from `() (*permissions.Manager, error)` to `(flags types.GlobalFlags) (*permissions.Manager, error)` — removes the `GetGlobalFlags()` internal call, consistent with how `getFileManager(ctx, flags)` works in files.go
- Nested subcommand groups (`PermAuditCmd`, `PermBulkCmd`) require their own struct with `cmd:""` tagged fields — no `Run()` method needed on the group struct itself
- `cmd:"create-link"` / `cmd:"remove-public"` / `cmd:"update-role"` / `cmd:"anyone-with-link"` override the default field-name-derived command name for hyphenated commands
- `default:"true"` works for bool fields in kong (send-notification flag)
- `default:"reader"` works for string fields in kong (role default on create-link)
- Positional args on nested commands work fine: `PermAuditUserCmd.Email string \`arg:""\``
- No flag collisions found between permissions domain flags and the 18 global flags

### Key delta from cobra
- Remove: all `var permXxxCmd = &cobra.Command{...}`, global flag vars, entire `init()`, all `func runPermXxx(...)` 
- Add: kong struct types + `Run(globals *Globals) error` methods
- stubs.go: remove the 5-line `PermissionsCmd` stub (type + Run)

### File size: 676 lines cobra → ~390 lines kong (42% reduction)

## Admin domain migration (2026-02-20)

### 3-level nesting pattern
Kong handles 3-level nesting (admin users list, admin groups members add) with nested struct embedding:
- `AdminCmd` → `AdminUsersCmd` / `AdminGroupsCmd` → leaf command structs
- `AdminGroupsCmd` embeds `AdminGroupsMembersCmd` for the 3rd level

### Stub cleanup
Previous migrations (permissions) may have already removed stubs from stubs.go. Check `git show <prev-commit> --stat` to verify what was already cleaned up before adding to stubs.go in git add.

### required:"" tag for positional args
For cobra MarkFlagRequired equivalents in kong, use `required:""` on struct fields.
Admin create commands use: `GivenName string ... required:""`

### No flag collisions in admin domain
All admin flags (domain, customer, query, limit, page-token, fields, paginate, order-by, given-name, family-name, password, suspended, org-unit-path, description, name, roles, role) are safe — none collide with globals.

### getAdminService helper
Kept verbatim from cobra version — only change is removal of cobra import; function signature and body unchanged.

## labels.go migration (8 commands, 2026-02-20)

### Nested subcommand pattern
`LabelsFileCmd` is a nested command group (not a leaf). It embeds 4 sub-structs with `cmd:""` tags — exactly like top-level `LabelsCmd`. Kong handles arbitrary nesting transparently.

### Flag scope isolation
In cobra, shared global vars (`labelsView`, `labelsFields`, etc.) were reused across multiple commands with identical flag names. In kong, each struct has its own fields — no sharing needed, no collision risk within a command.

### `--fields` dual meaning
`labelsFileApplyCmd` and `labksFileUpdateCmd` both used `--fields` to mean "JSON field values" while list/get commands used `--fields` for "API fields to return". In kong structs this is fine since each struct has its own `Fields string` field with different help text — no renaming required.

### No global flag collisions for labels domain
None of the labels flags (`customer`, `view`, `minimum-role`, `published-only`, `limit`, `page-token`, `fields`, `use-admin-access`, `language-code`, `hide-in-search`, `show-in-apply`) collide with globals.

### Stub removal: always remove before/alongside the rewrite
The `LabelsCmd redeclared` LSP error appeared immediately after the edit — the stub in `stubs.go` must be removed in the same change set. Order: edit labels.go first, then remove stub, then build.

## sync domain migration (2026-02-20)

### Cobra parent-with-RunE pattern → kong
The cobra `syncCmd` was both a parent container AND had its own `RunE` for bidirectional sync. In kong, the parent struct (`SyncCmd`) is a pure container; the bidirectional RunE was dropped. The 6 leaf subcommands (init, push, pull, status, list, remove) map cleanly.

### Shared helper refactoring
`runSyncWithMode` previously called `GetGlobalFlags()` internally and used package-level global flag vars (syncDelete, syncConflict, etc.). In kong, the signature became:
```go
func runSyncWithMode(flags types.GlobalFlags, configID string, mode diff.Mode, command string, planOnly bool, delete bool, conflictStr string, concurrency int, useChanges bool) error
```
Each `Run(globals *Globals)` method calls `globals.ToGlobalFlags()` then passes both flags and its own struct field values to the helper.

### SyncStatusCmd has no concurrency field
`syncStatusCmd` in cobra didn't register `--concurrency`, so the global was 0 for status runs. In kong, `SyncStatusCmd` has no `Concurrency` field and passes `0` to `runSyncWithMode`.

### No flag collisions
All sync-domain flags (exclude, conflict, direction, id, delete, concurrency, use-changes) are clear of all global flags.

## chat.go migration (14 commands, 3-level nesting)

### Pattern used
- `ChatCmd` → `ChatSpacesCmd`, `ChatMessagesCmd`, `ChatMembersCmd` (3-level nesting)
- Positional cobra `args[0]`, `args[1]` → kong `arg:""` struct fields
- All chat flags (limit, page-token, paginate, display-name, type, external-users, filter, text, thread, email, role) — NO collisions with global flags
- `getChatService()` helper preserved verbatim (just removed cobra import)
- `globals.ToGlobalFlags()` replaces `GetGlobalFlags()`

### Cobra → Kong arg mapping
- `args[0]` → `SpaceID string \`arg:""\``
- `args[0], args[1]` → `SpaceID string \`arg:""\`` + `MessageID/MemberID string \`arg:""\``

### stubs.go state
- Previous migrations (sheets, sync) already removed their stubs from the committed stubs.go
- Working tree had ChatCmd stub reverted; my edit fixed it back to match HEAD
- ChatCmd removal was needed to fix the working tree state, even though HEAD already lacked it

## Sheets Domain Migration (10 commands)

### Nested Values Subgroup
The `sheets values` subgroup requires `SheetsValuesCmd` as an intermediate struct with sub-commands as fields. Kong handles nested subcommands naturally when a struct field is tagged with `cmd:""`.

### Helper Function Refactoring
Old cobra pattern used global package-level variables (`sheetsValuesJSON`, etc.) that cobra flags bound to. The new kong pattern puts these as struct fields, requiring helper functions to be refactored from zero-arg (`readSheetValues()`) to parameterized (`readSheetValuesFrom(json, file string)`).

### Test Updates Required
When refactoring from global variable helpers to parameterized helpers, `*_test.go` files must be updated to pass arguments directly instead of setting globals. The test logic stays identical, just the calling convention changes.

### Range as Struct Field
`Range` is safe to use as a Go struct field name (not a keyword). Kong's `arg:""` tag works cleanly with it for positional args.

### Files Changed
- `internal/cli/sheets.go` — full rewrite (517 → ~430 lines)
- `internal/cli/stubs.go` — removed SheetsCmd stub
- `internal/cli/sheets_test.go` — updated to use parameterized helpers

## slides.go + docs.go migration (2026-02-20)

### Pattern applied
- Both files followed identical structure: List/Get/Read/Create/Update (Docs) + Replace (Slides)
- Helper functions that read global vars (`readSlidesRequests`, `readSlidesReplacements`, `readDocsRequests`) were renamed to `parse*` and refactored to accept explicit parameters instead of package-level globals
- No flag name collisions with globals were found for either domain

### Test file updates required
- `slides_test.go` and `docs_test.go` referenced old global vars (`slidesUpdateRequests`, etc.) and old function names
- Updated both test files to call `parseSlidesRequests(json, file)` / `parseSlidesReplacements(data, file)` / `parseDocsRequests(json, file)` with explicit params — no global state needed → simpler, thread-safe tests

### Stubs.go state after this migration
- `DocsCmd` and `SlidesCmd` stubs removed
- Remaining stubs: `FoldersCmd`, `DrivesCmd`, `CompletionCmd`
