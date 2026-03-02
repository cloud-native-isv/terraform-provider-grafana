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
			"annotations_permissions": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Annotations permissions for the dashboard.",
			},
			"api_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "API version of the dashboard.",
			},
			"can_admin": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user can admin the dashboard.",
			},
			"can_delete": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user can delete the dashboard.",
			},
			"can_edit": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user can edit the dashboard.",
			},
			"can_save": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user can save the dashboard.",
			},
			"can_star": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user can star the dashboard.",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp of the dashboard.",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User who created the dashboard.",
			},
			"expires": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiration timestamp of the dashboard.",
			},
			"folder_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the folder where the dashboard is located (deprecated, use folder_uid instead).",
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
			"folder_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the folder where the dashboard is located.",
			},
			"has_acl": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the dashboard has ACL (Access Control List).",
			},
			"is_folder": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the dashboard is a folder.",
			},
			"is_snapshot": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the dashboard is a snapshot.",
			},
			"is_starred": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the Grafana dashboard is starred.",
			},
			"provisioned": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the dashboard is provisioned.",
			},
			"provisioned_external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External ID for provisioned dashboard.",
			},
			"public_dashboard_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether public dashboard is enabled.",
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL slug of the dashboard (deprecated).",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the dashboard.",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp of the dashboard.",
			},
			"updated_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User who last updated the dashboard.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full URL of the dashboard.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numerical version of the Grafana dashboard.",
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

	detailsJSONBytes, err := json.Marshal(dashboard)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to marshal dashboard details for uid %q: %w", uid, err))
	}

	d.SetId(MakeOrgResourceID(orgID, uid))
	d.Set("uid", uid)

	// Set DashboardMeta fields
	if dashboard.Meta.AnnotationsPermissions != nil {
		annotationsPermissionsJSON, _ := json.Marshal(dashboard.Meta.AnnotationsPermissions)
		d.Set("annotations_permissions", string(annotationsPermissionsJSON))
	}
	d.Set("api_version", dashboard.Meta.APIVersion)
	d.Set("can_admin", dashboard.Meta.CanAdmin)
	d.Set("can_delete", dashboard.Meta.CanDelete)
	d.Set("can_edit", dashboard.Meta.CanEdit)
	d.Set("can_save", dashboard.Meta.CanSave)
	d.Set("can_star", dashboard.Meta.CanStar)

	// Handle time fields - only set if not zero
	if !dashboard.Meta.Created.IsZero() {
		d.Set("created", dashboard.Meta.Created.String())
	}
	d.Set("created_by", dashboard.Meta.CreatedBy)
	if !dashboard.Meta.Expires.IsZero() {
		d.Set("expires", dashboard.Meta.Expires.String())
	}
	d.Set("folder_id", dashboard.Meta.FolderID)
	d.Set("folder_title", dashboard.Meta.FolderTitle)
	d.Set("folder_uid", dashboard.Meta.FolderUID)
	d.Set("folder_url", dashboard.Meta.FolderURL)
	d.Set("has_acl", dashboard.Meta.HasACL)
	d.Set("is_folder", dashboard.Meta.IsFolder)
	d.Set("is_snapshot", dashboard.Meta.IsSnapshot)
	d.Set("is_starred", dashboard.Meta.IsStarred)
	d.Set("provisioned", dashboard.Meta.Provisioned)
	d.Set("provisioned_external_id", dashboard.Meta.ProvisionedExternalID)
	d.Set("public_dashboard_enabled", dashboard.Meta.PublicDashboardEnabled)
	d.Set("slug", dashboard.Meta.Slug)
	d.Set("type", dashboard.Meta.Type)
	if !dashboard.Meta.Updated.IsZero() {
		d.Set("updated", dashboard.Meta.Updated.String())
	}
	d.Set("updated_by", dashboard.Meta.UpdatedBy)
	d.Set("url", metaClient.GrafanaSubpath(dashboard.Meta.URL))
	d.Set("version", int(dashboard.Meta.Version))

	if err := d.Set("details", string(detailsJSONBytes)); err != nil {
		return diag.Errorf("error setting dashboard details attributes: %s", err)
	}

	return nil
}
