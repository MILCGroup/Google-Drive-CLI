# Contribute Advanced Drive APIs to gogcli

## TL;DR

> **Quick Summary**: Open 5 well-crafted feature request issues on steipete/gogcli proposing advanced Drive APIs that gdrv has already implemented: Changes, Permissions Auditing, Activity, Labels, and Admin SDK.
> 
> **Deliverables**:
> - 5 GitHub issues opened on steipete/gogcli
> - Each using Conventional Commits title format
> - Each under ~300 words with concrete `gog` CLI syntax
> - Each verified as OPEN after creation
> 
> **Estimated Effort**: Short
> **Parallel Execution**: YES - limited (sequential creation with 60s spacing, parallel research/drafting)
> **Critical Path**: Research overlap → Draft all 5 → File sequentially

---

## Context

### Original Request
Compare gdrv against steipete/gogcli, then contribute the features we've mastered as a series of issues on their repo.

### Interview Summary
**Key Discussions**:
- gdrv has 6 features gogcli lacks: Activity API, Changes API, Labels API, Permission Auditing, Admin SDK, Safety Layer
- gogcli has 4.3k stars, 69 open issues, none overlapping our advanced Drive APIs
- User wants to contribute upstream, starting with issues before PRs

**Research Findings**:
- gogcli ALREADY has `--dry-run` (global flag) + `confirmDestructive()` — Safety Layer issue dropped
- gogcli ALREADY has `groups` command via Cloud Identity API — Admin SDK must be carefully distinguished
- Stale PR #179 proposes "GAM feature parity for Workspace admin" — must be referenced
- Issue #232 (in-place file replace) already SHIPPED in v0.10.0 — don't cite as unresolved
- steipete engages with CONCISE issues (<300 words, concrete CLI syntax) — essays get ignored
- De facto issue format: Summary → Current behavior → Proposed → Use case → API notes → Related issues

### Metis Review
**Identified Gaps** (addressed):
- Safety Layer dropped as standalone issue (2/3 already implemented; idempotency mention folded into Permissions issue)
- Admin SDK reframed to distinguish Cloud Identity (existing) from Admin Directory API (proposed)
- Reordered by community impact: Changes → Permissions → Activity → Labels → Admin
- Issue length capped at ~300 words / ~2000 characters max
- All CLI examples use `gog` namespace, never `gdrv`
- Softer PR commitment language unless user explicitly commits

---

## Work Objectives

### Core Objective
Open 5 feature request issues on steipete/gogcli, each proposing a Google API that gdrv has proven in production, positioned to maximize maintainer engagement.

### Concrete Deliverables
- Issue 1: `feat(drive): add changes tracking (Changes API v3)` — sync, webhooks, change monitoring
- Issue 2: `feat(drive): add permission auditing and bulk operations` — security compliance
- Issue 3: `feat(drive): add activity query for audit trails (Activity API v2)` — compliance
- Issue 4: `feat(drive): add labels for structured file metadata (Labels API v2)` — taxonomy
- Issue 5: `feat(admin): add user and group management (Admin Directory API)` — provisioning

### Definition of Done
- [ ] 5 issues created and OPEN on steipete/gogcli
- [ ] Each title starts with `feat(`
- [ ] Each body contains `gog` CLI examples in code blocks
- [ ] Zero occurrences of `gdrv` in any issue body
- [ ] Each body < 2000 characters
- [ ] Each has Google API documentation link
- [ ] All 5 issue URLs collected and reported

### Must Have
- Conventional Commits title format (`feat(scope): description`)
- Concrete `gog` CLI syntax proposals in bash code blocks
- Google API documentation links
- Use case explanation for each feature
- Pre-flight overlap check before each issue creation

