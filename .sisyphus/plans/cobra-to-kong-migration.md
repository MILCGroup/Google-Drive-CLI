# Cobra → Kong CLI Framework Migration

## TL;DR

> **Quick Summary**: Full migration of gdrv's CLI layer from spf13/cobra (imperative builder pattern) to alecthomas/kong (declarative struct-tag pattern). Rewrites all 19 CLI files (~140 commands, ~261 flags) to use kong structs, replaces global state with DI, adds shell completion, and expands test coverage.
>
> **Deliverables**:
> - All `internal/cli/*.go` files rewritten for kong
> - `cmd/gdrv/main.go` rewritten with `kong.Parse` + `ctx.Run`
> - Global state (`globalFlags`, `logger`) eliminated in favor of `kong.Bind()` DI
> - Shell completion support (bash/zsh/fish) via carapace or kong-completion
> - All existing tests migrated + parse-level tests added for all command structs
> - cobra/pflag dependencies fully removed from go.mod
>
> **Estimated Effort**: Large (3-5 days with parallel execution)
> **Parallel Execution**: YES — 4 waves, up to 10 parallel tasks
> **Critical Path**: Task 1 → Task 3 → Task 6 → Task 7-16 → Task 17-20 → F1-F4

---

## Context

### Original Request
Compare spf13/cobra and alecthomas/kong CLI frameworks, evaluate migration feasibility for the gdrv Google Drive CLI project.

### Interview Summary
**Key Discussions**:
- Comprehensive 10-dimension comparison of both frameworks (architecture, type safety, DI, hooks, completion, testing, error handling, ecosystem, docs)
- Full codebase audit: 140 commands, 261 flags, 19 CLI files, 1 PersistentPreRunE hook, global state DI pattern
- User chose full rewrite (not incremental), days timeline, breaking changes OK, add completion, expand tests

**Research Findings**:
- Kong eliminates ~900 lines of flag registration boilerplate
- Kong's Bind() DI replaces global state pattern cleanly
- Kong hooks compose (vs cobra's PersistentPreRunE which overrides)
- Kong has zero runtime dependencies
- Shell completion needs external plugin (carapace or kong-completion)
- Kong lacks built-in doc generation (not needed — not currently used)
- Exit codes are dead code in current implementation (GetExitCode() never called)

### Metis Review
**Identified Gaps** (addressed):
- **Exit codes are dead code**: GetExitCode() maps 22 error codes to exit codes 10-99 but is never called. README claims codes 0-6. Plan includes fixing this with kong's ExitCoder interface.
- **cmd.Flags().Changed()**: auth.go:576-577 uses cobra-specific API to distinguish "explicitly set to empty" vs "not provided". Must use `*string` pointer types in kong.
- **Shared flag variables**: `filesParentID` bound to 4 commands. Each kong struct needs its own field.
- **Dual flag names**: `--doc` and `--doc-text` bound to same bool. Kong handles with `aliases:""` tag.
- **Logger nil-safety**: Logger is nil until hook runs. Error paths calling GetLogger() early need nil-safe handling.
- **JSON output schema**: The CLIOutput envelope (`schemaVersion`, `traceId`, `command`, `data`, `warnings`, `errors`) is the AI-agent contract. OutputWriter is already framework-agnostic — don't couple to kong.

---

## Work Objectives

### Core Objective
Replace cobra's imperative command/flag registration with kong's declarative struct-tag pattern across the entire CLI layer, while preserving all business logic (Manager layer) untouched.

### Concrete Deliverables
- 19 rewritten CLI files in `internal/cli/`
- Rewritten `cmd/gdrv/main.go`
- Shell completion command (`gdrv completion bash/zsh/fish`)
- Parse-level tests for all command struct definitions
- Migrated existing 4 test files
- Clean `go.mod` with no cobra/pflag references

### Definition of Done
- [ ] `go build -o bin/gdrv ./cmd/gdrv` succeeds
- [ ] `go test -v -race ./...` passes (existing + new tests)
- [ ] `go vet ./...` reports no issues
- [ ] `grep -r "spf13/cobra" go.mod go.sum` returns 0 matches
- [ ] `grep -r "spf13/pflag" go.mod go.sum` returns 0 matches
- [ ] `bin/gdrv --help` shows all commands
- [ ] `bin/gdrv files list --json` produces valid JSON with CLIOutput envelope
- [ ] `bin/gdrv completion bash | bash -n` succeeds (valid completion script)

### Must Have
- All 140 commands accessible via same command paths (e.g., `gdrv files list`, `gdrv auth login`)
- All 261 flags functional (names preserved by default; intentional renames documented)
- JSON output envelope (`CLIOutput`) byte-identical structure
- Exit codes 0-6 functioning per README contract (fix dead code)
- `--json` alias for `--output json` preserved
- `--output` defaults to `"json"` (AI-agent default)
- Short flags preserved: `-q`, `-v`, `-f`, `-y`
- `perm` alias for `permissions` command preserved
- All RunE business logic preserved (copy-paste into Run() methods — wiring change, not behavior change)
- Global flags (`--profile`, `--drive-id`, `--json`, `--quiet`, `--verbose`, `--debug`, `--strict`, `--no-cache`, `--cache-ttl`, `--include-shared-with-me`, `--config`, `--log-file`, `--dry-run`, `--force`, `--yes`) inherited by all subcommands

### Must NOT Have (Guardrails)
- Do NOT modify business logic (Manager layer, API client, auth, types, errors packages)
- Do NOT redesign the error/exit code taxonomy — just wire ExitCoder to existing constants
- Do NOT refactor OutputWriter internals — it works, just make it callable from Run() methods
- Do NOT add enum/xor/and validation tags during migration — follow-up work
- Do NOT over-engineer DI into BindToProvider() — keep getFileManager() helpers, call from Run()
- Do NOT improve help text while rewriting — copy Short/Long strings verbatim
- Do NOT restructure file layout — keep one file per domain
- Do NOT couple OutputWriter to kong types
- Do NOT add config file resolver (not currently used)
- Do NOT add doc generation (not currently used)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (`go test` with race detection)
- **Automated tests**: YES (Tests-after — migrate existing 4 test files + add parse-level coverage)
- **Framework**: `go test` (standard library)

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **CLI commands**: Use Bash — run `bin/gdrv <command> --help`, `bin/gdrv <command> [args]`, parse JSON output
- **Build verification**: Use Bash — `go build`, `go test -race`, `go vet`
- **Parse verification**: Use Bash — `go test -run TestParse` for struct-level tests
- **Completion**: Use Bash — `bin/gdrv completion bash | bash -n`

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — scaffolding + baseline):
├── Task 1: Pre-migration baseline capture (flag manifest + JSON output) [quick]
├── Task 2: Go module changes (add kong dep) [quick]
├── Task 3: Globals struct + CLI root struct + main.go rewrite [deep]
├── Task 4: OutputWriter + drive_helpers.go framework decoupling [quick]
└── Task 5: Smoke test — about + version commands in kong [quick]

Wave 2 (Pattern Establishment — sequential, sets template):
└── Task 6: files.go migration (14 commands, establishes ALL patterns) [deep]

