package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	cwsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/grafana/api"
	"github.com/grafana/terraform-provider-grafana/v4/internal/common"
)

func resourceDatasourceQuery() *common.Resource {
	resource := &schema.Resource{
		Description: `
Executes Grafana datasource queries declaratively and stores a stable machine-readable summary in Terraform state.

This resource supports two execution modes:
- managed: uses Grafana managed query endpoint with query targets
- proxy: uses datasource proxy endpoint with path/method/params/body

Only result summary is stored in state, raw response payload is not persisted.
`,
		CreateContext: createDatasourceQuery,
		ReadContext:   readDatasourceQuery,
		UpdateContext: updateDatasourceQuery,
		DeleteContext: deleteDatasourceQuery,
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Execution mode: managed or proxy.",
				ValidateFunc: validation.StringInSlice([]string{"managed", "proxy"}, false),
			},
			"datasource_id": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Grafana datasource numeric ID.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"time_range": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Query time range. Defaults to last 24h aligned to UTC minute boundary.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"from": {Type: schema.TypeString, Optional: true},
					"to":   {Type: schema.TypeString, Optional: true},
				}},
			},
			"managed_query": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Managed mode query configuration.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"instant": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Whether managed query is executed as instant query.",
					},
					"queries": {
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Description: "Managed query targets.",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"ref_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Reference ID for the query target.",
							},
							"expr": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Query expression.",
							},
							"datasource_uid": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Optional datasource UID per target.",
							},
							"interval_ms": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Interval in milliseconds.",
							},
							"max_data_points": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Max datapoints for the target.",
							},
							"raw": {
								Type:        schema.TypeMap,
								Optional:    true,
								Description: "Raw query payload fields.",
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						}},
					},
				}},
			},
			"proxy_query": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Proxy mode query configuration.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"path": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Datasource proxy path.",
					},
					"method": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "HTTP method: GET or POST.",
						ValidateFunc: validation.StringInSlice([]string{"GET", "POST", "get", "post"}, false),
					},
					"params": {
						Type:        schema.TypeMap,
						Optional:    true,
						Description: "URL query params for proxy call.",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"body": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "Optional JSON body for POST proxy call.",
						ValidateFunc: validation.StringIsJSON,
					},
				}},
			},
			"retry": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Retry policy. Exponential backoff with maximum 3 attempts.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"max_attempts": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(1, 3),
						Description:  "Maximum retry attempts, from 1 to 3.",
					},
				}},
			},
			"result_summary": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Stable machine-readable query result summary.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"mode":         {Type: schema.TypeString, Computed: true},
					"result_count": {Type: schema.TypeInt, Computed: true},
					"from":         {Type: schema.TypeString, Computed: true},
					"to":           {Type: schema.TypeString, Computed: true},
					"fingerprint":  {Type: schema.TypeString, Computed: true},
					"warnings": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				}},
			},
			"last_error": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Latest query error summary after retry exhaustion.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"category":     {Type: schema.TypeString, Computed: true},
					"message":      {Type: schema.TypeString, Computed: true},
					"retryable":    {Type: schema.TypeBool, Computed: true},
					"details_json": {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}

	return common.NewLegacySDKResource(
		common.CategoryGrafanaOSS,
		"grafana_datasource_query",
		common.NewResourceID(common.StringIDField("id")),
		resource,
	)
}

