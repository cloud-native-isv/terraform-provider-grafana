---
description: "Task list for adding Grafana Data Sources Data Source"
---

# Tasks: Grafana Data Sources Data Source

**Input**: Design documents from `.specify/specs/001-add-data-sources-ds/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included as per standard provider practice (acceptance tests).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Verify project structure and dependencies (Grafana Go Client)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

*No new foundational types required as we reuse existing `internal/resources/grafana` package structures.*

---

## Phase 3: User Story 1 - Retrieve All Data Sources (Priority: P1) 🎯 MVP

**Goal**: As a Terraform user, I want to retrieve a list of all Grafana data sources so that I can reference them or audit my configuration.

**Independent Test**: Define `data "grafana_data_sources" "all" {}` and verify output lists all data sources.

### Tests for User Story 1

- [x] T002 [US1] Create acceptance test file `internal/resources/grafana/data_source_data_sources_test.go` with `TestAccDataSourceDataSources` test case (verifies simple read).

### Implementation for User Story 1

- [x] T003 [US1] Create data source file `internal/resources/grafana/data_source_data_sources.go`.
- [x] T004 [US1] Implement `datasourceDataSources` function in `internal/resources/grafana/data_source_data_sources.go` with schema definition matching `data-model.md` (attributes: `data_sources` list, nested objects).
- [x] T005 [US1] Implement `datasourceDataSourcesRead` function in `internal/resources/grafana/data_source_data_sources.go` calling `client.Datasources.GetDataSources()`.
- [x] T006 [US1] Implement mapping logic in `datasourceDataSourcesRead` to convert API response to schema (handling `json_data_encoded` and `database_id`).
- [x] T007 [US1] Register new data source in `internal/resources/grafana/provider.go` (or wherever resources are registered - verify existing pattern).

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T008: Documentation
  - [x] Create `docs/data-sources/data_sources.md`
  - [x] Ensure it follows the [standard format](../../../docs/data-sources/data_source.md)

- [x] T009: Verification
  - [x] Run full acceptance test suite for this resource
  - [x] Ensure CI passes (if applicable/local runs pass)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately.
- **User Stories (Phase 3+)**: Depends on Setup.
- **Polish (Final Phase)**: Depends on User Story 1 being complete.

### User Story Dependencies

- **User Story 1 (P1)**: Independent.

### Within Each User Story

- Test scaffold created first (T002).
- Implementation file created (T003).
- Schema defined (T004).
- Read logic implemented (T005).
- Mapping logic refined (T006).
- Registration (T007).

### Parallel Opportunities

- Documentation updates (T008) can run in parallel with implementation if specs are stable.

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 3: User Story 1.
3. **STOP and VALIDATE**: Run acceptance tests.
4. Deploy/demo.

### Incremental Delivery

1. Foundation ready.
2. Add User Story 1 → Test independently → Deploy.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Ensure `json_data_encoded` handles arbitrary JSON correctly.
- Verify existing `provider` registration mechanism in codebase.