Wave 3 (Domain Migrations — MAXIMUM PARALLEL, 10 tasks):
├── Task 7: auth.go (8 commands, Changed() edge case) [deep]
├── Task 8: admin.go (15 commands, 3 nesting levels) [unspecified-high]
├── Task 9: permissions.go (14 commands, audit/bulk) [unspecified-high]
├── Task 10: chat.go (14 commands, spaces/messages/members) [unspecified-high]
├── Task 11: sheets.go (10 commands, batch operations) [unspecified-high]
├── Task 12: labels.go (8 commands, file operations) [unspecified-high]
├── Task 13: sync.go (6 commands, bidirectional) [unspecified-high]
├── Task 14: slides.go (6 cmd) + docs.go (5 cmd) [unspecified-high]
├── Task 15: folders.go (5 cmd) + drives.go (2 cmd) [quick]
└── Task 16: changes.go (4 cmd) + config.go (3 cmd) + activity.go (1 cmd) [quick]

Wave 4 (Tests + Completion + Cleanup — 4 tasks):
├── Task 17: Migrate 4 existing test files to kong patterns [unspecified-high]
├── Task 18: Add parse-level tests for ALL command structs [unspecified-high]
├── Task 19: Shell completion support (carapace or kong-completion) [unspecified-high]
└── Task 20: Remove cobra/pflag deps, go mod tidy, full build verify [quick]

Wave FINAL (After ALL — 4 parallel verification agents):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)

Critical Path: Task 1 → Task 3 → Task 6 → Tasks 7-16 → Tasks 17-20 → F1-F4
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 10 (Wave 3)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | — | 3, 6-16 (reference) | 1 |
| 2 | — | 3 | 1 |
| 3 | 2 | 5, 6, 7-16 | 1 |
| 4 | — | 6, 7-16 | 1 |
| 5 | 3 | 6 | 1 |
| 6 | 3, 4, 5 | 7-16 (pattern) | 2 |
| 7-16 | 3, 6 (pattern) | 17-20 | 3 |
| 17 | 7-16 | F1-F4 | 4 |
| 18 | 7-16 | F1-F4 | 4 |
| 19 | 3 | F1-F4 | 4 |
| 20 | 7-16, 17-19 | F1-F4 | 4 |
| F1-F4 | 20 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **5 tasks** — T1→`quick`, T2→`quick`, T3→`deep`, T4→`quick`, T5→`quick`
- **Wave 2**: **1 task** — T6→`deep`
- **Wave 3**: **10 tasks** — T7→`deep`, T8-T13→`unspecified-high`, T14→`unspecified-high`, T15-T16→`quick`
- **Wave 4**: **4 tasks** — T17-T19→`unspecified-high`, T20→`quick`
- **FINAL**: **4 tasks** — F1→`oracle`, F2→`unspecified-high`, F3→`unspecified-high`, F4→`deep`

---

## TODOs

> Implementation + Test = ONE Task. Never separate.
> EVERY task MUST have: Recommended Agent Profile + Parallelization info + QA Scenarios.

- [ ] 1. Pre-Migration Baseline Capture

  **What to do**:
  - Extract a complete flag manifest from the current cobra codebase: every command path, every flag name, type, default value, required status, short form
  - Run `bin/gdrv --help` and capture output for every top-level command + representative subcommands
  - Run `bin/gdrv files list --json 2>&1 || true` (and a few other commands) to capture JSON output envelope structure
  - Save all outputs as baseline files in `.sisyphus/evidence/` for post-migration diffing
  - Create a structured manifest file (JSON or markdown) listing all 140 commands and 261 flags

  **Must NOT do**:
  - Modify any source code
  - Change build configuration

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
    - No special skills needed — pure grep/read/bash operations

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5)
  - **Blocks**: Tasks 6-16 (reference material), Tasks 17-18 (test targets)
  - **Blocked By**: None

  **References**:
  **Pattern References**:
  - `internal/cli/root.go:72-87` — All 14 persistent flag definitions to capture
  - `internal/cli/files.go:138-202` — Example of local flag registration pattern (most complex domain)

  **API/Type References**:
  - `internal/types/cli.go` — GlobalFlags struct definition (all persistent flags)
  - `internal/cli/output.go` — CLIOutput envelope structure

  **Acceptance Criteria**:
  - [ ] Baseline manifest file exists at `.sisyphus/evidence/task-1-flag-manifest.json` or `.md`
  - [ ] Contains all 140 command paths
  - [ ] Contains all 261 flags with name, type, default, required, short form
  - [ ] JSON output samples captured for at least 3 commands

  **QA Scenarios**:

  ```
  Scenario: Flag manifest completeness
    Tool: Bash
    Preconditions: Project builds successfully
    Steps:
      1. Build binary: go build -o bin/gdrv ./cmd/gdrv
      2. Run grep -c "cobra.Command{" internal/cli/*.go to count command definitions
      3. Verify manifest contains same count of entries
    Expected Result: Manifest entry count matches grep count (approximately 140)
    Evidence: .sisyphus/evidence/task-1-flag-manifest.json

  Scenario: JSON output baseline capture
    Tool: Bash
    Preconditions: Binary built, auth may not be configured (expect error responses)
    Steps:
      1. Run bin/gdrv version and capture output
      2. Run bin/gdrv files list --json 2>&1 || true and capture output
      3. Verify captured files are non-empty
    Expected Result: Output files contain valid text/JSON (even if error responses)
    Evidence: .sisyphus/evidence/task-1-json-baseline.txt
  ```

  **Commit**: NO (evidence files only, not committed)

---

- [ ] 2. Go Module Changes — Add Kong Dependency

  **What to do**:
  - Run `go get github.com/alecthomas/kong@latest`
  - Run `go get github.com/jotaen/kong-completion@latest` (or evaluate carapace — pick the simpler option for shell completion)
  - Do NOT remove cobra/pflag yet (they're still imported — removal happens in Task 20)
  - Verify `go mod tidy` succeeds
  - Verify `go build ./...` still succeeds with both deps present

  **Must NOT do**:
  - Remove cobra/pflag imports (still in use until migration completes)
  - Modify any source files

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5)
  - **Blocks**: Task 3 (needs kong importable)
  - **Blocked By**: None

  **References**:
  **External References**:
  - `https://github.com/alecthomas/kong` — Kong module path
  - `https://github.com/jotaen/kong-completion` — Shell completion plugin

  **Acceptance Criteria**:
  - [ ] `go.mod` contains `github.com/alecthomas/kong`
  - [ ] `go build ./...` succeeds
  - [ ] `go mod tidy` exits cleanly

  **QA Scenarios**:

  ```
  Scenario: Kong dependency added successfully
    Tool: Bash
    Preconditions: Clean working tree
    Steps:
      1. Run go get github.com/alecthomas/kong@latest
      2. Run go mod tidy
      3. Run go build ./...
      4. Grep go.mod for "alecthomas/kong"
    Expected Result: Build succeeds, kong appears in go.mod
    Evidence: .sisyphus/evidence/task-2-go-mod.txt

  Scenario: Existing tests still pass
    Tool: Bash
    Preconditions: Kong added to go.mod
    Steps:
      1. Run go test -race ./...
    Expected Result: All existing tests pass (0 failures)
    Failure Indicators: Any test failure or compilation error
    Evidence: .sisyphus/evidence/task-2-test-results.txt
  ```

  **Commit**: YES (group with Task 3)
  - Message: `build(deps): add kong CLI framework dependency`
  - Files: `go.mod`, `go.sum`
  - Pre-commit: `go build ./...`

---

