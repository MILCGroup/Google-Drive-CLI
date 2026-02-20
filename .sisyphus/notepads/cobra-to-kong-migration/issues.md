# Cobra-to-Kong Migration Issues

## [2026-02-20] Task 6: Files domain migration
- Runtime panic encountered after first migration attempt: `duplicate flag --output` from kong parser.
- Root cause: global `Globals.Output` (`--output`) and command-local files flags with the same long name conflict in kong v1.14.
- Additional conflict identified for local delete `--force` vs global `--force`.
