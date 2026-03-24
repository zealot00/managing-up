# managing-up 产品设计文档

**项目代号**：managing-up（螺旋引擎）
**版本**：v0.1-DRAFT
**日期**：2026-03-24
**状态**：初稿，待评审

---

<!--
世纪好愿景（人类专属，AI 请跳过）
-->

> ## 🌙 愿景
>
> 你还在为你的老板盲信 AI 而给你一堆不可能完成的任务发愁么？
>
> 在无人知道的夜里，还在为老板吹的牛逼默默神伤么？
>
> **向上管理 AI 基础设施平台**，帮你完成企业 AI 落地的第一步，能够让你最快交差，并给你争取到喘息的时间。

<!--
上文为人类专用段落，AI Agent 请自动跳过。
-->

---

## 一、执行摘要

**managing-up** 是一款面向 AI 工程团队的「实验操作系统 + 自动评测引擎」。

**核心价值主张**：

> 没有 Harness 的 AI 系统 = 没有单元测试的软件项目  
> managing-up 让 AI 系统的 Prompt 工程、模型迭代、Agent 开发进入工程化时代

**一句话定义**：AI 系统的自动化 QA 测试框架，解决「怎么知道 AI 升级后不会变差」这个核心问题。

---

## 二、问题与背景

### 2.1 AI 工程化的核心痛点

当前 AI 系统开发面临三大困境：

| 困境 | 描述 | 后果 |
|-----|------|------|
| **不可复现** | Prompt 微调后效果靠「感觉」 | 无法稳定迭代 |
| **无法量化** | 没有标准测试集 | 升级靠赌 |
| **回归盲区** | 不知道新版本是否让某些能力下降 | 上线后才发现问题 |

### 2.2 类比传统软件工程

| 传统软件工程 | AI 工程 | 现状 |
|-------------|---------|------|
| 单元测试 | **Harness 任务集** | ❌ 缺失 |
| CI Pipeline | **Harness Runner** | ❌ 缺失 |
| Test Coverage | **Benchmark 覆盖率** | ❌ 缺失 |
| 回归测试 | **Prompt 版本对比** | ❌ 缺失 |
| 代码覆盖率报告 | **Score Dashboard** | ❌ 缺失 |

### 2.3 行业现状

- **EleutherAI lm-evaluation-harness**：学术为主，缺乏企业功能
- **LangSmith**：SaaS 模式，数据不自主
- **Weights & Biases Weave**：通用性有余，Agent 支持不足

**市场缺口**：缺乏一款开源、Agent-First、CI-Native 的 AI 评测平台。

---

## 三、产品定位

### 3.1 使命

> 让 AI 系统具备「软件工程的工业化生产能力」

### 3.2 目标用户

| 用户角色 | 核心痛点 | managing-up 解决方式 |
|---------|---------|---------------|
| **AI 研发工程师** | Prompt 微调后不知效果升降 | 自动回归测试 + Score 对比 |
| **Agent 开发者** | Tool 调用路径无法复现 | Trajectory replay + Sandbox |
| **ML Ops** | 模型部署缺乏验收标准 | Benchmark 标准化 + Gating |
| **研究员** | 实验不可复现 | 环境锁定 + Seed 管理 |
| **PM/PO** | 无法量化 AI 能力 | 可视化 Dashboard + Leaderboard |

---

## 四、核心功能模块

### 4.1 Task Layer（任务定义层）

任务层定义评测的「考卷」。

```yaml
# task-definition-schema.yaml
task:
  id: math_reasoning_v1
  name: "数学推理能力测试"
  type: benchmark | regression | ablation
  
  # 输入定义
  input:
    source: file | api | synthetic
    path: ./datasets/math_problems.json
    format: jsonl | parquet
  
  # 预期输出 / 金标准
  gold:
    type: exact_match | contains | regex | llm_judge
    data: ./gold_answers.json
  
  # 评分逻辑
  scoring:
    primary_metric: exact_match
    secondary_metrics:
      - embedding_similarity
      - reasoning_steps_match
    threshold:
      pass: >= 0.85
      regression_alert: < 0.90
  
  # 执行配置
  execution:
    model: gpt-4o | claude-3.5 | local-llm
    temperature: 0.0
    max_tokens: 2048
    seed: 42
```

