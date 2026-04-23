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
│  ┌──────────────────────────────────────────────────┐     │
│  │           MCP Router (无状态接入层)                  │     │
│  │  Ingress → Task Parser → Match Engine → Egress  │     │
│  │  + Prometheus Metrics (/metrics)                  │     │
│  └──────────────────────────────────────────────────┘     │
│                                                             │
│  ┌──────────────────────────────────────────────────┐     │
│  │           Skill Repository (企业级仓库)            │     │
│  │  Registry + Market + Dependencies + SOP Reference  │     │
│  └──────────────────────────────────────────────────┘     │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │    Tasks     │  │  Experiments │  │  Evaluations │    │
│  │   任务定义    │  │   实验编排    │  │   评分引擎    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
│  ┌─────────────────────────────────────────────┐          │
│  │              LLM Gateway                     │          │
│  │  (OpenAI/Anthropic 兼容, 限流, 熔断, 预算)  │          │
│  └─────────────────────────────────────────────┘          │
│                                                             │
│  ┌─────────────┐  ┌──────────────────────────────────┐    │
│  │    Redis    │  │        PostgreSQL                 │    │
│  │  (限流/熔断) │  │  (数据持久化, API Keys, 预算)     │    │
│  └─────────────┘  └──────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

**managing-up 不运行在 AI 应用的运行时路径上**，而是作为**旁路（Side-Car）**的质量保障系统。

> 有数据才能拒绝不切实际的需求。没有数据，你拒绝不了PPT。

### 架构图

![Architecture](docs/architecture.png)

### 部署图

![Deployment](docs/deployment.png)

---

## 快速开始 Quick Start

### 1. 配置数据库

```bash
cd apps/api
# 编辑 .env 文件
cat > .env << EOF
PORT=8080
DB_DRIVER=postgres
DATABASE_URL=postgresql://postgres:pass@localhost:5432/managing-up
EOF
```

### 2. 运行迁移和种子数据

```bash
cd apps/api
go run cmd/migrate/main.go   # 创建表结构
go run cmd/seed/main.go       # 插入测试数据 + 默认用户
```

默认账号：`admin` / `admin`

### 3. 启动后端

```bash
cd apps/api
go run cmd/server/main.go
# Server running at http://localhost:8080
```

### 4. 启动前端

```bash
cd apps/web
npm install
npm run dev
# Frontend at http://localhost:3000
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PORT` | API 监听端口 | `8080` |
| `DB_DRIVER` | 数据库驱动 (必须 postgres) | - |
| `DATABASE_URL` | PostgreSQL 连接字符串 | - |
| `LLM_PROVIDER` | 默认 LLM 提供商 | `ollama` |
| `LLM_MODEL` | 默认 LLM 模型 | `deepseek-r1-tool-calling:7b` |
| `LLM_BASE_URL` | LLM API 地址 | `http://localhost:11434` |
| `LLM_API_KEY` | LLM API Key | - |
| `NEXT_PUBLIC_API_BASE_URL` | 前端 API 地址 | `http://localhost:8080` |
| `ORCHESTRATOR_SKIP_AUTH` | 临时绕过 Orchestrator JWT 认证（开发模式） | `false` |
| `ORCHESTRATOR_JWT_SECRET` | Orchestrator JWT 密钥 | - |
| `ORCHESTRATOR_JWT_ISSUER` | Orchestrator JWT 签发者 | - |
| `GATEWAY_SCANNER_BUFFER_SIZE` | SSE 流扫描器初始 buffer 大小 | `10485760` (10MB) |
| `GATEWAY_SCANNER_MAX_BUFFER_SIZE` | SSE 流扫描器最大 buffer 大小 | `52428800` (50MB) |
| `GATEWAY_MAX_TOKEN_ESTIMATE` | 请求 token 估算上限（DoS防护） | `1000000` (1M) |
| `GATEWAY_ENCRYPTION_KEY` | API Key 加密密钥（32-byte base64） | - |

---

## API 端点 API Endpoints

### 认证 Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/logout` | 用户登出 |
| GET | `/api/v1/auth/me` | 获取当前用户信息 |

