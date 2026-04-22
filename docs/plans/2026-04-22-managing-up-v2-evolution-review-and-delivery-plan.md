# Managing-Up v2.0 评审与开发计划

**日期**: 2026-04-22  
**输入**: `Managing-Up v2.0 架构演进设计文档：Agent 智能操作系统与治理枢纽`  
**结论**: 方向正确，但当前版本不宜直接按原文落地。应先完成“执行治理内核收敛”，再按 P0/P1/P2 分阶段推进。

## 1. 评审结论

v2.0 的核心判断是成立的: `Registry-first` 应升级为 `Runtime + Governance-first`。  
但当前仓库实际状态仍处于 `v1.0 骨架 + 局部能力雏形`：

- 已有执行引擎、Trace、Evaluator、MCP Router、Skill Market 基础设施。
- 尚未形成统一的治理闭环，尤其缺少:
  - 网关前置合规检查与策略决策层
  - Skill 版本发布与回归门禁联动
  - 远程 Memory Hub 与 RAG 注入链路
  - Any-to-MCP 的声明式适配与响应压缩

因此，v2.0 不应理解为“新增几个模块”，而应理解为“把现有运行时、路由层、评测层、数据层收敛成一个受治理的中间件内核”。

## 2. 关键发现

### F1. “统一执行平面”方向正确，但当前代码仍是分层拼接，不是统一内核

代码现状：

- `apps/api/internal/engine/engine.go` 已实现执行状态机和 Trace 发射。
- `apps/api/internal/service/mcp_router_service.go` 仍是简单匹配服务，只有 `taskTypes + tags -> best server`。
- `apps/api/internal/server/handlers/mcp_router.go` 路由请求只消费 `structured.task_type/tags`，没有 Pre-flight policy、审批、风险分级、会话态。

结论：

- v2.0 不能直接把 Router 和 Engine “绑在一起”。
- 必须先定义一个明确的 `Execution Governance Contract`，把以下能力抽到统一中间层：
  - 请求归一化
  - 风险/合规判定
  - 路由决策
  - Trace session 建立
  - 执行后评测回填

### F2. “闭环治理”是 v2.0 的主轴，但当前发布链路还没打通

代码现状：

- `apps/api/internal/evaluator/runner.go` 已有评测执行器和多种 evaluator。
- `apps/api/internal/service/skill_enterprise_service.go` 只覆盖依赖、评分、市场查询。
- `apps/api/internal/server/server.go` 中 `repoToSkillRepoAdapter` 的多个企业能力方法仍返回空实现，说明 Skill Market 企业链路尚未真正接通。

结论：

- Regression Gate 应被定义为 v2.0 的 P0，而不是附属功能。
- 如果没有 “版本变更 -> 自动评测 -> 阈值判断 -> promote gate” 这条链，治理不会闭环。

### F3. Memory Hub 仍处于“数据结构存在、系统能力不存在”的阶段

代码现状：

- `apps/api/internal/engine/context/memory.go` 只有进程内 `MemoryContext`，没有远程持久化、租户隔离、检索接口。
- 当前迁移中没有 `memory_cells`、`memory_indexes`、`memory_links` 等支撑表。

结论：

- v2.0 可以做 Memory Hub，但必须收敛目标。
- 第一阶段只做 `remote memory store + retrieval metadata + session binding`，不要一开始就把长期记忆、向量检索、自动 RAG 编排全部打包。

### F4. Any-to-MCP 适配层可做，但“响应裁剪”不能建立在当前 tokenizer 假设上

代码现状：

- `apps/api/internal/engine/tokenizer.go` 只是 `chars / 4` 级别估算，不具备稳定的语义压缩依据。
- 当前代码中也没有通用 Swagger/OpenAPI 到 MCP adapter 的声明式编排框架。

结论：

- Bridge Adapter 可作为 P2。
- 其中应拆为两个子模块：
  - 声明式 REST adapter 生成
  - 响应摘要/裁剪策略
- 不建议把“语义压缩”直接绑定到现有 tokenizer；应先落 JSON selector、字段白名单、大小阈值裁剪，再考虑摘要模型。

### F5. v2.0 数据模型需要避免与现有 `execution_traces` 重叠

代码现状：

