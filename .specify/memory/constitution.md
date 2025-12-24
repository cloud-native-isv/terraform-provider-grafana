<!--
Sync Impact Report:
- Version change: 1.0.0 -> 1.0.1
- List of modified principles: None (Templates aligned)
- Added sections: None
- Removed sections: None
- Templates requiring updates:
  - /.specify/templates/plan-template.md (✅ updated)
  - /.specify/templates/tasks-template.md (✅ updated)
- Follow-up TODOs: None
-->

# Terraform Provider for Grafana Constitution

## Core Principles

### Generated Documentation
Documentation is generated via `tfplugindocs`. Manual edits to `docs/` are forbidden. All documentation changes must originate from schema descriptions or examples.

### Testing Mandate
Acceptance tests (`make testacc`) are required for all resources and data sources. Unit tests are required for internal logic and helper functions. Tests must pass before merging.

### Terraform Standards
Adhere to HashiCorp SDK/Framework standards. Follow best practices for provider development as outlined in the "Extending Terraform" documentation.

### Go Tooling
Use standard Go tools. Code must be formatted with `go fmt`, vetted with `go vet`, and generated with `go generate`.

## Development Workflow

### Local Development
Use `dev_overrides` in `.terraformrc` for local testing. Build the binary with `go build`.

### Release Process
Releases are automated via GitHub Actions and GoReleaser. Semantic versioning is strictly followed.

## Governance

### Compliance
All PRs must verify compliance with these principles. Complexity must be justified.

### Amendments
Amendments require documentation and approval.

**Version**: 1.0.1 | **Ratified**: 2025-12-15 | **Last Amended**: 2025-12-24
