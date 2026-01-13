# Quickstart

Add the `grafana_data_sources` data source to your Terraform configuration to retrieve all data sources.

## Configuration

```hcl
data "grafana_data_sources" "all" {}

output "all_datasources" {
  value = data.grafana_data_sources.all.data_sources
}
```

## Example Output

```json
[
  {
    "id": 1,
    "uid": "P123456",
    "name": "Prometheus",
    "type": "prometheus",
    "is_default": true,
    ...
  }
]
```
