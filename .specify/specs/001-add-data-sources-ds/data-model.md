# Data Model

## Grafana Data Sources Data Source

### Schema

| Attribute | Type | Computed | Description |
|-----------|------|----------|-------------|
| `data_sources` | `list(object)` | `true` | List of Grafana data sources. |
| `id` | `string` | `true` | The ID of this data source (placeholder usually). |

### Nested Object: `data_source`

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | `int64` | The numeric ID from the database (mapped from API `id`). |
| `org_id` | `int64` | The organization ID. |
| `uid` | `string` | The unique identifier (UID). |
| `name` | `string` | The name of the data source. |
| `type` | `string` | The data source type (e.g., `prometheus`). |
| `type_logo_url` | `string` | URL to the type logo. |
| `access_mode` | `string` | Access mode (`proxy` or `direct`). |
| `url` | `string` | The URL of the data source. |
| `username` | `string` | The username (if applicable). |
| `database_name` | `string` | The database name. |
| `basic_auth_enabled` | `bool` | Whether basic auth is enabled. |
| `is_default` | `bool` | Whether this is the default data source. |
| `json_data_encoded` | `string` | JSON encoded string of configuration data. |
| `read_only` | `bool` | Whether the data source is read-only. |

*Note: Sensitive fields like `password` are typically empty or not returned fully by API, checking existing implementation behavior.*