### 核心 Core

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/healthz` | 健康检查 |
| GET | `/api/v1/meta` | API 元数据 |
| GET | `/api/v1/dashboard` | Dashboard 统计 |
| GET | `/api/v1/tip` | 获取登录页名言 (随机) |

### 技能 Skills

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | 列出技能 |
| POST | `/api/v1/skills` | 创建技能 |
| GET | `/api/v1/skills/{id}` | 技能详情 |
| GET | `/api/v1/skills/{id}/spec` | 下载 YAML spec |
| GET | `/api/v1/skill-versions` | 列出技能版本 |
| POST | `/api/v1/generate-skill` | 从 SOP 生成技能 |

### 执行 Executions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/executions` | 列出执行记录 |
| POST | `/api/v1/executions` | 触发执行 |
| GET | `/api/v1/executions/{id}` | 执行详情 |
| POST | `/api/v1/executions/{id}/approve` | 审批/拒绝 |
| GET | `/api/v1/approvals` | 列出审批 |

### 任务 Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks` | 列出任务 |
| POST | `/api/v1/tasks` | 创建任务 |
| GET | `/api/v1/tasks/{id}` | 任务详情 |
| PUT | `/api/v1/tasks/{id}` | 更新任务 |
| DELETE | `/api/v1/tasks/{id}` | 删除任务 |
| GET | `/api/v1/metrics` | 列出指标定义 |
| GET | `/api/v1/task-executions` | 任务执行列表 |
| GET | `/api/v1/task-executions/{id}` | 任务执行详情 |

### 实验 Experiments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/experiments` | 列出实验 |
| POST | `/api/v1/experiments` | 创建实验 |
| GET | `/api/v1/experiments/{id}` | 实验结果 |
| GET | `/api/v1/experiments/{id}/compare` | 对比实验 |
| POST | `/api/v1/check-regression` | 回归检测 |

### 回放 Replay

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/replay-snapshots` | 回放快照列表 |
| GET | `/api/v1/replay-snapshots/{id}` | 快照详情 |

### LLM Gateway (OpenAI/Anthropic 兼容)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/models` | 可用模型列表 |
| POST | `/v1/chat/completions` | OpenAI 兼容接口 |
| POST | `/v1/messages` | Anthropic 兼容接口 |
| POST | `/v1/embeddings` | Embeddings 接口 |
| GET | `/api/v1/gateway/keys` | 列出 Gateway API Key |
| POST | `/api/v1/gateway/keys` | 创建 Gateway API Key |
| DELETE | `/api/v1/gateway/keys/{id}` | 撤销 API Key |
| GET | `/api/v1/gateway/usage` | 使用统计 (按 Provider/Model) |
| GET | `/api/v1/gateway/usage/users` | 使用统计 (按用户) |

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

---

## LLM Gateway

OpenAI 和 Anthropic 兼容的 LLM 代理接口，支持多提供商、API Key 认证、使用统计和费用追踪。

### 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/models` | List available models |
| POST | `/v1/chat/completions` | OpenAI-compatible chat completions |
| POST | `/v1/messages` | Anthropic-compatible messages |
| POST | `/v1/embeddings` | Embeddings |

### 使用示例

```bash
# List models
curl http://localhost:8080/v1/models

# OpenAI chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $OPENAI_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hi"}]}'

# Anthropic messages
curl -X POST http://localhost:8080/v1/messages \
  -H "x-api-key: $ANTHROPIC_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-haiku-3","messages":[{"role":"user","content":"Hi"}]}'
```

### Gateway 管理

```bash
# 创建 API Key
curl -X POST http://localhost:8080/api/v1/gateway/keys \
  -H "Content-Type: application/json" \
  -d '{"name":"my-key"}'

# 列出 Keys
curl http://localhost:8080/api/v1/gateway/keys

# 查看使用统计
curl http://localhost:8080/api/v1/gateway/usage?from=2026-04-01&to=2026-04-30

# 查看用户级别使用统计 (admin)
curl http://localhost:8080/api/v1/gateway/usage/users
```

### Gateway 功能

- **多提供商路由**: 自动根据模型名路由到对应提供商
- **API Key 认证**: Bearer Token / x-api-key
- **使用统计**: 按 Provider/Model/User 聚合
- **费用追踪**: 基于模型定价自动计算成本
- **限流**: 每 Key 每分钟请求限制
- **重试机制**: 指数退避自动重试

---

## MCP Servers (Model Context Protocol)

MCP Server 管理 API，允许 AI Agent 发现和使用外部 MCP 服务器提供的工具。

