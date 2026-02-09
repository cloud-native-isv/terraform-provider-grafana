<!--
Sync Impact Report:
- Version change: 1.0.1 -> 1.1.0
- List of modified principles:
  - Generated Documentation (clarified source-of-truth and update path)
  - Testing Mandate (clarified acceptance vs unit expectations)
  - Terraform Standards (clarified framework/SDK guidance)
  - Go Tooling (expanded to linting/go generate expectations)
- Added sections:
  - Project Context & Tech Stack
  - Development Workflow & Quality Gates
- Removed sections: None
- Templates requiring updates:
  - /.specify/templates/plan-template.md (✅ updated)
  - /.specify/templates/tasks-template.md (✅ updated)
  - /.specify/templates/spec-template.md (✅ no change needed)
- Follow-up TODOs: None
-->

# Terraform Provider for Grafana Constitution

## Core Principles

### 以特性（Feature）为核心的开发理念
Feature 列表是项目的长期主干与单一事实来源。
任何 spec → plan → tasks → implement 阶段都必须复核 Feature 的新增/合并/拆分/删除。
所有 Feature 变更必须记录并可追溯到对应的 spec/plan 依据。

### Generated Documentation
Documentation is generated via `tfplugindocs` (go generate). Manual edits to
`docs/` are forbidden. Documentation changes MUST originate from schema
descriptions, examples, or templates.

### Testing Mandate
Acceptance tests (`make testacc*`) are required for all resources and data
sources. Unit tests are required for internal logic and helper functions. All
tests MUST pass before merging.

### Terraform Standards
Adhere to HashiCorp SDK/Framework standards. Follow provider development best
practices in the "Extending Terraform" documentation.

### Go Tooling
Use standard Go tools. Code MUST be formatted with `go fmt`, vetted with
`go vet`, generated with `go generate`, and linted via `golangci-lint`.

### Release & Versioning
Semantic versioning is mandatory. Releases are produced by GoReleaser via
GitHub Actions, and tags MUST correspond to released versions.

## Project Context & Tech Stack

**Summary**: This repository implements the Terraform provider for Grafana and
Grafana Cloud, exposing resources and data sources for infrastructure as code.

**Primary Stack**: Go 1.25.3, Terraform Plugin SDK v2 and Plugin Framework,
Grafana APIs, OpenAPI clients, tfplugindocs, GoReleaser.

**Project Structure (high level)**:
- cmd/, main.go: provider binaries and generators
- internal/: provider implementation
- pkg/: generation and shared utilities
- docs/, templates/, examples/: generated docs inputs
- scripts/, tools/: repo automation

## Development Workflow & Quality Gates

### Local Development
Use `dev_overrides` in `.terraformrc` for local testing. Build with `go build`
or `make build`.

### Documentation
Docs are generated with `go generate ./...` (see `tfplugindocs`). Never edit
`docs/` manually.

### Tests
Use `make testacc`, `make testacc-oss-docker`, or related targets for
acceptance tests. Unit tests are required for internal helpers.

### Linting
Run `golangci-lint` (dockerized target) and fix issues before merge.

### Release Process
Releases are automated via GitHub Actions and GoReleaser. Tag via `make release`
and follow the release checklist in README.

## Governance

### Compliance
All PRs MUST verify compliance with these principles. Any added complexity
must be justified in the PR description or spec/plan artifacts.

### Amendments
Amendments require a documented rationale, reviewer approval, and a semantic
version bump. MAJOR for incompatible removals/redefinitions, MINOR for new
principles/sections, PATCH for clarifications.

### Review Expectations
Constitution compliance is reviewed during planning and before merge.

**Version**: 1.1.0 | **Ratified**: 2025-12-15 | **Last Amended**: 2026-02-09