- [ ] 3. Globals Struct + CLI Root Struct + main.go Rewrite

  **What to do**:
  - Define a `Globals` struct in `internal/cli/root.go` containing all 14 persistent flags as struct fields with kong tags:
    ```go
    type Globals struct {
        Profile            string          `help:"Authentication profile" default:"default" name:"profile"`
        DriveID            string          `help:"Shared Drive ID" name:"drive-id"`
        Output             string          `help:"Output format" default:"json" enum:"json,table" name:"output"`
        Quiet              bool            `help:"Suppress non-essential output" short:"q" name:"quiet"`
        Verbose            bool            `help:"Enable verbose logging" short:"v" name:"verbose"`
        Debug              bool            `help:"Enable debug output" name:"debug"`
        Strict             bool            `help:"Convert warnings to errors" name:"strict"`
        NoCache            bool            `help:"Bypass path resolution cache" name:"no-cache"`
        CacheTTL           int             `help:"Path cache TTL in seconds" default:"300" name:"cache-ttl"`
        IncludeSharedWithMe bool           `help:"Include shared-with-me items" name:"include-shared-with-me"`
        Config             string          `help:"Path to configuration file" name:"config"`
        LogFile            string          `help:"Path to log file" name:"log-file"`
        DryRun             bool            `help:"Preview without changes" name:"dry-run"`
        Force              bool            `help:"Force operation" short:"f" name:"force"`
        Yes                bool            `help:"Answer yes to all prompts" short:"y" name:"yes"`
        JSON               bool            `help:"JSON output (alias for --output json)" name:"json"`
    }
    ```
  - Define the root `CLI` struct with all top-level commands as embedded fields:
    ```go
    type CLI struct {
        Globals
        Version    VersionCmd    `cmd:"" help:"Print version number"`
        Files      FilesCmd      `cmd:"" help:"File operations"`
        Auth       AuthCmd       `cmd:"" help:"Authentication commands"`
        // ... all 16 domains
    }
    ```
  - Implement `BeforeApply` hook on Globals to replace PersistentPreRunE:
    - Validate global flags (handle `--json` alias)
    - Initialize logger with config from flags
    - Store logger in a way accessible to Run() methods (via Bind or on Globals struct)
  - Rewrite `cmd/gdrv/main.go`:
    ```go
    func main() {
        var cli CLI
        ctx := kong.Parse(&cli,
            kong.Name("gdrv"),
            kong.Description("Google Drive CLI"),
            kong.Vars{"version": version.Version},
            kong.Bind(&cli.Globals),
        )
        err := ctx.Run(&cli.Globals)
        ctx.FatalIfErrorf(err)
    }
    ```
  - Keep old cobra root command and init() temporarily (other domains still use it) — but mark as deprecated with TODO comments
  - Implement `ToGlobalFlags()` on Globals to convert to existing `types.GlobalFlags` for backward compat with Manager layer

  **Must NOT do**:
  - Migrate any domain commands yet (files, auth, etc.)
  - Change GlobalFlags type definition in internal/types/
  - Modify OutputWriter

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Core architectural change — defines patterns all other tasks depend on
  - **Skills**: [`golang-pro`]
    - `golang-pro`: Kong integration requires Go idioms, struct embedding, interface satisfaction

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 1, 2, 4)
  - **Parallel Group**: Wave 1
  - **Blocks**: Tasks 5, 6, 7-16 (ALL domain migrations depend on this)
  - **Blocked By**: Task 2 (kong must be in go.mod)

  **References**:
  **Pattern References**:
  - `internal/cli/root.go` — Current root command, PersistentPreRunE hook (lines 22-60), global flags (lines 71-87), Execute/GetGlobalFlags/GetLogger functions
  - `cmd/gdrv/main.go` — Current entry point calling cli.Execute()

  **API/Type References**:
  - `internal/types/cli.go` — `GlobalFlags` struct (the conversion target for Globals.ToGlobalFlags())
  - `internal/logging/logger.go` — Logger interface and NewLogger constructor
  - `pkg/version/version.go` — Version variables injected at build time

  **External References**:
  - Kong README — Bind() DI pattern: `kong.Bind(&globals)` for injecting into Run() methods
  - Kong README — BeforeApply hook for pre-command initialization
  - Kong README — kong.Vars for build-time version injection

  **WHY Each Reference Matters**:
  - `root.go` PersistentPreRunE (line 30-59) — This EXACT logic must be ported to BeforeApply/AfterApply hook. Study the logging config assembly.
  - `types/cli.go` GlobalFlags — The Manager layer consumes this type. Globals must convert to it cleanly.

  **Acceptance Criteria**:
  - [ ] `Globals` struct defined with all 14 persistent flag fields + kong tags
  - [ ] Root `CLI` struct defined with command fields (initially empty/stub types)
  - [ ] `main.go` uses `kong.Parse` instead of `rootCmd.Execute()`
  - [ ] `Globals.ToGlobalFlags()` converts to `types.GlobalFlags`
  - [ ] Logger initialization works via hook (BeforeApply or AfterApply)
  - [ ] `go build ./cmd/gdrv` succeeds (even if most commands are stubs)

  **QA Scenarios**:

  ```
  Scenario: Kong parse works with global flags
    Tool: Bash
    Preconditions: Binary built successfully
    Steps:
      1. Run go build -o bin/gdrv ./cmd/gdrv
      2. Run bin/gdrv --help
      3. Verify output contains "gdrv" and lists available commands
      4. Run bin/gdrv --profile test --json version (or any available command)
    Expected Result: Help output shows kong-style formatting with global flags listed. Command runs without panic.
    Failure Indicators: "unknown flag" error, panic, empty help output
    Evidence: .sisyphus/evidence/task-3-help-output.txt

  Scenario: Global flags parsing and conversion
    Tool: Bash
    Preconditions: Binary built
    Steps:
      1. Run bin/gdrv --debug --verbose --quiet --profile myprofile version
      2. Verify no parse errors (exit code 0)
    Expected Result: Flags parsed without error, command executes
    Evidence: .sisyphus/evidence/task-3-global-flags.txt
  ```

  **Commit**: YES
  - Message: `refactor(cli): define kong CLI root struct and rewrite main.go entry point`
  - Files: `cmd/gdrv/main.go`, `internal/cli/root.go`
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 4. OutputWriter + drive_helpers.go Framework Decoupling

  **What to do**:
  - Review `internal/cli/output.go` — verify it has NO cobra imports. If it does, remove them.
  - Review `internal/cli/drive_helpers.go` — verify it has NO cobra imports. If it does, refactor to accept plain parameters instead of `*cobra.Command`.
  - Ensure `NewOutputWriter()` accepts `types.OutputFormat` (or string) and booleans — no cobra types
  - Ensure `ResolveFileID()`, `GetPathResolver()`, `GetResolveOptions()` in root.go accept plain types, not cobra types
  - These helper functions will be called from kong `Run()` methods — they must be framework-agnostic

  **Must NOT do**:
  - Change OutputWriter's internal logic or JSON envelope structure
  - Change any Manager-layer code

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5)
  - **Blocks**: Tasks 6-16 (all domain migrations use these helpers)
  - **Blocked By**: None

  **References**:
  **Pattern References**:
  - `internal/cli/output.go` — OutputWriter with WriteSuccess/WriteError/WriteJSON methods
  - `internal/cli/drive_helpers.go` — Helper functions for auth, API client setup
  - `internal/cli/root.go:120-186` — ResolveFileID, GetPathResolver, GetResolveOptions

  **Acceptance Criteria**:
  - [ ] `grep -r "spf13/cobra" internal/cli/output.go internal/cli/drive_helpers.go` returns 0 matches
  - [ ] `go build ./...` succeeds
  - [ ] Existing tests still pass

  **QA Scenarios**:

  ```
  Scenario: No cobra imports in helper files
    Tool: Bash
    Steps:
      1. Run grep -r "spf13/cobra" internal/cli/output.go internal/cli/drive_helpers.go
      2. Run grep -r "spf13/pflag" internal/cli/output.go internal/cli/drive_helpers.go
    Expected Result: 0 matches for both greps
    Evidence: .sisyphus/evidence/task-4-no-cobra-helpers.txt

  Scenario: Build still works
    Tool: Bash
    Steps:
      1. Run go build ./...
      2. Run go test -race ./internal/cli/...
    Expected Result: Build and tests pass
    Evidence: .sisyphus/evidence/task-4-build-verify.txt
  ```

  **Commit**: YES (group with Task 3)
  - Message: `refactor(cli): decouple helpers from cobra framework types`
  - Files: `internal/cli/output.go`, `internal/cli/drive_helpers.go`

