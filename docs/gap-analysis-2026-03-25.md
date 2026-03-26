# Agent Harness 差距分析报告

**生成日期:** 2026-03-25  
**分析依据:** 代码库全面扫描（6 个并行探索任务）  
**对比基准:** `docs/product-spec.md` — Agent Harness 产品规格书 v1.0

---

## 一、总体评估

| 维度 | 产品计划 Phase 0 | 当前实现 | 差距 |
|------|------------------|---------|------|
| **Domain 1: Benchmark/Task** | 基础 Task Registry | CRUD + 简单过滤 | ⚠️ 缺版本控制、capability tagging、difficulty 精细化、hidden set |
| **Domain 2: Experiment Orchestration** | 基础 Sweep Engine | task×agent 矩阵并行 | ⚠️ 缺 hyperparameter sweep、分布式 runner（单 worker）、budget scheduler |
| **Domain 3: Agent Execution & Replay** | Trajectory Capture + Replay Bus | 11 event types + snapshot replay | ⚠️ 缺 full replay bus、mock 环境、policy variant runner |
| **Domain 4: Evaluation & Capability** | Metric Plugin Framework | 2 种 metric (exact_match, semantic_similarity) | ❌ 缺 judge model、statistical metric、capability graph、variance analyzer |
| **Domain 5: Analytics/Observability** | Capability Dashboard | 6 指标卡（静态）| ❌ 缺 radar view、trend line、failure mining、cost intelligence |
| **Domain 6: Governance/CI** | CI Gate + RBAC | Approval workflow（人工审批）| ❌ 缺 CI Gate API、GitHub integration、benchmark integrity、RBAC 多租户 |

---

## 二、按 Domain 详细差距

### Domain 1: Benchmark / Task Management

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| Task Registry API + UI | ✅ `POST/GET /tasks` + `/tasks` 页面 | 缺 `dataset_version`、`capability_tags`、`is_hidden` 字段 |
| Task Builder（从 trace 生成）| ❌ 无 | 完全缺失 — 没有 `POST /api/v1/tasks/from-trace` |
| Skill-level Benchmark Pack | ❌ 无 | 缺预置 benchmark packs（tool_selection_pack, planning_pack, memory_pack）|
| Sampling Policy Engine | ❌ 无 | 缺 stratified sampling、entropy-aware sampling |
| Task 扩展字段（product-spec 定义）| ⚠️ 部分实现 | 缺 `task_type`、`input_source`、`execution_model`、`execution_temperature` 等完整字段 |

**数据库差距：**
- `tasks` 表已扩展（migration 0007）但 UI 只暴露 name/description/tags/difficulty
- 缺 `is_hidden`（hidden set support）
- 缺 `dataset_version` 和 `input_format`

---

### Domain 2: Experiment Orchestration

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| Sweep Engine（model/prompt/temp grid）| ❌ 无 | 完全缺失 — 没有 hyperparameter sweep |
| Distributed Runner | ⚠️ 单 worker | 只有 1 个 worker（2s polling）— 缺 >1 worker 支持 |
| Budget Scheduler | ❌ 无 | 缺 token budget、time budget、priority queue |
| Deterministic Mode | ⚠️ Seed 字段存在 | `execution_seed` 字段已定义但未在线执行层面强制 |

**当前实现细节：**
- `ExperimentService.RunExperiment()` 用 semaphore 控制 10 并发 worker
- 但 `Worker` 是单实例，缺分布式能力
- Experiment 支持 task_ids × agent_ids 矩阵，但只跑已有 agent_id（非动态 variant）

---

### Domain 3: Agent Execution & Replay

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| Trajectory Capture | ⚠️ 11 event types | 缺 `model`、`tokens_used`、`latency_ms` 丰富字段 |
| Replay Bus | ⚠️ Snapshot + DeterministicRNG | 缺 full trajectory replay（从 DB 回放）、tool mocking、environment snapshot |
| Skill Sandbox | ❌ 无 | 缺 deterministic execution、fake API、latency injection |
| Policy Variant Runner | ❌ 无 | 没有 A/B/C 多 policy 并行对比 |
| Trajectory Search Visualization | ⚠️ trace timeline UI 存在 | 缺 decision tree、failure path 可视化 |

**Tool Gateway 差距：**
- 当前实现：HTTP adapter + Calculator mock
- 缺 real tool gateway（shell/SQL safety、key management）
- 缺 latency injection、fake API 能力

---

### Domain 4: Evaluation & Capability Modeling

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| Metric Plugin Framework | ⚠️ 2 种 type | 缺 `judge_model`（LLM 评分）、`statistical` metric |
| Capability Graph | ❌ 无 | 没有分数 → capability → product KPI 的映射 |
| Variance Analyzer | ❌ 无 | 缺 multi-run 分布、confidence interval、significance test |
| Capability Diff Engine | ❌ 无 | 没有"这次改动哪里变聪明/变笨"的自动分析 |

**当前 metric types（只 2 种）：**
- `exact_match`
- `semantic_similarity`

**产品计划要求（5+ 种）：**
- deterministic
- judge_model
- statistical
- llm_judge
- 等等

---

### Domain 5: Analytics & Observability

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| Capability Dashboard (Radar View) | ❌ 无 | 只有 6 指标卡，没有 radar view |
| Trend Line | ❌ 无 | 没有 time-series capability graph |
| Regression Diff | ⚠️ `POST /check-regression` 存在 | 但没有可视化 diff 面板 |
| Failure Mining | ❌ 无 | 没有 trajectory clustering、dead-end detection |
| Cost Intelligence | ❌ 无 | 没有 cost per success、marginal gain curve |
| Entropy Budget Panel | ❌ 无 | 完全没有 trajectory variance/policy divergence 洞察 |