### MCP 工作流

```
注册 MCP Server → 审批测试 → 批准上线 → Agent 使用工具
     ↓               ↓           ↓           ↓
   pending      validation   approved    tools available
```

### API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/mcp-servers` | 列出所有 MCP Server |
| POST | `/api/v1/mcp-servers` | 注册新 MCP Server |
| GET | `/api/v1/mcp-servers/{id}` | 获取详情 |
| PUT | `/api/v1/mcp-servers/{id}` | 更新配置 |
| DELETE | `/api/v1/mcp-servers/{id}` | 删除 |
| POST | `/api/v1/mcp-servers/{id}/approve` | 审批/拒绝 |

### 注册 MCP Server

```bash
# 注册 stdio 传输的 MCP Server
curl -X POST http://localhost:8080/api/v1/mcp-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "filesystem",
    "description": "File system operations",
    "transport_type": "stdio",
    "command": "npx",
    "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
  }'

# 注册 HTTP/SSE 传输的 MCP Server
curl -X POST http://localhost:8080/api/v1/mcp-servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "remote-api",
    "description": "Remote HTTP MCP Server",
    "transport_type": "sse",
    "url": "https://mcp.example.com/sse"
  }'
```

### 审批 MCP Server

```bash
# 批准 (会自动验证连接)
curl -X POST http://localhost:8080/api/v1/mcp-servers/{id}/approve \
  -H "Content-Type: application/json" \
  -d '{"decision": "approved", "approver": "admin"}'

# 拒绝
curl -X POST http://localhost:8080/api/v1/mcp-servers/{id}/approve \
  -H "Content-Type: application/json" \
  -d '{"decision": "rejected", "approver": "admin", "note": "Security concern: untrusted command"}'
```

### MCP Server 状态

| Status | 说明 |
|--------|------|
| `pending` | 新注册，待审批 |
| `approved` | 已批准，可使用 |
| `rejected` | 已拒绝 |
| `disabled` | 已禁用 |

### 安全特性

- **命令白名单**: stdio 传输仅允许已知安全命令 (npx, node, python, docker, kubectl, gh, git, curl, wget)
- **参数验证**: 拒绝包含 shell 元字符的参数
- **Header 验证**: HTTP 传输头不能包含 CRLF 注入
- **连接验证**: 批准前自动验证 MCP 服务器可访问

---

## MCP Router (智能路由)

MCP Router 是无状态接入层，基于任务类型+标签智能路由到最优 MCP 服务。

### 特性

- **智能路由**: 基于 task_type + tags 精确匹配
- **信任评分**: 按 trust_score + use_count 排序选择最优服务
- **Prometheus Metrics**: 请求速率、延迟分布、错误率监控
- **审批自动同步**: MCP Server 审批通过后自动加入路由池

### API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/router/mcp/route` | 路由请求 |
| GET | `/api/v1/router/mcp/catalog` | 路由目录 |
| GET | `/api/v1/router/mcp/match` | 匹配查询（调试用） |
| GET | `/metrics` | Prometheus metrics |

---

## MCP Hub & SDK

MCP Hub 是 MCP Server 的注册与管理中心，提供 SDK 支持多语言开发 MCP Server。

### SDK 列表

| Language | Package | Location |
|----------|---------|----------|
| Go | `github.com/zealot/managing-up/sdk/mcp` | [`sdk/mcp/`](sdk/mcp/) |
| Python | `managing-up-sdk` | [`sdk/python/`](sdk/python/) |
| TypeScript | `@managing-up/mcp-sdk` | [`sdk/typescript/`](sdk/typescript/) |
| Rust | `managing-up-mcp-sdk` | [`sdk/rust/`](sdk/rust/) |

### SDK 特性

- **自动注册**: MCP Server 启动时自动注册到平台
- **Metrics 收集**: 内置请求量、延迟、错误率统计
- **Prometheus 兼容**: 支持 `/metrics` 端点
- **多传输模式**: 支持 Stdio 和 HTTP/SSE

### 环境变量

| Variable | Description | Default |
|----------|-------------|---------|
| `MANAGING_UP_PLATFORM_URL` | 平台 API 地址 | `http://localhost:8080` |
| `MANAGING_UP_NAME` | MCP Server 名称 | (required) |
| `MANAGING_UP_VERSION` | MCP Server 版本 | `1.0.0` |
| `MANAGING_UP_TOKEN` | 注册认证 Token | - |
| `MANAGING_UP_TRANSPORT_TYPE` | 传输类型 (`stdio`/`http`) | `http` |

### Go 示例

```go
import (
    "context"
    sdk "github.com/zealot/managing-up/sdk/mcp"
    "github.com/mark3labs/mcp-go/mcp"
)

func main() {
    server := sdk.NewServer(sdk.Config{
        Name:        "my-mcp",
        Version:     "1.0.0",
        Description: "My MCP Server",
    })

    server.AddTool(mcp.NewTool(
        "hello",
        mcp.WithDescription("Say hello"),
        mcp.WithString("name", mcp.Required()),
    ), helloHandler)

    ctx := context.Background()
    server.Register(ctx)
    server.StartHTTP(ctx, ":8080")
}
```

### Python 示例

```python
from managing_up_sdk import MCPServer, MCPServerConfig

server = MCPServer.from_env()
server.add_tool("hello", "Say hello", {}, hello_handler)

asyncio.run(server.register())
asyncio.run(server.serve_stdio())
```

### TypeScript 示例

```typescript
import { MCPServer } from '@managing-up/mcp-sdk';

const server = new MCPServer({
    name: 'my-mcp',
    version: '1.0.0',
});

server.addTool({
    name: 'hello',
    description: 'Say hello',
    inputSchema: { type: 'object' },
    handler: async (args) => `Hello, ${args.name}!`,
});

await server.register();
```

### 路由请求示例

```bash
# 路由请求
curl -X POST http://localhost:8080/api/v1/router/mcp/route \
  -H "Content-Type: application/json" \
  -H "X-Agent-ID: openclaw-v1" \
  -d '{
    "task": {
      "description": "帮我审查这段 Python 代码的性能问题",
      "structured": {
        "task_type": "code_review",
        "language": "python",
        "complexity": "high"
      }
    }
  }'
```

### Prometheus Metrics

```
# HELP mcp_router_requests_total Total MCP router requests
# TYPE mcp_router_requests_total counter
mcp_router_requests_total{agent="openclaw-v1",task_type="code_review",status="success"} 1523

# HELP mcp_router_request_duration_seconds MCP router request latency
# TYPE mcp_router_request_duration_seconds histogram

# HELP mcp_router_match_failures_total Route match failures
# TYPE mcp_router_match_failures_total counter
mcp_router_match_failures_total{reason="no_matching_server"} 23
```

---

## Gateway Sessions & Policy Hook

Gateway Sessions 提供请求追踪和策略决策能力，形成"请求 → Policy Check → 路由 → 执行 → 评测回填"的闭环。

### 特性

- **Session 追踪**: 每个请求创建唯一 Session，记录完整生命周期
- **Pre-flight Policy**: 请求执行前进行风险/合规判定
- **Policy 规则引擎**: 支持 task_type、tags、risk_level 条件匹配
- **自动化决策**: 高风险任务自动标记为需要审批

### API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/gateway/sessions` | 列出 Gateway Sessions |
| GET | `/api/v1/gateway/sessions/{id}` | Session 详情 |

### Session 数据结构

```json
{
  "id": "sess_abc123",
  "session_type": "router",
  "agent_id": "openclaw-v1",
  "correlation_id": "corr_xyz789",
  "task_intent": {
    "task_type": "code_review",
    "tags": ["security", "python"],
    "raw_description": "审查这段代码"
  },
  "risk_level": "medium",
  "policy_decision": {
    "allowed": true,
    "policy_id": "default",
    "reasons": []
  },
  "status": "active",
  "started_at": "2026-04-22T10:30:00Z"
}
```

### 风险等级

| Level | 说明 | 默认行为 |
|-------|------|---------|
| `low` | 低风险操作 | 自动放行 |
| `medium` | 中等风险 | 日志记录 |
| `high` | 高风险操作 | 需要审批 |
| `critical` | 极高风险 | 阻止执行 |

### Policy 规则配置

```bash
# 环境变量配置默认策略规则
DEFAULT_POLICY_RULES='[{"condition":"task_type:delete","action":"deny","reason":"Delete operations require approval"}]'
```

---

## Capability Snapshots & Regression Gate

Capability Snapshots 记录技能版本的能力评测结果，Regression Gate 确保只有通过评测的版本才能发布。