---

- [ ] 5. Smoke Test — About + Version Commands in Kong

  **What to do**:
  - Migrate `about.go` (1 command — simplest) to kong struct pattern:
    ```go
    type AboutCmd struct{}
    func (cmd *AboutCmd) Run(globals *Globals) error { /* existing logic */ }
    ```
  - Migrate `version` command (in root.go) to kong:
    - Use `kong.Vars{"version": version.Version}` and `VersionFlag` or custom VersionCmd
  - Wire both into the root CLI struct
  - Remove the old cobra about/version command definitions
  - Verify the full kong parse → dispatch → execution flow works end-to-end

  **Must NOT do**:
  - Migrate any other domain

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential after Task 3
  - **Blocks**: Task 6
  - **Blocked By**: Task 3

  **References**:
  **Pattern References**:
  - `internal/cli/about.go` — Current about command (simplest command in codebase)
  - `internal/cli/root.go:62-69` — Current version command

  **External References**:
  - Kong README — `kong.Vars{"version": ...}` for version display

  **Acceptance Criteria**:
  - [ ] `bin/gdrv about` works (shows API capabilities or expected output)
  - [ ] `bin/gdrv version` works (shows version string)
  - [ ] `bin/gdrv --help` lists both commands

  **QA Scenarios**:

  ```
  Scenario: About command works via kong dispatch
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv about --help
      3. bin/gdrv about (may fail on auth — that's OK, should not panic)
    Expected Result: Help shows about command description. Execution either succeeds or returns auth error (not panic/crash).
    Evidence: .sisyphus/evidence/task-5-about-smoke.txt

  Scenario: Version command works
    Tool: Bash
    Steps:
      1. bin/gdrv version
    Expected Result: Prints version string (e.g., "dev" or version tag)
    Failure Indicators: "unknown command" error, panic
    Evidence: .sisyphus/evidence/task-5-version-smoke.txt
  ```

  **Commit**: YES (group with Task 3 commit)
  - Message: `refactor(cli): migrate about and version commands to kong`
  - Files: `internal/cli/about.go`, `internal/cli/root.go`

---

- [ ] 6. files.go Migration — Pattern Establishment (14 commands)

  **What to do**:
  - This is the MOST CRITICAL task — it establishes the pattern ALL other domain tasks will follow
  - Define `FilesCmd` struct with all 14 subcommands as nested structs:
    - `List`, `Get`, `Upload`, `Download`, `Delete`, `Copy`, `Move`, `Trash`, `Restore`, `Revisions` (with nested `Download`, `Restore`), `ListTrashed`, `ExportFormats`
  - Convert all 17 local flags to struct field tags:
    - `filesParentID` → per-command `Parent string` field (NOT shared — each command gets its own)
    - `filesQuery`, `filesLimit`, `filesPageToken`, `filesOrderBy`, `filesIncludeTrashed`, `filesFields`, `filesPaginate` → list-specific fields
    - `filesOutput` → download-specific field
    - `filesPermanent`, `filesForce` → delete-specific fields
    - `filesDownloadDoc` → download-specific field. Handle dual flag names (`--doc` / `--doc-text`) using `aliases:"doc-text"` or define both
  - Convert `cobra.ExactArgs(1)` → `arg:""` tags on struct fields (typed, not string index)
  - Convert `cobra.ExactArgs(2)` → two `arg:""` fields
  - Convert `MarkFlagRequired("parent")` → `required:""` tag
  - Convert `MarkFlagRequired("output")` → `required:""` tag
  - Implement `Run(globals *Globals) error` methods on each subcommand struct:
    - Call `globals.ToGlobalFlags()` to get `types.GlobalFlags`
    - Call existing `getFileManager()` helper (or equivalent)
    - Copy existing `runFilesXxx()` function body into Run() method
    - The business logic is IDENTICAL — only the wiring changes
  - Wire `FilesCmd` into root `CLI` struct
  - Remove old cobra filesCmd variables, init() function, and runXxx functions

  **Must NOT do**:
  - Change any Manager method signatures or behavior
  - Refactor getFileManager() helper — just call it from Run()
  - Add validation beyond what cobra currently enforces (no new enum/xor/and tags)
  - Improve help text — copy Short/Long strings verbatim

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 14 commands, most complex domain, establishes patterns for all others
  - **Skills**: [`golang-pro`]
    - `golang-pro`: Struct embedding, kong tags, method dispatch patterns

  **Parallelization**:
  - **Can Run In Parallel**: NO — sequential (pattern establishment)
  - **Parallel Group**: Wave 2 (solo)
  - **Blocks**: Tasks 7-16 (all domain migrations follow this pattern)
  - **Blocked By**: Tasks 3, 4, 5

  **References**:
  **Pattern References**:
  - `internal/cli/files.go` — ALL 724 lines. The ENTIRE file is the migration source. Key sections:
    - Lines 16-116: 14 cobra.Command definitions → become 14 nested structs
    - Lines 118-136: 17 flag variables → become struct fields
    - Lines 138-202: init() with flag registration + AddCommand → deleted entirely
    - Lines 205-230: getFileManager() helper → keep as-is, call from Run()
    - Lines 232-724: runFilesXxx() functions → copy body into Run() methods

  **API/Type References**:
  - `internal/files/manager.go` — Manager methods called by Run() (Upload, Download, List, etc.)
  - `internal/types/api.go` — RequestContext used in all operations

  **External References**:
  - Kong README: `arg:""` tag for positional arguments
  - Kong README: nested `cmd:""` structs for subcommands
  - Kong README: `aliases:""` for dual flag names

  **WHY Each Reference Matters**:
  - `files.go` lines 138-202 show the EXACT flag-to-variable mappings that become struct tags
  - `files.go` lines 205-230 show the DI helper pattern that Run() methods will call
  - `files.go` lines 232-724 contain the business logic that is copy-pasted into Run() methods

  **Acceptance Criteria**:
  - [ ] `FilesCmd` struct defined with all 14 subcommand structs
  - [ ] All 17 local flags converted to struct field tags
  - [ ] All Run() methods implemented (business logic preserved)
  - [ ] `go build ./cmd/gdrv` succeeds
  - [ ] `bin/gdrv files --help` lists all 14 subcommands
  - [ ] `bin/gdrv files list --help` shows all list flags
  - [ ] No cobra imports remain in files.go

  **QA Scenarios**:

  ```
  Scenario: Files list command help
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv files --help
      3. Verify output contains: list, get, upload, download, delete, copy, move, trash, restore, revisions, list-trashed, export-formats
      4. bin/gdrv files list --help
      5. Verify output contains flags: --parent, --query, --limit, --page-token, --order-by, --include-trashed, --fields, --paginate
    Expected Result: All 14 subcommands listed, all flags present with correct names
    Failure Indicators: Missing commands, "unknown flag" errors, missing flags
    Evidence: .sisyphus/evidence/task-6-files-help.txt

  Scenario: Files list command execution (may fail on auth)
    Tool: Bash
    Steps:
      1. bin/gdrv files list --json 2>&1 || true
      2. Verify output is valid JSON (even if error response)
      3. Verify JSON contains "command" field
    Expected Result: JSON output with CLIOutput envelope structure (even for auth errors)
    Evidence: .sisyphus/evidence/task-6-files-json.txt

  Scenario: No cobra imports remain
    Tool: Bash
    Steps:
      1. grep -c "spf13/cobra" internal/cli/files.go
    Expected Result: 0 matches
    Evidence: .sisyphus/evidence/task-6-no-cobra.txt
  ```

  **Commit**: YES
  - Message: `refactor(cli): migrate files domain (14 commands) to kong struct pattern`
  - Files: `internal/cli/files.go`
  - Pre-commit: `go build ./cmd/gdrv && go vet ./internal/cli/...`

