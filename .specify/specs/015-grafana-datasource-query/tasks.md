# Tasks: Grafana Datasource Query Resource

**Input**: Design documents from `/cws_data/terraform-provider-grafana/.specify/specs/015-grafana-datasource-query/`
**Prerequisites**: plan.md, requirements.md, data-model.md, contracts/query-resource.yaml, quickstart.md

**Tests**: 需要同时包含 unit 与 acceptance，覆盖每个用户故事的独立可验证路径。

**Organization**: 任务按用户故事分组，确保每个故事可独立实现与验证。

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立资源骨架与文档输入结构

- [X] T001 创建资源实现文件骨架 `internal/resources/grafana/resource_datasource_query.go`
- [X] T002 创建资源辅助文件骨架 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T003 [P] 创建示例目录与示例文件 `examples/resources/grafana_datasource_query/resource.tf`
- [X] T004 [P] 在 `internal/resources/grafana/resources.go` 预留资源注册入口 `resourceDatasourceQuery()`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 完成所有用户故事共享的底座能力（未完成前禁止进入 US 实现）

- [X] T005 定义资源配置/state 结构体与 schema 基础字段于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T006 实现配置规范化与确定性哈希函数于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T007 [P] 实现时间范围默认值与 UTC 分钟对齐逻辑于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T008 [P] 实现结果摘要与 fingerprint 计算逻辑于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T009 实现 mode/字段组合校验框架于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T010 实现 cws-lib-go 查询请求转换层（managed/proxy）于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T011 实现统一错误归类与 Terraform diagnostics 映射于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T012 实现指数退避重试编排（最多 3 次）于 `internal/resources/grafana/resource_datasource_query_helpers.go`

**Checkpoint**: 基础能力就绪，可进入用户故事并行实现

---

## Phase 3: User Story 1 - 声明式执行数据源查询 (Priority: P1) 🎯 MVP

**Goal**: 声明资源即可执行查询并输出稳定摘要，重复执行无无意义变更。

**Independent Test**: 仅实现本阶段时，`terraform apply/refresh/plan` 可完成 managed 查询，且等价结果不产生 drift。

### Tests for User Story 1 (MANDATORY)

- [X] T013 [P] [US1] 新增摘要与哈希稳定性单元测试 `internal/resources/grafana/resource_datasource_query_test.go`
- [X] T014 [P] [US1] 新增 managed 模式请求转换单元测试 `internal/resources/grafana/resource_datasource_query_test.go`
- [ ] T015 [US1] 新增声明式查询 acceptance 用例（成功+重复执行稳定）`internal/resources/grafana/resource_datasource_query_acc_test.go`

### Implementation for User Story 1

- [X] T016 [US1] 实现 managed 模式 Create/Read 执行路径于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T017 [US1] 将查询结果写入 `result_summary` 而非原始响应于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T018 [US1] 实现等价摘要短路更新（避免无意义 state 变化）于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T019 [US1] 完成资源 ID 基于规范化配置哈希写入逻辑于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T020 [US1] 在聚合入口注册资源并确保 provider 可见 `internal/resources/grafana/resources.go`

**Checkpoint**: US1 可独立交付（MVP）

---

## Phase 4: User Story 2 - 查询失败可诊断 (Priority: P2)

**Goal**: 失败路径返回清晰、可分类、可操作的诊断信息。

**Independent Test**: 使用无效参数与权限受限场景触发失败，错误可区分 validation / authorization / upstream / network / throttling。

### Tests for User Story 2 (MANDATORY)

- [X] T021 [P] [US2] 新增字段级校验失败测试 `internal/resources/grafana/resource_datasource_query_test.go`
- [X] T022 [P] [US2] 新增错误分类映射测试（含 401/403/429/5xx/网络异常）`internal/resources/grafana/resource_datasource_query_test.go`
- [ ] T023 [US2] 新增失败诊断 acceptance 用例 `internal/resources/grafana/resource_datasource_query_acc_test.go`

### Implementation for User Story 2

- [X] T024 [US2] 完善输入校验错误信息（字段级、可操作提示）于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T025 [US2] 实现 cws-lib-go 错误到 Terraform diagnostics 的分类映射于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T026 [US2] 在失败响应中补充请求上下文摘要（不含敏感原始结果）于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T027 [US2] 实现瞬时失败重试后的最终错误落盘策略（last_error）于 `internal/resources/grafana/resource_datasource_query.go`