### Must NOT Have (Guardrails)
- NO references to `gdrv` command syntax in issue bodies (use `gog` throughout)
- NO references to cobra, RequestContext, Manager pattern, or gdrv internals
- NO essay-length issues (>2000 chars body) — steipete ignores them
- NO promises to submit PRs (use softer language: "I've implemented this in another Drive CLI and can share details")
- NO claiming Safety Layer is novel (gogcli already has `--dry-run` + `confirmDestructive()`)
- NO citing Issue #232 as unresolved demand (it shipped in v0.10.0)
- NO filing all issues simultaneously (minimum 60-second spacing)
- NO proposing changes to existing commands (new subcommands only)
- NO full auth/scope discussions (one-line scope note only)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: N/A (no code changes)
- **Automated tests**: None
- **Framework**: N/A

### QA Policy
Every task verifies its issue was created correctly using `gh issue view`.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Issue creation**: Use Bash (`gh issue create`) — Create issue, verify state
- **Content validation**: Use Bash (`gh issue view --json`) — Assert fields match

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — research + drafting):
├── Task 1: Pre-flight overlap check for all 5 keywords [quick]
├── Task 2: Draft Issue 1 — Changes API [writing]
├── Task 3: Draft Issue 2 — Permissions Auditing [writing]
├── Task 4: Draft Issue 3 — Activity API [writing]
├── Task 5: Draft Issue 4 — Labels API [writing]
├── Task 6: Draft Issue 5 — Admin SDK [writing]

Wave 2 (After Wave 1 — sequential filing with spacing):
├── Task 7: File Issue 1 — Changes API (depends: 1, 2) [quick]
├── Task 8: File Issue 2 — Permissions Auditing (depends: 1, 3, 7) [quick]
├── Task 9: File Issue 3 — Activity API (depends: 1, 4, 8) [quick]
├── Task 10: File Issue 4 — Labels API (depends: 1, 5, 9) [quick]
├── Task 11: File Issue 5 — Admin SDK (depends: 1, 6, 10) [quick]

Wave FINAL (After ALL tasks — verification):
├── Task F1: Verify all 5 issues are OPEN and correctly formatted [quick]
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|---|---|---|---|
| 1 | — | 7-11 | 1 |
| 2 | — | 7 | 1 |
| 3 | — | 8 | 1 |
| 4 | — | 9 | 1 |
| 5 | — | 10 | 1 |
| 6 | — | 11 | 1 |
| 7 | 1, 2 | 8 | 2 |
| 8 | 1, 3, 7 | 9 | 2 |
| 9 | 1, 4, 8 | 10 | 2 |
| 10 | 1, 5, 9 | 11 | 2 |
| 11 | 1, 6, 10 | F1 | 2 |
| F1 | 7-11 | — | FINAL |

### Agent Dispatch Summary

- **Wave 1**: **6 tasks** — T1 → `quick`, T2-T6 → `writing`
- **Wave 2**: **5 tasks** — T7-T11 → `quick` (sequential, 60s spacing)
- **FINAL**: **1 task** — F1 → `quick`

---

## TODOs

- [ ] 1. Pre-flight Overlap Check

  **What to do**:
  - For each of the 5 keywords, run `gh search issues --repo steipete/gogcli "<keyword>"` to verify no existing issues overlap:
    - `"drive changes API"` or `"changes tracking"` or `"sync changes"`
    - `"permission audit"` or `"bulk permissions"` or `"permission analysis"`
    - `"drive activity"` or `"activity API"` or `"audit trail"`
    - `"drive labels API"` or `"file labels"` or `"metadata labels"`
    - `"admin directory"` or `"admin SDK"` or `"user management admin"`
  - If ANY overlap found: report it and skip that issue, adjusting the plan
  - If no overlap: confirm all 5 are clear to file

  **Must NOT do**:
  - Do NOT file any issues during this task — research only

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
    - No specialized skills needed — simple gh CLI commands

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2-6)
  - **Blocks**: Tasks 7, 8, 9, 10, 11
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gogcli existing issues list: `gh issue list --repo steipete/gogcli --state all --limit 100`

  **External References**:
  - GitHub CLI search docs: `https://cli.github.com/manual/gh_search_issues`

  **Acceptance Criteria**:

  **QA Scenarios**:

  ```
  Scenario: All 5 feature areas have zero existing issues
    Tool: Bash (gh search)
    Steps:
      1. gh search issues --repo steipete/gogcli "drive changes API" --json title,number
      2. gh search issues --repo steipete/gogcli "permission audit" --json title,number
      3. gh search issues --repo steipete/gogcli "drive activity API" --json title,number
      4. gh search issues --repo steipete/gogcli "drive labels API" --json title,number
      5. gh search issues --repo steipete/gogcli "admin directory API" --json title,number
      6. For each: assert result count == 0 (or results are clearly unrelated)
    Expected Result: No overlapping issues found for any of the 5 feature areas
    Evidence: .sisyphus/evidence/task-1-overlap-check.json
  ```

  **Commit**: NO

