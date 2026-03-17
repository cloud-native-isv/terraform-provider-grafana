package grafana

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	cwsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/grafana/api"
	"github.com/grafana/terraform-provider-grafana/v4/internal/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	defaultTimeFrom = "now-24h"
	defaultTimeTo   = "now"
)

type queryTimeRange struct {
	From string
	To   string
}

type queryResultSummary struct {
	Mode        string
	ResultCount int
	From        string
	To          string
	Fingerprint string
	Warnings    []string
}

type queryExecutionResult struct {
	Summary queryResultSummary
	Error   *cwsapi.QueryError
}

func normalizeTimeRange(input map[string]any, now time.Time) queryTimeRange {
	result := queryTimeRange{From: defaultTimeFrom, To: defaultTimeTo}
	if len(input) > 0 {
		if from, ok := input["from"].(string); ok && strings.TrimSpace(from) != "" {
			result.From = strings.TrimSpace(from)
		}
		if to, ok := input["to"].(string); ok && strings.TrimSpace(to) != "" {
			result.To = strings.TrimSpace(to)
		}
	}
	if result.From == defaultTimeFrom && result.To == defaultTimeTo {
		alignedNow := now.UTC().Truncate(time.Minute)
		result.From = alignedNow.Add(-24 * time.Hour).Format(time.RFC3339)
		result.To = alignedNow.Format(time.RFC3339)
	}
	return result
}

func readSingleBlock(d *schema.ResourceData, key string) map[string]any {
	raw, ok := d.GetOk(key)
	if !ok {
		return map[string]any{}
	}
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return map[string]any{}
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return item
}

func readNestedList(block map[string]any, key string) []map[string]any {
	raw, ok := block[key]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		asMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, asMap)
	}
	return result
}

func normalizeMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = normalizeAny(value)
	}
	return result
}

func normalizeAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeMap(typed)
	case []any:
		result := make([]any, len(typed))
		for idx := range typed {
			result[idx] = normalizeAny(typed[idx])
		}
		return result
	case []string:
		result := make([]string, len(typed))
		copy(result, typed)
		sort.Strings(result)
		asAny := make([]any, len(result))
		for idx := range result {
			asAny[idx] = result[idx]
		}
		return asAny
	default:
		return typed
	}
}

func hashDeterministic(input any) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func normalizeResourceConfig(d *schema.ResourceData, tr queryTimeRange) map[string]any {
	result := map[string]any{
		"mode":          d.Get("mode").(string),
		"datasource_id": d.Get("datasource_id").(int),
		"time_range": map[string]any{
			"from": tr.From,
			"to":   tr.To,
		},
	}
	if managed := readSingleBlock(d, "managed_query"); len(managed) > 0 {
		result["managed_query"] = normalizeMap(managed)
	}
	if proxy := readSingleBlock(d, "proxy_query"); len(proxy) > 0 {
		result["proxy_query"] = normalizeMap(proxy)
	}
	if retry := readSingleBlock(d, "retry"); len(retry) > 0 {
		result["retry"] = normalizeMap(retry)
	}
	return result
}

func buildResultSummary(mode string, tr queryTimeRange, queryResult *cwsapi.QueryResult) (queryResultSummary, error) {
	resultCount := len(queryResult.Results)
	warnings := make([]string, len(queryResult.Warnings))
	copy(warnings, queryResult.Warnings)
	sort.Strings(warnings)

	fingerprintInput := map[string]any{
		"mode":         mode,
		"from":         tr.From,
		"to":           tr.To,
		"result_count": resultCount,
		"warnings":     warnings,
		"results":      normalizeAny(queryResult.Results),
	}
	fingerprint, err := hashDeterministic(fingerprintInput)
	if err != nil {
		return queryResultSummary{}, err
	}

	return queryResultSummary{
		Mode:        mode,
		From:        tr.From,
		To:          tr.To,
		ResultCount: resultCount,
		Warnings:    warnings,
		Fingerprint: fingerprint,
	}, nil
}