- `apps/api/migrations/0005_add_trace_and_tasks.up.sql` 已创建 `execution_traces`。
- `apps/api/internal/repository/postgres/repository.go` 已写入和查询 trace。
- v2.0 草案新增 `mcp_gateway_sessions`，但没有定义它与 `execution_traces`、`executions`、`mcp_router_logs` 的关系。

结论：

- 不应新建一个与 trace 平行的日志体系。
- 推荐关系：
  - `mcp_gateway_sessions`: 会话头
  - `execution_traces`: 事件流
  - `mcp_router_logs`: 路由决策日志
  - 三者通过 `session_id/execution_id/correlation_id` 串联

## 3. 建议的 v2.0 收敛架构

### 3.1 目标形态

v2.0 建议收敛为四层：

1. `Gateway Control Plane`
   - 请求归一化
   - Task intent 解析
   - Pre-flight policy
   - Route decision

2. `Execution Runtime`
   - Skill execution
   - Approval wait state
   - Tool/MCP invocation
   - Trace emission

3. `Governance Hub`
   - SOP/compliance binding
   - Regression Gate
   - Capability snapshots
   - promote/reject decision

4. `Memory Hub`
   - session memory
   - cross-session state
   - retrieval metadata
   - RAG injection hooks

### 3.2 建议新增的统一协议

先定义统一内部协议，再开始大规模改代码：

- `GatewaySession`
- `ExecutionIntent`
- `PolicyDecision`
- `RouteDecision`
- `EvaluationGateResult`
- `MemoryBinding`

如果没有这层协议，v2.0 很容易继续演化成多个 handler/service/repo 的局部拼装。

## 4. 开发优先级

### P0: 治理内核收敛

目标：让 “请求 -> 执行 -> 评测 -> 发布” 形成最小闭环。

包含：

- 定义统一 session/decision 数据模型
- 为 MCP Router 增加 Pre-flight policy hook
- 为 Skill promote 增加 Regression Gate
- 打通 capability snapshot 回填
- 补齐当前 enterprise adapter/repository 的空实现

为什么是 P0：

- 这是 v2.0 是否成立的分水岭。
- 没有它，Memory/RAG/Bridge Adapter 都只是新增功能，不是“操作系统内核”。

### P1: Memory Hub MVP

目标：提供可治理、可追踪、跨会话的远程记忆，而不是仅有进程内上下文。

包含：

- `memory_cells`、`memory_cell_bindings`
- execution/session 级 memory 绑定
- 基于 metadata 的检索
- 网关自动注入 memory context

### P2: Bridge Adapter

目标：降低旧系统接入成本。

包含：

- Swagger/OpenAPI 导入
- REST -> MCP adapter 模板
- 请求参数插值
- 响应字段过滤、裁剪、摘要

### P3: 高阶治理增强

目标：把内核从“可用”拉到“企业级”。

包含：

- 熔断与流量治理策略统一到网关控制面
- trace/session 的可视化
- 多租户 memory 隔离
- 风险策略版本化

## 5. 建议分期

## Phase 0: 架构收敛与补债

周期：3-5 天

交付：

- 明确内部协议和数据关系图
- 明确 `execution_traces / mcp_router_logs / gateway_sessions` 关系
- 盘点并补齐当前空实现

涉及目录：

- `apps/api/internal/service`
- `apps/api/internal/server`
- `apps/api/internal/repository`
- `docs/`

完成标准：

- 评审通过一份统一时序图和 ER 图
- P0 所需 repository/service contract 固定

## Phase 1: Unified Governance Kernel

周期：1-2 周

交付：

- `mcp_gateway_sessions`
- gateway pre-flight policy
- SOP/compliance guard
- promote 前自动评测门禁
- `skill_capability_snapshots`

涉及目录：

- `apps/api/migrations`
- `apps/api/internal/service`
- `apps/api/internal/server/handlers`
- `apps/api/internal/evaluator`
- `apps/api/internal/repository/postgres`
- `apps/web/app/mcp-router`
- `apps/web/app/skills`

完成标准：

- Skill 版本变更后可自动触发回归评测
- 未过线版本不可 promote
- 路由与执行可被同一 session 追踪

## Phase 2: Memory Hub MVP

周期：1 周

交付：

