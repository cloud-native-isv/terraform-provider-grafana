resource "grafana_datasource_query" "managed_cpu" {
  mode          = "managed"
  datasource_id = 1

  managed_query {
    queries = [
      {
        ref_id = "A"
        expr   = "sum(rate(node_cpu_seconds_total[5m]))"
      }
    ]
  }

  time_range {
    from = "now-24h"
    to   = "now"
  }
}
