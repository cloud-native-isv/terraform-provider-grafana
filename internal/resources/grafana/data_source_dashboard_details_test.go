package grafana_test

import (
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/terraform-provider-grafana/v4/internal/testutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceDashboardDetails_basic(t *testing.T) {
	testutils.CheckOSSTestsEnabled(t)

	var dashboard models.DashboardFullWithMeta
	checks := []resource.TestCheckFunc{
		dashboardCheckExists.exists("grafana_dashboard.test", &dashboard),
		resource.TestCheckResourceAttr(
			"data.grafana_dashboard_details.from_uid", "title", "Production Overview Details",
		),
		resource.TestCheckResourceAttrPair(
			"data.grafana_dashboard_details.from_uid", "dashboard_id", "grafana_dashboard.test", "dashboard_id",
		),
		resource.TestCheckResourceAttr(
			"data.grafana_dashboard_details.from_uid", "uid", "test-ds-dashboard-details-uid",
		),
		resource.TestCheckResourceAttrSet("data.grafana_dashboard_details.from_uid", "config_json"),
		resource.TestCheckResourceAttrSet("data.grafana_dashboard_details.from_uid", "meta_json"),
		resource.TestCheckResourceAttrSet("data.grafana_dashboard_details.from_uid", "details_json"),
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: testutils.ProtoV5ProviderFactories,
		CheckDestroy:             dashboardCheckExists.destroyed(&dashboard, nil),
		Steps: []resource.TestStep{
			{
				Config: testutils.TestAccExample(t, "data-sources/grafana_dashboard_details/data-source.tf"),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}