---

- [ ] 7. auth.go Migration (8 commands, Changed() edge case)

  **What to do**:
  - Define `AuthCmd` struct with subcommands: `Login`, `Logout`, `ServiceAccount`, `Status`, `Device`, `Profiles`, `Diagnose`
  - Convert all auth flags to struct field tags (`--scopes`, `--no-browser`, `--preset`, `--client-id`, `--client-secret`, `--key-file`, `--impersonate-user`, `--wide`, `--refresh-check`)
  - **CRITICAL EDGE CASE**: `auth.go` uses `cmd.Flags().Changed("client-id")` and `cmd.Flags().Changed("client-secret")` (approximately line 576-577) to distinguish "explicitly set to empty string" from "flag not provided". In kong, use `*string` pointer types:
    ```go
    ClientID     *string `help:"OAuth client ID" name:"client-id"`
    ClientSecret *string `help:"OAuth client secret" name:"client-secret"`
    ```
    Then check `cmd.ClientID != nil` instead of `cmd.Flags().Changed("client-id")`
  - Convert `MarkFlagRequired("key-file")` → `required:""` on ServiceAccount.KeyFile field
  - Follow the exact pattern established by Task 6 (files.go)
  - Copy `runAuthXxx()` function bodies into Run() methods

  **Must NOT do**:
  - Change auth business logic or OAuth flow
  - Modify internal/auth/ package

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Changed() edge case requires careful pointer-type handling
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 8-16)
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6 (pattern)

  **References**:
  **Pattern References**:
  - `internal/cli/auth.go` — ALL 651 lines. Key sections:
    - Lines 20-73: 8 cobra.Command definitions
    - Lines 75-85: Flag variables
    - Lines 87-111: init() with flag registration
    - Lines 114-651: runAuthXxx() functions
  - `internal/cli/files.go` (post-Task-6) — The kong pattern to follow

  **API/Type References**:
  - `internal/auth/manager.go` — Auth Manager methods called by Run()

  **WHY Each Reference Matters**:
  - `auth.go` line ~576-577: `cmd.Flags().Changed()` calls — THE critical edge case. Search for all `Changed(` calls in this file.

  **Acceptance Criteria**:
  - [ ] All 8 auth commands accessible via `bin/gdrv auth <cmd>`
  - [ ] `*string` pointer types used for client-id and client-secret
  - [ ] `bin/gdrv auth --help` lists all subcommands
  - [ ] `bin/gdrv auth service-account --help` shows --key-file as required
  - [ ] No cobra imports in auth.go

  **QA Scenarios**:

  ```
  Scenario: Auth commands listed
    Tool: Bash
    Steps:
      1. bin/gdrv auth --help
      2. Verify: login, logout, service-account, status, device, profiles, diagnose all listed
    Expected Result: All 7 subcommands shown (+ auth parent help)
    Evidence: .sisyphus/evidence/task-7-auth-help.txt

  Scenario: Changed() edge case handled
    Tool: Bash
    Steps:
      1. grep -n "Changed(" internal/cli/auth.go (should return 0 — cobra pattern removed)
      2. grep -n "\*string" internal/cli/auth.go (should find ClientID/ClientSecret pointer types)
    Expected Result: No Changed() calls, pointer types present
    Evidence: .sisyphus/evidence/task-7-changed-edge-case.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate auth domain to kong`
  - Files: `internal/cli/auth.go`

---

- [ ] 8. admin.go Migration (15 commands, 3-level nesting)

  **What to do**:
  - Define `AdminCmd` with deeply nested structure:
    ```
    AdminCmd → Users → List, Get, Create, Delete, Update, Suspend, Unsuspend
            → Groups → List, Get, Create, Delete, Update
                     → Members → List, Add, Remove
    ```
  - Convert all admin flags (domain, customer, query, limit, page-token, order-by, fields, paginate, given-name, family-name, password, suspended, org-unit-path, description, role, roles)
  - Convert 3 `MarkFlagRequired` calls (given-name, family-name, password on users create)
  - Follow Task 6 pattern for Run() methods
  - Copy `runAdminXxx()` function bodies

  **Must NOT do**:
  - Change internal/admin/ package

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 15 commands with 3-level nesting requires careful struct composition
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/admin.go` — ALL lines. 15 commands, ~600 lines
  - `internal/admin/manager.go` — Admin Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv admin users list --help` shows all flags
  - [ ] `bin/gdrv admin groups members list --help` works (3-level nesting)
  - [ ] No cobra imports in admin.go

  **QA Scenarios**:

  ```
  Scenario: 3-level nesting works
    Tool: Bash
    Steps:
      1. bin/gdrv admin --help
      2. bin/gdrv admin users --help
      3. bin/gdrv admin groups members --help
    Expected Result: Each level shows its subcommands correctly
    Evidence: .sisyphus/evidence/task-8-admin-nesting.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate admin domain to kong`
  - Files: `internal/cli/admin.go`

---

- [ ] 9. permissions.go Migration (14 commands, audit/bulk)

  **What to do**:
  - Define `PermissionsCmd` with nested audit/bulk subcommand groups:
    ```
    PermissionsCmd → List, Create, Update, Remove, CreateLink
                   → Audit → Public, External, AnyoneWithLink, User
                   → Analyze, Report
                   → Bulk → RemovePublic, UpdateRole
                   → Search
    ```
  - Preserve `perm` alias: use `aliases:"perm"` or `name:"permissions" aliases:"perm"` on the parent command
  - Convert all permission flags (type, role, email, domain, folder-id, recursive, from-role, to-role, etc.)
  - Convert 7 `MarkFlagRequired` calls (type, role on create; role on update; internal-domain on audit external; folder-id/from-role/to-role on bulk update-role)
  - Follow Task 6 pattern

  **Must NOT do**:
  - Change internal/permissions/ package

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/permissions.go` — ALL lines. 14 commands with audit/bulk nested groups
  - `internal/permissions/` — Permissions Manager and audit functions

  **Acceptance Criteria**:
  - [ ] `bin/gdrv permissions --help` lists all subcommands
  - [ ] `bin/gdrv perm --help` also works (alias)
  - [ ] `bin/gdrv permissions audit --help` shows audit subcommands
  - [ ] `bin/gdrv permissions bulk update-role --help` shows required flags
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Permissions alias works
    Tool: Bash
    Steps:
      1. bin/gdrv perm --help
      2. bin/gdrv permissions --help
    Expected Result: Both produce identical help output
    Evidence: .sisyphus/evidence/task-9-perm-alias.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate permissions domain to kong`
  - Files: `internal/cli/permissions.go`

