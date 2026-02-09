# Data Model: Update Grafana SDKs

## Entities

### SDKVersion
- **Purpose**: 记录目标 SDK 版本与来源。
- **Fields**:
  - `name` (string, required): `grafana-app-sdk` 或 `grafana-plugin-sdk-go`
  - `target_version` (string, required): 目标版本号（例如 `v0.50.1`）
  - `source_url` (string, required): 对应发布页链接
  - `current_version` (string, optional): 升级前版本（从 go.mod 读取）

### CompatibilityExpectation
- **Purpose**: 描述升级后保持兼容的范围。
- **Fields**:
  - `scope` (string, required): 关键使用场景范围（如构建、核心资源/数据源操作）
  - `must_hold` (bool, required): 必须保持兼容
  - `verification_notes` (string, optional): 验证方式说明

## Relationships
- `SDKVersion` -> `CompatibilityExpectation`：升级对应的兼容性期望。

## Validation Rules
- `target_version` 必须与需求指定版本完全一致。
- `source_url` 必须指向官方发布页。
- `must_hold` 必须为 `true`。

## Records

### SDKVersion

| name | current_version | target_version | source_url | dependency_location |
| --- | --- | --- | --- | --- |
| grafana-app-sdk | v0.48.1 | v0.50.1 | https://github.com/grafana/grafana-app-sdk/releases/tag/v0.50.1 | go.mod require block |
| grafana-plugin-sdk-go | v0.275.0 | v0.287.0 | https://github.com/grafana/grafana-plugin-sdk-go/releases/tag/v0.287.0 | go.mod require block (indirect) |

### CompatibilityExpectation

| scope | must_hold | verification_notes |
| --- | --- | --- |
| 标准构建与关键资源/数据源最小使用路径 | true | 依据 quickstart.md 执行构建与关键场景验证 |