- `memory_cells`
- memory CRUD/service
- execution/session memory binding
- metadata retrieval + gateway injection

涉及目录：

- `apps/api/migrations`
- `apps/api/internal/engine/context`
- `apps/api/internal/service`
- `apps/api/internal/server`
- `apps/web/app`

完成标准：

- 无状态 Agent 可跨会话读取指定 memory scope
- 所有 memory 读写具备 trace 和审计信息

## Phase 3: Bridge Adapter MVP

周期：1-2 周

交付：

- OpenAPI 导入器
- adapter template schema
- interpolateInput 扩展
- response optimizer

涉及目录：

- `apps/api/internal/gateway`
- `apps/api/internal/engine`
- `apps/api/internal/service`
- `apps/web/app/mcp-router`

完成标准：

- 至少 1 个企业 REST 系统可通过模板导入并被 Agent 调用
- 大响应可经过可配置裁剪后返回

## 6. 详细开发任务拆解

### Track A. 数据模型

优先级：`P0`

- 新增 `mcp_gateway_sessions`
- 新增 `skill_capability_snapshots`
- 新增 `memory_cells`
- 视需要新增 `memory_bindings`
- 为 `mcp_router_logs` 增加 `session_id`
- 为 `execution_traces` 增加可选 `session_id/correlation_id`

验收：

- ER 图可解释完整生命周期
- 不出现双份 trace truth source

### Track B. Gateway Control Plane

优先级：`P0`

- 在 Router handler/service 前增加 request normalization
- 增加 policy decision hook
- 引入 `Compliance Guard`
- 引入 session 创建与 trace 绑定

验收：

- 高风险/合规 Skill 能进入审批态
- 所有路由请求有 session 头记录

### Track C. Regression Gate

优先级：`P0`

- Skill version 变更时关联 TestCases 与指标
- 调用 `EvaluationRunner`
- 将结果写入 snapshot
- 在 promote 时校验阈值

验收：

- promote API 在低于阈值时明确拒绝
- UI 可看到最近一次 gate 结果

### Track D. Memory Hub MVP

优先级：`P1`

- 定义 memory scope: `execution / session / agent / tenant`
- CRUD + query service
- gateway 注入 memory context

验收：

- 同一 agent/session 可读取前次关键状态
- 可配置禁用 memory auto-injection

### Track E. Bridge Adapter

优先级：`P2`

- OpenAPI/Swagger parser
- template registry
- request mapping
- response field selector
- response summarizer

验收：

- 模板可声明输入映射和输出裁剪规则
- 非结构化大 JSON 不直接全量返回给 Agent

## 7. 推荐任务顺序

1. 先补齐企业能力空实现，清掉“文档存在、代码未接通”的假闭环。
2. 再定义 `session + policy + gate` 的统一 contract。
3. 之后推进 migration 与 repo/service 改造。
4. 然后接 UI 可视化。
5. 最后再做 Memory Hub 和 Bridge Adapter。

## 8. 主要风险

- 最大风险不是功能复杂，而是边界继续发散。
- 如果 Router、Engine、Evaluator、Memory 继续各自演进，v2.0 会再次退化成并列模块集合。
- 如果先做 Memory 或 Bridge Adapter，而治理闭环未完成，系统会先变大、后失控。
- 当前部分企业功能接口仍为空实现，若不先补齐，任何 v2.0 计划都会建立在错误完成度判断上。

## 9. 建议的里程碑验收

### M1: 治理闭环成立

- 一个 Skill 新版本提交后自动评测
- 结果写入 snapshot
- promote 被 gate 控制
- trace/session 全链路可查

### M2: Memory Hub 可用

- Agent 可跨会话取回状态
- 可审计、可隔离、可关闭

### M3: 旧系统接入可复制

- 通过模板快速导入 1 个 REST 系统
- Agent 可稳定调用
- 返回上下文成本受控

## 10. 最终建议

这份 v2.0 设计不需要推翻，但需要重排优先级：

- 先做 `Governance Kernel`
- 再做 `Memory Hub MVP`
- 最后做 `Bridge Adapter`

对当前仓库而言，最重要的不是继续扩写“OS 愿景”，而是先把已有 Router、Engine、Evaluator、Skill Enterprise 能力收敛成一个可证明有效的治理内核。
