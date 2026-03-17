package grafana_test

import (
	"regexp"
	"testing"

	"github.com/grafana/terraform-provider-grafana/v4/internal/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceQuery_modeValidation(t *testing.T) {
	testutils.CheckOSSTestsEnabled(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "grafana_datasource_query" "invalid" {
  mode          = "managed"
  datasource_id = 1

  proxy_query {
    path   = "/api/v1/query"
    method = "GET"
  }
}
`,
				ExpectError: regexp.MustCompile("managed mode requires managed_query block"),
			},
		},
	})
}

func TestAccDatasourceQuery_proxyValidation(t *testing.T) {
	testutils.CheckOSSTestsEnabled(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "grafana_datasource_query" "invalid" {
  mode          = "proxy"
  datasource_id = 1

  proxy_query {
    path   = ""
    method = "GET"
  }
}
`,
				ExpectError: regexp.MustCompile("proxy_query.path is required"),
			},
		},
	})
}
