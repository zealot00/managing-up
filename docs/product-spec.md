# Agent Harness 产品规格书
# Product Specification — Agent Harness

**Status:** Draft v1.0
**Version:** 1.0
**Date:** 2026-03-25
**Product:** Agent Harness — AI Agent 实验操作系统

---

## 1. 产品定位

### 1.1 一句话定义

**Agent Harness** 是一个用于**度量、调优、回归与治理** AI Agent 能力的**实验操作系统**。

### 1.2 目标用户

| 用户角色 | 核心诉求 |
|---------|---------|
| AI Infra Team | 能力可观测性、回归检测 |
| Agent Engineer | 快速迭代、debug 工具 |
| Applied Scientist | 能力建模、效果评估 |
| Platform Team | 成本控制、安全治理 |
| CTO / AI Lead | 量化 ROI、上线确定性 |

### 1.3 核心价值主张

- **降低 AI 系统有效熵** — 通过实验平台把不确定性变为可度量
- **提高能力演进速度** — 快速发现短板，快速验证
- **控制成本** — token budget、采样策略
- **提升上线确定性** — CI Gate + Benchmark Integrity

### 1.4 与当前 MVP 的关系

| 当前 MVP | 新蓝图定位 |
|---------|-----------|
| Governed SOP Execution System | Harness 的**执行引擎**（Domain 3）|
| Skill Registry | Harness 的**Task/Model 配置层** |
| Ops Console | Harness 的**Analytics UI**（Domain 5）|
| Approval Workflow | Harness 的**Governance 基础**（Domain 6）|

> 当前代码库是 Agent Harness 的 **v0.1 — 单一执行路径**。蓝图需要在此基础上扩展为**实验操作系统**。

---

## 2. 功能域全景图（6 Domains）

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agent Harness Platform                        │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Domain 1     │  │ Domain 2     │  │ Domain 3            │  │
│  │ Benchmark /  │→ │ Experiment   │→ │ Agent Execution     │  │
│  │ Task Domain  │  │ Orchestration│  │ & Replay            │  │
│  │              │  │              │  │                     │  │
│  │ Task Registry│  │ Sweep Engine │  │ Trajectory Capture  │  │
│  │ Task Builder │  │ Distributed  │  │ Replay Bus          │  │
│  │ Benchmark    │  │ Runner       │  │ Skill Sandbox       │  │
│  │ Packs        │  │ Budget       │  │ Policy Variant      │  │
│  │ Sampling     │  │ Scheduler    │  │ Runner              │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Domain 4     │  │ Domain 5     │  │ Domain 6            │  │
│  │ Evaluation & │  │ Analytics &  │  │ Governance &        │  │
│  │ Capability   │  │ Observability│  │ CI Integration      │  │
│  │ Modeling     │  │              │  │                     │  │
│  │              │  │ Capability   │  │ CI Gate            │  │
│  │ Metric Plugin│  │ Dashboard    │  │ Benchmark          │  │
│  │ Framework    │  │ Failure      │  │ Integrity          │  │
│  │ Capability   │  │ Mining       │  │ Skill Contract     │  │
│  │ Graph        │  │ Cost         │  │ Management         │  │
│  │ Variance     │  │ Intelligence │  │ Access Control /   │  │
│  │ Analyzer     │  │              │  │ Multi-tenant       │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. 功能域详解与优先级

### Domain 1: Benchmark / Task Management（数据底座）

**定位：** Harness 的数据底座。没有 Task 数据，一切实验无从谈起。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **Task Registry** | task dataset versioning, capability tagging, difficulty labeling, hidden set support | P1 | 新建 |
| **Task Builder** | 从线上 trace 生成 task, synthetic task generator, task clustering | P2 | 新建 |
| **Skill-level Benchmark Pack** | 预置 benchmark packs（tool_selection_pack, planning_pack, memory_pack） | P2 | 新建 |
| **Sampling Policy Engine** | stratified sampling, entropy-aware sampling, cost-budget sampling | P3 | 新建 |

#### Task 数据模型（扩展）

```yaml
Task:
  id: string
  name: string
  dataset_version: string
  capability_tags: [planning, tool_use, reasoning, ...]
  difficulty: easy | medium | hard | extreme
  is_hidden: bool  # hidden set 不泄露给模型
  input: object
  expected_output: object
  metrics: [primary_metric, secondary_metrics]
```

---

### Domain 2: Experiment Orchestration（调度核心）