---

- [ ] 10. chat.go Migration (14 commands, spaces/messages/members)

  **What to do**:
  - Define `ChatCmd` with 3-level nesting:
    ```
    ChatCmd → Spaces → List, Get, Create, Delete
           → Messages → List, Get, Create, Update, Delete
           → Members → List, Get, Create, Delete
    ```
  - Convert all chat flags (display-name, type, external-users, text, thread, filter, email, role, limit, page-token, paginate)
  - Follow Task 6 pattern

  **Must NOT do**:
  - Change internal/chat/ package

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/chat.go` — ALL lines. 14 commands across 3 subgroups

  **Acceptance Criteria**:
  - [ ] All 3 subgroups accessible (spaces, messages, members)
  - [ ] `bin/gdrv chat spaces create --help` shows flags
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Chat command tree
    Tool: Bash
    Steps:
      1. bin/gdrv chat --help
      2. bin/gdrv chat spaces --help
      3. bin/gdrv chat messages --help
      4. bin/gdrv chat members --help
    Expected Result: Each group lists its subcommands
    Evidence: .sisyphus/evidence/task-10-chat-tree.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate chat domain to kong`
  - Files: `internal/cli/chat.go`

---

- [ ] 11. sheets.go Migration (10 commands, batch operations)

  **What to do**:
  - Define `SheetsCmd` with nested values subgroup:
    ```
    SheetsCmd → List, Create, Get, BatchUpdate
             → Values → Get, Update, Append, Clear
    ```
  - Convert all sheets flags (parent, query, limit, page-token, order-by, fields, paginate, values, values-file, value-input-option, requests, requests-file)
  - Follow Task 6 pattern

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/sheets.go` — ALL lines. 10 commands
  - `internal/sheets/manager.go` — Sheets Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv sheets --help` lists all subcommands
  - [ ] `bin/gdrv sheets values --help` lists get, update, append, clear
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Sheets values subgroup works
    Tool: Bash
    Steps:
      1. bin/gdrv sheets values --help
    Expected Result: Lists get, update, append, clear subcommands
    Evidence: .sisyphus/evidence/task-11-sheets-values.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate sheets domain to kong`
  - Files: `internal/cli/sheets.go`

---

- [ ] 12. labels.go Migration (8 commands, file operations)

  **What to do**:
  - Define `LabelsCmd` with nested file operations:
    ```
    LabelsCmd → List, Get, Create, Publish, Disable
             → File → List, Apply, Update, Remove
    ```
  - Convert all labels flags (view, customer, fields, label fields JSON)
  - Note: labels.go already has some kong-like inline RunE logic — check if cobra features are used beyond basic pattern
  - Follow Task 6 pattern

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/labels.go` — ALL lines. 8 commands with file subgroup

  **Acceptance Criteria**:
  - [ ] `bin/gdrv labels --help` and `bin/gdrv labels file --help` work
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Labels file subgroup
    Tool: Bash
    Steps:
      1. bin/gdrv labels file --help
    Expected Result: Lists list, apply, update, remove subcommands
    Evidence: .sisyphus/evidence/task-12-labels-file.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate labels domain to kong`
  - Files: `internal/cli/labels.go`

---

- [ ] 13. sync.go Migration (6 commands)

  **What to do**:
  - Define `SyncCmd` struct:
    ```
    SyncCmd → Init, Push, Pull, Status, List, Remove
    ```
  - Convert all sync flags (local-dir, remote-folder-id, sync-direction, conflict-strategy, exclude-patterns, include-patterns, max-concurrent, etc.)
  - Follow Task 6 pattern

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/sync.go` — ALL lines. 6 commands

  **Acceptance Criteria**:
  - [ ] `bin/gdrv sync --help` lists all 6 subcommands
  - [ ] No cobra imports

  **QA Scenarios**:

  ```
  Scenario: Sync commands listed
    Tool: Bash
    Steps:
      1. bin/gdrv sync --help
    Expected Result: init, push, pull, status, list, remove all listed
    Evidence: .sisyphus/evidence/task-13-sync-help.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate sync domain to kong`
  - Files: `internal/cli/sync.go`

---

- [ ] 14. slides.go (6 commands) + docs.go (5 commands) Migration

  **What to do**:
  - These two domains have nearly identical structure (list, get, create, read, update + slides has replace)
  - Define `SlidesCmd` struct: List, Get, Create, Read, Update, Replace
  - Define `DocsCmd` struct: List, Get, Create, Read, Update
  - Convert all flags for both (parent, query, limit, page-token, order-by, fields, paginate, requests, requests-file, data, file)
  - Follow Task 6 pattern for both files

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Two files but structurally similar — efficient to handle together
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/slides.go` — 6 commands
  - `internal/cli/docs.go` — 5 commands
  - `internal/slides/manager.go`, `internal/docs/manager.go` — Manager methods

  **Acceptance Criteria**:
  - [ ] `bin/gdrv slides --help` lists all 6 subcommands
  - [ ] `bin/gdrv docs --help` lists all 5 subcommands
  - [ ] `bin/gdrv slides replace --help` shows --data and --file flags
  - [ ] No cobra imports in either file

  **QA Scenarios**:

  ```
  Scenario: Slides and docs commands
    Tool: Bash
    Steps:
      1. bin/gdrv slides --help
      2. bin/gdrv docs --help
    Expected Result: Both list their subcommands correctly
    Evidence: .sisyphus/evidence/task-14-slides-docs.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate slides and docs domains to kong`
  - Files: `internal/cli/slides.go`, `internal/cli/docs.go`

---

- [ ] 15. folders.go (5 commands) + drives.go (2 commands) Migration

  **What to do**:
  - Define `FoldersCmd`: Create, List, Delete, Move, Get
  - Define `DrivesCmd`: List, Get
  - Convert all folder flags (name, parent, recursive, limit, page-token, paginate, fields)
  - Convert drives flags (limit, page-token, paginate, fields)
  - Follow Task 6 pattern

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small domains, straightforward migration
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/folders.go` — 5 commands
  - `internal/cli/drives.go` — 2 commands

  **Acceptance Criteria**:
  - [ ] `bin/gdrv folders --help` and `bin/gdrv drives --help` work
  - [ ] No cobra imports in either file

  **QA Scenarios**:

  ```
  Scenario: Folders and drives
    Tool: Bash
    Steps:
      1. bin/gdrv folders --help
      2. bin/gdrv drives --help
    Expected Result: Both list their subcommands
    Evidence: .sisyphus/evidence/task-15-folders-drives.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate folders and drives domains to kong`
  - Files: `internal/cli/folders.go`, `internal/cli/drives.go`

---

