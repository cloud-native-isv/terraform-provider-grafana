package grafana_test

import (
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/terraform-provider-grafana/v4/internal/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceDataSourceDetails_basic(t *testing.T) {
	testutils.CheckOSSTestsEnabled(t)

	var dataSource models.DataSource
	checks := []resource.TestCheckFunc{
		datasourceCheckExists.exists("grafana_data_source.prometheus", &dataSource),
		resource.TestMatchResourceAttr("data.grafana_data_source_details.from_uid", "id", defaultOrgIDRegexp),
		resource.TestCheckResourceAttr("data.grafana_data_source_details.from_uid", "name", "prometheus-ds-details-test"),
		resource.TestCheckResourceAttr("data.grafana_data_source_details.from_uid", "uid", "prometheus-ds-details-test-uid"),
		resource.TestCheckResourceAttr("data.grafana_data_source_details.from_uid", "type", "prometheus"),
		resource.TestCheckResourceAttr("data.grafana_data_source_details.from_uid", "basic_auth_enabled", "true"),
		resource.TestCheckResourceAttrSet("data.grafana_data_source_details.from_uid", "version"),
		resource.TestCheckResourceAttrSet("data.grafana_data_source_details.from_uid", "with_credentials"),
		resource.TestCheckResourceAttrSet("data.grafana_data_source_details.from_uid", "details"),
		resource.TestCheckResourceAttrSet("data.grafana_data_source_details.from_uid", "database_id"),
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		CheckDestroy:             datasourceCheckExists.destroyed(&dataSource, nil),
		Steps: []resource.TestStep{
			{
				Config: testutils.TestAccExample(t, "data-sources/grafana_data_source_details/data-source.tf"),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}