- [ ] 2. Draft Issue 1 — Drive Changes API v3

  **What to do**:
  - Draft a GitHub issue body (<300 words, <2000 chars) proposing Drive Changes API v3 support
  - Title: `feat(drive): add changes tracking for sync and automation (Changes API v3)`
  - Structure: Summary → Current State → Proposed Commands → Use Cases → API Notes → Related
  - Proposed CLI interface (use `gog` namespace):
    ```bash
    gog drive changes start-token                    # Get starting page token
    gog drive changes list --token <token>           # List changes since token
    gog drive changes list --token <token> --max 50  # With pagination
    gog drive changes watch --token <token> --webhook-url <url>  # Set up webhook
    gog drive changes stop <channelId> <resourceId>  # Stop webhook
    ```
  - Mention use cases: sync tools, incremental backups, change monitoring, automation triggers
  - Include API link: https://developers.google.com/drive/api/v3/reference/changes
  - Mention required scope: standard Drive scope (no additional scopes needed)
  - Closing note: "I've implemented this in another Drive CLI and can share implementation details if helpful."
  - Save draft to `.sisyphus/drafts/gogcli-issue-1-changes.md`

  **Must NOT do**:
  - Do NOT use `gdrv` syntax anywhere
  - Do NOT exceed 2000 characters in body
  - Do NOT propose changes to existing `gog drive` commands
  - Do NOT include full auth/scope discussion

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3-6)
  - **Blocks**: Task 7
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gdrv Changes implementation: `internal/changes/manager.go` — reference for proposed command structure
  - gdrv Changes CLI: `internal/cli/changes.go` — reference for flags and subcommands

  **External References**:
  - Google Drive Changes API v3: https://developers.google.com/drive/api/v3/reference/changes
  - gogcli existing Drive commands: `gog drive --help` (ls, search, upload, download, mkdir, rename, move, delete, permissions, share, drives)
  - Example well-received gogcli issue: Issue #220 on steipete/gogcli — short, concrete, got fast positive response

  **WHY Each Reference Matters**:
  - `internal/changes/manager.go`: Shows the command structure and flags that work in practice — adapt to `gog` namespace
  - Issue #220: Template for concise issue format that steipete actually engages with

  **Acceptance Criteria**:
  - [ ] Draft saved to `.sisyphus/drafts/gogcli-issue-1-changes.md`
  - [ ] Body < 2000 characters
  - [ ] Contains `gog drive changes` commands in code block
  - [ ] Contains API documentation link
  - [ ] Zero occurrences of `gdrv`

  **QA Scenarios**:

  ```
  Scenario: Draft meets length and content requirements
    Tool: Bash (wc + grep)
    Steps:
      1. wc -c .sisyphus/drafts/gogcli-issue-1-changes.md — assert < 2000
      2. grep -c "gog drive changes" — assert >= 1
      3. grep -c "gdrv" — assert == 0
      4. grep -c "developers.google.com" — assert >= 1
    Expected Result: Draft under 2000 chars, contains gog syntax, no gdrv, has API link
    Evidence: .sisyphus/evidence/task-2-draft-validation.txt
  ```

  **Commit**: NO

