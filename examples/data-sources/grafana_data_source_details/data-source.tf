resource "grafana_data_source" "prometheus" {
  type                = "prometheus"
  name                = "prometheus-ds-details-test"
  uid                 = "prometheus-ds-details-test-uid"
  url                 = "https://my-instance.com"
  basic_auth_enabled  = true
  basic_auth_username = "username"

  json_data_encoded = jsonencode({
    httpMethod        = "POST"
    prometheusType    = "Mimir"
    prometheusVersion = "2.4.0"
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "password"
  })
}

data "grafana_data_source_details" "from_uid" {
  uid = grafana_data_source.prometheus.uid
}