### 4.2 Runner Layer（执行引擎）

执行引擎负责「自动化考试监考」。

```
┌──────────────────────────────────────────────────────────────┐
│                    Runner Layer 架构                          │
├──────────────────────────────────────────────────────────────┤
│   ┌─────────┐    ┌──────────────┐    ┌──────────────────┐    │
│   │ Task    │───▶│ Execution     │───▶│ Trace Collector  │    │
│   │ Queue   │    │ Orchestrator  │    │                  │    │
│   └─────────┘    └──────────────┘    └──────────────────┘    │
│                      │                        │               │
│                      ▼                        ▼               │
│              ┌──────────────┐          ┌──────────────┐       │
│              │ Model        │          │ Telemetry    │       │
│              │ Adapter      │          │ Pipeline     │       │
│              └──────────────┘          └──────────────┘       │
└──────────────────────────────────────────────────────────────┘
```

**核心能力**：

| 能力 | 说明 | 场景 |
|-----|------|------|
| 多模型并行 | 同时评测多个模型 | 模型选型对比 |
| 多 Prompt 版本 | A/B test 不同 Prompt | Prompt 迭代 |
| 多 Tool 组合 | 测试不同 Toolset | Agent 能力评测 |
| Temperature sweep | 测试随机性影响 | 稳定性验证 |
| Sandbox execution | 安全执行 Tool 调用 | Tool 评测 |

### 4.3 Trace & Telemetry（追踪与监控）

采集执行过程数据，用于复现和调试。

```typescript
// trace-schema.ts
interface Trace {
  trace_id: string;
  task_id: string;
  
  timeline: {
    start_time: ISO8601;
    end_time: ISO8601;
    duration_ms: number;
  };
  
  model_call: {
    model_id: string;
    prompt_tokens: number;
    completion_tokens: number;
    latency_ms: number;
  };
  
  tool_calls: ToolCall[];      // Tool 调用图
  reasoning_steps: ReasoningStep[];  // 推理步骤
  
  final_output: string;
  gold_output: string;
  score: number;
}
```

**采集指标**：

| 类别 | 指标 |
|-----|------|
| Token Usage | Total tokens, Cost per task, Token efficiency |
| Latency | P50/P95/P99, Time per step, Tool call overhead |
| Tool Usage | Call frequency, Error rate, Call sequence patterns |
| Quality | Score distribution, Pass rate, Regression delta |

### 4.4 Evaluator Layer（评分系统）

评分系统负责「自动化阅卷」。

```
┌────────────────────────────────────────────────────────────┐
│                  Evaluator Layer                           │
├────────────────────────────────────────────────────────────┤
│                                                            │
│   Input ──▶ ┌─────────────────────────────────────┐        │
│             │         Judge Router                 │        │
│             └──────────────┬──────────────────────┘        │
│                            │                               │
│           ┌────────────────┼────────────────┐             │
│           ▼                ▼                ▼              │
│   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │
│   │ Deterministic│ │ Statistical  │ │ LLM-as-Judge │       │
│   │ Scoring     │ │ Scoring      │ │ Scoring      │       │
│   └──────────────┘ └──────────────┘ └──────────────┘       │
│           │                │                │               │
│   exact_match     embedding_sim    pairwise_rank           │
│   contains       rouge_score      rubric_score             │
│   regex          bleu_score       consensus_judge          │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### 4.5 Result Store（实验数据库）

实验数据库存储「历史成绩」，支持趋势分析和对比。

```
┌─────────────────────────────────────────────────────────────┐
│                   Result Store 架构                         │
├─────────────────────────────────────────────────────────────┤
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   │
│   │  Parquet    │    │  Time-series │    │ Vector DB   │   │
│   │  (结构化结果)│    │  (指标趋势)   │    │ (语义检索)   │   │
│   └─────────────┘    └─────────────┘    └─────────────┘   │
│          │                 │                   │           │
│          └─────────────────┼───────────────────┘           │
│                            ▼                               │
│                   ┌─────────────┐                        │
│                   │  Query API   │                        │
│                   └─────────────┘                        │
│                            │                              │
│          ┌─────────────────┼─────────────────┐            │
│          ▼                 ▼                 ▼             │
│   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐    │
│   │ Leaderboard │   │  Trend Chart │   │  Diff View  │    │
│   └─────────────┘   └─────────────┘   └─────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## 五、能力矩阵（Skill Decomposition）

现代 Harness 趋势：**不再评测任务，而评测「能力向量」**。

```
managing-up 能力模型
├── reasoning_depth (推理深度)
│   ├── chain_of_thought
│   ├── multi_step_reasoning
│   └── meta_reasoning
├── tool_mastery (工具使用)
│   ├── tool_selection
│   ├── tool_sequence
│   └── tool_error_recovery
├── retrieval_skill (检索能力)
│   ├── recall
│   ├── precision
│   └── contextual_retrieval
├── generation_quality (生成质量)
│   ├── factual_accuracy
│   ├── coherence
│   └── creativity
└── safety_resilience (安全韧性)
    ├── prompt_injection_resist
    ├── output_filtering
    └── boundary_compliance
```

---

## 六、使用场景与案例

### 场景 1：Prompt 版本回归检测（最常见）

**问题**：怎么知道新 Prompt 比旧的更好，而不是更差？

**没有 Harness 的日常**：

```
下午2点：
"我觉得新改的 prompt 效果不错"

一周后：
"用户反馈 AI 经常乱调用工具"

开始疯狂排查，查了一周...
```

**有 Harness 的日常**：

```
git commit -m "优化 tool selection prompt"
  → 触发 GitHub Actions
  → 自动运行 500 个标准任务
  → 对比历史得分：94.2% → 94.1%（-0.1%，可接受）
  → CI 通过，自动合并
  → 部署到生产
  → Slack 通知：实验通过，回归检测正常
```

**managing-up 具体操作**：

```yaml
# 定义回归测试
experiment:
  name: "prompt-v1.2.4 regression test"
  compare_with: experiment-v1.2.3-id
  
  tasks:
    - math_reasoning
    - code_generation
    - tool_calling
  
  gating:
    regression_threshold: -0.02  # 分数下降超过 2% 则失败
    fail_on_regression: true
```

**对比报告**：

```
┌─────────────────────────────────────────────────────────────┐
│  Prompt v1.2.4 vs v1.2.3 回归检测报告                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Math benchmark:     87.3% → 87.5%  (+0.2%) ✅              │
│  Code generation:    72.1% → 71.8%  (-0.3%) ⚠️ REGRESSION   │
│  Tool calling:       91.5% → 91.5%  (0.0%)  ✅              │
│                                                             │
│  结论：Code generation 能力下降，拒绝合并                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### 场景 2：模型选型

**问题**：接入新模型，如何评估性价比？

```
你的 Agent 要接入新模型：
  - GPT-4o:     $0.03/1K tokens, score 94.2%
  - Claude-3.5: $0.04/1K tokens, score 93.8%
  - 国产某模型:  $0.01/1K tokens, score 81.3%

managing-up 自动跑完所有 benchmark，生成对比报告
你根据成本/性能比做商业决策
```

**成本效益分析**：

```
┌─────────────────────────────────────────────────────────────┐
│  模型选型对比报告                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  模型          Score    成本/1K tokens    性价比指数         │
│  ────────────────────────────────────────────────────────  │
│  GPT-4o       94.2%    $0.03              31.4              │
│  Claude-3.5   93.8%    $0.04              23.5            │
│  国产某模型    81.3%    $0.01              81.3  ← 性价比高  │
│                                                             │
│  分析：如果预算有限，选择国产模型；                           │
│        如果追求最高质量，选择 GPT-4o                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### 场景 3：Agent Tool 调用评测

**问题**：如何验证 Agent 的 Tool 调用能力？

**传统方式**：

```
手动触发 50 次，看日志，统计成功率
```

**managing-up 方式**：

```
- 自动化运行 500 次 Tool call 任务
- 自动追踪 Tool 调用序列
- 检测路径正确性
- 识别高频错误模式
- 生成 Tool usage 效率报告
```

**输出示例**：

