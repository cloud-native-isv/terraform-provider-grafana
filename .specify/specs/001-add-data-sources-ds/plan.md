# Implementation Plan: Grafana Data Sources Data Source

**Branch**: `001-add-data-sources-ds` | **Date**: 2026-01-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `.specify/specs/001-add-data-sources-ds/spec.md`

## Summary

Implement a new Terraform data source `grafana_data_sources` to retrieve detailed information about all available Grafana data sources. This involves adding `internal/resources/grafana/data_source_data_sources.go`. The data source will use the `GetDataSources` method from the Grafana Go client to fetch the list and expose it via a Terraform schema with mapped attributes.

## Technical Context

**Language/Version**: Go (1.23+ matching project)
**Primary Dependencies**: 
- `github.com/hashicorp/terraform-plugin-sdk/v2` (SDKv2)
- `github.com/grafana/grafana-openapi-client-go` (Client)
**Storage**: N/A (Read-only data source)
**Testing**: Acceptance tests using `resource.TestCase`
**Target Platform**: Terraform Provider (Cross-platform)
**Project Type**: Terraform Provider
**Performance Goals**: Standard API latency
**Constraints**: Read-only access to Grafana API
**Scale/Scope**: Single data source, potential for large response payloads (pagination handled by client if applicable, or single list).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Compliance**:

- **Library-First**: N/A (Provider implementation)
- **CLI Interface**: N/A (Terraform HCL)
- **Test-First**: Acceptance tests will be written.
- **Integration Testing**: Tests verify against real Grafana instance.
- **Observability**: Standard SDK diagnostics.
- **Simplicity**: Single file addition, straightforward mapping.

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this feature)

```text
.specify/specs/001-add-data-sources-ds/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/resources/grafana/
└── data_source_data_sources.go
```

**Structure Decision**: Place the new data source implementation in `internal/resources/grafana/` alongside existing `data_source*.go` files, following the project's convention.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | | |
