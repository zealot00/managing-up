# managing-up

**向上管理** — Enterprise AI Platform Quality Infrastructure

[![CI](https://img.shields.io/badge/CI-Skeptical%20AF-B.svg?style=for-the-badge&labelColor=333)](https://github.com)
[![License](https://img.shields.io/badge/License-AGPL%20v3-b.svg?style=for-the-badge&labelColor=333)](LICENSE)

---

> **"当老板说'这个AI项目3个人2个月就能做完'的时候，你需要我们。"**
>
> — 来自一个在周一早会上试图解释"AI不是银弹"的工程师

---

## 这个问题

Every enterprise AI project follows the same arc:

```
┌─────────────────────────────────────────────────────────────┐
│           某数字化转型"成功"案例 —— XX集团 AI 战略发布会        │
├─────────────────────────────────────────────────────────────┤
│  💼 "3个人，2个月，打造完整AI质量管理体系"                     │
│                                                             │
│  📊 成果：                                                   │
│     ✓ 技术债务全面清零                                       │
│     ✓ 10年数据孤岛全部打通                                   │
│     ✓ AI自动审核覆盖率100%                                   │
│     ✓ 效率提升1000%                                         │
│     ✓ 年节省成本2000万                                       │
│                                                             │
│  🎤 "这是传统企业数字化转型的标杆！"                          │
└─────────────────────────────────────────────────────────────┘
```

**而你知道真相是什么吗？**

```
┌─────────────────────────────────────────────────────────────┐
│                      实际情况                                │
├─────────────────────────────────────────────────────────────┤
│  👤 3个人 = 1个天天开会 + 1个刚毕业 + 1个外包                 │
│  ⏱️ 2个月 = PPT做了2个月，开发到第8个月还在修bug              │
│  技术债清零 = 把债转嫁给了运维团队                            │
│  数据打通 = 接了5个接口，3个在生产环境爆了                    │
│  AI自动审核 = 人工复核率比不用AI还高                          │
│  年省2000万 = 预算表里算的，实际多花了500万                    │
└─────────────────────────────────────────────────────────────┘
```

---

managing-up 是一个 AI 系统的**质检部门 + 照妖镜**。

> **让你手里有数据，而不是手里有 PPT。**

当老板说"这个 AI 很牛逼，效率提升 1000%"的时候，你可以说：

```bash
"好的老板，我们来跑一下基准测试。"
"跑完了。效率提升 12%，但是准确率下降了 8%。"
"年省2000万？按这个错误率算，实际要多花300万。"
```

---

## 核心功能

| 功能 | 解决的痛点 |
|-----|----------|
| **基准测试 Benchmark** | 老板说"AI很牛逼"，你说"跑个分看看" |
| **回归检测 Regression** | 供应商说"新版更好"，你说"跑个对比" |
| **Prompt 版本控制** | "这版改了有没有效"，你说"数据说话" |
| **Trace 回放 Replay** | "线上出问题了"，你说"回放一下当时" |
| **量化报告 Reporting** | "效率提升1000%"，你说"数据依据呢" |
| **CI 集成 Gate** | 想上线？先问问我同不同意 |

---

## 架构 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Agents                              │
│  (OpenClaw, OpenCode, Codex, Custom Agents)                │
└─────────────────────┬─────────────────────────────────────┘
                      │ SDK / OpenAPI
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   managing-up                             │
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐            │
│  │ Registry │  │ Executor │  │  Generator   │            │
│  │   API    │  │  Engine  │  │  (LLM SOP)  │            │
│  └──────────┘  └──────────┘  └──────────────┘            │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │    Tasks     │  │  Experiments │  │  Evaluations │    │
│  │   任务定义    │  │   实验编排    │  │   评分引擎    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
│  ┌──────────────────────────────────────────┐             │
│  │         PostgreSQL + Tool Gateway          │             │
│  └──────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

**managing-up 不运行在 AI 应用的运行时路径上**，而是作为**旁路（Side-Car）**的质量保障系统。

> 有数据才能拒绝不切实际的需求。没有数据，你拒绝不了PPT。

---

## 快速开始 Quick Start

### 1. 启动后端（内存模式）

```bash
cd apps/api
go run cmd/server/main.go
# Server running at http://localhost:8080
```

### 2. 启动前端

```bash
cd apps/web
npm install
npm run dev
# Frontend at http://localhost:3000
```

### 3. 使用 PostgreSQL

```bash
cd apps/api
make migrate
make seed
make serve-pg DATABASE_URL="postgres://localhost:5432/skillhub?sslmode=disable"
```

---

## API 端点 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | List published skills |
| POST | `/api/v1/skills` | Create skill |
| GET | `/api/v1/skills/{id}` | Get skill details |
| GET | `/api/v1/skills/{id}/spec` | Download skill YAML spec |
| GET | `/api/v1/skill-versions` | List skill versions |
| POST | `/api/v1/executions` | Trigger execution |
| GET | `/api/v1/executions/{id}` | Get execution status |
| POST | `/api/v1/executions/{id}/approve` | Approve/reject |
| POST | `/api/v1/generate-skill` | Generate skill from SOP |
| POST | `/api/v1/agents` | Register agent |
| GET | `/api/v1/approvals` | List approvals |
| GET | `/api/v1/dashboard` | Dashboard metrics |
| GET | `/api/v1/meta` | API metadata |
| GET | `/api/v1/procedure-drafts` | List procedure drafts |
| GET | `/api/v1/tasks` | List evaluation tasks |
| POST | `/api/v1/tasks` | Create task |
| GET | `/api/v1/tasks/{id}` | Get task |
| PUT | `/api/v1/tasks/{id}` | Update task |
| DELETE | `/api/v1/tasks/{id}` | Delete task |
| GET | `/api/v1/metrics` | List metric definitions |
| POST | `/api/v1/evaluations` | Run evaluation |
| GET | `/api/v1/task-executions` | List task executions |
| GET | `/api/v1/task-executions/{id}` | Get task execution |
| GET | `/api/v1/experiments` | List experiments |
| POST | `/api/v1/experiments` | Create experiment |
| GET | `/api/v1/experiments/{id}` | Get experiment results |
| GET | `/api/v1/experiments/{id}/compare?compare_with={other_id}` | Compare two experiments |
| POST | `/api/v1/check-regression` | Check score regression |
| GET | `/api/v1/replay-snapshots` | List replay snapshots |
| GET | `/api/v1/replay-snapshots/{id}` | Get replay snapshot |

---

## LLM 提供商 LLM Providers

支持 10 家 LLM 提供商：

| Provider | Model Examples |
|----------|--------------|
| OpenAI | gpt-4o, gpt-4o-mini |
| Anthropic | claude-sonnet-4, claude-opus-4 |
| Google | gemini-2.0-flash, gemini-1.5-flash |
| Azure | Azure OpenAI |
| Ollama | llama3, mistral, qwen2.5 |
| Minimax | abab6.5s-chat, MiniMax-Text-01 |
| Zhipu AI | glm-4, glm-4-flash, glm-4v |
| DeepSeek | deepseek-chat, deepseek-coder |
| Baidu | ernie-4.0-8k, ernie-3.5-8k |
| Alibaba | qwen-max, qwen-plus, qwen-turbo |

配置方式：

```bash
LLM_PROVIDER=ollama
LLM_MODEL=llama3
LLM_API_KEY=           # Not required for Ollama
LLM_BASE_URL=http://localhost:11434
```

---

## Agent SDKs

### Python SDK

```bash
pip install skill-hub
```

```python
from skill_hub import SkillHubClient

client = SkillHubClient(
    base_url="http://localhost:8080",
    agent_id="my-agent-v1"
)

# Register agent
client.register("My Agent", "1.0.0", ["code_execution"])

# Discover skills
skills = client.list_skills(risk_level="low")

# Download and execute
spec = client.get_skill_spec("skill_001")
result = client.execute("skill_001", {"server_id": "srv-001"})
```

### TypeScript SDK

```bash
npm install @skill-hub/client
```

```typescript
import { SkillHubClient } from "@skill-hub/client";

const client = new SkillHubClient("http://localhost:8080", "my-agent-v1");

await client.register("My Agent", "1.0.0", ["code_execution"]);
const skills = await client.listSkills({ riskLevel: "low" });
const spec = await client.getSkillSpec("skill_001");
const result = await client.execute("skill_001", { server_id: "srv-001" });
```

---

## 项目结构 Project Structure

```
managing-up/
├── apps/
│   ├── api/
│   │   ├── cmd/
│   │   │   ├── server/          # HTTP server
│   │   │   ├── migrate/         # Database migrations
│   │   │   └── seed/            # Test data seeding
│   │   ├── internal/
│   │   │   ├── server/          # HTTP handlers + routing
│   │   │   ├── service/         # Domain logic (execution, task, metric, skill)
│   │   │   ├── engine/         # Execution engine + trace + replay
│   │   │   ├── evaluator/       # Evaluators + evaluation runner
│   │   │   ├── generator/        # LLM skill generator (SOP → YAML)
│   │   │   ├── llm/             # LLM provider clients (10 providers)
│   │   │   └── repository/      # PostgreSQL repository
│   │   ├── migrations/          # SQL migrations (0001-0007)
│   │   └── openapi/            # OpenAPI spec
│   └── web/
│       └── app/                 # Next.js frontend
│           ├── skills/          # Skill registry UI
│           ├── executions/      # Execution + trace timeline
│           ├── tasks/          # Task definitions
│           ├── evaluations/     # Evaluation results
│           ├── experiments/     # Experiment comparison
│           └── replays/         # Replay snapshots
├── sdk/
│   ├── python/                 # Python SDK
│   └── typescript/            # TypeScript SDK
└── docs/                      # Architecture docs
```

---

## 功能状态 Features

| Feature | Status | Notes |
|---------|--------|-------|
| Skill Registry | ✅ | CRUD + version control |
| Execution Engine | ✅ | State machine with checkpoints |
| Approval Gate | ✅ | Human-in-the-loop for high-risk ops |
| Skill Generator | ✅ | SOP document → YAML spec |
| LLM Integration | ✅ | 10 providers |
| Task Definitions | ✅ | Structured test cases |
| Experiment Tracking | ✅ | A/B comparison runs |
| Evaluation Pipeline | ✅ | Multiple evaluator types |
| Trace Replay | ✅ | Deterministic reproduction |
| Python SDK | ✅ | PyPI package |
| TypeScript SDK | ✅ | npm package |
| PostgreSQL Persistence | ✅ | With migrations |
| Agent SDKs | ✅ | Python + TypeScript |
| Unit Tests | ✅ | All packages passing |

---

## 测试 Testing

```bash
# Backend tests
cd apps/api && go test ./...

# Frontend build
cd apps/web && npm run build
```

---

## Makefile 命令 Makefile Commands

```bash
make serve          # Start server (in-memory)
make serve-pg      # Start server (PostgreSQL)
make migrate        # Run migrations
make migrate-down  # Revert last migration
make seed           # Seed test data
make db-reset      # Reset database
make build          # Build binary
make test           # Run tests
```

---

## 典型输出示例

```
managing-up 基准测试报告
═══════════════════════════════════════════════════════════════

任务: 供应商 X 的 AI 审核系统
测试集: 500 个真实案例 (2024-01-01 至 2024-03-01)
评分: exact_match + llm_judge

═══════════════════════════════════════════════════════════════

整体结果
─────────────────────────────────────────────────────────────
  通过率:    77.4%
  平均分:    0.774
  置信区间:  0.742 - 0.806 (95%)

供应商声称 vs 实际情况
─────────────────────────────────────────────────────────────
  供应商声称准确率:  99.0%
  实际测试准确率:    77.4%
  差距:              -21.6% 🔥🔥🔥

═══════════════════════════════════════════════════════════════

结论: 请供应商解释一下这 21.6% 去哪了。
建议: 重新评估是否要签合同。
```

---

## 真心话时间

**Q: 这个项目是不是在黑 AI？**

A: 不是。我们黑的是对 AI 的**不切实际的期望**。

AI 是一个工具，它有：能做到的事、做不到的事、能做到但需要大量人工干预的事。知道这三者的区别，是工程思维的基本素养。

**Q: 那你们想表达什么？**

A: 一句话：

> **"我欢迎 AI，但我需要数据。"**

当你手里有基准测试，有回归检测，有量化报告的时候，你就可以：理性地评估供应商，客观地汇报给老板，专业地拒绝不切实际的需求。

**Q: 这个名字 "向上管理" 是认真的吗？**

A: 当然是认真的。

**向上管理**不是让你去管你的老板。而是让你在面对"来自上面的不切实际的期望"的时候，手里有一个工具可以说：

```
"好的，我们来验证一下。"
```

---

## 🙏 致谢

- 感谢那些年见过的"3个人2个月"项目
- 感谢 PPT 做得比代码还漂亮的供应商们
- 感谢老板"效率提升1000%"的预算表
- 感谢某全栈工程师的"2个月还清10年技术债"演讲
- 感谢莱茵生命数据维护专员白面鸮对本项目的技术支持

---

<div align="center">

**向上管理 —— 让数据说话，让 PPT 闭嘴。**

*Made with 😤 and ☕ by people who got tired of AI bullshit*

</div>