### 特性

- **自动化评测**: Skill 版本变更时自动触发回归测试
- **阈值门禁**: 未通过阈值的版本不可 promote
- **历史追踪**: 可查看任意版本的历史评测结果
- **多指标评分**: 支持多种评测指标聚合

### API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/snapshots` | 获取 Skill 最新 Snapshot |
| GET | `/api/v1/snapshots/list` | 列出 Skill 的所有 Snapshots |

### Snapshot 数据结构

```json
{
  "id": "snap_xyz123",
  "skill_id": "skill_code_review",
  "version": "1.2.0",
  "snapshot_type": "regression_gate",
  "dataset_id": "ds_benchmark_v2",
  "run_id": "run_456",
  "metrics": {
    "accuracy": 0.85,
    "precision": 0.82,
    "recall": 0.88,
    "f1": 0.85
  },
  "overall_score": 0.85,
  "passed": true,
  "gate_policy_id": "default_gate",
  "evaluated_at": "2026-04-22T12:00:00Z"
}
```

### Regression Gate 工作流

```
版本提交 → 自动触发评测 → Snapshot 记录结果 → Promote 检查 Snapshot
                                              ↓
                                    passed=true? → 允许 Promote
                                         ↓
                                    passed=false → 拒绝 Promote
```

### 查看 Snapshot 页面

前端页面: `/skills/snapshots?skill_id=your_skill_id`

---

## Memory Hub

Memory Hub 提供跨会话的记忆能力，支持 Agent 在多轮对话中保持上下文。

### 特性

- **多 Scope 支持**: execution / session / agent / tenant
- **CRUD 操作**: 创建、读取、更新、删除记忆单元
- **自动注入**: Gateway 自动将 Memory Context 注入请求
- **元数据检索**: 基于 tags、key 进行检索

### Memory Cell 数据结构

```json
{
  "id": "mem_abc123",
  "scope": "session",
  "agent_id": "openclaw-v1",
  "session_id": "sess_xyz789",
  "execution_id": "exec_123",
  "key": "user_preferences",
  "value": {"theme": "dark", "language": "zh-CN"},
  "value_type": "json",
  "tags": ["preferences", "ui"],
  "created_at": "2026-04-22T10:30:00Z",
  "updated_at": "2026-04-22T14:00:00Z"
}
```

### 内存 Scope 说明

| Scope | 生命周期 | 用途 |
|-------|---------|------|
| `execution` | 单次执行 | 步骤间传递中间结果 |
| `session` | 会话期间 | 跨 Agent 调用保持上下文 |
| `agent` | Agent 级别 | Agent 的长期偏好设置 |
| `tenant` | 租户级别 | 组织级共享知识 |

---

## Bridge Adapter

Bridge Adapter 提供 Any-to-MCP 的适配能力，支持将 REST API 快速转换为 MCP 工具。

### 特性

- **OpenAPI 导入**: 从 OpenAPI/Swagger spec 自动生成 Adapter 模板
- **请求插值**: 支持输入参数映射到 REST 请求
- **响应裁剪**: 字段白名单、大小限制、摘要生成
- **多格式支持**: JSON selector、字段过滤、列表截断

### OpenAPI 导入示例

```bash
# 导入 OpenAPI spec
curl -X POST http://localhost:8080/api/v1/gateway/bridge/import \
  -H "Content-Type: application/json" \
  -d @openapi_spec.json

# 返回 Adapter 模板
{
  "id": "adapter_123",
  "name": "My REST API",
  "base_url": "https://api.example.com",
  "endpoints": [...],
  "input_mapping": {...},
  "output_rules": [...]
}
```

### 响应优化规则

| 规则类型 | 说明 | 示例 |
|---------|------|------|
| `pick` | 只返回指定字段 | `{"type":"pick","fields":["id","name"]}` |
| `omit` | 排除指定字段 | `{"type":"omit","fields":["internal_id"]}` |
| `truncate` | 截断字符串字段 | `{"type":"truncate","fields":["description"],"max_length":500}` |
| `summarize` | 截断数组列表 | `{"type":"summarize","max_items":10}` |

---

## 前端页面说明

### 页面路由

