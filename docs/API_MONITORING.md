# API Monitoring Summary

This document lists all Google APIs monitored for drift detection.

## REST APIs (Discovery API)

Monitored via [`.github/workflows/api-drift.yml`](../.github/workflows/api-drift.yml)
Schedule: Weekly, Sundays at 2 AM UTC

### Active Monitoring (13 APIs)

| Service | Version | Priority | Description | Status |
|---------|---------|----------|-------------|--------|
| drive | v3 | 1 | File storage and management | ✅ Active |
| gmail | v1 | 1 | Email operations | ✅ Active |
| calendar | v3 | 1 | Calendar and event management | ✅ Active |
| people | v1 | 2 | Contact and profile management | ✅ Active |
| slides | v1 | 2 | Presentation management | ✅ Active |
| docs | v1 | 2 | Document management | ✅ Active |
| sheets | v4 | 2 | Spreadsheet management | ✅ Active |
| chat | v1 | 2 | Chat and messaging | ✅ Active |
| admin | directory_v1 | 2 | User and group management | ✅ Active |
| cloudidentity | v1 | 2 | Group management | ✅ Active |
| script | v1 | 3 | Script project management | ✅ Active |
| tasks | v1 | 3 | Task list management | ✅ Active |

### Pending (Service-Specific Discovery)

These APIs use non-standard discovery URLs and require custom handling:

| Service | Version | Issue | Discovery URL |
|---------|---------|-------|---------------|
| drivelabels | v2 | Custom URL | `https://drivelabels.googleapis.com/$discovery/rest` |
| forms | v1 | Custom URL | `https://forms.googleapis.com/$discovery/rest` |

## gRPC APIs (Protocol Buffers)

Monitored via [`.github/workflows/grpc-drift.yml`](../.github/workflows/grpc-drift.yml)
Schedule: Weekly, Sundays at 3 AM UTC

### Active Monitoring (4 APIs)

| Service | Version | Description | Proto Path |
|---------|---------|-------------|------------|
| monitoring | v3 | Cloud Monitoring | `google/monitoring/v3` |
| logging | v2 | Cloud Logging | `google/logging/v2` |
| iam | admin/v1 | IAM Admin | `google/iam/admin/v1` |
| ai | generativelanguage/v1 | Generative Language API | `google/ai/generativelanguage/v1` |

## Used But Not Yet Monitored

These APIs are imported in the codebase but not yet added to drift detection:

### REST APIs
- **drive labels v2** - Used in `internal/labels/` (requires custom discovery URL)

### gRPC APIs
All currently imported gRPC APIs are now monitored.

## Adding New APIs

### REST API
1. Add entry to `apis.yaml`:
   ```yaml
   - service: <name>
     version: <version>
     discovery_url: https://www.googleapis.com/discovery/v1/apis/<name>/<version>/rest
     owner: "@<team>"
     priority: <1-3>
     description: "<description>"
   ```
2. Run `./discovery-checker -config apis.yaml -generate -output .`
3. Commit the new snapshot
4. Push to trigger initial CI run

### gRPC API
1. Add proto path to `.github/workflows/grpc-drift.yml` in the `SERVICES` list
2. Workflow will auto-detect and create initial import on next run

## Risk Classification

Changes are classified into three categories:

### Additive (Auto-merge allowed)
- New optional fields
- New methods
- New optional parameters
- New scopes

### Risky (Requires owner approval)
- New required parameters
- Changed path templates
- Removed scopes
- Type changes in request/response

### Breaking (Blocks pipeline)
- Removed methods
- Removed required fields
- New required fields in request
- Changed auth requirements
- Incompatible type changes

## Workflow Behavior

### On Drift Detection
1. REST APIs: Generates JSON snapshots, creates PR with updates
2. gRPC APIs: Copies proto files, regenerates Go code, creates PR

### Notifications
- Breaking changes: Creates urgent GitHub issue
- Risky changes: Requires manual owner approval
- Additive changes: Auto-merge if tests pass

## Snapshot Locations

- **REST APIs**: `third_party/google/discovery/<service>/<version>.json`
- **gRPC APIs**: `third_party/googleapis/<proto_path>/`

## References

- [Discovery API Docs](https://developers.google.com/discovery)
- [Google APIs Repository](https://github.com/googleapis/googleapis)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
