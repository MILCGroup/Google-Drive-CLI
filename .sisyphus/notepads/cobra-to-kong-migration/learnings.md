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
