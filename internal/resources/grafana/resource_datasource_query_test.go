package grafana

import (
	"errors"
	"net/url"
	"testing"
	"time"

	cwsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/grafana/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func newDatasourceQueryTestData(t *testing.T, input map[string]any) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceDatasourceQuery().Schema.Schema, input)
}

func TestDatasourceQuery_DefaultTimeRangeAligned(t *testing.T) {
	now := time.Date(2026, 3, 17, 8, 45, 23, 0, time.UTC)
	tr := normalizeTimeRange(map[string]any{}, now)
	require.Equal(t, "2026-03-16T08:45:00Z", tr.From)
	require.Equal(t, "2026-03-17T08:45:00Z", tr.To)
}

func TestDatasourceQuery_HashDeterministic(t *testing.T) {
	left := map[string]any{
		"mode": "managed",
		"time_range": map[string]any{
			"to":   "2026-03-17T08:45:00Z",
			"from": "2026-03-16T08:45:00Z",
		},
		"managed_query": map[string]any{
			"queries": []any{map[string]any{"ref_id": "A", "expr": "up"}},
		},
	}
	right := map[string]any{
		"managed_query": map[string]any{
			"queries": []any{map[string]any{"expr": "up", "ref_id": "A"}},
		},
		"time_range": map[string]any{
			"from": "2026-03-16T08:45:00Z",
			"to":   "2026-03-17T08:45:00Z",
		},
		"mode": "managed",
	}

	leftHash, err := hashDeterministic(normalizeAny(left))
	require.NoError(t, err)
	rightHash, err := hashDeterministic(normalizeAny(right))
	require.NoError(t, err)
	require.Equal(t, leftHash, rightHash)
}

func TestDatasourceQuery_BuildManagedRequest(t *testing.T) {
	d := newDatasourceQueryTestData(t, map[string]any{
		"mode":          "managed",
		"datasource_id": 12,
		"managed_query": []any{map[string]any{
			"instant": true,
			"queries": []any{map[string]any{
				"ref_id":          "A",
				"expr":            "up",
				"datasource_uid":  "prom",
				"interval_ms":     1000,
				"max_data_points": 100,
				"raw": map[string]any{
					"legend": "cpu",
				},
			}},
		}},
	})

	request, err := buildManagedRequest(d, queryTimeRange{From: "2026-03-16T08:45:00Z", To: "2026-03-17T08:45:00Z"})
	require.NoError(t, err)
	require.Equal(t, cwsapi.ModeManaged, request.Mode)
	require.EqualValues(t, 12, request.DatasourceID)
	require.Len(t, request.Queries, 1)
	require.Equal(t, "A", request.Queries[0].RefID)
	require.Equal(t, "up", request.Queries[0].Expr)
}

func TestDatasourceQuery_BuildProxyRequest(t *testing.T) {
	d := newDatasourceQueryTestData(t, map[string]any{
		"mode":          "proxy",
		"datasource_id": 5,
		"proxy_query": []any{map[string]any{
			"path":   "/api/v1/query",
			"method": "post",
			"params": map[string]any{"query": "up"},
			"body":   `{"time":"now"}`,
		}},
	})

	request := buildProxyRequest(d)
	require.Equal(t, cwsapi.ModeProxy, request.Mode)
	require.EqualValues(t, 5, request.DatasourceID)
	require.Equal(t, "POST", request.Method)
	require.Equal(t, "/api/v1/query", request.ProxyPath)
	require.Equal(t, "up", request.Params["query"])
	require.Equal(t, "now", request.Body["time"])
}

func TestDatasourceQuery_ValidateModeCombination(t *testing.T) {
	d := newDatasourceQueryTestData(t, map[string]any{
		"mode":          "managed",
		"datasource_id": 1,
		"proxy_query": []any{map[string]any{
			"path":   "/api/v1/query",
			"method": "GET",
		}},
	})

	diags := validateDatasourceQueryCombination(d)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Detail, "managed mode requires managed_query block")
}

func TestDatasourceQuery_ErrorClassification(t *testing.T) {
	queryErr := classifyQueryExecutionError(&cwsapi.QueryError{Category: cwsapi.CategoryAuthorization, Message: "forbidden", Retryable: false})
	require.Equal(t, cwsapi.CategoryAuthorization, queryErr.Category)

	networkErr := classifyQueryExecutionError(&url.Error{Err: errors.New("connection reset")})
	require.Equal(t, cwsapi.CategoryNetwork, networkErr.Category)
	require.True(t, networkErr.Retryable)
}