- [ ] 3. Draft Issue 2 — Permission Auditing and Bulk Operations

  **What to do**:
  - Draft a GitHub issue body (<300 words, <2000 chars) proposing permission auditing capabilities
  - Title: `feat(drive): add permission auditing and bulk operations`
  - Structure: Summary → Current State (acknowledge existing `drive permissions/share/unshare`) → Proposed Commands → Use Cases → API Notes → Related
  - Proposed CLI interface:
    ```bash
    gog drive audit public                           # Find files with public access
    gog drive audit external --domain example.com    # Find externally shared files
    gog drive audit user user@example.com            # Audit a specific user's access
    gog drive bulk remove-public --parent <folderId> --dry-run  # Bulk remove public links
    gog drive bulk update-role --parent <folderId> --from writer --to reader --dry-run
    ```
  - Frame as ADDITIVE to existing `drive share`/`unshare`/`permissions` commands
  - Use cases: security compliance, external sharing audits, bulk permission cleanup, risk assessment
  - Note: Uses existing Drive Permissions API (no additional scopes)
  - Mention idempotency: "Bulk operations could benefit from idempotency tracking to safely resume interrupted runs"
  - Reference existing `confirmDestructive()` pattern for bulk operations
  - Closing note: "I've implemented this in another Drive CLI and can share implementation details if helpful."

  **Must NOT do**:
  - Do NOT propose replacing existing permission commands
  - Do NOT frame Safety Layer as novel
  - Do NOT exceed 2000 characters

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4-6)
  - **Blocks**: Task 8
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gdrv Permissions implementation: `internal/permissions/manager.go` — audit, analyze, bulk operations
  - gdrv Permissions CLI: `internal/cli/permissions.go` — flags and subcommand structure
  - gogcli existing Drive permissions: `gog drive permissions`, `gog drive share`, `gog drive unshare`

  **External References**:
  - Google Drive Permissions API: https://developers.google.com/drive/api/v3/reference/permissions
  - gogcli Issue #291: "Security: Add confirmDestructive() guards for dangerous operations" — shows safety awareness

  **Acceptance Criteria**:
  - [ ] Draft saved to `.sisyphus/drafts/gogcli-issue-2-permissions.md`
  - [ ] Body < 2000 characters
  - [ ] Acknowledges existing `drive share`/`unshare` commands
  - [ ] Contains `gog drive audit` or `gog drive bulk` in code block
  - [ ] Zero occurrences of `gdrv`

  **QA Scenarios**:

  ```
  Scenario: Draft meets requirements and acknowledges existing commands
    Tool: Bash (wc + grep)
    Steps:
      1. wc -c .sisyphus/drafts/gogcli-issue-2-permissions.md — assert < 2000
      2. grep -c "gog drive" — assert >= 2
      3. grep -c "gdrv" — assert == 0
      4. grep -ci "existing\|already\|current" — assert >= 1 (acknowledges existing commands)
    Expected Result: Draft acknowledges existing permission commands, proposes additive features
    Evidence: .sisyphus/evidence/task-3-draft-validation.txt
  ```

  **Commit**: NO

