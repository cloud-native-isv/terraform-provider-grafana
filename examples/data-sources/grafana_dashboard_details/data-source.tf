resource "grafana_dashboard" "test" {
  config_json = jsonencode({
    id            = 12345,
    uid           = "test-ds-dashboard-details-uid"
    title         = "Production Overview Details",
    tags          = ["templated"],
    timezone      = "browser",
    schemaVersion = 16,
    version       = 0,
    refresh       = "25s"
  })
}

data "grafana_dashboard_details" "from_id" {
  depends_on = [
    grafana_dashboard.test
  ]

  dashboard_id = grafana_dashboard.test.dashboard_id
}
