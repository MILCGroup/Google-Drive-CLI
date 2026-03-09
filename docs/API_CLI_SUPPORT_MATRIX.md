# CLI and API Support Matrix

This document maps which APIs have CLI commands vs which are monitored for drift.

## REST APIs

| API | CLI Command | CLI Status | Monitored | Drift Status | Notes |
|-----|-------------|------------|-----------|--------------|-------|
| **Drive** | `drive`, `files`, `folders`, `drives` | ✅ Full | ✅ Yes | ✅ Active | Core API with full support |
| **Gmail** | `gmail` | ✅ Full | ✅ Yes | ✅ Active | Complete email operations |
| **Calendar** | `calendar` | ✅ Full | ✅ Yes | ✅ Active | Events and free/busy |
| **People** | `people` | ✅ Full | ✅ Yes | ✅ Active | Contacts and directory |
| **Slides** | `slides` | ✅ Full | ✅ Yes | ✅ Active | Presentations |
| **Docs** | `docs` | ✅ Full | ✅ Yes | ✅ Active | Documents |
| **Sheets** | `sheets` | ✅ Full | ✅ Yes | ✅ Active | Spreadsheets |
| **Chat** | `chat` | ✅ Full | ✅ Yes | ✅ Active | Spaces, messages, members |
| **Admin** | `admin` | ✅ Full | ✅ Yes | ✅ Active | Directory, users, groups |
| **Cloud Identity** | `groups` | ✅ Full | ✅ Yes | ✅ Active | Groups (distinct from admin groups) |
| **Apps Script** | `appscript` | ✅ Full | ✅ Yes | ✅ Active | Script projects |
| **Tasks** | `tasks` | ✅ Full | ✅ Yes | ✅ Active | Task lists |
| **Forms** | `forms` | ⚠️ Limited | ❌ No | ⏸️ Custom discovery | Only get/responses/create commands |
| **Drive Labels** | `labels` | ⚠️ Limited | ❌ No | ⏸️ Custom discovery | Basic label operations |

### Missing CLI for Monitored APIs: None ✅

All 13 monitored REST APIs have CLI commands implemented.

### Missing Monitoring for CLI APIs:
- **Forms** - Uses `forms.googleapis.com/$discovery/rest` (custom URL)
- **Drive Labels** - Uses `drivelabels.googleapis.com/$discovery/rest` (custom URL)

## gRPC APIs

| API | CLI Command | CLI Status | Monitored | Drift Status | Notes |
|-----|-------------|------------|-----------|--------------|-------|
| **Monitoring** | `monitoring` | ⚠️ Limited | ✅ Yes | ✅ Active | Uses gRPC, not REST discovery |
| **Logging** | `logging` | ⚠️ Limited | ✅ Yes | ✅ Active | Uses gRPC, not REST discovery |
| **IAM Admin** | `iamadmin` | ⚠️ Limited | ✅ Yes | ✅ Active | Uses gRPC, not REST discovery |
| **AI/Generative** | `ai` | ⚠️ Limited | ✅ Yes | ✅ Active | Uses gRPC, not REST discovery |
| **Meet** | `meet` | ⚠️ Limited | ❌ No | ❌ Not in CI | Uses gRPC, needs proto monitoring |

### Notes on gRPC APIs:
- All gRPC APIs use **Protocol Buffers** (not Discovery API)
- CLI commands are mostly stubs/limited implementations
- gRPC drift detection workflow monitors 4 main APIs
- **Meet API** is missing from drift detection - should be added

## Summary

### ✅ Complete (CLI + Monitoring)
- Drive, Gmail, Calendar, People, Slides, Docs, Sheets
- Chat, Admin, Cloud Identity, Apps Script, Tasks
- Monitoring, Logging, IAM Admin, AI (gRPC)

### ⚠️ Partial (CLI only, no monitoring)
- Forms (needs custom discovery handler)
- Drive Labels (needs custom discovery handler)
- Meet (gRPC - needs proto monitoring added)

### ❌ No CLI (not applicable)
- Changes API (uses Drive scopes, no separate CLI)
- Activity API (uses Drive scopes, no separate CLI)

## Recommendations

### High Priority
1. **Add Meet API to gRPC drift detection** - CLI exists but no monitoring

### Medium Priority  
2. **Add custom discovery URL handler** for Forms and Drive Labels
   - Modify discovery-checker to support service-specific discovery URLs
   - Currently blocks on standard Discovery API 404s

### Low Priority
3. **Enhance CLI implementations** for gRPC APIs (currently stub commands)