**当前 Dashboard（只 6 静态指标）：**
- Active Skills
- Published Versions
- Running Executions
- Waiting Approvals
- Success Rate
- Avg Duration

---

### Domain 6: Governance & CI Integration

| 产品计划功能 | 当前状态 | 差距 |
|------------|---------|------|
| CI Gate API | ❌ 无 | 没有 `POST /api/v1/ci-gates` |
| CI Gate Rules Engine | ❌ 无 | 没有 benchmark threshold → block merge |
| GitHub Integration | ❌ 无 | 没有 GitHub Check Runs |
| Benchmark Integrity Check | ❌ 无 | 没有 contamination detection |
| Access Control / Multi-tenant | ⚠️ 基础 auth 存在 | 没有 team 隔离、RBAC per team |
| Skill Contract Management | ❌ 无 | 没有 schema versioning、backward compatibility |

---

## 三、Phase 1 优先级差距

### Phase 1: Benchmark Foundation

| 功能 | 产品计划交付物 | 当前实现 | 差距 |
|------|--------------|---------|------|
| Task Registry API + UI | `POST/GET /api/v1/tasks` + Task List Page | ✅ 已有 | 完整 |
| Task Builder（从 trace 生成）| `POST /api/v1/tasks/from-trace` | ❌ 无 | 完全缺失 |
| Metric Plugin Framework | `metric: deterministic / judge_model / statistical` | ⚠️ 只有 deterministic + semantic_similarity | 缺 judge_model + statistical |
| Basic Sweep Engine | `POST /api/v1/experiments` 支持多-variant | ⚠️ 矩阵执行存在但无 variant 支持 | 缺 variant/sweep |

**验收标准差距：** *"一个 Task 可以被创建、执行、打分"* — 基本流程通，但打分只有 2 种 metric（非计划中的 3 种）

---

## 四、基础设施差距

| 组件 | 产品计划 | 当前实现 | 差距 |
|------|---------|---------|------|
| Worker Pool | 分布式多 worker | 单 worker（2s polling）| ❌ |
| LLM Provider | 10 家 | ✅ 10 家 | 完整 |
| Agent SDK | Python + TypeScript | ✅ 都有 | 完整 |
| OpenAPI Spec | Agent-facing 完整 | ⚠️ 只覆盖 agent 操作的子集 | 不完整 |
| 数据库迁移 | 8 个 migration | ✅ 8 个 | 完整 |
| 测试 | - | 83 tests | 完整 |

---

## 五、关键发现总结

### ✅ 已完整实现（Phase 0 基线）

1. **单执行引擎** — pending→running→waiting_approval→succeeded/failed 状态机
2. **Tool Gateway（Mock）** — HTTP adapter + Calculator
3. **Skill Registry + Versions** — CRUD + YAML spec
4. **Dashboard（基础）** — 6 指标卡
5. **Approval Workflow** — 人工审批节点
6. **Trace Events** — 11 event types，DB 持久化
7. **JWT + CORS 认证** — 基础安全
8. **LLM Provider 集成** — 10 家提供商
9. **Experiment infrastructure** — task×agent 矩阵并行（10 worker）
10. **Replay Snapshot** — DeterministicRNG + snapshot

### ❌ 严重缺失（Phase 1 核心功能）

1. **Task Builder** — 从 trace 自动生成 task
2. **Judge Model Metric** — LLM 作为评判者
3. **Capability Graph** — 分数 → capability → KPI 映射
4. **Sweep Engine** — hyperparameter grid search
5. **CI Gate** — benchmark threshold → block merge
6. **Radar Dashboard** — capability 多维可视化
7. **Tool Gateway Real Implementation** — shell/SQL safety

### ⚠️ 部分实现（需要深化）

1. **Experiment** — 有基础设施但缺 sweep/variant
2. **Deterministic Mode** — 字段存在但未强制
3. **Trace Visualization** — 有 timeline 但缺 decision tree
4. **Metric Framework** — 2/5 种类型

---

## 六、建议优先级

| 优先级 | 功能 | 所属 Domain | 理由 |
|--------|------|------------|------|
| **P0** | Task Builder（从 trace 生成）| D1 | 数据底座，没有 task 就没有实验 |
| **P0** | Judge Model Metric | D4 | 核心评分能力，exact_match 不足以评估 LLM 输出质量 |
| **P1** | Basic Sweep Engine | D2 | Phase 1 核心交付物 |
| **P1** | Capability Diff API | D4 | 自动回答"变聪明/变笨" |
| **P2** | CI Gate API | D6 | 企业门槛 |
| **P2** | Radar Dashboard | D5 | Phase 5 基础 |
| **P2** | Tool Gateway Real Impl | D3 | High priority 在 todo-plan.md |

---

## 附录：代码库关键文件索引

| 组件 | 关键文件 |
|------|---------|
| 执行引擎 | `apps/api/internal/engine/engine.go` |
| Worker | `apps/api/internal/engine/worker.go` |
| Tool Gateway | `apps/api/internal/engine/tool_gateway.go` |
| Trace | `apps/api/internal/engine/trace.go` |
| Replay | `apps/api/internal/engine/replay.go` |
| Experiment Service | `apps/api/internal/service/experiment.go` |
| Task Service | `apps/api/internal/service/task.go` |
| Metric Service | `apps/api/internal/service/metric.go` |
| LLM Providers | `apps/api/internal/llm/` |
| 数据库 Schema | `apps/api/migrations/` |
| Repository | `apps/api/internal/repository/postgres/repository.go` |
| API Handlers | `apps/api/internal/server/server.go` |
| 前端页面 | `apps/web/app/` |
| 前端组件 | `apps/web/components/` |

---

*本报告基于 2026-03-25 对代码库的全面扫描分析生成*
