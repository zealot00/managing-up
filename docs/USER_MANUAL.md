# Managing-Up v2.1 用户手册

> **版本**: 2.1 | **更新日期**: 2026-05-18

---

## 目录

1. [概述](#概述)
2. [LLM Gateway 多模型网关](#llm-gateway-多模型网关)
3. [Provider Fallback 灾备降级](#provider-fallback-灾备降级)
4. [MCP Server 管理与工具调用](#mcp-server-管理与工具调用)
5. [MCP Proxy 代理网关](#mcp-proxy-代理网关)
6. [MCP Router 智能路由](#mcp-router-智能路由)
7. [Gateway Sessions 会话追踪](#gateway-sessions-会话追踪)
8. [Policy Hook 策略钩子](#policy-hook-策略钩子)
9. [Capability Snapshots 能力快照](#capability-snapshots-能力快照)
10. [Regression Gate 回归门禁](#regression-gate-回归门禁)
11. [Memory Hub 记忆中心](#memory-hub-记忆中心)
12. [Bridge Adapter 桥接适配器](#bridge-adapter-桥接适配器)
13. [Sweep Engine 超参矩阵](#sweep-engine-超参矩阵)
14. [Skill Repository 技能仓库](#skill-repository-技能仓库)
15. [Orchestrator 编排引擎](#orchestrator-编排引擎)
16. [用户设置与偏好](#用户设置与偏好)
17. [前端页面使用指南](#前端页面使用指南)
18. [环境变量参考](#环境变量参考)
19. [故障排除](#故障排除)

---

## 概述

Managing-Up v2.1 是企业级 AI 质量基础设施平台，提供完整的 AI 服务治理闭环：

- **多模型网关**: OpenAI/Anthropic 兼容的 LLM 代理，10 家提供商、限流、熔断、预算、计费
- **灾备降级**: Provider Fallback 自动切换，多级 fallback chain，熔断器保护
- **MCP 生态**: Server 管理 + Proxy 代理 + Router 路由 + 权限控制，完整的 MCP 全链路
- **治理内核**: Session 追踪 → Policy Check → 路由决策 → 执行 → 评测 → 发布
- **记忆能力**: 跨会话状态保持和检索
- **快速集成**: Bridge Adapter 支持 Any-to-MCP
- **超参扫描**: Sweep Engine 支持 Model × Temperature × MaxTokens × Prompt 矩阵实验

---

## LLM Gateway 多模型网关

### 什么是 LLM Gateway

LLM Gateway 是 OpenAI 和 Anthropic 兼容的代理接口，统一管理多家 LLM 提供商的请求、认证、计费和监控。

### 支持的提供商

| Provider | 模型示例 |
|----------|---------|
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

### 核心功能

- **多提供商路由**: 自动根据模型名路由到对应提供商
- **API Key 认证**: Bearer Token / x-api-key 双协议支持
- **使用统计**: 按 Provider/Model/User 聚合
- **费用追踪**: 基于模型定价自动计算成本
- **限流**: Redis 分布式每 Key 每分钟请求限制
- **熔断器**: Redis 实现的指数退避熔断保护
- **预算控制**: 原子 check-and-decrement 预算检查
- **重试机制**: 指数退避自动重试 + Provider Fallback 灾备
- **流式响应**: SSE 流式传输，支持超大 buffer 配置
- **API Key 加密**: AES-GCM 加密存储

### 兼容接口

| 端点 | 协议 |
|------|------|
| `/v1/chat/completions` | OpenAI Chat Completions |
| `/v1/messages` | Anthropic Messages |
| `/v1/embeddings` | Embeddings |
| `/v1/models` | 模型列表 |

### 管理接口

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| GET/POST | `/api/v1/gateway/keys` | 列出/创建 API Key | RequireAuth |
| DELETE | `/api/v1/gateway/keys/{id}` | 撤销 API Key | RequireAuth |
| GET | `/api/v1/gateway/providers` | 列出提供商 | RequireAuth |
| GET | `/api/v1/gateway/providers/{id}` | 提供商详情 | RequireAuth |
| GET | `/api/v1/gateway/usage` | 使用统计 | RequireAuth |
| GET | `/api/v1/gateway/usage/users` | 用户级统计 | RequireAuth |
| GET | `/api/v1/gateway/budget` | 预算管理 | RequireAuth |

---

## Provider Fallback 灾备降级

### 什么是 Provider Fallback

当主用 LLM 提供商不可用（宕机、限流、超时）时，自动切换到备用提供商，确保服务持续可用。

### 工作流程

```
请求到达 → 主 Provider 调用 → 失败?
  ↓ 否                        ↓ 是
  返回结果               熔断器记录失败 → 尝试 Fallback Provider
                              ↓ 成功?
                         是 → 返回结果
                         否 → 尝试下一个 Fallback → 全部失败则报错
```

### Fallback Chain 配置

#### 方式一：数据库配置（推荐，支持热更新）

通过管理页面 `/fallback-chains` 或 API 配置：

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| GET | `/api/v1/admin/fallback-chains` | 列出所有 Fallback Chain | RequireAuth |
| POST | `/api/v1/admin/fallback-chains` | 创建 Chain | RequireAuth |
| GET | `/api/v1/admin/fallback-chains/{id}` | Chain 详情 | RequireAuth |
| PUT | `/api/v1/admin/fallback-chains/{id}` | 更新 Chain | RequireAuth |
| DELETE | `/api/v1/admin/fallback-chains/{id}` | 删除 Chain | RequireAuth |

创建后自动热加载到 Router，无需重启。

#### 方式二：环境变量配置

```bash
GATEWAY_FALLBACK_CHAINS='{"gpt-4o": ["anthropic:claude-sonnet-4", "ollama:qwen2.5"]}'
```

格式：`{模型名: ["provider:model", "provider:model"]}` 的 JSON 映射。

### 熔断器

- 自动检测 Provider 失败
- 指数退避恢复（Redis 实现）
- 熔断打开时跳过该 Provider，直接尝试下一个

### 注意事项

- DB 配置的 Chain 在进程重启后自动加载
- DB Chain 覆盖同名 model 的 env 配置，但保留 env 中独有的 model
- Chain 中已禁用的 Target 会被自动跳过

---

## MCP Server 管理与工具调用

### MCP Server 生命周期

```
注册 MCP Server → 审批测试 → 批准上线 → Agent 发现工具 → 权限检查 → 调用工具
     ↓               ↓           ↓            ↓              ↓           ↓
   pending      validation   approved    tools list      perm check   invoke
```

### 管理接口

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| GET/POST | `/api/v1/mcp-servers` | 列出/注册 MCP Server | RequireAuth |
| GET/PUT/DELETE | `/api/v1/mcp-servers/{id}` | 详情/更新/删除 | RequireAuth |
| POST | `/api/v1/mcp-servers/{id}/approve` | 审批/拒绝 | RequireAuth |
| GET | `/api/v1/mcp-servers/{id}/tools` | 列出工具 | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/resources` | 列出资源 | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/resources/read` | 读取资源 | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/resources/templates` | 资源模板 | OptionalAuth |
| POST | `/api/v1/mcp-servers/{id}/resources/subscribe` | 订阅资源 | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/prompts` | 列出 Prompt | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/prompts/{name}` | 获取 Prompt | OptionalAuth |
| GET | `/api/v1/mcp-servers/{id}/health` | 健康检查(单个) | OptionalAuth |
| GET | `/api/v1/mcp-servers/health` | 健康检查(全部) | OptionalAuth |
| GET | `/api/v1/mcp/tools` | 列出所有工具 | OptionalAuth |

### 工具调用

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| POST | `/api/v1/mcp/invoke` | 调用 MCP 工具 | OptionalAuth |
| POST | `/api/v1/mcp/invoke/stream` | 流式调用 MCP 工具 | OptionalAuth |

### 权限管理

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| POST | `/api/v1/mcp/permissions` | 授权 MCP Server 访问 | RequireAuth |
| GET | `/api/v1/mcp/permissions/list` | 列出权限 | OptionalAuth |
| DELETE | `/api/v1/mcp/permissions/{id}` | 撤销权限 | RequireAuth |

### 传输模式

| 类型 | 说明 |
|------|------|
| `stdio` | 本地进程，通过 stdin/stdout 通信 |
| `sse` | 远程 HTTP/SSE 传输 |

### 安全特性

- **命令白名单**: stdio 仅允许 npx, node, python, docker, kubectl, gh, git, curl, wget
- **参数验证**: 拒绝包含 shell 元字符的参数
- **Header 验证**: 阻止 CRLF 注入
- **权限检查**: 调用工具前验证 user/api_key 对该 MCP Server 的访问权限

---

## MCP Proxy 代理网关

### 什么是 MCP Proxy

MCP Proxy 是完整的 MCP 协议代理，拦截 Agent 对 MCP Server 的所有请求，在代理层实现认证、权限过滤和工具命名空间隔离。

### 核心能力

- **认证集成**: 复用 Gateway API Key 认证，Agent 使用同一个 Key 访问 LLM 和 MCP
- **权限过滤**: 仅返回用户有权限访问的 MCP Server 的工具列表
- **工具命名空间**: 工具名格式为 `{server_name}__{tool_name}`，避免不同 MCP Server 工具名冲突
- **协议代理**: 透明代理 MCP JSON-RPC 请求到后端 MCP Server

### 接入方式

```
Agent → MCP Proxy (认证 + 过滤) → 后端 MCP Servers
```

端点: `/mcp`（标准 MCP 协议）

---

## MCP Router 智能路由

### 什么是 MCP Router

MCP Router 是无状态接入层，基于任务类型 + 标签智能路由到最优 MCP 服务。

### 特性

- **智能路由**: 基于 task_type + tags 精确匹配
- **信任评分**: 按 trust_score + use_count 排序选择最优服务
- **Prometheus Metrics**: 请求速率、延迟分布、错误率监控
- **Session 历史**: 请求级会话追踪

### 接口

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| POST | `/api/v1/mcp-router/route` | 路由请求 | OptionalAuth |
| GET | `/api/v1/mcp-router/catalog` | 路由目录 | OptionalAuth |
| GET | `/api/v1/mcp-router/match` | 匹配查询(调试) | OptionalAuth |
| GET | `/api/v1/mcp-router/sessions` | Session 历史 | 无 |
| GET | `/metrics` | Prometheus metrics | 无 |

---

## Gateway Sessions 会话追踪

### 什么是 Gateway Session

Gateway Session 是每个请求在网关层的唯一标识，记录从请求到响应的完整生命周期。

### 工作流程

```
Agent 请求 → 创建 Session → Policy Check → 路由决策 → 执行 → Session 结束
```

### API 调用

```bash
# 列出所有 Session
curl http://localhost:8080/api/v1/gateway/sessions

# 按 Agent 过滤
curl "http://localhost:8080/api/v1/gateway/sessions?agent_id=openclaw-v1"
```

---

## Policy Hook 策略钩子

### 内置风险等级

| 等级 | 说明 | 默认行为 |
|------|------|---------|
| `low` | 查看、搜索等低风险操作 | 自动放行 |
| `medium` | 创建、更新等中等风险 | 日志记录后放行 |
| `high` | 删除、部署等高风险 | 标记需要审批 |
| `critical` | 支付、管理等极高风险 | 阻止执行 |

### 配置

```bash
DEFAULT_POLICY_RULES='[
  {"condition": "task_type:delete", "action": "deny", "reason": "Delete operations require approval"}
]'
```

---

## Capability Snapshots 能力快照

### 什么是 Capability Snapshot

Snapshot 是 Skill 版本的评测结果记录，包含评分、指标、通过/失败状态。

### API 调用

```bash
# 获取最新 Snapshot
curl "http://localhost:8080/api/v1/snapshots?skill_id=YOUR_SKILL_UUID&version=1.0.0"

# 列出所有 Snapshot
curl "http://localhost:8080/api/v1/snapshots/list?skill_id=YOUR_SKILL_UUID"
```

> **注意**: v2.1 起 skills.id 为 UUID 格式，不再使用 `skill_001` 等文本 ID。

---

## Regression Gate 回归门禁

### 工作流程

```
开发者提交新版本 → 自动触发评测 → 创建 Snapshot → Promote 检查 → passed=true 允许 / passed=false 拒绝
```

### 通过 Gate 的条件

1. 存在对应版本的 Snapshot
2. `passed` 字段为 `true`
3. `overall_score` 达到阈值要求

---

## Memory Hub 记忆中心

### 内存 Scope

| Scope | 生命周期 | 适用场景 |
|-------|---------|---------|
| `execution` | 单次执行 | 步骤间传递中间结果 |
| `session` | 会话期间 | 跨 Agent 调用保持上下文 |
| `agent` | Agent 级别 | Agent 的长期偏好设置 |
| `tenant` | 租户级别 | 组织级共享知识 |

### API 调用

```bash
# 存储记忆
curl -X POST http://localhost:8080/api/v1/memory \
  -H "Content-Type: application/json" \
  -d '{"scope":"session","agent_id":"openclaw-v1","session_id":"sess_abc123","key":"user_preferences","value":{"theme":"dark"}}'

# 检索记忆
curl "http://localhost:8080/api/v1/memory?scope=session&session_id=sess_abc123"
```

---

## Bridge Adapter 桥接适配器

### 完整工作流

```
1. 导入 OpenAPI Spec
       ↓
2. 生成 Adapter 模板 (endpoints + mapping)
       ↓
3. 配置响应优化规则
       ↓
4. 注册为 MCP Server
       ↓
5. Agent 即可调用
```

### 响应优化规则

| 规则类型 | 说明 | 示例 |
|---------|------|------|
| `pick` | 只返回指定字段 | `{"type":"pick","fields":["id","name"]}` |
| `omit` | 排除指定字段 | `{"type":"omit","fields":["internal_id"]}` |
| `truncate` | 截断字符串字段 | `{"type":"truncate","fields":["description"],"max_length":500}` |
| `summarize` | 截断数组列表 | `{"type":"summarize","max_items":10}` |

---

## Sweep Engine 超参矩阵

### 什么是 Sweep Engine

Sweep Engine 支持多维超参数组合实验，自动生成 Model × Temperature × MaxTokens × Prompt 的笛卡尔积矩阵。

### 接口

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sweeps` | 列出 Sweep |
| POST | `/api/v1/sweeps/create` | 创建 Sweep |
| GET | `/api/v1/sweeps/{id}` | Sweep 详情 |
| DELETE | `/api/v1/sweeps/delete/{id}` | 删除 Sweep |
| GET | `/api/v1/sweeps/matrix/{id}` | 获取矩阵结果 |

---

## Skill Repository 技能仓库

### 特性

- **SOP 参照**: 每个 Skill 可关联标准操作规程版本
- **依赖管理**: Skill 间依赖声明与解析
- **评分系统**: 用户评分 + 信任评分
- **市场发现**: 按类别浏览和搜索 Skill
- **多来源创建**: 手动创建 / 上传 / CLI 工具 / AI Agent 生成

### 接口

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | 列出 Skills |
| POST | `/api/v1/skills` | 创建 Skill |
| GET | `/api/v1/skills/{id}` | Skill 详情 |
| GET | `/api/v1/skills/{id}/spec` | 下载 Spec YAML |
| GET | `/api/v1/skills/market` | 浏览市场 |
| GET | `/api/v1/skills/search` | 搜索 Skills |
| POST | `/api/v1/skills/{id}/rate` | 评分 |
| GET | `/api/v1/skills/{id}/dependencies` | 查看依赖 |
| POST | `/api/v1/skills/resolve-deps` | 解析依赖 |
| GET | `/api/v1/skill-versions` | 列出版本 |

---

## Orchestrator 编排引擎

CLI 编排 API，用于远程增强提取、Skill 版本管理、测试编排和门禁评估。

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/healthz` | 健康检查 |
| POST | `/v1/runs` | 创建编排 Run |
| GET | `/v1/runs/{id}` | Run 状态 |
| GET | `/v1/runs/{id}/artifacts` | Run 产物 |
| POST | `/v1/extraction/enhance` | 增强提取 |
| GET/POST | `/v1/skills` | 列出/创建 Skill |
| GET | `/v1/skills/{id}` | Skill 详情 |
| GET | `/v1/skills/{id}/versions` | 版本列表 |
| POST | `/v1/skills/{id}/versions` | 创建版本 |
| POST | `/v1/skills/{id}/promote` | 发布版本 |
| POST | `/v1/tests/runs` | 创建测试 Run |
| POST | `/v1/gates/evaluate` | 评估 Gate |
| GET | `/v1/policies/{id}` | 获取 Policy |

---

## 用户设置与偏好

### 个人资料

| Method | Endpoint | Description | 认证 |
|--------|----------|-------------|------|
| GET | `/api/v1/user/profile` | 获取个人资料 | RequireAuth |
| PUT | `/api/v1/user/password` | 修改密码 | RequireAuth |
| GET/PUT | `/api/v1/user/preferences` | 获取/更新偏好 | RequireAuth |

### 偏好设置

- **语言**: 中文 / English 切换
- **侧边栏折叠**: 记忆侧边栏展开/折叠状态

偏好保存到数据库，跨会话持久化。

---

## 前端页面使用指南

| 路径 | 功能 | 说明 |
|------|------|------|
| `/dashboard` | 仪表盘 | 统计概览 |
| `/skills` | 技能管理 | CRUD + 市场 + 依赖 |
| `/executions` | 执行记录 | 执行追踪 + 审批 |
| `/tasks` | 任务定义 | 结构化测试用例 |
| `/evaluations` | 评估引擎 | 指标 + 评分 + 运行 |
| `/experiments` | 实验对比 | A/B 对比 |
| `/approvals` | 审批中心 | 人工审批 |
| `/replays` | 回放快照 | 确定性回放 |
| `/gateway` | Gateway 管理 | API Key + 用量统计 + 提供商 |
| `/mcp` | MCP Server 管理 | CRUD + 审批 + 健康检查 |
| `/mcp-router` | MCP Router 仪表盘 | 路由目录 + 监控 + Session |
| `/mcp-router/metrics` | Prometheus 监控 | 请求延迟 + 错误率 |
| `/mcp-router/sessions` | Session 历史 | 路由会话追踪 |
| `/fallback-chains` | Fallback Chain 管理 | CRUD + 热更新 + 启停 |
| `/policies` | 策略管理 | 规则编辑 + 版本控制 |
| `/sweeps` | 超参矩阵 | 创建矩阵实验 + 结果可视化 |
| `/profile` | 个人资料 | 修改密码 |
| `/preferences` | 偏好设置 | 语言 + 侧边栏 |

---

## 环境变量参考

### Gateway & Fallback

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GATEWAY_FALLBACK_CHAINS` | Fallback Chain JSON 配置 | - |
| `GATEWAY_OLLAMA_BASE_URL` | Ollama 默认地址 | - |
| `GATEWAY_SCANNER_BUFFER_SIZE` | SSE 流 scanner 初始 buffer | 10485760 (10MB) |
| `GATEWAY_SCANNER_MAX_BUFFER_SIZE` | SSE 流 scanner 最大 buffer | 52428800 (50MB) |
| `GATEWAY_MAX_TOKEN_ESTIMATE` | 请求 token 估算上限 | 1000000 |
| `GATEWAY_ENCRYPTION_KEY` | API Key 加密密钥 (32-byte base64) | - |

### Auth

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `AUTH_JWT_SECRET` | JWT 签名密钥 | `dev-secret-change-in-production` |

### Redis

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `REDIS_ADDR` | Redis 地址 | - |
| `REDIS_PASSWORD` | Redis 密码 | - |
| `REDIS_DB` | Redis 数据库编号 | 0 |

### Gateway Sessions & Policy

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SESSION_RETENTION_DAYS` | Session 保留天数 | 30 |
| `DEFAULT_POLICY_RULES` | 默认策略规则 JSON | [] |

### Memory Hub

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ENABLE_MEMORY_INJECTION` | 启用自动注入 | false |
| `MEMORY_INJECTION_SCOPE` | 注入 scope | session |
| `MEMORY_TTL_DAYS` | 默认 TTL | 7 |

### Orchestrator

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ORCHESTRATOR_JWT_SECRET` | JWT 密钥 | - |
| `ORCHESTRATOR_JWT_ISSUER` | JWT 签发者 | - |
| `ORCHESTRATOR_SKIP_AUTH` | 开发模式跳过认证 | false |

---

## 故障排除

### Session 未找到

- 检查 `agent_id` 是否正确
- 确认请求已通过 Gateway 处理

### Snapshot 未通过

- 查看 `overall_score` 是否达到阈值
- 检查 `metrics` 中的详细指标

### Memory 检索为空

- 确认 `scope`、`session_id`、`agent_id` 参数正确
- 检查 Memory 是否已过期

### Fallback 未生效

- 确认 Fallback Chain 已创建且 `is_enabled=true`
- 确认 Target 的 `is_enabled=true`
- 检查主 Provider 的熔断器是否已打开
- 重启后 DB Chain 会自动加载

### API Key 创建失败

- 确认 `GATEWAY_ENCRYPTION_KEY` 已配置（32-byte base64）

---

## 相关文档

- [架构设计](architecture.md)
- [API 参考](api-reference.md)
- [部署指南](deployment.md)