- [ ] 4. Draft Issue 3 — Drive Activity API v2

  **What to do**:
  - Draft a GitHub issue body (<300 words, <2000 chars) proposing Drive Activity API v2 support
  - Title: `feat(drive): add activity query for audit trails (Activity API v2)`
  - Structure: Summary → Current State → Proposed Commands → Use Cases → API Notes → Related
  - MUST proactively distinguish from Gmail History API / `gmail watch serve` — "This is the Drive Activity API, separate from Gmail history"
  - Proposed CLI interface:
    ```bash
    gog drive activity query                          # Recent activity across Drive
    gog drive activity query --file <fileId>          # Activity for specific file
    gog drive activity query --folder <folderId>      # Activity for folder tree
    gog drive activity query --user user@example.com  # Activity by user
    gog drive activity query --actions edit,share --from 2026-01-01T00:00:00Z
    ```
  - Use cases: compliance auditing, security monitoring, access tracking, incident investigation
  - Include API link: https://developers.google.com/drive/activity/v2
  - Note additional scope: `https://www.googleapis.com/auth/drive.activity.readonly`

  **Must NOT do**:
  - Do NOT confuse with Gmail History API
  - Do NOT exceed 2000 characters

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-3, 5-6)
  - **Blocks**: Task 9
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gdrv Activity implementation: `internal/activity/manager.go` — query structure and filtering
  - gdrv Activity CLI: `internal/cli/activity.go` — flags and output format

  **External References**:
  - Google Drive Activity API v2: https://developers.google.com/drive/activity/v2
  - API reference: https://developers.google.com/drive/activity/v2/reference/rest/v2/activity/query

  **Acceptance Criteria**:
  - [ ] Draft saved to `.sisyphus/drafts/gogcli-issue-3-activity.md`
  - [ ] Body < 2000 characters
  - [ ] Distinguishes from Gmail History API
  - [ ] Contains `gog drive activity` commands
  - [ ] Zero occurrences of `gdrv`

  **QA Scenarios**:

  ```
  Scenario: Draft distinguishes Activity API from Gmail
    Tool: Bash (grep)
    Steps:
      1. grep -ci "not gmail\|separate from gmail\|drive activity API\|distinct from" .sisyphus/drafts/gogcli-issue-3-activity.md — assert >= 1
      2. grep -c "gdrv" — assert == 0
      3. wc -c — assert < 2000
    Expected Result: Draft clearly distinguishes Drive Activity from Gmail history
    Evidence: .sisyphus/evidence/task-4-draft-validation.txt
  ```

  **Commit**: NO

- [ ] 5. Draft Issue 4 — Drive Labels API v2

  **What to do**:
  - Draft a GitHub issue body (<300 words, <2000 chars) proposing Drive Labels API v2 support
  - Title: `feat(drive): add labels for structured file metadata (Labels API v2)`
  - MUST proactively clarify: "This is the Drive Labels API v2 for structured file metadata, not Gmail labels (which gogcli already handles)"
  - Proposed CLI interface:
    ```bash
    gog drive labels list                             # List available label schemas
    gog drive labels get <labelId>                    # Get label schema details
    gog drive labels file list <fileId>               # Labels applied to a file
    gog drive labels file apply <fileId> <labelId> --fields '{"key":"value"}'
    gog drive labels file remove <fileId> <labelId>
    ```
  - Use cases: enterprise content classification, retention policies, project tagging, workflow automation
  - Include API link: https://developers.google.com/drive/labels/overview
  - Note additional scope: `https://www.googleapis.com/auth/drive.labels`

  **Must NOT do**:
  - Do NOT confuse with Gmail labels (gogcli already has `gmail labels`)
  - Do NOT propose admin-only label creation in the initial issue (keep it scoped to apply/read)

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-4, 6)
  - **Blocks**: Task 10
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gdrv Labels implementation: `internal/labels/manager.go` — label operations
  - gdrv Labels CLI: `internal/cli/labels.go` — command structure

  **External References**:
  - Google Drive Labels API v2: https://developers.google.com/drive/labels/overview
  - API reference: https://developers.google.com/drive/labels/reference/rest

  **Acceptance Criteria**:
  - [ ] Draft saved to `.sisyphus/drafts/gogcli-issue-4-labels.md`
  - [ ] Body < 2000 characters
  - [ ] Explicitly distinguishes from Gmail labels
  - [ ] Contains `gog drive labels` commands
  - [ ] Zero occurrences of `gdrv`

  **QA Scenarios**:

  ```
  Scenario: Draft distinguishes Drive Labels from Gmail labels
    Tool: Bash (grep)
    Steps:
      1. grep -ci "not gmail labels\|separate from gmail\|file metadata\|structured metadata" .sisyphus/drafts/gogcli-issue-4-labels.md — assert >= 1
      2. grep -c "gdrv" — assert == 0
      3. wc -c — assert < 2000
    Expected Result: Draft clearly distinguishes Drive Labels from Gmail labels
    Evidence: .sisyphus/evidence/task-5-draft-validation.txt
  ```

  **Commit**: NO

