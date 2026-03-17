# Implementation Plan: Grafana Datasource Query Resource

**Branch**: `015-grafana-datasource-query` | **Date**: 2026-03-16 | **Spec**: `/cws_data/terraform-provider-grafana/.specify/specs/015-grafana-datasource-query/requirements.md`
**Input**: Specification from `/cws_data/terraform-provider-grafana/.specify/specs/015-grafana-datasource-query/requirements.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

新增 `grafana_datasource_query` Terraform 资源，在资源生命周期中调用 `pkg/cws-lib-go` 子模块提供的 Grafana datasource query API 能力，支持 managed/proxy 两种查询模式，并将查询结果以稳定机读摘要写入 state。实现重点是：输入校验、可分类错误、瞬时失败重试、等价结果下状态稳定（避免无意义 diff）、以及基于规范化配置生成确定性 `id`。

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.25.x（provider 主仓）；复用 `pkg/cws-lib-go` 现有 Go API 封装  
**Primary Dependencies**: Terraform Plugin SDK v2（legacy 资源）、Terraform Plugin Framework（provider 并存架构）、`pkg/cws-lib-go/lib/cloud/grafana/api` 查询客户端  
**Storage**: Terraform state（仅保存机读摘要，不保存原始大结果体）  
**Testing**: Go unit tests（摘要/哈希/校验/错误分类），Acceptance tests（资源创建+刷新+失败路径）  
**Target Platform**: Terraform Provider（Linux/macOS/Windows 运行 Terraform）
**Project Type**: 多包 Go provider（`internal/resources/*` + `pkg/provider/*`）  
**Performance Goals**: 单次查询额外本地处理开销可忽略；大结果场景保持 state 可控（摘要化）  
**Constraints**: 必须复用 cws-lib-go 查询能力；`docs/` 不可手工修改；默认结果非敏感；瞬时错误指数退避最多 3 次  
**Scale/Scope**: 1 个新资源 + provider 资源注册 + 资源文档生成输入 + 对应测试

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

- **Feature-Centric Development**: PASS。当前 feature 归属 `003`，无新增/合并/拆分 feature；需同步更新 `.specify/memory/features.md` 与 `.specify/memory/features/003.md`。
- **Generated Documentation**: PASS。仅更新 schema/示例/模板输入，后续通过 `go generate ./...` 生成文档；不手改 `docs/`。
- **Testing Mandate**: PASS。计划包含 acceptance + unit，覆盖成功、失败、重试、稳定性。
- **Terraform Standards**: PASS。新资源按现有 legacy SDK 资源风格接入 `internal/resources/grafana` 与 `pkg/provider/resources.go`。
- **Go Tooling**: PASS。实现阶段执行 `go fmt`、`go vet`、`golangci-lint` 与相关测试。
- **Release & Versioning**: PASS。本次为特性增量，变更纳入常规版本发布流程。

### Additional Constraints from `$ARGUMENTS`

- 本次调用未提供额外 `$ARGUMENTS`，无附加硬约束。

### Post-Design Gate (after Phase 1 artifacts)

- 数据模型覆盖资源配置、请求、摘要输出、错误分类与状态转换。
- 契约定义 managed/proxy 行为、重试语义与错误响应。
- 快速上手文档覆盖最小可运行路径与验证命令。

**Gates Status**: ✅ All gates pass

## Project Structure

### Documentation (this spec)

```text
.specify/specs/[REQUIREMENTS_KEY]/
├── plan.md              # This file (/speckit.plan command output)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this spec. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
internal/resources/grafana/
├── resource_datasource_query.go          # 新增资源定义（schema + CRUD/Read）
├── resource_datasource_query_test.go     # 新增单元测试
└── resources.go                          # 注册资源

pkg/provider/
└── resources.go                          # 聚合注册（无需新模块）

pkg/cws-lib-go/lib/cloud/grafana/api/
├── query.go                              # 已有：managed/proxy query API
└── errors.go                             # 已有：错误分类

examples/resources/grafana_datasource_query/
└── resource.tf                           # 新增示例输入（用于文档生成）

.specify/specs/015-grafana-datasource-query/
├── plan.md
├── data-model.md
├── quickstart.md
└── contracts/query-resource.yaml
```

**Structure Decision**: 采用现有 provider 分层，资源落在 `internal/resources/grafana` 并接入统一聚合；查询执行复用 `pkg/cws-lib-go`，避免复制 HTTP 逻辑。

## Phase 0: Research Review & Context

- `$ARGUMENTS` 解析：本次调用无附加参数。
- `research.md` 检查：`/cws_data/terraform-provider-grafana/.specify/specs/015-grafana-datasource-query/research.md` 不存在。
- 通过现有仓库文档与内存消除关键不确定性：
  - README 与 docs 约束：文档必须由 `go generate` 生成，禁止手改 `docs/`。
  - 资源接入路径：`internal/resources/grafana/resources.go` + `pkg/provider/resources.go`。
  - 查询能力来源：`pkg/cws-lib-go/lib/cloud/grafana/api/query.go` 与 `errors.go`。
  - Feature 基线：`003` 已定义为 Datasource Query Resource（当前 Planned）。

## Phase 1: Design & Contracts

1. 抽取并落盘数据模型到 `data-model.md`（资源配置、请求、摘要、错误对象、状态迁移）。
2. 生成契约到 `contracts/query-resource.yaml`（managed/proxy 请求与统一响应）。
3. 生成 `quickstart.md`（最小配置、执行、校验、失败诊断）。
4. Phase 1 后复核宪章门禁并确认无违规。

## Phase 2: Implementation Planning

1. 在 `internal/resources/grafana` 新增 `grafana_datasource_query` 资源：
  - 输入 schema（mode、datasource、query payload、time_range、执行选项）
  - 默认时间窗（`now-24h` 到 `now`，UTC 分钟对齐）
  - 规范化配置哈希生成确定性 `id`
  - 调用 cws-lib-go 执行查询，结果摘要入 state
2. 增加错误分类与重试编排：校验、认证/授权、上游、网络；`429/5xx/短超时` 指数退避最多 3 次。
3. 状态稳定策略：按规范化摘要比较，等价结果不触发无意义更新。
4. 补充测试：
  - unit：哈希稳定性、时间默认、摘要计算、错误分类
  - acceptance：成功路径、空结果、失败可诊断、重复执行无 drift
5. 补充文档输入（示例/描述），执行 `go generate ./...` 同步生成 docs。

## Complexity Tracking

N/A