| 路径 | 功能 | 说明 |
|------|------|------|
| `/gateway/sessions` | Session 历史 | 查看所有 Gateway Session 记录和 Policy 决策 |
| `/skills/snapshots` | Snapshot 历史 | 查看 Skill 版本的评测结果和 Regression Gate 状态 |
| `/evaluations` | 评估引擎 | 管理任务、执行、指标，可运行评估和查看结果 |
| `/mcp-router` | MCP 路由仪表盘 | 查看路由目录、服务器状态、信任评分 |

### 使用示例

**查看 Session 历史:**
1. 访问 `/gateway/sessions`
2. 在搜索框输入 Agent ID 过滤
3. 点击 Session 查看详情和 Policy 决策

**查看 Snapshot:**
1. 访问 `/skills/snapshots`
2. 输入 Skill ID 查询
3. 查看 PASSED/FAILED 状态和评分

**运行评估:**
1. 访问 `/evaluations`
2. 点击"运行评估"选择任务
3. 等待执行完成查看结果

---

## Skill Repository (技能仓库)

企业级 Skill 仓库，支持依赖管理、版本控制、市场发现和 SOP 参照。

### 特性

- **SOP 参照**: 每个 Skill 可关联标准操作规程版本
- **依赖管理**: Skill 间依赖声明与解析
- **评分系统**: 用户评分 + 信任评分
- **多来源创建**: 手动创建 / 上传 / CLI 工具 / AI Agent 生成

### API 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | 列出 Skills |
| GET | `/api/v1/skills/{id}` | Skill 详情 |
| GET | `/api/v1/skills/{id}/spec` | 下载 Skill Spec YAML |
| GET | `/api/v1/skills/market` | 浏览市场 |
| GET | `/api/v1/skills/search` | 搜索 Skills |
| POST | `/api/v1/skills/{id}/rate` | 评分 |
| GET | `/api/v1/skills/{id}/dependencies` | 查看依赖 |
| POST | `/api/v1/skills/{id}/versions` | 创建版本 |
| POST | `/api/v1/skills/{id}/rollback/{ver}` | 回滚版本 |

### Skill Spec 示例

```yaml
name: code-review-gpt
version: 1.0.0
risk_level: medium
description: AI 代码审查技能
inputs:
  - name: code
    type: string
    required: true
steps:
  - id: review
    type: tool
    tool_ref: openai-gpt4
    with:
      model: gpt-4o
      task: code_review

# SOP 参照
sop_reference:
  sop_id: SOP-DEV-001
  sop_name: 代码审查标准操作规程
  sop_version: "2.1"
  sop_section: "4.3 自动化审查"
  compliance_required: true

# 企业级扩展
enterprise:
  category: code_analysis
  tags: [ai, review, security]
  dependencies:
    - skill_id: sandbox-runtime
      version_constraint: ">=1.0.0"
  trust_score: 0.95
  verified: true
```

---

## 独立 Build

API 可独立构建：

```bash
# 单独构建 API
cd apps/api
go build ./...

# 运行测试
go test ./...

# 迁移数据库
go run cmd/migrate/main.go
```

---

## sop-to-skill Orchestrator API

CLI 编排 API，用于远程增强提取、Skill 版本管理、测试编排和门禁评估。