- [ ] 6. Draft Issue 5 — Admin Directory API

  **What to do**:
  - Draft a GitHub issue body (<300 words, <2000 chars) proposing Admin Directory API support
  - Title: `feat(admin): add user and group management (Admin Directory API)`
  - MUST acknowledge existing `groups` command (Cloud Identity API) and distinguish Admin Directory API
  - MUST reference stale PR #179 ("GAM feature parity") — "This relates to the scope discussed in PR #179"
  - Proposed CLI interface:
    ```bash
    gog admin users list --domain example.com
    gog admin users get user@example.com
    gog admin users create user@example.com --given "John" --family "Doe" --password "..."
    gog admin users suspend user@example.com
    gog admin groups list --domain example.com
    gog admin groups members list group@example.com
    gog admin groups members add group@example.com user@example.com --role MEMBER
    ```
  - Frame as complementing existing `groups` command: "The existing `gog groups` command uses Cloud Identity for group listing. This proposal adds Admin Directory API for full user provisioning and group management with domain-wide admin capabilities."
  - Use cases: user provisioning, onboarding/offboarding, group management, audit
  - Note: Requires service account with domain-wide delegation (matches existing SA support in gogcli)
  - Scopes: `admin.directory.user`, `admin.directory.group`

  **Must NOT do**:
  - Do NOT ignore existing `groups` command
  - Do NOT ignore PR #179
  - Do NOT propose OU management, device management, or admin roles (scope creep)
  - Do NOT exceed 2000 characters

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-5)
  - **Blocks**: Task 11
  - **Blocked By**: None

  **References**:

  **Pattern References**:
  - gdrv Admin implementation: `internal/admin/manager.go` — user/group CRUD operations
  - gdrv Admin CLI: `internal/cli/admin.go` — command structure and flags
  - gogcli existing groups: `gog groups list`, `gog groups members` (Cloud Identity API)

  **External References**:
  - Admin SDK Directory API: https://developers.google.com/admin-sdk/directory/reference/rest/v1/users
  - gogcli PR #179: Stale PR proposing "GAM feature parity for Workspace admin"
  - gogcli existing groups via Cloud Identity: `gog groups list` / `gog groups members`

  **WHY Each Reference Matters**:
  - PR #179: Must be referenced to show awareness of existing conversation
  - `gog groups`: Must be acknowledged so the issue doesn't look uninformed

  **Acceptance Criteria**:
  - [ ] Draft saved to `.sisyphus/drafts/gogcli-issue-5-admin.md`
  - [ ] Body < 2000 characters
  - [ ] References PR #179
  - [ ] Acknowledges existing `groups` command
  - [ ] Distinguishes Admin Directory from Cloud Identity
  - [ ] Contains `gog admin` commands
  - [ ] Zero occurrences of `gdrv`

  **QA Scenarios**:

  ```
  Scenario: Draft references PR #179 and existing groups command
    Tool: Bash (grep)
    Steps:
      1. grep -c "#179\|PR 179" .sisyphus/drafts/gogcli-issue-5-admin.md — assert >= 1
      2. grep -ci "existing.*groups\|groups.*command\|cloud identity" — assert >= 1
      3. grep -c "gdrv" — assert == 0
      4. wc -c — assert < 2000
    Expected Result: Draft references PR #179 and distinguishes from existing groups
    Evidence: .sisyphus/evidence/task-6-draft-validation.txt
  ```

  **Commit**: NO

