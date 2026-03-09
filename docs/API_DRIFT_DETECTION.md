# API Drift Detection

This document describes the CI-based API drift detection system for gdrv, which monitors Google API specifications and generates typed code when changes are detected.

## Overview

Instead of runtime discovery (which has reliability and performance concerns), gdrv uses a **static code generation** approach with automated drift detection:

1. **Discovery documents** are fetched weekly via CI
2. **Drift detection** compares committed snapshots against current API specs
3. **Risk classification** categorizes changes as additive/risky/breaking
4. **Auto-generation** creates Pull Requests with updated type definitions
5. **Human review** ensures breaking changes are handled appropriately

## Architecture

```
┌──────────────────┐
│  Google          │
│  Discovery API   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  CI Job          │
│  (weekly)        │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Drift Check     │
│  + Classify      │
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌──────────┐
│  No     │ │  Yes     │
│  Changes│ │  Generate│
└─────────┘ └────┬─────┘
                 │
                 ▼
        ┌────────────────┐
        │  Create PR     │
        │  + Report      │
        └────────────────┘
```

## Schedule

**When**: Sundays at 2:00 AM UTC

**Timezone Notes**:
- PST (Pacific Standard Time, winter): Saturday 6:00 PM
- PDT (Pacific Daylight Time, summer): Saturday 7:00 PM
- EST (Eastern Standard Time, winter): Saturday 9:00 PM
- EDT (Eastern Daylight Time, summer): Saturday 10:00 PM
- GMT/BST (London): 2:00 AM / 3:00 AM

The timing ensures the job runs during off-hours in US timezones while being early enough in European mornings for quick response if needed.

## Security & Reliability Measures

### HTTPS Enforcement
All discovery URLs must use HTTPS. The CI workflow validates this before fetching:

```yaml
- name: Verify security - HTTPS only
  run: |
    if grep -E 'http://' apis.yaml; then
      echo "Error: HTTP URLs found in apis.yaml"
      exit 1
    fi
```

### Snapshot Normalization
Discovery documents are normalized before comparison to avoid noise:
- **Stable JSON ordering**: Maps, schemas, and methods sorted alphabetically
- **Volatile field removal**: `etag`, `revision`, timestamps excluded
- **Canonical formatting**: Consistent indentation and escaping
- **Empty vs omitted**: Empty arrays normalized consistently

### Change Classification

Changes are classified with a bias toward "risky" over "breaking" when ambiguous:

#### Additive (Auto-merge allowed)
- New optional fields
- New methods
- New optional parameters
- New OAuth scopes
- Enum value additions (configurable)

#### Risky (Requires owner approval)
- New required parameters
- Changed path templates
- Removed OAuth scopes
- Type changes (int32→int64, etc.)
- Format changes (date-time patterns)
- Fields becoming required
- Schema tightening (additionalProperties)
- Integer/number precision changes

#### Breaking (Blocks pipeline)
- Removed methods
- Removed required fields
- Required field additions to request
- HTTP method changes
- Authentication requirement changes
- Enum value removals
- Incompatible type changes (string↔integer)

### Edge Cases Handled

**Enum Expansions**: Classified as **risky** (not additive) by default because consumers may validate enum values. If your code uses plain string types without validation, you can set `EnumExpansionIsAdditive: true` to treat expansions as additive.

**Integer Formats**: int32→int64 is risky (widening), int64→int32 is risky (narrowing). int32→string is breaking.

**additionalProperties**: Schema tightening is risky (may reject previously valid payloads), schema relaxing is additive.

**OAuth Scope Changes**: Always at least "risky" - removed scopes may break existing auth flows.

### Toolchain Version Pinning

For reproducibility, all tools are pinned to specific versions:
- Go: `1.23.6` (specific patch version)
- goimports: `v0.24.0`

## Components

### 1. Allowlist (`apis.yaml`)

Defines which APIs are monitored:

