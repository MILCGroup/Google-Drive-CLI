# Cobra-to-Kong Migration Problems

## [2026-02-20] Task 6: Files domain migration
- Unresolved compatibility gap: desired cobra-equivalent local flags (`--output`, local `--force`) cannot coexist with current global flags in kong v1.14 without changing root/global flag names.
- Unresolved parity gap: ideal mixed node shape (`files revisions <file-id>` plus nested subcommands on same struct) is not supported directly by kong v1.14.