- [ ] 7. File Issue 1 — Drive Changes API v3

  **What to do**:
  - Verify no new overlapping issues appeared since Task 1 check: `gh search issues --repo steipete/gogcli "drive changes"`
  - Read the draft from `.sisyphus/drafts/gogcli-issue-1-changes.md`
  - File the issue: `gh issue create --repo steipete/gogcli --title "feat(drive): add changes tracking for sync and automation (Changes API v3)" --body-file .sisyphus/drafts/gogcli-issue-1-changes.md`
  - Capture the issue number and URL
  - Verify: `gh issue view <N> --repo steipete/gogcli --json title,state,url`
  - Assert state == OPEN

  **Must NOT do**:
  - Do NOT modify the draft at this stage (if issues found, report and skip)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 — Sequential (first in sequence)
  - **Blocks**: Task 8
  - **Blocked By**: Tasks 1, 2

  **Acceptance Criteria**:
  - [ ] Issue created on steipete/gogcli
  - [ ] Issue state is OPEN
  - [ ] Title matches expected format

  **QA Scenarios**:

  ```
  Scenario: Issue 1 created and verified
    Tool: Bash (gh)
    Steps:
      1. gh issue create --repo steipete/gogcli --title "feat(drive): ..." --body-file ...
      2. Capture output URL and issue number
      3. gh issue view <N> --repo steipete/gogcli --json state,title
      4. Assert .state == "OPEN"
      5. Assert .title starts with "feat(drive):"
    Expected Result: Issue created with OPEN state and correct title
    Evidence: .sisyphus/evidence/task-7-issue-1-created.json
  ```

  **Commit**: NO

- [ ] 8. File Issue 2 — Permission Auditing (60s after Task 7)

  **What to do**:
  - Wait 60 seconds after Task 7 completes (rate limiting courtesy)
  - Verify no overlap: `gh search issues --repo steipete/gogcli "permission audit"`
  - File: `gh issue create --repo steipete/gogcli --title "feat(drive): add permission auditing and bulk operations" --body-file .sisyphus/drafts/gogcli-issue-2-permissions.md`
  - Capture issue number, verify OPEN state

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 — Sequential (second)
  - **Blocks**: Task 9
  - **Blocked By**: Tasks 1, 3, 7

  **Acceptance Criteria**:
  - [ ] Issue created and OPEN
  - [ ] 60+ seconds after Task 7

  **QA Scenarios**:

  ```
  Scenario: Issue 2 created with proper spacing
    Tool: Bash (gh)
    Steps:
      1. sleep 60
      2. gh issue create --repo steipete/gogcli --title "feat(drive): ..." --body-file ...
      3. gh issue view <N> --repo steipete/gogcli --json state — assert OPEN
    Expected Result: Issue created, OPEN, with 60s delay
    Evidence: .sisyphus/evidence/task-8-issue-2-created.json
  ```

  **Commit**: NO

- [ ] 9. File Issue 3 — Activity API v2 (60s after Task 8)

  **What to do**:
  - Wait 60 seconds after Task 8
  - Verify no overlap: `gh search issues --repo steipete/gogcli "drive activity"`
  - File: `gh issue create --repo steipete/gogcli --title "feat(drive): add activity query for audit trails (Activity API v2)" --body-file .sisyphus/drafts/gogcli-issue-3-activity.md`
  - Capture issue number, verify OPEN state

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 — Sequential (third)
  - **Blocks**: Task 10
  - **Blocked By**: Tasks 1, 4, 8

  **Acceptance Criteria**:
  - [ ] Issue created and OPEN
  - [ ] 60+ seconds after Task 8

  **QA Scenarios**:

  ```
  Scenario: Issue 3 created
    Tool: Bash (gh)
    Steps:
      1. sleep 60
      2. gh issue create --repo steipete/gogcli --title "feat(drive): ..." --body-file ...
      3. gh issue view <N> --json state — assert OPEN
    Expected Result: Issue created and OPEN
    Evidence: .sisyphus/evidence/task-9-issue-3-created.json
  ```

  **Commit**: NO