- [ ] 16. changes.go (4 commands) + config.go (3 commands) + activity.go (1 command) Migration

  **What to do**:
  - Define `ChangesCmd`: StartPageToken, List, Watch, Stop
  - Define `ConfigCmd`: Show, Set, Reset
  - Define `ActivityCmd`: Query
  - Convert all flags:
    - Changes: page-token (required on list/watch), webhook-url (required on watch), drive-id, include-removed, etc.
    - Config: key/value args
    - Activity: file-id, folder-id, ancestor-name, start-time, end-time, action-types, user, limit, page-token
  - Convert 3 `MarkFlagRequired` calls on changes commands
  - Follow Task 6 pattern

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small domains, 8 total commands
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Tasks 17-20
  - **Blocked By**: Task 6

  **References**:
  - `internal/cli/changes.go` — 4 commands with required flags
  - `internal/cli/config.go` — 3 commands
  - `internal/cli/activity.go` — 1 command

  **Acceptance Criteria**:
  - [ ] `bin/gdrv changes --help`, `bin/gdrv config --help`, `bin/gdrv activity --help` all work
  - [ ] No cobra imports in any of the 3 files

  **QA Scenarios**:

  ```
  Scenario: Changes, config, activity commands
    Tool: Bash
    Steps:
      1. bin/gdrv changes --help
      2. bin/gdrv config --help
      3. bin/gdrv activity --help
    Expected Result: All list their subcommands
    Evidence: .sisyphus/evidence/task-16-changes-config-activity.txt
  ```

  **Commit**: YES (group with Wave 3)
  - Message: `refactor(cli): migrate changes, config, activity domains to kong`
  - Files: `internal/cli/changes.go`, `internal/cli/config.go`, `internal/cli/activity.go`

---

- [ ] 17. Migrate Existing 4 Test Files to Kong Patterns

  **What to do**:
  - Migrate these existing test files:
    - `internal/cli/admin_test.go`
    - `internal/cli/sheets_test.go`
    - `internal/cli/slides_test.go`
    - `internal/cli/docs_test.go`
  - Also check/migrate:
    - `internal/cli/output_test.go` — OutputWriter tests (probably framework-agnostic, verify)
    - `internal/cli/config_parse_bool_test.go` — Config tests (verify)
    - `internal/cli/auth_error_test.go` — Auth error tests (verify)
  - Replace any cobra-specific test patterns:
    - `rootCmd.SetArgs(...)` → `kong.Must(&cli).Parse(args)`
    - `rootCmd.Execute()` → `parser.Parse(args)` + `ctx.Run()`
    - `cmd.SetOut(buf)` → `kong.Writers(buf, buf)`
  - Use kong test pattern:
    ```go
    func TestAdminUsersCreate(t *testing.T) {
        var cli CLI
        parser := kong.Must(&cli,
            kong.Name("test"),
            kong.Exit(func(int) { t.Fatal("unexpected exit") }),
        )
        _, err := parser.Parse([]string{"admin", "users", "create", "user@example.com", "--given-name", "Test", "--family-name", "User", "--password", "Pass123"})
        assert.NoError(t, err)
        assert.Equal(t, "Test", cli.Admin.Users.Create.GivenName)
    }
    ```
  - Ensure all existing test assertions are preserved (same coverage, different framework)

  **Must NOT do**:
  - Delete any existing test coverage
  - Change what is being tested — only how it's tested

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 7 test files to migrate with framework-specific patterns
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 18, 19, 20)
  - **Blocks**: Task 20 (must pass before final cleanup)
  - **Blocked By**: Tasks 7-16 (all domains must be migrated)

  **References**:
  - `internal/cli/admin_test.go`, `internal/cli/sheets_test.go`, `internal/cli/slides_test.go`, `internal/cli/docs_test.go` — Current test files
  - `internal/cli/output_test.go`, `internal/cli/config_parse_bool_test.go`, `internal/cli/auth_error_test.go` — Possibly framework-agnostic tests to verify

  **Acceptance Criteria**:
  - [ ] `go test -v -race ./internal/cli/...` passes
  - [ ] No cobra imports in any test file
  - [ ] Test count is equal or greater than before migration

  **QA Scenarios**:

  ```
  Scenario: All CLI tests pass
    Tool: Bash
    Steps:
      1. go test -v -race ./internal/cli/...
    Expected Result: All tests pass, 0 failures
    Failure Indicators: Any FAIL line in output
    Evidence: .sisyphus/evidence/task-17-test-results.txt

  Scenario: No cobra in test files
    Tool: Bash
    Steps:
      1. grep -r "spf13/cobra" internal/cli/*_test.go
    Expected Result: 0 matches
    Evidence: .sisyphus/evidence/task-17-no-cobra-tests.txt
  ```

  **Commit**: YES
  - Message: `test(cli): migrate existing CLI tests to kong patterns`
  - Files: `internal/cli/*_test.go`
  - Pre-commit: `go test -race ./internal/cli/...`

---

- [ ] 18. Add Parse-Level Tests for ALL Command Structs

  **What to do**:
  - Create `internal/cli/parse_test.go` (or per-domain `*_parse_test.go` files)
  - For EVERY command struct, add at least one parse test that verifies:
    - Command path resolves correctly (e.g., `["files", "list"]` → `FilesCmd.List`)
    - Required flags are enforced (parse fails without them)
    - Default values are applied correctly
    - Positional args bind to correct fields
    - Short flags work (`-q`, `-v`, `-f`, `-y`)
  - Add a comprehensive "all commands reachable" test:
    ```go
    func TestAllCommandsReachable(t *testing.T) {
        commands := [][]string{
            {"files", "list"},
            {"files", "get", "FILEID"},
            {"auth", "login"},
            // ... all 140 command paths
        }
        for _, args := range commands {
            _, err := parser.Parse(args)
            // Should parse without error (validation errors OK, parse errors NOT OK)
        }
    }
    ```
  - Add global flag inheritance test: verify `--json`, `--quiet`, `--profile` work on every subcommand
  - Add `--json` alias test: verify `--json` sets Output to "json"

  **Must NOT do**:
  - Test business logic (Manager calls) — only parse-level behavior
  - Duplicate coverage from Task 17

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Comprehensive coverage across all 140 commands
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: Task 20
  - **Blocked By**: Tasks 7-16

  **References**:
  - Task 1 output — Flag manifest (`.sisyphus/evidence/task-1-flag-manifest.json`) for complete list of commands and flags to test
  - All migrated `internal/cli/*.go` files — The kong struct definitions to test

  **Acceptance Criteria**:
  - [ ] Parse test exists for every domain (files, auth, admin, permissions, sheets, docs, slides, folders, changes, labels, chat, sync, config, about)
  - [ ] "All commands reachable" test covers all 140 command paths
  - [ ] Global flag inheritance verified
  - [ ] Required flag enforcement verified for all 17 required flags
  - [ ] `go test -v -race ./internal/cli/...` passes

  **QA Scenarios**:

  ```
  Scenario: Parse tests pass
    Tool: Bash
    Steps:
      1. go test -v -race -run TestParse ./internal/cli/...
      2. go test -v -race -run TestAllCommands ./internal/cli/...
    Expected Result: All parse tests pass
    Evidence: .sisyphus/evidence/task-18-parse-tests.txt

  Scenario: Test coverage increased
    Tool: Bash
    Steps:
      1. go test -cover ./internal/cli/... | tail -1
      2. Verify coverage percentage
    Expected Result: Coverage > 30% (parse tests cover struct definitions + flag parsing)
    Evidence: .sisyphus/evidence/task-18-coverage.txt
  ```

  **Commit**: YES
  - Message: `test(cli): add parse-level tests for all kong command structs`
  - Files: `internal/cli/parse_test.go` (or `*_parse_test.go`)
  - Pre-commit: `go test -race ./internal/cli/...`

---

