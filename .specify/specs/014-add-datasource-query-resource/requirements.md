# Requirements Specification: Grafana Datasource Query Resource

**Requirement Branch**: `014-add-datasource-query-resource`  
**Created**: 2026-03-13  
**Status**: Draft  
**Input**: User description: "grafana_datasource_query resource,这个resource 调用pkg/cws-lib-go子模块提供的grafana datasource query相关API进行查询。"

## Feature Traceability *(mandatory)*

- Feature index impact: Update
- Affected Feature IDs: 018
- Evidence links: .specify/specs/014-add-datasource-query-resource/requirements.md
- Follow-up action in `.specify/memory/features.md`: Update Feature 018 latest spec reference and last updated date

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

### User Story 1 - Create Query Definition (Priority: P1)

As a Terraform user, I can define a datasource query resource with query payload, time range, and execution mode so query intent is managed as code.

**Why this priority**: Without declarative query definition, users cannot include query workflows in infrastructure automation.

**Independent Test**: Create a minimal resource configuration and run plan/apply; the query executes and returns structured result data in state outputs.

**Acceptance Scenarios**:

1. **Given** valid Grafana connection settings and a valid datasource identifier, **When** the user applies a configuration containing a managed query definition, **Then** the resource runs the query and stores normalized query result data.
2. **Given** a query definition missing required fields for the selected mode, **When** the user runs plan or apply, **Then** the system returns a validation error explaining missing inputs.

---

### User Story 2 - Re-run on Input Changes (Priority: P2)

As a Terraform user, I can trigger re-execution when query inputs change so state always reflects the latest query response for the declared parameters.

**Why this priority**: Reliable drift-free query output is required for repeatable automation and downstream references.

**Independent Test**: Apply once, change time window or query content, apply again, verify result output is refreshed and previous stale values are replaced.

**Acceptance Scenarios**:

1. **Given** an existing query resource in state, **When** the user updates query text or time-range inputs, **Then** the next apply re-executes the query and updates result fields.
2. **Given** no effective input change, **When** the user runs apply again, **Then** the resource does not perform unnecessary update actions.

---

### User Story 3 - Diagnose Query Failures (Priority: P3)

As a Terraform user, I can receive categorized, actionable errors for authentication, validation, network, and upstream query failures.

**Why this priority**: Fast diagnosis reduces failed pipeline time and repeated trial-and-error.

**Independent Test**: Execute with invalid credentials and invalid query payload; confirm user-visible errors include clear failure category and remediation direction.

**Acceptance Scenarios**:

1. **Given** invalid access credentials, **When** apply is executed, **Then** the resource fails with an authorization-category error.
2. **Given** datasource-side query rejection, **When** apply is executed, **Then** the resource fails with an upstream-category error and includes context for troubleshooting.

---

### Edge Cases

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right edge cases.
-->

- Datasource exists but returns an empty result set for the selected time range.
- Query returns very large result payloads that exceed practical state size constraints.
- Managed mode query requires time filters but caller omits time range inputs.
- Proxy mode receives an unsupported HTTP method or missing path.
- Query resource references a datasource identifier that is deleted after initial apply.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: 系统必须提供 `grafana_datasource_query` 资源，用于声明并执行 Grafana datasource 查询。
- **FR-002**: 资源必须支持两类查询模式（统一查询模式与代理查询模式），并在配置层面强制校验各自必填输入。
- **FR-003**: 资源必须允许用户声明查询时间范围，并在执行时将该时间范围应用到查询请求。
- **FR-004**: 资源必须将查询结果以稳定、可比较的结构保存到状态中，避免无意义漂移。
- **FR-005**: 当查询输入发生有效变化时，资源必须重新执行查询并刷新状态结果。
- **FR-006**: 当外部服务返回失败时，资源必须输出可归类的错误类型（至少覆盖校验、鉴权、网络、上游失败）。
- **FR-007**: 资源必须支持通过环境配置或显式参数使用现有 Grafana 连接凭据。
- **FR-008**: 资源必须在 datasource 不存在、无权限或查询语法错误等情况下给出清晰错误信息，不得静默成功。
- **FR-009**: 资源必须支持导入后读取现有资源状态并在后续计划中保持一致行为。

### Key Entities *(include if requirement involves data)*

- **Datasource Query Resource**: 表示单个可声明查询单元，关键属性包括查询模式、datasource 标识、查询内容、时间范围、执行选项。
- **Query Request Definition**: 用户声明的查询输入，包含查询目标、路径/方法或统一查询参数、可选请求体。
- **Query Result Snapshot**: 每次执行后保存的结果快照，包含结果数据、引用标识、执行元信息。
- **Error Classification**: 查询失败时的标准化错误分类及消息，用于用户诊断和自动化处理。

## Assumptions

- 默认同一资源一次仅执行一个逻辑查询请求集合，不在本阶段包含批量资源编排。
- 默认查询结果用于配置编排与审计用途，不作为高频近实时数据采集通道。
- 默认对敏感字段采用最小暴露原则，错误与状态中不回显不必要凭据内容。

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 90% 以上的有效查询配置可在一次 apply 内成功完成并产出可读取结果。
- **SC-002**: 对于缺失必填项的配置，100% 在 plan/apply 阶段被拦截并返回明确校验错误。
- **SC-003**: 常见失败场景（鉴权失败、网络失败、上游失败）100% 映射到明确错误类别。
- **SC-004**: 在输入未变更的连续两次 apply 中，资源计划结果保持稳定（无无效更新）。
- **SC-005**: 用户可在 10 分钟内完成首次资源配置并成功执行一次查询（基于 quickstart 流程验证）。

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
