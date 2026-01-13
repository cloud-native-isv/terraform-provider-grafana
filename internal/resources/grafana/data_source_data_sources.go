package grafana

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/grafana/terraform-provider-grafana/v4/internal/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceDataSources() *common.DataSource {
	schema := &schema.Resource{
		Description: "Get details about all Grafana Datasources.",
		ReadContext: datasourceDataSourcesRead,
		Schema: map[string]*schema.Schema{
			"org_id": orgIDAttribute(),
			"data_sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The numeric ID.",
						},
						"org_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"uid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type_logo_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"basic_auth_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_default": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"json_data_encoded": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"read_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
	return common.NewLegacySDKDataSource(common.CategoryGrafanaOSS, "grafana_data_sources", schema)
}

func datasourceDataSourcesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, orgID := OAPIClientFromNewOrgResource(meta, d)

	resp, err := client.Datasources.GetDataSources()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(orgID, 10))

	dataSources := make([]map[string]any, len(resp.GetPayload()))
	for i, ds := range resp.GetPayload() {
		jsonDataEncoded := ""
		if ds.JSONData != nil {
			if b, err := json.Marshal(ds.JSONData); err == nil {
				jsonDataEncoded = string(b)
			}
		}

		dataSources[i] = map[string]any{
			"database_id":        ds.ID,
			"org_id":             ds.OrgID,
			"uid":                ds.UID,
			"name":               ds.Name,
			"type":               ds.Type,
			"type_logo_url":      ds.TypeLogoURL,
			"access_mode":        ds.Access,
			"url":                ds.URL,
			"username":           ds.User,
			"database_name":      ds.Database,
			"basic_auth_enabled": ds.BasicAuth,
			"is_default":         ds.IsDefault,
			"json_data_encoded":  jsonDataEncoded,
			"read_only":          ds.ReadOnly,
		}
	}

	if err := d.Set("data_sources", dataSources); err != nil {
		return diag.Errorf("error setting data_sources attribute: %s", err)
	}

	return nil
}