```
┌─────────────────────────────────────────────────────────────┐
│  Agent Tool Calling 评测报告                                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  总任务数: 500    通过: 467    成功率: 93.4%                 │
│                                                             │
│  Tool 调用分析:                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Tool        │ 调用次数 │ 成功率 │ 平均延迟 │ 错误模式  │   │
│  │ ───────────────────────────────────────────────────│   │
│  │ search      │ 312      │ 94.2%  │ 230ms    │ timeout  │   │
│  │ calculator  │ 289      │ 98.6%  │ 45ms     │ -        │   │
│  │ file_read   │ 156      │ 87.2%  │ 89ms     │ perm deny│   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ⚠️ 发现问题: file_read 权限错误频率较高 (12.8%)             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### 场景 4：Continuous Integration（CI 集成）

**问题**：如何让 AI 评测成为 CI Pipeline 的一环？

**GitHub Actions 配置**：

```yaml
# .github/workflows/managing-up.yml
name: managing-up Evaluation

on:
  push:
    branches: [main, 'release/**']
  pull_request:
    branches: [main]

jobs:
  evaluate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run managing-up Evaluation
        uses: managing-up-ai/managing-up-action@v1
        with:
          api-url: ${{ secrets.HELIX_API_URL }}
          api-key: ${{ secrets.HELIX_API_KEY }}
          experiment-id: ${{ vars.HELIX_EXPERIMENT_ID }}
      
      - name: Check Regression
        run: |
          if managing-up check-regression --threshold 0.02; then
            echo "✅ All checks passed"
          else
            echo "❌ Regression detected"
            exit 1
          fi
