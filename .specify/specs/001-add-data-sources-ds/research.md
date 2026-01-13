# Research

## Technical Decisions

### 1. Data Source Schema Attributes
**Decision**: Map the `models.DataSource` struct fields to the Terraform schema attributes as defined in the spec, ensuring snake_case keys.
**Rationale**: Consistency with Terraform conventions and provider standards.
**Source**: `internal/resources/grafana/resource_data_source.go` and `internal/resources/grafana/data_source_data_source.go` show existing mappings.

### 2. ID Handling
**Decision**: Map API `id` (int64) to `database_id` (int) in Terraform.
**Rationale**: Avoid conflict with reserved `id` attribute.

### 3. JSON Data Handling
**Decision**: Use `json_data_encoded` string attribute.
**Rationale**: `jsonData` from API is a map of arbitrary types. Serializing to string avoids schema complexity and type issues.

### 4. Implementation Location
**Decision**: `internal/resources/grafana/data_source_data_sources.go`.
**Rationale**: Follows pattern `data_source_{resource_name}.go`.