- [ ] 19. Shell Completion Support

  **What to do**:
  - Evaluate and integrate shell completion (pick ONE):
    - **Option A**: `jotaen/kong-completion` — lighter, 14 stars, bash/zsh/fish
    - **Option B**: `carapace-sh/carapace` — heavier, 1162 stars, multi-shell
    - Recommendation: Start with kong-completion for simplicity
  - Add `CompletionCmd` to root CLI struct:
    ```go
    Completion completionCmd `cmd:"" help:"Generate shell completion script" hidden:""`
    ```
  - Implement completion subcommands: `gdrv completion bash`, `gdrv completion zsh`, `gdrv completion fish`
  - Each should write a valid completion script to stdout
  - Mark the completion command as `hidden:""` so it doesn't clutter `--help`

  **Must NOT do**:
  - Add powershell completion (not requested, can add later)
  - Make completion a required dep — it should be optional/graceful

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Third-party integration with shell-specific output
  - **Skills**: [`golang-pro`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: Task 20
  - **Blocked By**: Task 3 (root CLI struct must exist)

  **References**:
  **External References**:
  - `https://github.com/jotaen/kong-completion` — kong-completion plugin
  - `https://github.com/carapace-sh/carapace` — carapace multi-shell completion

  **Acceptance Criteria**:
  - [ ] `bin/gdrv completion bash` outputs valid bash completion script
  - [ ] `bin/gdrv completion zsh` outputs valid zsh completion script
  - [ ] `bin/gdrv completion fish` outputs valid fish completion script
  - [ ] `bin/gdrv completion bash | bash -n` succeeds (valid syntax)

  **QA Scenarios**:

  ```
  Scenario: Bash completion is valid
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. bin/gdrv completion bash > /tmp/gdrv_completion.bash
      3. bash -n /tmp/gdrv_completion.bash
    Expected Result: Exit code 0 (valid bash syntax)
    Evidence: .sisyphus/evidence/task-19-bash-completion.txt

  Scenario: Completion command hidden from help
    Tool: Bash
    Steps:
      1. bin/gdrv --help
      2. Check output does NOT contain "completion" in command list
    Expected Result: Completion command not visible in help (hidden)
    Evidence: .sisyphus/evidence/task-19-hidden.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): add shell completion support (bash/zsh/fish)`
  - Files: `internal/cli/completion.go` (new file)
  - Pre-commit: `go build ./cmd/gdrv`

---

- [ ] 20. Remove Cobra/Pflag Dependencies + Final Build Verification

  **What to do**:
  - Remove ALL remaining cobra and pflag imports from every file:
    - `grep -r "spf13/cobra" internal/ cmd/` should return 0 matches
    - `grep -r "spf13/pflag" internal/ cmd/` should return 0 matches
  - Run `go mod tidy` to remove unused dependencies
  - Verify `go.mod` no longer contains `github.com/spf13/cobra` or `github.com/spf13/pflag`
  - Run full build and test suite:
    - `go build -o bin/gdrv ./cmd/gdrv`
    - `go test -v -race -cover ./...`
    - `go vet ./...`
  - Verify binary runs correctly:
    - `bin/gdrv --help`
    - `bin/gdrv version`
    - `bin/gdrv files --help`
  - Clean up any TODO comments from Task 3 about deprecated cobra code

  **Must NOT do**:
  - Change any business logic
  - Add new features

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Cleanup task — grep, delete imports, go mod tidy
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential after Tasks 17-19
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 17, 18, 19

  **References**:
  - All `internal/cli/*.go` files — scan for remaining cobra/pflag imports
  - `go.mod`, `go.sum` — dependency files to clean

  **Acceptance Criteria**:
  - [ ] `grep -r "spf13/cobra" internal/ cmd/ go.mod go.sum` returns 0 matches
  - [ ] `grep -r "spf13/pflag" internal/ cmd/ go.mod go.sum` returns 0 matches
  - [ ] `go build -o bin/gdrv ./cmd/gdrv` succeeds
  - [ ] `go test -v -race ./...` all pass
  - [ ] `go vet ./...` no issues
  - [ ] `bin/gdrv --help` shows all commands

  **QA Scenarios**:

  ```
  Scenario: Cobra fully removed
    Tool: Bash
    Steps:
      1. grep -r "spf13/cobra" internal/ cmd/ go.mod go.sum || echo "CLEAN"
      2. grep -r "spf13/pflag" internal/ cmd/ go.mod go.sum || echo "CLEAN"
    Expected Result: Both return "CLEAN" (0 matches)
    Evidence: .sisyphus/evidence/task-20-cobra-removed.txt

  Scenario: Full build and test suite
    Tool: Bash
    Steps:
      1. go build -o bin/gdrv ./cmd/gdrv
      2. go test -v -race -cover ./...
      3. go vet ./...
      4. bin/gdrv --help
      5. bin/gdrv version
    Expected Result: Build succeeds, all tests pass, vet clean, help shows all commands
    Failure Indicators: Any compilation error, test failure, or missing command
    Evidence: .sisyphus/evidence/task-20-final-build.txt
  ```

  **Commit**: YES
  - Message: `build(cli): remove cobra/pflag dependencies, complete kong migration`
  - Files: `go.mod`, `go.sum`, any files with remaining cobra imports
  - Pre-commit: `go build ./cmd/gdrv && go test -race ./... && go vet ./...`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Rejection → fix → re-run.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go build ./...`, `go vet ./...`, `go test -race ./...`. Review all changed files for: dead code, unused imports, commented-out cobra references, `as any` patterns. Check AI slop: excessive comments, over-abstraction, generic names. Verify kong struct tags are well-formed.
  Output: `Build [PASS/FAIL] | Vet [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Run `go build -o bin/gdrv ./cmd/gdrv`. Test every top-level command (`files`, `auth`, `admin`, `permissions`, `sheets`, `docs`, `slides`, `folders`, `changes`, `labels`, `chat`, `sync`, `config`, `about`, `version`). Verify `--help` output, `--json` flag, error handling. Test shell completion. Save evidence.
  Output: `Commands [N/N pass] | Help [N/N] | JSON [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes to business logic files.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Wave 1**: `build(cli): add kong dependency and define CLI root struct` — go.mod, go.sum, internal/cli/root.go, cmd/gdrv/main.go
- **Wave 2**: `refactor(cli): migrate files domain to kong struct pattern` — internal/cli/files.go
- **Wave 3**: `refactor(cli): migrate all remaining domains to kong` — internal/cli/*.go (all domain files)
- **Wave 4**: `test(cli): migrate and expand CLI test coverage` + `feat(cli): add shell completion support` + `build(cli): remove cobra/pflag dependencies`
- **Final**: `chore(cli): cobra-to-kong migration complete`

---

## Success Criteria

### Verification Commands
```bash
go build -o bin/gdrv ./cmd/gdrv           # Expected: successful build, no errors
go test -v -race -cover ./...              # Expected: all tests pass
go vet ./...                               # Expected: no issues
grep -r "spf13/cobra" go.mod go.sum        # Expected: 0 matches
grep -r "spf13/pflag" go.mod go.sum        # Expected: 0 matches
bin/gdrv --help                            # Expected: all commands listed
bin/gdrv files list --help                 # Expected: all flags shown
bin/gdrv version                           # Expected: version string
bin/gdrv completion bash | bash -n         # Expected: valid bash syntax
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass (existing + new)
- [ ] cobra/pflag fully removed from go.mod
- [ ] Shell completion functional
- [ ] JSON output envelope preserved
- [ ] Exit codes 0-6 functional per README