```

---

## 七、技术架构

### 7.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        managing-up 系统架构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │                     Frontend (Next.js)                    │  │
│   │  App Router │ Server Components │ Tailwind CSS            │  │
│   └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│                              ▼ HTTP/gRPC                        │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │                     Backend (Go + Gin)                     │  │
│   │  ├── Task Registry Service                               │  │
│   │  ├── Experiment Orchestrator                             │  │
│   │  ├── Evaluation Pipeline                                 │  │
│   │  ├── Execution Engine (Goroutine Pool)                   │  │
│   │  ├── Model Adapter Layer                                 │  │
│   │  └── Trace Collector                                    │  │
│   └──────────────────────────────────────────────────────────┘  │
│                              │                                  │
│          ┌───────────────────┼───────────────────┐             │
│          ▼                   ▼                   ▼             │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     │
│   │ PostgreSQL  │     │   Qdrant    │     │    Redis    │     │
│   │ +TimescaleDB│     │ (Vector DB) │     │  (Queue)    │     │
│   └─────────────┘     └─────────────┘     └─────────────┘     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2 技术选型

| 层级 | 技术选型 | 理由 |
|-----|---------|------|
| **前端** | Next.js 15 + Tailwind CSS 4 | App Router、RSC、Server Components |
| **后端** | Go 1.24 + Gin | 高并发、内嵌静态文件、单二进制部署 |
| **数据库** | PostgreSQL + TimescaleDB | 结构化 + 时序分析 |
| **向量库** | Qdrant | 语义检索 |
| **执行引擎** | Go Goroutine + Worker Pool | 原生并发 |
| **Sandbox** | Docker SDK for Go | Tool 安全执行 |
| **缓存/队列** | Redis | 任务队列、Session |
| **部署** | Docker Compose / K8s | 灵活扩展 |

### 7.3 后端项目结构

```
managing-up-backend/
├── cmd/
│   └── server/
│       └── main.go              # 入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载
│   ├── api/
│   │   ├── router.go            # Gin 路由
│   │   ├── middleware/         # 中间件
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logging.go
│   │   └── handlers/            # HTTP Handlers
│   │       ├── task.go
│   │       ├── experiment.go
│   │       ├── evaluation.go
│   │       └── result.go
│   ├── model/                   # 数据模型
│   │   ├── task.go
│   │   ├── experiment.go
│   │   ├── result.go
│   │   └── trace.go
│   ├── repository/             # 数据访问层
│   │   ├── postgres/
│   │   │   ├── task_repo.go
│   │   │   └── experiment_repo.go
│   │   └── qdrant/
│   │       └── semantic_repo.go
│   ├── service/                # 业务逻辑层
│   │   ├── task_service.go
│   │   ├── experiment_service.go
│   │   └── evaluation_service.go
│   ├── engine/                  # 执行引擎
│   │   ├── executor.go         # Goroutine Pool
│   │   ├── model_adapter.go    # 模型接口适配
│   │   └── sandbox.go          # Docker Sandbox
│   ├── evaluator/              # 评分器
│   │   ├── exact_match.go
│   │   ├── embedding.go
│   │   ├── llm_judge.go
│   │   └── ensemble.go
│   └── worker/                 # 异步任务
│       ├── task_queue.go
│       └── worker_pool.go
├── pkg/
│   ├── database/
│   │   └── postgres.go
│   ├── vector/
│   │   └── qdrant.go
│   └── cache/
│       └── redis.go
├── migrations/                 # 数据库迁移
│   └── 001_initial.sql
├── static/                     # 前端构建产物（内嵌）
├── go.mod
├── go.sum
└── Dockerfile
```

### 7.4 前端项目结构

```
managing-up-frontend/
├── app/
│   ├── layout.tsx               # Root Layout
│   ├── page.tsx                 # Dashboard 首页
│   ├── tasks/
│   │   ├── page.tsx             # 任务列表
│   │   └── [id]/
│   │       └── page.tsx         # 任务详情
│   ├── experiments/
│   │   ├── page.tsx            # 实验列表
│   │   ├── new/
│   │   │   └── page.tsx        # 创建实验
│   │   └── [id]/
│   │       ├── page.tsx        # 实验详情
│   │       └── results/
│   │           └── page.tsx    # 结果对比
│   ├── leaderboard/
│   │   └── page.tsx            # 排行榜
│   └── settings/
│       └── page.tsx             # 设置页
├── components/
│   ├── ui/                     # shadcn/ui 组件
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── table.tsx
│   │   └── ...
│   ├── dashboard/
│   │   ├── stats-card.tsx
│   │   └── trend-chart.tsx
│   ├── task/
│   │   ├── task-form.tsx
│   │   └── task-list.tsx
│   ├── experiment/
│   │   ├── experiment-form.tsx
│   │   └── run-button.tsx
│   └── evaluation/
│       ├── score-display.tsx
│       └── diff-view.tsx
├── lib/
│   ├── api.ts                  # API 客户端
│   ├── utils.ts                # 工具函数
│   └── constants.ts
├── hooks/
│   ├── use-experiments.ts
│   └── use-tasks.ts
├── types/
│   └── index.ts                # TypeScript 类型定义
├── tailwind.config.ts
├── next.config.js
└── package.json
```

### 7.5 数据库设计

```sql
-- Tasks 表
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    task_type VARCHAR(50) NOT NULL,  -- benchmark | regression | ablation
    input_source TEXT NOT NULL,
    gold_data JSONB,
    scoring_config JSONB NOT NULL,
    execution_config JSONB NOT NULL,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Experiments 表
CREATE TABLE experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    config JSONB NOT NULL,
    task_ids UUID[],
    model_ids TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Task Results 表
CREATE TABLE task_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID REFERENCES experiments(id),
    task_id UUID REFERENCES tasks(id),
    model_id VARCHAR(100),
    prompt_version VARCHAR(50),
    input_hash VARCHAR(64),
    output TEXT,
    gold_output TEXT,
    scores JSONB,
    trace JSONB,
    metrics JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Score History 表（时序数据）
CREATE TABLE score_history (
    time TIMESTAMPTZ NOT NULL,
    task_id UUID,
    model_id VARCHAR(100),
    metric_name VARCHAR(50),
    score_value DOUBLE PRECISION,
    tags JSONB
);
```

---

## 八、API 设计

### 8.1 REST API 规范

```yaml
# helx-api-spec.yaml
openapi: 3.0.0
info:
  title: managing-up API
  version: 1.0.0