func buildManagedRequest(d *schema.ResourceData, tr queryTimeRange) (*cwsapi.ManagedQueryRequest, error) {
	managed := readSingleBlock(d, "managed_query")
	queries := readNestedList(managed, "queries")
	resultQueries := make([]cwsapi.QueryTarget, 0, len(queries))
	for _, query := range queries {
		target := cwsapi.QueryTarget{
			RefID:         fmt.Sprintf("%v", query["ref_id"]),
			Expr:          fmt.Sprintf("%v", query["expr"]),
			DatasourceUID: fmt.Sprintf("%v", query["datasource_uid"]),
			Raw:           normalizeMap(toStringAnyMap(query["raw"])),
		}
		if target.RefID == "" {
			continue
		}
		if interval, ok := query["interval_ms"].(int); ok {
			target.IntervalMs = int64(interval)
		}
		if maxDataPoints, ok := query["max_data_points"].(int); ok {
			target.MaxDataPoints = int64(maxDataPoints)
		}
		resultQueries = append(resultQueries, target)
	}

	instant := false
	if value, ok := managed["instant"].(bool); ok {
		instant = value
	}

	return &cwsapi.ManagedQueryRequest{
		Mode:         cwsapi.ModeManaged,
		DatasourceID: int64(d.Get("datasource_id").(int)),
		From:         tr.From,
		To:           tr.To,
		Instant:      instant,
		Queries:      resultQueries,
	}, nil
}

func buildProxyRequest(d *schema.ResourceData) *cwsapi.ProxyQueryRequest {
	proxy := readSingleBlock(d, "proxy_query")
	request := &cwsapi.ProxyQueryRequest{
		Mode:         cwsapi.ModeProxy,
		DatasourceID: int64(d.Get("datasource_id").(int)),
		ProxyPath:    fmt.Sprintf("%v", proxy["path"]),
		Method:       strings.ToUpper(fmt.Sprintf("%v", proxy["method"])),
	}
	if params, ok := proxy["params"].(map[string]any); ok && len(params) > 0 {
		request.Params = make(map[string]string, len(params))
		for key, value := range params {
			request.Params[key] = fmt.Sprintf("%v", value)
		}
	}
	if body := strings.TrimSpace(fmt.Sprintf("%v", proxy["body"])); body != "" && body != "%!v(<nil>)" {
		var bodyObject map[string]any
		if err := json.Unmarshal([]byte(body), &bodyObject); err == nil {
			request.Body = bodyObject
		}
	}
	return request
}

func toStringAnyMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	result, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return result
}

func classifyQueryExecutionError(err error) *cwsapi.QueryError {
	if err == nil {
		return nil
	}
	queryErr := cwsapi.CategorizeQueryError(err)
	if queryErr != nil {
		return queryErr
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return &cwsapi.QueryError{
			Category:  cwsapi.CategoryNetwork,
			Message:   urlErr.Error(),
			Retryable: true,
			Details: map[string]any{
				"error": urlErr.Err.Error(),
			},
		}
	}

	return &cwsapi.QueryError{
		Category:  cwsapi.CategoryUnknown,
		Message:   err.Error(),
		Retryable: false,
	}
}

func diagnosticsFromQueryError(prefix string, reqContext map[string]any, qerr *cwsapi.QueryError) diag.Diagnostics {
	detailParts := []string{qerr.Message}
	if len(reqContext) > 0 {
		if payload, err := json.Marshal(reqContext); err == nil {
			detailParts = append(detailParts, fmt.Sprintf("request_context=%s", string(payload)))
		}
	}
	if len(qerr.Details) > 0 {
		if payload, err := json.Marshal(qerr.Details); err == nil {
			detailParts = append(detailParts, fmt.Sprintf("details=%s", string(payload)))
		}
	}
	summary := fmt.Sprintf("%s (%s)", prefix, qerr.Category)
	return diag.Diagnostics{diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   strings.Join(detailParts, " | "),
	}}
}

func maxRetryAttempts(d *schema.ResourceData) int {
	retry := readSingleBlock(d, "retry")
	attempts := 3
	if value, ok := retry["max_attempts"].(int); ok {
		attempts = value
	}
	if attempts < 1 {
		attempts = 1
	}
	if attempts > 3 {
		attempts = 3
	}
	return attempts
}