**Checkpoint**: US1 + US2 均可独立验证

---

## Phase 5: User Story 3 - 查询范围与行为可控 (Priority: P3)

**Goal**: 支持 managed/proxy 两种模式及可控时间窗口，行为与声明一致。

**Independent Test**: 分别配置 managed/proxy 与默认/自定义 time_range，结果窗口与执行模式符合预期。

### Tests for User Story 3 (MANDATORY)

- [X] T028 [P] [US3] 新增默认 time_range（now-24h~now，UTC 分钟对齐）测试 `internal/resources/grafana/resource_datasource_query_test.go`
- [X] T029 [P] [US3] 新增 proxy 模式请求构造测试 `internal/resources/grafana/resource_datasource_query_test.go`
- [ ] T030 [US3] 新增模式切换与时间窗口 acceptance 用例 `internal/resources/grafana/resource_datasource_query_acc_test.go`

### Implementation for User Story 3

- [X] T031 [US3] 实现 proxy 模式执行路径（GET/POST、params/body）于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T032 [US3] 实现 mode 互斥块校验（managed_query/proxy_query）于 `internal/resources/grafana/resource_datasource_query.go`
- [X] T033 [US3] 实现 time_range 默认值注入与标准化透传于 `internal/resources/grafana/resource_datasource_query_helpers.go`
- [X] T034 [US3] 在摘要中持久化时间边界与模式信息于 `internal/resources/grafana/resource_datasource_query.go`

**Checkpoint**: US1/US2/US3 全量能力可独立验证

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 完成文档、质量门禁与回归收敛

- [X] T035 [P] 更新资源示例与描述输入 `examples/resources/grafana_datasource_query/resource.tf` 与 `internal/resources/grafana/resource_datasource_query.go`
- [ ] T036 运行文档生成并校验无手工编辑 `go generate ./...`（影响 `docs/resources/` 生成结果）
- [X] T037 运行单元测试回归 `go test ./internal/resources/grafana/...`
- [X] T038 运行 provider 关联测试回归 `go test ./pkg/provider/...`
- [ ] T039 运行 acceptance 目标验证 `make testacc`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2 → Phase 3/4/5 → Phase 6
- 用户故事阶段均依赖 Phase 2 完成
- US2、US3 不阻塞 US1 的 MVP 交付

### User Story Dependencies

- **US1 (P1)**: 仅依赖 Foundational，可最先交付
- **US2 (P2)**: 依赖 Foundational，可在 US1 完成后增量交付
- **US3 (P3)**: 依赖 Foundational，可与 US2 并行推进

### Within Each User Story

- 先测试任务（T013-015 / T021-023 / T028-030），再实现任务
- 同一文件上的实现任务顺序执行，跨文件可并行

## Parallel Opportunities

- Setup: T003 与 T004 可并行
- Foundational: T007 与 T008 可并行
- US1: T013 与 T014 可并行
- US2: T021 与 T022 可并行
- US3: T028 与 T029 可并行
- Polish: T035 与测试执行准备可并行

## Parallel Example: User Story 1

- 并行启动: T013 `internal/resources/grafana/resource_datasource_query_test.go`
- 并行启动: T014 `internal/resources/grafana/resource_datasource_query_test.go`
- 串行接续: T016 → T017 → T018 → T019 → T020

## Parallel Example: User Story 2

- 并行启动: T021 `internal/resources/grafana/resource_datasource_query_test.go`
- 并行启动: T022 `internal/resources/grafana/resource_datasource_query_test.go`
- 串行接续: T024 → T025 → T026 → T027

## Parallel Example: User Story 3

- 并行启动: T028 `internal/resources/grafana/resource_datasource_query_test.go`
- 并行启动: T029 `internal/resources/grafana/resource_datasource_query_test.go`
- 串行接续: T031 → T032 → T033 → T034

## Implementation Strategy

### MVP First (US1)

1. 完成 Phase 1 与 Phase 2
2. 仅交付 Phase 3（US1）
3. 执行 T015 + T037 验证 MVP

### Incremental Delivery

1. US1 上线后再补 US2 失败诊断
2. 最后补 US3 模式与时间窗口控制
3. 每个故事完成后均可独立演示与验收

### Format Validation

- 所有任务均满足 `- [ ] Txxx ...` 清单格式
- 所有用户故事任务均包含 `[US1]/[US2]/[US3]` 标签
- 并行任务均显式标记 `[P]`
- 每个任务描述均包含明确文件路径