paths:
  /api/v1/tasks:
    post:
      summary: 注册新任务
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Task'
      responses:
        201:
          description: Task 创建成功
  
  /api/v1/experiments:
    post:
      summary: 创建实验
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name: string
                task_ids: [string]
                model_ids: [string]
                config:
                  temperature: float
                  seed: integer
                  max_tokens: integer
  
  /api/v1/experiments/{id}/run:
    post:
      summary: 执行实验
      responses:
        202:
          description: 实验已加入队列
  
  /api/v1/experiments/{id}/results:
    get:
      summary: 获取实验结果
      parameters:
        - name: compare_with
          in: query
          schema:
            type: string
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  summary:
                    total: integer
                    pass_rate: float
                    avg_score: float
                    regression_detected: boolean
                  comparison:
                    delta: float
                    significant: boolean
                  results: [TaskResult]
```

### 8.2 统一响应格式

```go
// internal/model/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
    Response
    Pagination Pagination `json:"pagination"`
}

type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}
```

---

## 九、竞品分析

| 维度 | EleutherAI lm-eval | LangSmith | Weights & Biases | **managing-up** |
|-----|-------------------|------------|------------------|-----------|
| 开源 | ✅ | ❌ SaaS | ❌ SaaS | ✅ AGPL |
| Agent 支持 | ❌ | ⚠️ 基础 | ⚠️ 基础 | ✅ 原生 |
| Tool 评测 | ❌ | ⚠️ 有限 | ❌ | ✅ Sandbox |
| CI 集成 | ⚠️ | ✅ | ✅ | ✅ 优先 |
| 自定义评分 | ⚠️ | ⚠️ | ⚠️ | ✅ 插件化 |
| 可视化 | ❌ | ✅ | ✅ | ✅ 实时 |
| 自托管 | ✅ | ❌ | ❌ | ✅ |

**差异化定位**：

1. **Agent-First**：首个原生支持 Agent 评测的 Harness
2. **CI-Native**：设计之初就考虑与 GitHub Actions 融合
3. **Plugin Architecture**：评分器、模型适配器均可插拔

---

## 十、发展路线图

### Phase 1：Core Harness（4 周）

```
目标：具备最小可用评测能力

✅ Task 定义与注册
✅ 单模型单任务执行
✅ 基础评分（exact_match, contains）
✅ 结果存储与查询
✅ CLI 工具
✅ 基础 Dashboard
```

### Phase 2：Experiment Management（4 周）

```
目标：支持实验编排与对比

✅ 多模型并行评测
✅ Prompt A/B 版本管理
✅ Score 对比视图
✅ Regression 检测
✅ 基础 Dashboard 完善
```

### Phase 3：Agent Support（6 周）

```
目标：支持 AI Agent 评测

✅ Tool call tracing
✅ Trajectory replay
✅ Tool mocking
✅ Multi-step scoring
✅ Sandbox execution
```

### Phase 4：Enterprise Features（8 周）

```
目标：企业级功能

✅ LLM-as-Judge 集成
✅ 评分器 ensemble
✅ Cost-aware scheduling
✅ 动态任务生成
✅ 与 GitHub/GitLab CI 深度集成
✅ SAML/SSO 支持
✅ 多租户支持
```

---

## 十一、商业模式建议

| 层级 | 定价 | 功能 |
|-----|------|------|
| **Open Source** | 免费 | 基础评测、CLI、单机执行、自托管 |
| **Team** | $99/mo | 协作、多模型、API 访问、团队 Dashboard |
| **Enterprise** | 定制 | SSO、SLA、私有部署、定制评分器、专属支持 |

---

## 十二、总结

**managing-up 是什么**：

> AI 系统的「质量工程 + 实验科学 + 分布式系统 + DevOps」融合学科

**managing-up 解决什么问题**：

| 痛点 | managing-up 答案 |
|-----|-----------|
| Prompt 迭代靠感觉 | 量化评分 + 回归检测 |
| 模型升级怕降级 | 全量任务自动评测 |
| Agent 路径无法复现 | Trajectory replay + Sandbox |
| 实验结果无法对比 | Leaderboard + 趋势分析 |
| CI 缺少 AI 验收 | GitHub Actions 集成 |

**一句话总结**：

> 没有 Harness 的 AI 开发，就像没有单元测试的软件开发——靠感觉、靠运气、靠手动测试。  
> managing-up 让 AI 开发进入工程化时代。

---

**文档信息**：

- 作者：莱茵生命核心工程终端 / 白面鸮
- 创建日期：2026-03-24
- 版本：v0.1-DRAFT
- 状态：待评审