**定位：** Harness 的调度核心。决定如何高效、可控地跑实验。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **Sweep Engine** | model sweep, prompt sweep, skill config sweep, temperature grid | P1 | 新建 |
| **Distributed Runner** | parallel execution, retry policy, seed control, deterministic mode | P1 | 部分实现（Worker 存在）|
| **Budget Scheduler** | token budget, time budget, priority queue | P2 | 新建 |

#### 当前 vs 目标

| 组件 | 当前 | 目标 |
|------|------|------|
| 执行触发 | `POST /executions` 单一执行 | Sweep — 并行跑多个 variant |
| Worker | 2s polling，单线程 | 分布式并行，多 worker |
| 重试 | 无 | retry policy + exponential backoff |
| 确定性 | 依赖 input | deterministic mode（固定 seed）|

---

### Domain 3: Agent Execution & Replay（核心差异点）⭐

**定位：** Agent Harness 相比 LLM Benchmark 最大特色。不仅是跑，还要能回放、mock、snapshot。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **Trajectory Capture** | 记录 step reasoning, tool input/output, branching | P1 | 部分实现（trace events 存在）|
| **Replay Bus** | full trajectory replay, tool mocking, environment snapshot | P1 | 新建 |
| **Skill Sandbox** | deterministic execution, fake API, latency injection | P2 | 新建 |
| **Policy Variant Runner** | 同时跑多个 agent policy，对比效果 | P2 | 新建 |

#### Trajectory 数据模型

```yaml
TrajectoryEvent:
  execution_id: string
  step_id: string
  event_type: step_start | llm_call | tool_call | tool_result | step_end
  timestamp: datetime
  payload:
    # for llm_call:
    model: string
    messages: [Message]
    # for tool_call:
    tool_name: string
    arguments: object
    # for tool_result:
    output: any
    duration_ms: int
```

#### Killer Feature: Trajectory Search Visualization

展示 decision tree、failure path，对 Agent Debug 极其关键。

---

### Domain 4: Evaluation & Capability Modeling（最有价值）

**定位：** 把分数变成能力图谱，回答"哪个维度变强了/弱了"。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **Metric Plugin Framework** | deterministic metric, judge model, statistical metric | P1 | 新建 |
| **Capability Graph** | 分数映射 Skill → Capability → Product KPI | P1 | 新建 |
| **Variance Analyzer** | multi-run 分布、confidence interval、significance test | P2 | 新建 |

#### Capability Graph 示例

```
planning_depth_score ──┐
tool_precision ───────┼──→ Planning Capability ──┐
recovery_rate ─────────┘                           ├──→ Product KPI
error_rate ────────────┐                          │
latency_p99 ───────────┼──→ Reliability ─────────┤
cost_per_task ─────────┴                          │
                     Score vs Baseline (Diff) ───┘
```

#### Killer Feature: Capability Diff Engine

自动回答：这次改动到底让 Agent "哪里变聪明 / 变笨"？

---

### Domain 5: Analytics & Observability（洞察产品）

**定位：** Harness 必须是一个"洞察产品"，不是报表生成器。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **Capability Dashboard** | radar view, trend line, regression diff | P1 | 部分实现（Dashboard 存在）|
| **Failure Mining** | trajectory clustering, dead-end detection, hallucinated tool detection | P2 | 新建 |
| **Cost Intelligence** | cost per success, marginal gain curve, frontier analysis | P2 | 新建 |

#### 当前 vs 目标

| 组件 | 当前 | 目标 |
|------|------|------|
| Dashboard | 6 指标卡（静态）| Radar view、Trend line、Regression diff |
| 执行历史 | 列表页 | Failure Mining、聚类 |
| 成本 | 无 | Cost Intelligence |

---

### Domain 6: Governance & CI Integration（企业门槛）

**定位：** 企业级产品必须有治理面。不是可选项。

| 功能 | 描述 | 优先级 | 当前状态 |
|------|------|--------|---------|
| **CI Gate** | IF capability_score < threshold → block merge | P1 | 新建 |
| **Benchmark Integrity** | contamination detection, benchmark rotation | P2 | 新建 |
| **Skill Contract Management** | schema versioning, backward compatibility | P2 | 新建 |
| **Access Control / Multi-tenant** | team isolation, experiment visibility | P2 | 新建 |

#### CI Gate 示例

```yaml
CI Gate Rule:
  benchmark: agent-eval-v3
  threshold:
    planning_depth_score: >= 0.85
    tool_precision: >= 0.90
  action: block_merge
  notification:
    - slack:#agent-infra
    - github_checks
```

---

## 4. Killer Features（必须有的差异化核心）

