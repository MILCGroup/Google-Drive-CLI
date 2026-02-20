# Cobra-to-Kong Migration Decisions

## [2026-02-20] Task 6: Files domain migration
- Kept `internal/cli/root.go` untouched per migration constraints; addressed kong incompatibilities inside files command structs only.
- Implemented revisions listing as `FilesRevisionsListCmd` under `FilesRevisionsCmd` with `default:"withargs"` because kong v1.14 forbids positional args on branching command nodes.
- Renamed conflicting local flags to non-global names so parser initializes and help/build/tests run successfully.
