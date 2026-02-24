# Cobra-to-Kong Migration Issues

## [2026-02-20] Task 6: Files domain migration
- Runtime panic encountered after first migration attempt: `duplicate flag --output` from kong parser.
- Root cause: global `Globals.Output` (`--output`) and command-local files flags with the same long name conflict in kong v1.14.
- Additional conflict identified for local delete `--force` vs global `--force`.

## [2026-02-20] Task 6 follow-up: attempted flag parity restore
- Reproduced failure when restoring local flags to cobra names: `FilesDownloadCmd.Output: duplicate flag --output`.
- Repro command: `go build -o bin/gdrv ./cmd/gdrv && bin/gdrv files download --help` (panic occurs at runtime parser initialization).
- Observed behavior confirms kong v1.14 rejects duplicate long flag names inherited from globals; local shadowing of parent flags is not accepted.


## [2026-02-20] Task: Auth domain migration to kong
- No migration blockers encountered for auth command flags; required `--key-file` constraint works with kong `required:""` tag.
- LSP diagnostics on changed files are clean after migration.