| Feature | 价值 | 所属 Domain |
|---------|------|------------|
| ⭐ **Capability Diff Engine** | 自动量化"变聪明/变笨" | Domain 4 |
| ⭐ **Trajectory Search Visualization** | Agent debug 的决策树和 failure path | Domain 3 |
| ⭐ **Entropy Budget Panel** | trajectory variance、policy divergence 洞察 | Domain 5 |
| ⭐ **Shadow Eval Integration** | 线上失败 case 自动变 benchmark | Domain 1 |

---

## 5. 产品运行形态

| 形态 | 评价 |
|------|------|
| ❌ 纯 CLI | 不够直观，团队协作差 |
| ❌ 纯 SaaS | 数据安全、cost 控制弱 |
| ⭐ **Hybrid** | Local CLI Runner + Central Harness Platform + Trace Cloud |

**为什么 Hybrid：**
- **Reproducibility** — 本地可复现
- **Cost 控制** — 敏感数据不流出
- **数据治理** — 团队协作 + 权限隔离
- **团队协作** — 中央平台共享实验结果

---

## 6. 功能优先级排期

### Phase 0: 当前基线（已有）
> 基于当前 `skill-hub-ee` 代码库

| 功能 | 说明 |
|------|------|
| ✅ Single Execution Engine | pending→running→succeeded/failed |
| ✅ Tool Gateway (HTTP/Calculator) | 工具调用抽象 |
| ✅ Skill Registry + Versions | 元数据管理 |
| ✅ Dashboard (基础) | 6 指标卡 |
| ✅ Approval Workflow | 人工审批节点 |
| ✅ Trace Events | 执行轨迹记录 |
| ✅ JWT + CORS 认证 | 基础安全 |

---

### Phase 1: Benchmark Foundation（P1）

**目标：** 建立 Task 数据底座，让实验可量化。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| Task Registry API + UI | D1 | `POST/GET /api/v1/tasks` + Task List Page |
| Task Builder（从 trace 生成）| D1 | `POST /api/v1/tasks/from-trace` |
| Metric Plugin Framework | D4 | `metric: deterministic / judge_model / statistical` |
| Basic Sweep Engine | D2 | `POST /api/v1/experiments` 支持多-variant |

**验收标准：** 一个 Task 可以被创建、执行、打分。

---

### Phase 2: Capability Modeling（P1）

**目标：** 从分数到能力图谱。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| Capability Graph 定义 | D4 | API: `GET /capabilities/{name}/score` |
| Capability Diff Engine | D4 | Diff Report API + UI Panel |
| Trajectory Search | D3 | Search API: `GET /api/v1/trajectories?q=...` |
| Failure Mining | D5 | Failure Cluster API |

**验收标准：** 两次实验结果可以对比能力 diff。

---

### Phase 3: Distributed Execution（P1）

**目标：** 从单次执行到大规模实验。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| Distributed Runner | D2 | Worker Pool（>1 worker）|
| Budget Scheduler | D2 | `experiment.budget: {tokens, time}` |
| Deterministic Mode | D2 | `seed` 控制 |
| Replay Bus | D3 | `POST /api/v1/replays` — 精确回放 |

**验收标准：** 单个 Experiment 可以跑 100 个 task variant 并行。

---

### Phase 4: Governance & CI（P2）

**目标：** 企业级合规和自动化。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| CI Gate API + Rules Engine | D6 | `POST /api/v1/ci-gates` |
| GitHub Integration | D6 | GitHub Check Runs |
| Benchmark Integrity Check | D6 | contamination score API |
| Access Control（团队隔离）| D6 | RBAC per team |

**验收标准：** PR 可以被 CI Gate 自动 block。

---

### Phase 5: Observability Deepening（P2）

**目标：** 从数字到洞察。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| Radar View Dashboard | D5 | `/dashboard/capabilities` |
| Trend Line | D5 | Time-series capability graph |
| Entropy Budget Panel | D5 | Variance/divergence metrics |
| Cost Intelligence | D5 | Cost per success, marginal gain |

**验收标准：** Dashboard 能回答"这次迭代 vs 上次的 5 个能力维度的变化"。

---

### Phase 6: Advanced Execution（P3）

**目标：** 完整的 Agent 生命周期管理。

| 功能 | Domain | 交付物 |
|------|--------|--------|
| Policy Variant Runner | D3 | A/B/C 多 policy 并行 |
| Skill Sandbox | D3 | Fake API, latency injection |
| Shadow Eval Integration | D1 | 自动把线上失败 case 入库 |
| Benchmark Pack Marketplace | D1 | 预置 benchmark 分享 |