func validateDatasourceQueryCombination(d *schema.ResourceData) diag.Diagnostics {
	mode := d.Get("mode").(string)
	managed := readSingleBlock(d, "managed_query")
	proxy := readSingleBlock(d, "proxy_query")

	if mode == "managed" {
		if len(managed) == 0 {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid mode configuration", Detail: "managed mode requires managed_query block"}}
		}
		if len(proxy) > 0 {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid mode configuration", Detail: "managed mode does not allow proxy_query block"}}
		}
		queries := readNestedList(managed, "queries")
		if len(queries) == 0 {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid managed_query", Detail: "managed_query.queries must contain at least one query"}}
		}
		for idx, query := range queries {
			refID := strings.TrimSpace(fmt.Sprintf("%v", query["ref_id"]))
			expr := strings.TrimSpace(fmt.Sprintf("%v", query["expr"]))
			rawMap := toStringAnyMap(query["raw"])
			if refID == "" {
				return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid managed_query", Detail: fmt.Sprintf("managed_query.queries[%d].ref_id is required", idx)}}
			}
			if expr == "" && len(rawMap) == 0 {
				return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid managed_query", Detail: fmt.Sprintf("managed_query.queries[%d] requires expr or raw", idx)}}
			}
		}
		return nil
	}

	if mode == "proxy" {
		if len(proxy) == 0 {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid mode configuration", Detail: "proxy mode requires proxy_query block"}}
		}
		if len(managed) > 0 {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid mode configuration", Detail: "proxy mode does not allow managed_query block"}}
		}
		path := strings.TrimSpace(fmt.Sprintf("%v", proxy["path"]))
		if path == "" {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid proxy_query", Detail: "proxy_query.path is required"}}
		}
		method := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", proxy["method"])))
		if method != "GET" && method != "POST" {
			return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid proxy_query", Detail: "proxy_query.method must be GET or POST"}}
		}
		body := strings.TrimSpace(fmt.Sprintf("%v", proxy["body"]))
		if body != "" && body != "%!v(<nil>)" {
			var bodyMap map[string]any
			if err := json.Unmarshal([]byte(body), &bodyMap); err != nil {
				return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid proxy_query", Detail: "proxy_query.body must be valid JSON"}}
			}
		}
		return nil
	}

	return diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "invalid mode configuration", Detail: "mode must be managed or proxy"}}
}

func executeDatasourceQuery(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if diags := validateDatasourceQueryCombination(d); len(diags) > 0 {
		return diags
	}

	tr := normalizeTimeRange(readSingleBlock(d, "time_range"), time.Now())
	if err := writeDeterministicID(d, tr); err != nil {
		return diag.FromErr(err)
	}

	queryClient, err := buildDatasourceQueryClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	attempts := maxRetryAttempts(d)
	var result *queryExecutionResult
	execResult := &queryExecutionResult{}
	mode := d.Get("mode").(string)
	if mode == "managed" {
		managedReq, err := buildManagedRequest(d, tr)
		if err != nil {
			return diag.FromErr(err)
		}
		resp, queryErr := executeWithRetry(ctx, attempts, func() (*cwsapi.QueryResult, error) {
			return queryClient.ExecuteManagedQuery(managedReq)
		})
		if queryErr != nil {
			execResult.Error = queryErr
		} else {
			summary, summaryErr := buildResultSummary(mode, tr, resp)
			if summaryErr != nil {
				return diag.FromErr(summaryErr)
			}
			execResult.Summary = summary
		}
		result = execResult
	} else {
		proxyReq := buildProxyRequest(d)
		resp, queryErr := executeWithRetry(ctx, attempts, func() (*cwsapi.QueryResult, error) {
			return queryClient.ExecuteProxyQuery(proxyReq)
		})
		if queryErr != nil {
			execResult.Error = queryErr
		} else {
			summary, summaryErr := buildResultSummary(mode, tr, resp)
			if summaryErr != nil {
				return diag.FromErr(summaryErr)
			}
			execResult.Summary = summary
		}
		result = execResult
	}

	if result.Error != nil {
		setLastErrorState(d, result.Error)
		contextSummary := requestContextSummary(d, tr)
		contextSummary["attempts"] = attempts
		return diagnosticsFromQueryError("datasource query execution failed", contextSummary, result.Error)
	}

	setLastErrorState(d, nil)
	if currentFingerprint(d) != result.Summary.Fingerprint {
		setResultSummaryState(d, result.Summary)
	}
	return nil
}

func createDatasourceQuery(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return executeDatasourceQuery(ctx, d, meta)
}

func readDatasourceQuery(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return executeDatasourceQuery(ctx, d, meta)
}

func updateDatasourceQuery(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return executeDatasourceQuery(ctx, d, meta)
}

func deleteDatasourceQuery(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId("")
	return nil
}
