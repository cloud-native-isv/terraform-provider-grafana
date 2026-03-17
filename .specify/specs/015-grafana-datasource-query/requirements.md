# Requirements Specification: Grafana Datasource Query Resource

**Requirement Branch**: `015-grafana-datasource-query`  
**Created**: 2026-03-16  
**Status**: Draft  
**Input**: User description: "grafana_datasource_query resource,这个resource 调用pkg/cws-lib-go子模块提供的grafana datasource query相关API进行查询。"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - 声明式执行数据源查询 (Priority: P1)

作为平台维护者，我可以在基础设施配置中声明一个查询资源，并在执行后获得稳定、可消费的查询结果输出。

**Why this priority**: 这是本特性的核心价值，直接决定该资源是否可用于自动化查询场景。

**Independent Test**: 仅实现该用户故事时，可通过创建资源并读取结果字段完成端到端验证，且可直接交付“声明式查询”价值。

**Acceptance Scenarios**:

1. **Given** 用户已提供有效 Grafana 连接信息与数据源标识，**When** 用户创建查询资源，**Then** 系统成功执行查询并在资源状态中保存结果。
2. **Given** 用户再次执行相同配置，**When** 上游数据未变化，**Then** 系统返回一致结果且不产生非预期变更。

---

### User Story 2 - 查询失败可诊断 (Priority: P2)

作为平台维护者，当查询失败时，我可以看到明确的失败类别与上下文信息，从而快速定位配置或权限问题。

**Why this priority**: 失败可诊断性决定了该资源在生产环境中的可运维性。

**Independent Test**: 使用无效查询参数触发失败，验证返回的错误信息是否可区分输入问题、认证问题与上游问题。

**Acceptance Scenarios**:

1. **Given** 查询参数不完整或无效，**When** 用户创建或刷新资源，**Then** 系统返回可理解且可操作的校验错误。

---

### User Story 3 - 查询范围与行为可控 (Priority: P3)

作为平台维护者，我可以声明查询模式与时间范围，并让资源按预期执行，满足不同观测与排障场景。

**Why this priority**: 行为可控可以提升资源在多数据源与多时间窗口场景下的适配能力。

**Independent Test**: 分别配置不同查询模式和时间范围，验证查询结果与预期窗口一致。

**Acceptance Scenarios**:

1. **Given** 用户指定查询模式与时间范围，**When** 资源执行查询，**Then** 返回结果仅覆盖声明范围并符合所选模式的响应语义。

---

### Edge Cases

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right edge cases.
-->

- 查询返回空数据集时，资源应成功完成并明确标记结果为空，而不是视为失败。
- 查询结果体量较大时，资源仅持久化机读摘要，不将原始结果写入 state，并保持摘要字段结构稳定。
- 查询模式与输入参数组合不合法时，资源应在执行前阻断并返回可操作提示。
- 认证通过但数据源权限不足时，资源应返回权限相关错误并保留请求上下文。

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: 系统必须提供 `grafana_datasource_query` 资源，允许用户声明目标数据源与查询请求内容。
- **FR-002**: 系统必须支持至少两类查询执行模式：统一查询模式与代理查询模式，并在输入中显式区分。
- **FR-003**: 系统必须支持查询时间范围输入；当用户未提供 `time_range` 时，默认使用最近 24 小时（`from=now-24h,to=now`），并使用 UTC 且按分钟边界对齐。
- **FR-004**: 系统必须在资源状态中输出可机读的查询结果，且输出结构在同一模式下保持稳定。
- **FR-005**: 系统必须在查询失败时返回可分类的错误信息，至少覆盖校验、认证/授权、上游失败与网络失败。
- **FR-006**: 系统必须在资源生命周期执行中复用项目已有的 Grafana 查询能力，以确保与现有 CLI 查询行为一致。
- **FR-007**: 系统必须在输入无效时阻止执行并返回明确的字段级错误说明。
- **FR-008**: 用户必须能够通过刷新资源重新执行查询，并获得当前时点的最新结果。
- **FR-009**: 系统必须避免在等价查询结果下产生无意义状态变更。
- **FR-010**: 系统必须将查询结果字段默认作为非敏感数据存储与输出，不对结果内容自动标记 `Sensitive`。
- **FR-011**: 对于瞬时失败（至少包括 `429`、`5xx`、短暂网络超时），系统必须采用指数退避重试，最多 3 次；超过上限后返回可分类失败。
- **FR-012**: 资源 `id` 必须由规范化查询配置生成确定性哈希；同一等价配置必须生成相同 `id`，配置变更后必须生成新 `id`。
- **FR-013**: 系统必须仅将查询结果的机读摘要写入状态，不持久化原始结果体；摘要至少包含结果计数、时间边界与稳定比较所需标识。

### Key Entities *(include if requirement involves data)*

- **Datasource Query Resource**: 用户声明式配置对象，包含数据源标识、执行模式、查询参数、时间范围与执行选项。
- **Query Request**: 单次查询请求载荷，包含查询目标、范围边界与请求元信息。
- **Query Result**: 查询返回的机读结果集合，包含结果数据、结果标识与可追踪元数据。
- **Query Error**: 标准化错误对象，包含错误类别、可读消息与定位上下文。

## Assumptions

- 查询资源是只读用途，不负责创建或修改 Grafana 中的仪表板与数据源配置。
- 查询执行依赖调用方已具备合法连接信息与目标数据源访问权限。
- 查询结果以“可重复消费”为优先目标，默认不承诺长期历史快照存储。
- 当上游返回空结果时，系统视为成功响应而非异常。

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 在验收场景中，用户可在 5 分钟内完成资源声明并成功获得首次查询结果。
- **SC-002**: 针对有效输入的查询执行成功率达到 99% 及以上（不含上游服务不可用时段）。
- **SC-003**: 对于输入错误场景，100% 返回可分类且可操作的错误信息。
- **SC-004**: 在 20 组重复执行验证中，等价结果场景的非预期状态变更次数为 0。

## Clarifications

### Session 2026-03-16

- Q: 当用户未提供 `time_range` 时，`grafana_datasource_query` 的默认时间范围应采用哪种策略？ → A: 默认最近 24 小时（`from=now-24h,to=now`，UTC，分钟对齐）。
- Q: `grafana_datasource_query` 的结果输出敏感性策略应如何定义？ → A: 所有结果字段都非敏感（明文入 state）。
- Q: 查询遇到瞬时失败（如 `429`、`5xx`、短暂网络超时）时，资源应采用哪种重试策略？ → A: 指数退避重试，最多 3 次。
- Q: `grafana_datasource_query` 作为只读查询资源，其 `id` 应采用哪种生成策略？ → A: 规范化配置哈希（确定性）。
- Q: 当查询结果过大可能导致 Terraform state 膨胀时，资源应采用哪种处理策略？ → A: 仅存摘要，不存原始结果。