### 端点

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/healthz` | Health check |
| POST | `/v1/runs` | Create orchestration run |
| GET | `/v1/runs/{runId}` | Get run status |
| GET | `/v1/runs/{runId}/artifacts` | List artifacts |
| POST | `/v1/extraction/enhance` | Enhanced extraction |
| POST | `/v1/extraction/compare` | Compare extractions |
| GET | `/v1/skills` | List skills |
| POST | `/v1/skills` | Create skill |
| GET | `/v1/skills/{id}` | Get skill detail |
| GET | `/v1/skills/{id}/versions` | List versions |
| POST | `/v1/skills/{id}/versions` | Create version |
| GET | `/v1/skills/{id}/diff` | Diff versions |
| POST | `/v1/skills/{id}/rollback` | Rollback version |
| POST | `/v1/skills/{id}/promote` | Promote version |
| POST | `/v1/tests/runs` | Create test run |
| POST | `/v1/gates/evaluate` | Evaluate gate |
| GET | `/v1/policies/{id}` | Get policy |

详细文档见 [docs/api-reference.md](docs/api-reference.md)

---

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
│   │   │   ├── seed/            # Test data seeding
│   │   │   ├── hashpw/          # Password hashing utility
│   │   │   └── updatepw/        # Password reset utility
│   │   ├── internal/
│   │   │   ├── server/          # HTTP handlers + routing
│   │   │   │   └── handlers/    # Auth handler
│   │   │   ├── service/         # Domain logic
│   │   │   ├── engine/          # Execution engine + trace + replay
│   │   │   │   └── executors/  # MCP client + registry
│   │   │   ├── evaluator/       # Evaluators + evaluation runner
│   │   │   ├── generator/       # LLM skill generator (SOP → YAML)
│   │   │   ├── gateway/         # LLM Gateway (OpenAI/Anthropic compatible)
│   │   │   ├── llm/             # LLM provider clients (10 providers)
│   │   │   ├── orchestrator/    # SOP-to-Skill orchestrator
│   │   │   ├── seh/             # SEH module
│   │   │   └── repository/      # PostgreSQL repository
│   │   └── migrations/          # SQL migrations (0001-0012)
│   └── web/
│       └── app/                 # Next.js frontend
│           ├── components/      # Reusable UI components
│           ├── context/         # Auth context
│           ├── dashboard/       # Dashboard pages
│           ├── gateway/         # Gateway management + usage stats
│           ├── login/           # Login page (Soviet post-punk style)
│           ├── skills/          # Skill registry UI
│           ├── executions/      # Execution + trace timeline
│           ├── tasks/           # Task definitions
│           ├── evaluations/     # Evaluation results
│           ├── experiments/     # Experiment comparison
│           └── replays/         # Replay snapshots
├── sdk/
│   ├── python/                  # Python SDK
│   └── typescript/              # TypeScript SDK
└── docs/                        # Architecture docs
```

---

## 功能状态 Features

| Feature | Status | Notes |
|---------|--------|-------|
| MCP Server Management | ✅ | CRUD + approve workflow + validation |
| MCP Router | ✅ | 无状态接入层 + 智能路由 + Prometheus |
| Skill Registry | ✅ | CRUD + version control |
| Skill Repository | ✅ | Enterprise extensions + market + dependencies |
| Execution Engine | ✅ | State machine with checkpoints |
| Approval Gate | ✅ | Human-in-the-loop for high-risk ops |
| Skill Generator | ✅ | SOP document → YAML spec |
| LLM Integration | ✅ | 10 providers |
| LLM Gateway | ✅ | OpenAI/Anthropic compatible, streaming, cost tracking |
| Task Definitions | ✅ | Structured test cases |
| Experiment Tracking | ✅ | A/B comparison runs |
| Evaluation Pipeline | ✅ | Multiple evaluator types |
| Trace Replay | ✅ | Deterministic reproduction |
| Python SDK | ✅ | PyPI package |
| TypeScript SDK | ✅ | npm package |
| PostgreSQL Persistence | ✅ | With migrations (required) |
| Agent SDKs | ✅ | Python + TypeScript |
| Unit Tests | ✅ | All packages passing |
| Admin Panel UI | ✅ | Sidebar layout, dark theme |
| Usage Statistics | ✅ | Token ranking, cost tracking, charts |
| Login Tips | ✅ | Database-driven quotes/tips |
| TanStack Query | ✅ | Declarative data fetching + caching |
| React Hook Form + Zod | ✅ | Real-time inline form validation |
| Bulk Actions | ✅ | Checkbox selection + batch operations |
| Data Pagination | ✅ | LoadMore on Executions/Tasks/Evaluations |
| Data Formatters | ✅ | Relative time, duration, text truncation |
| Skeleton Loading | ✅ | ListSkeleton/CardGridSkeleton + keepPreviousData |
| MCP Router Dashboard | ✅ | /mcp-router 路由概览仪表盘 |
| MCP Router Metrics | ✅ | /mcp-router/metrics Prometheus 监控 |
| Skill Market | ✅ | /skills/market 市场浏览 |
| My Skills | ✅ | /skills/my-skills 我的 Skills |
| Gateway Sessions | ✅ | /gateway/sessions Session 历史 + Policy 决策 |
| Skill Snapshots | ✅ | /skills/snapshots 评测结果 + Regression Gate |
| Evaluations Dashboard | ✅ | /evaluations 评估引擎 Dashboard 布局 |
| Memory Hub | ✅ | 跨会话记忆存储与检索 |
| Bridge Adapter | ✅ | OpenAPI 导入 + 响应优化 |

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