func executeWithRetry(ctx context.Context, attempts int, call func() (*cwsapi.QueryResult, error)) (*cwsapi.QueryResult, *cwsapi.QueryError) {
	var lastErr *cwsapi.QueryError
	for idx := 1; idx <= attempts; idx++ {
		result, err := call()
		if err == nil {
			return result, nil
		}
		lastErr = classifyQueryExecutionError(err)
		if !lastErr.Retryable || idx == attempts {
			break
		}
		waitMs := 100 * (1 << (idx - 1))
		timer := time.NewTimer(time.Duration(waitMs) * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, classifyQueryExecutionError(ctx.Err())
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func setResultSummaryState(d *schema.ResourceData, summary queryResultSummary) {
	_ = d.Set("result_summary", []any{map[string]any{
		"mode":         summary.Mode,
		"from":         summary.From,
		"to":           summary.To,
		"result_count": summary.ResultCount,
		"fingerprint":  summary.Fingerprint,
		"warnings":     summary.Warnings,
	}})
}

func setLastErrorState(d *schema.ResourceData, qerr *cwsapi.QueryError) {
	if qerr == nil {
		_ = d.Set("last_error", nil)
		return
	}
	detailsJSON := ""
	if len(qerr.Details) > 0 {
		if payload, err := json.Marshal(qerr.Details); err == nil {
			detailsJSON = string(payload)
		}
	}
	_ = d.Set("last_error", []any{map[string]any{
		"category":     string(qerr.Category),
		"message":      qerr.Message,
		"retryable":    qerr.Retryable,
		"details_json": detailsJSON,
	}})
}

func currentFingerprint(d *schema.ResourceData) string {
	raw, ok := d.GetOk("result_summary")
	if !ok {
		return ""
	}
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	entry, ok := items[0].(map[string]any)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", entry["fingerprint"])
}

func requestContextSummary(d *schema.ResourceData, tr queryTimeRange) map[string]any {
	ctx := map[string]any{
		"mode":          d.Get("mode").(string),
		"datasource_id": d.Get("datasource_id").(int),
		"from":          tr.From,
		"to":            tr.To,
	}
	if ctx["mode"] == "managed" {
		managed := readSingleBlock(d, "managed_query")
		ctx["query_count"] = len(readNestedList(managed, "queries"))
	}
	if ctx["mode"] == "proxy" {
		proxy := readSingleBlock(d, "proxy_query")
		ctx["path"] = fmt.Sprintf("%v", proxy["path"])
		ctx["method"] = strings.ToUpper(fmt.Sprintf("%v", proxy["method"]))
	}
	return ctx
}

func buildDatasourceQueryClient(meta any) (*cwsapi.GrafanaClient, error) {
	client := meta.(*common.Client)
	cfg := &cwsapi.Config{URL: client.GrafanaAPIURL}

	if client.GrafanaAPIConfig != nil {
		if client.GrafanaAPIConfig.APIKey != "" {
			cfg.Token = client.GrafanaAPIConfig.APIKey
		}
		if client.GrafanaAPIConfig.BasicAuth != nil {
			cfg.Username = client.GrafanaAPIConfig.BasicAuth.Username()
			password, hasPassword := client.GrafanaAPIConfig.BasicAuth.Password()
			if hasPassword {
				cfg.Password = password
			}
		}
		if client.GrafanaAPIConfig.TLSConfig != nil {
			cfg.NoCheckCertificate = client.GrafanaAPIConfig.TLSConfig.InsecureSkipVerify
		}
	}

	return cwsapi.NewClient(cfg)
}

func writeDeterministicID(d *schema.ResourceData, tr queryTimeRange) error {
	normalized := normalizeResourceConfig(d, tr)
	hash, err := hashDeterministic(normalized)
	if err != nil {
		return err
	}
	if len(hash) > 32 {
		hash = hash[:32]
	}
	d.SetId(hash)
	return nil
}

func mapRetryAttemptToString(attempts int) string {
	return strconv.Itoa(attempts)
}
