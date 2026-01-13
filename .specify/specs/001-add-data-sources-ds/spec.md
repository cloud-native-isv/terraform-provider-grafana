# Feature Specification: Grafana Data Sources Data Source

**Feature Branch**: `001-add-data-sources-ds`
**Created**: 2026-01-13
**Status**: Draft
**Feature**: [Grafana Data Sources Management](../../memory/features/006.md)
**Input**: User description: "新建一个名为grafana_data_sources的terraform数据源对象(放到internal/resources/grafana/data_source_data_sources.go文件中)，这个数据源通过调用client.Datasources.GetDataSources()方法获取所有的grafana datasource信息，GetDataSources接口的返回值Payload格式为..."

## User Scenarios & Testing

### User Story 1 - Retrieve All Data Sources (Priority: P1)

As a Terraform user, I want to retrieve a list of all Grafana data sources so that I can reference them or audit my configuration.

**Why this priority**: Core functionality requested.

**Independent Test**:
1. Create a Terraform configuration.
2. Define `data "grafana_data_sources" "all" {}`.
3. Output the results.
4. Apply and verify the list matches actual datasources.

**Acceptance Scenarios**:

1. **Given** a Grafana instance with datasources, **When** I query `grafana_data_sources`, **Then** I see all datasources with correct attributes.

### Edge Cases

- **No Data Sources**: When no data sources exist, the list should be empty (not error).
- **API Error**: If the API call fails (e.g. 401 Unauthorized), Terraform should report the error clearly.
- **Missing Fields**: If optional fields in the API response are missing, the schema should handle them (e.g. empty strings or defaults).

## Requirements

### Functional Requirements

- **FR-001**: Implement `grafana_data_sources` data source.
- **FR-002**: Place code in `internal/resources/grafana/data_source_data_sources.go`.
- **FR-003**: Use `client.Datasources.GetDataSources()` to fetch data.
- **FR-004**: Map API response fields to schema: `database_id` (mapped from `id`), `orgId`, `uid`, `name`, `type`, `typeLogoUrl`, `access`, `url`, `password`, `user`, `database`, `basicAuth`, `isDefault`, `jsonData`, `readOnly`.
- **FR-005**: Handle `jsonData` as encoded JSON string (schema attribute: `json_data_encoded`) to reliably support arbitrary content.
- **FR-006**: Ensure schema uses snake_case keys (e.g., `org_id`, `type_logo_url`).

### Key Entities

- **DataSource**: Represents a Grafana Data Source.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Terraform `plan` and `apply` succeed for the data source.
- **SC-002**: Read data matches API response.

## Clarifications

### Session 2026-01-13

- Q: How should `jsonData` be exposed given it can contain arbitrary types? → A: Option A - Encoded String (`json_data_encoded`).
- Q: How should the API's numerical `id` be exposed to avoid conflict with Terraform's reserved `id`? → A: Option A - Map to `database_id`.

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