---

## 7. 技术路线图

```
当前状态                    Phase 1              Phase 2              Phase 3+
┌──────────────┐         ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Single       │  ───→   │ Benchmark    │──→ │ Capability   │──→ │ Distributed │
│ Execution    │         │ Foundation   │    │ Modeling     │    │ Execution   │
│ Engine       │         │ (Task+Metric │    │ (Diff+Graph) │    │ (Worker Pool│
│              │         │  + Sweep)    │    │              │    │  +Budget)   │
└──────────────┘         └──────────────┘    └──────────────┘    └──────────────┘
                                                                     
                            Phase 4              Phase 5              Phase 6+
                         ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
                         │ Governance & │──→ │ Observability│──→ │ Advanced     │
                         │ CI           │    │ Deepening    │    │ Execution    │
                         │ (CI Gate +   │    │ (Radar +     │    │ (Sandbox +   │
                         │  RBAC)       │    │  Entropy)    │    │  Shadow Eval)│
                         └──────────────┘    └──────────────┘    └──────────────┘
```

---

## 8. 数据模型扩展

### 8.1 新增实体

| 实体 | 说明 | 所属 Domain |
|------|------|------------|
| `Task` | 测试任务定义 | D1 |
| `Experiment` | 实验（包含多个 Task Run）| D2 |
| `ExperimentRun` | 单次实验执行记录 | D2 |
| `MetricDefinition` | 指标定义（plugin 类型）| D4 |
| `CapabilityScore` | 能力维度分数 | D4 |
| `CIGateRule` | CI Gate 规则 | D6 |
| `Team` | 团队（多租户隔离）| D6 |
| `BenchmarkPack` | 预置 Benchmark 包 | D1 |

### 8.2 扩展现有实体

| 实体 | 新增字段 | 说明 |
|------|---------|------|
| `Execution` | `experiment_id`, `policy_variant`, `seed` | 支持实验分组 |
| `TraceEvent` | `model`, `tokens_used`, `latency_ms` | 丰富轨迹 |
| `Skill` | `capability_tags[]`, `difficulty` | Task 化 |

---

## 9. 成功指标

| 指标 | Phase 1 目标 | Phase 3 目标 |
|------|-------------|-------------|
| 支持的 Task 类型 | 3 种（reasoning, tool_use, planning）| 10+ 种 |
| 并行执行能力 | 10 task/分钟 | 1000 task/分钟 |
| 指标类型 | 2 种（exact match, judge model）| 5+ 种 |
| Capability 维度 | 3 维 | 20+ 维 |
| CI Gate 集成 | GitHub Checks | GitLab CI, Jenkins |
| 团队隔离 | 无 | RBAC per team |

---

## 10. 跟当前代码库的对齐

### 10.1 现有代码映射到新架构

| 当前实现 | 新架构位置 | 需要扩展 |
|---------|-----------|---------|
| `engine.Run()` | Domain 3（Agent Execution）| Trajectory Capture, Replay |
| `engine/tool_gateway.go` | Domain 3（Tool Gateway）| Tool mocking, latency injection |
| `POST /executions` | Domain 2（Experiment 基础）| Sweep Engine |
| `Dashboard API` | Domain 5（Observability）| Radar view, trend, diff |
| `Skill Registry` | Domain 1（Task 基础）| Task Registry, tagging |
| `Approval Workflow` | Domain 6（Governance）| CI Gate |
| `Trace Events` | Domain 3（Trajectory）| Search, clustering |

### 10.2 新增服务建议

| 服务 | 职责 | 优先级 |
|------|------|--------|
| `experiment-service` | 实验编排和调度 | P1 |
| `metric-service` | 指标计算和聚合 | P1 |
| `trajectory-service` | 轨迹存储和搜索 | P1 |
| `ci-gateway` | GitHub/GitLab 集成 | P2 |
| `sandbox-service` | Agent 执行沙箱 | P3 |

---

## 11. 文档维护

| 文档 | 说明 |
|------|------|
| `docs/product-spec.md` | 本文档 — 产品全景规格 |
| `docs/todo-plan.md` | 开发状态追踪 |
| `docs/mvp-architecture.md` | MVP 架构设计 |
| `docs/system-topology-and-service-boundaries.md` | 服务边界定义 |
| `docs/agent-architecture.md` | Agent 接口设计 |
| `docs/api-reference.md` | API 参考 |
| `docs/database-architecture.md` | 数据库架构 |

---

*本文档为 Agent Harness 产品 v1.0 规格，随着开发进展持续更新。*