- [ ] 10. File Issue 4 — Labels API v2 (60s after Task 9)

  **What to do**:
  - Wait 60 seconds after Task 9
  - Verify no overlap: `gh search issues --repo steipete/gogcli "drive labels API"`
  - File: `gh issue create --repo steipete/gogcli --title "feat(drive): add labels for structured file metadata (Labels API v2)" --body-file .sisyphus/drafts/gogcli-issue-4-labels.md`
  - Capture issue number, verify OPEN state

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 — Sequential (fourth)
  - **Blocks**: Task 11
  - **Blocked By**: Tasks 1, 5, 9

  **Acceptance Criteria**:
  - [ ] Issue created and OPEN

  **QA Scenarios**:

  ```
  Scenario: Issue 4 created
    Tool: Bash (gh)
    Steps:
      1. sleep 60
      2. gh issue create --repo steipete/gogcli --title "feat(drive): ..." --body-file ...
      3. gh issue view <N> --json state — assert OPEN
    Expected Result: Issue created and OPEN
    Evidence: .sisyphus/evidence/task-10-issue-4-created.json
  ```

  **Commit**: NO

- [ ] 11. File Issue 5 — Admin Directory API (60s after Task 10)

  **What to do**:
  - Wait 60 seconds after Task 10
  - Verify no overlap: `gh search issues --repo steipete/gogcli "admin directory"`
  - File: `gh issue create --repo steipete/gogcli --title "feat(admin): add user and group management (Admin Directory API)" --body-file .sisyphus/drafts/gogcli-issue-5-admin.md`
  - Capture issue number, verify OPEN state

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 — Sequential (fifth/last)
  - **Blocks**: Task F1
  - **Blocked By**: Tasks 1, 6, 10

  **Acceptance Criteria**:
  - [ ] Issue created and OPEN

  **QA Scenarios**:

  ```
  Scenario: Issue 5 created
    Tool: Bash (gh)
    Steps:
      1. sleep 60
      2. gh issue create --repo steipete/gogcli --title "feat(admin): ..." --body-file ...
      3. gh issue view <N> --json state — assert OPEN
    Expected Result: Issue created and OPEN
    Evidence: .sisyphus/evidence/task-11-issue-5-created.json
  ```

  **Commit**: NO

---

## Final Verification Wave

- [ ] F1. **Verify All Issues Created and Formatted**

  **What to do**:
  - Run `gh issue list --repo steipete/gogcli --author @me --state open --json number,title,url` to get all created issues
  - For each issue, run `gh issue view <N> --repo steipete/gogcli --json title,body,state,url`
  - Assert: all 5 issues have state `OPEN`
  - Assert: all 5 titles start with `feat(`
  - Assert: no issue body contains the string `gdrv`
  - Assert: each issue body contains at least one `gog` command in a code block
  - Assert: each issue body contains a Google API documentation link
  - Collect all 5 URLs into a summary report

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **QA Scenarios**:

  ```
  Scenario: All 5 issues exist and are OPEN
    Tool: Bash (gh)
    Steps:
      1. gh issue list --repo steipete/gogcli --author @me --state open --json number,title,url
      2. Assert: JSON array length == 5
      3. Assert: each title starts with "feat("
    Expected Result: 5 issues listed, all OPEN
    Evidence: .sisyphus/evidence/task-F1-all-issues-open.json

  Scenario: No issue contains "gdrv"
    Tool: Bash (gh + grep)
    Steps:
      1. For each issue number, gh issue view <N> --repo steipete/gogcli --json body
      2. grep -c "gdrv" — must return 0 for each
    Expected Result: Zero occurrences of "gdrv" across all 5 issue bodies
    Evidence: .sisyphus/evidence/task-F1-no-gdrv-references.txt
  ```

  **Commit**: NO

---

## Commit Strategy

No commits — this plan only creates GitHub issues on an external repository.

---

## Success Criteria

### Verification Commands
```bash
gh issue list --repo steipete/gogcli --author @me --state open --json number,title,url
# Expected: 5 issues, all with feat( prefix titles
```

### Final Checklist
- [ ] 5 issues created on steipete/gogcli
- [ ] All issues OPEN
- [ ] All titles use Conventional Commits format
- [ ] All bodies under 2000 characters
- [ ] All bodies contain `gog` CLI examples
- [ ] Zero `gdrv` references in any issue
- [ ] All URLs reported to user
