package grafana

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/terraform-provider-grafana/v4/internal/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceDashboardDetails() *common.DataSource {
	schema := &schema.Resource{
		Description: `
Datasource for retrieving complete details for a single Grafana dashboard from Dashboard HTTP API by dashboard ID.

* [Official documentation](https://grafana.com/docs/grafana/latest/dashboards/)
* [Dashboard HTTP API](https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/)
`,
		ReadContext: dataSourceReadDashboardDetails,
		Schema: map[string]*schema.Schema{
			"org_id": orgIDAttribute(),
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The uid of the Grafana dashboard to fetch.",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The title of the Grafana dashboard.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numerical version of the Grafana dashboard.",
			},
			"folder_title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The title of the folder where the Grafana dashboard is found.",
			},
			"folder_uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UID of the folder where the Grafana dashboard is found.",
			},
			"is_starred": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the Grafana dashboard is starred.",
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL slug of the dashboard (deprecated).",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full URL of the dashboard.",
			},
			"config": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The complete dashboard model JSON.",
			},
			"details": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The complete Dashboard HTTP API response JSON, including dashboard and meta.",
			},
		},
	}

	return common.NewLegacySDKDataSource(common.CategoryGrafanaOSS, "grafana_dashboard_details", schema)
}

func dataSourceReadDashboardDetails(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	metaClient := meta.(*common.Client)
	client, orgID := OAPIClientFromNewOrgResource(meta, d)

	uid := d.Get("uid").(string)
	if uid == "" {
		return diag.FromErr(fmt.Errorf("`uid` must be provided"))
	}

	dashboardResp, err := client.Dashboards.GetDashboardByUID(uid)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get dashboard details for uid %q: %w", uid, err))
	}

	dashboard := dashboardResp.GetPayload()
	model, ok := dashboard.Dashboard.(map[string]any)
	if !ok {
		return diag.FromErr(fmt.Errorf("unexpected dashboard model for uid %q", uid))
	}

	configJSONBytes, err := json.Marshal(model)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal dashboard model for uid %q: %w", uid, err))
	}

	detailsJSONBytes, err := json.Marshal(dashboard)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal dashboard details for uid %q: %w", uid, err))
	}

	version := 0
	if rawVersion, ok := model["version"].(float64); ok {
		version = int(rawVersion)
	}

	title := ""
	if rawTitle, ok := model["title"].(string); ok {
		title = rawTitle
	}

	d.SetId(MakeOrgResourceID(orgID, uid))
	d.Set("uid", uid)
	d.Set("title", title)
	d.Set("version", version)
	d.Set("folder_title", "")
	d.Set("folder_uid", dashboard.Meta.FolderUID)
	d.Set("is_starred", dashboard.Meta.IsStarred)
	d.Set("slug", dashboard.Meta.Slug)
	d.Set("url", metaClient.GrafanaSubpath(dashboard.Meta.URL))
	d.Set("config", string(configJSONBytes))
	if err := d.Set("details", string(detailsJSONBytes)); err != nil {
		return diag.Errorf("error setting dashboard details attributes: %s", err)
	}

	return nil
}
