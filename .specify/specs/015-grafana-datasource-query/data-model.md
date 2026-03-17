# Data Model: Grafana Datasource Query Resource

## Entities

### DatasourceQueryResource
Terraform 资源配置与状态的聚合对象。

**Config Fields**
- `mode` (enum): `managed` | `proxy`（必填）
- `datasource_id` (int64): 目标数据源 ID（必填）
- `time_range` (object):
  - `from` (string): Grafana 相对时间或 epoch 毫秒
  - `to` (string): Grafana 相对时间或 epoch 毫秒
  - 默认：`from=now-24h`、`to=now`（UTC，分钟对齐）
- `managed_query` (object, mode=managed 必填):
  - `queries` ([]QueryTarget): 查询目标数组，至少 1 条
  - `instant` (bool, optional)
- `proxy_query` (object, mode=proxy 必填):
  - `path` (string)
  - `method` (string): `GET` | `POST`
  - `params` (map[string]string, optional)
  - `body` (dynamic/json, optional)
- `retry` (object, optional):
  - `max_attempts` (int): 默认 3，最大 3
  - `backoff` (enum): `exponential`（固定策略）

**State Fields**
- `id` (string): 规范化配置哈希（确定性）
- `result_summary` (object):
  - `result_count` (int)
  - `from` (string)
  - `to` (string)
  - `fingerprint` (string): 稳定比较指纹
  - `mode` (string)
  - `warnings` ([]string, optional)
- `last_error` (object, optional): 最近一次失败摘要（仅诊断，不阻断后续成功覆盖）

### QueryTarget
managed 模式下单条查询目标。

**Fields**
- `ref_id` (string, required)
- `expr` (string, optional)
- `datasource_uid` (string, optional)
- `interval_ms` (int64, optional)
- `max_data_points` (int64, optional)
- `raw` (map[string]any, optional)

**Validation Rules**
- `ref_id` 必填
- `expr` 与 `raw` 至少其一存在

### QueryExecutionRequest
发送给 `pkg/cws-lib-go` 的标准请求模型（按 mode 分流）。

**Managed**
- 对应 `api.ManagedQueryRequest`

**Proxy**
- 对应 `api.ProxyQueryRequest`

### QueryExecutionError
查询失败标准化对象。

**Fields**
- `category`: `validation|authorization|upstream|network|throttling|unknown`
- `message`: 可读错误
- `details`: 附加上下文（状态码、字段等）
- `retryable`: 是否可重试

## Relationships

- `DatasourceQueryResource` 1..1 `QueryExecutionRequest`
- `DatasourceQueryResource` 1..N `QueryTarget`（仅 managed）
- `DatasourceQueryResource` 1..1 `result_summary`
- `DatasourceQueryResource` 0..1 `QueryExecutionError`（失败路径）

## Validation Matrix

- `mode=managed`:
  - `managed_query.queries` 必填且非空
  - `proxy_query` 必须为空
- `mode=proxy`:
  - `proxy_query.path` 与 `proxy_query.method` 必填
  - `managed_query` 必须为空
- `datasource_id` 必填且 > 0
- `time_range` 若显式提供，`from <= to`（按统一时间解析比较）

## State Transitions

1. `Planned`：读取配置并归一化
2. `Validated`：完成模式校验 + 字段校验
3. `Executing`：调用 cws-lib-go API（含重试）
4. `Summarized`：生成 `result_summary` + `fingerprint`
5. `Persisted`：写入 state（若 fingerprint 未变化则保持稳定）
6. `Failed`：返回分类错误，state 不写入原始结果体

## Idempotency Strategy

- `id` 使用规范化配置计算哈希（忽略与语义无关的序列化差异）
- `result_summary.fingerprint` 基于规范化摘要计算
- 当新旧 `fingerprint` 相同，保持状态稳定，避免无意义变更