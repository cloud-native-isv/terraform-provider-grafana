package grafana_test

import (
	"testing"

	"github.com/grafana/terraform-provider-grafana/v4/internal/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDataSources_basic(t *testing.T) {
	testutils.CheckOSSTestsEnabled(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "grafana_data_source" "test" {
  name = "test-ds-for-list"
  type = "prometheus"
  url  = "http://prometheus.example.com"
}

data "grafana_data_sources" "all" {
  depends_on = [grafana_data_source.test]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.grafana_data_sources.all", "data_sources.#"),
				),
			},
		},
	})
}