```yaml
apis:
  - service: drive
    version: v3
    owner: "@drive-team"
    
  - service: gmail
    version: v1
    owner: "@gmail-team"
    
  - service: calendar
    version: v3
    owner: "@calendar-team"
```

### 2. Discovery Checker (`cmd/discovery-checker/`)

CLI tool that runs in CI:

```bash
# Check for drift
discovery-checker -config apis.yaml -check

# Generate code from discovery docs
discovery-checker -config apis.yaml -generate -output .
```

### 3. Snapshots (`third_party/google/discovery/`)

Committed JSON snapshots of discovery documents:

```
third_party/google/discovery/
├── drive/
│   └── v3.json
├── gmail/
│   └── v1.json
└── calendar/
    └── v3.json
```

### 4. Generated Code (`internal/google/generated/`)

- **Types**: Go structs for request/response payloads
- **Descriptors**: Machine-readable endpoint metadata
- **Executor**: Handwritten client that uses the descriptors

## Risk Classification

Changes are classified into three tiers:

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
- Changed authentication requirements
- Incompatible type changes

## Usage

### Manual Drift Check

```bash
# Build the checker
go build -o discovery-checker ./cmd/discovery-checker

# Check all APIs
./discovery-checker -config apis.yaml -check -verbose

# Check specific API
./discovery-checker -config apis.yaml -check -service drive
```

### Generate Code Locally

```bash
# Generate for all APIs
./discovery-checker -config apis.yaml -generate -output .

# Generate for specific API
./discovery-checker -config apis.yaml -generate -service drive -output .

# Format generated code
gofmt -w internal/google/generated/
goimports -w internal/google/generated/
```

### CI Workflow

The drift detection runs:
- **Weekly** (Sundays at 2 AM UTC)
- **On demand** (via workflow_dispatch)

When drift is detected:
1. Changes are classified by risk
2. Code is generated and formatted
3. A PR is created with a detailed report
4. Breaking changes create urgent issues

## Directory Structure

```
gdrv/
├── apis.yaml                    # API allowlist and configuration
├── cmd/
│   └── discovery-checker/       # CI tool for drift detection
│       ├── main.go
│       └── templates/           # Go templates for code generation
├── internal/
│   └── discovery/               # Discovery API client (CI-only)
│       ├── client.go
│       ├── types.go
│       └── cache.go
├── internal/google/
│   └── generated/               # Generated code
│       ├── drive/
│       │   ├── types.go         # Request/response structs
│       │   └── descriptors.go   # Endpoint metadata
│       ├── gmail/
│       └── calendar/
└── third_party/
    └── google/
        └── discovery/           # Pinned discovery snapshots
            ├── drive/v3.json
            ├── gmail/v1.json
            └── calendar/v3.json
```

## Migration from Runtime Discovery

The previous runtime discovery approach (`gdrv api` command) has been removed in favor of this static approach. Benefits:

- **Reliability**: No runtime dependency on Discovery API availability
- **Performance**: No discovery fetches during CLI execution
- **Type Safety**: Full IDE autocomplete and compile-time checking
- **Reviewability**: Changes go through normal PR review process
- **Reproducibility**: Pinned snapshots ensure consistent behavior

## Future Enhancements

### Phase 2
- Generate complete method stubs with validation
- Auto-generate CLI commands from descriptors
- Expand allowlist to Admin, Chat, People APIs
- Automated breaking change migration guides

### Phase 3
- Cross-API consistency checks
- Usage analytics to detect unused fields
- Automated deprecation warnings
- Integration with Google's API deprecation notices

## Contributing

When adding a new API to the allowlist:

1. Add entry to `apis.yaml`
2. Run `./discovery-checker -config apis.yaml -generate`
3. Verify generated code compiles: `go build ./...`
4. Run tests: `go test ./internal/google/generated/...`
5. Commit the initial snapshot

## References

- [Google Discovery Service](https://developers.google.com/discovery)
- [Discovery Document Format](https://developers.google.com/discovery/v1/reference/apis)
- [Go Generate Best Practices](https://go.dev/blog/generate)
