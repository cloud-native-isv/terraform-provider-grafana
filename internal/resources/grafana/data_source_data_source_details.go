package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/grafana/terraform-provider-grafana/v4/internal/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceDataSourceDetails() *common.DataSource {
	schema := &schema.Resource{
		Description: "Get detailed information about a Grafana data source by UID using the Data Source HTTP API.",
		ReadContext: datasourceDataSourceDetailsRead,
		Schema: common.CloneResourceSchemaForDatasource(resourceDataSource().Schema, map[string]*schema.Schema{
			"org_id": orgIDAttribute(),
			"uid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UID of the Grafana data source to fetch.",
			},
			"database_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numeric ID of the data source.",
			},
			"access_control": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The access control permissions for the data source.",
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"secure_json_fields": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The secure JSON fields configured for this data source.",
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"type_logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The data source type logo URL.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The version of the data source.",
			},
			"with_credentials": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether credentials are included in requests.",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the data source is read only.",
			},
			"details": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The JSON-encoded string of the datasource JSONData field.",
			},
			"json_data_encoded":                      nil,
			"secure_json_data_encoded":               nil,
			"http_headers":                           nil,
			"private_data_source_connect_network_id": nil,
		}),
	}

	return common.NewLegacySDKDataSource(common.CategoryGrafanaOSS, "grafana_data_source_details", schema)
}

func datasourceDataSourceDetailsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, _ := OAPIClientFromNewOrgResource(meta, d)

	uid := d.Get("uid").(string)
	if uid == "" {
		return diag.FromErr(fmt.Errorf("uid must be set"))
	}

	resp, err := client.Datasources.GetDataSourceByUID(uid)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp == nil || resp.GetPayload() == nil {
		return diag.Errorf("unexpected state, API response is nil")
	}

	dataSource := resp.GetPayload()

	details := ""
	if dataSource.JSONData != nil {
		detailsBytes, marshalErr := json.Marshal(dataSource.JSONData)
		if marshalErr != nil {
			return diag.FromErr(fmt.Errorf("failed to marshal datasource JSONData: %w", marshalErr))
		}
		details = string(detailsBytes)
	}

	d.SetId(MakeOrgResourceID(dataSource.OrgID, dataSource.UID))
	d.Set("uid", dataSource.UID)
	d.Set("org_id", strconv.FormatInt(dataSource.OrgID, 10))
	d.Set("name", dataSource.Name)
	d.Set("type", dataSource.Type)
	d.Set("url", dataSource.URL)
	d.Set("username", dataSource.User)
	d.Set("access_mode", dataSource.Access)
	d.Set("database_name", dataSource.Database)
	d.Set("basic_auth_enabled", dataSource.BasicAuth)
	d.Set("basic_auth_username", dataSource.BasicAuthUser)
	d.Set("is_default", dataSource.IsDefault)

	d.Set("database_id", dataSource.ID)
	d.Set("access_control", dataSource.AccessControl)
	d.Set("secure_json_fields", dataSource.SecureJSONFields)
	d.Set("type_logo_url", dataSource.TypeLogoURL)
	d.Set("version", int(dataSource.Version))
	d.Set("with_credentials", dataSource.WithCredentials)
	d.Set("read_only", dataSource.ReadOnly)
	d.Set("details", details)

	return nil
}
