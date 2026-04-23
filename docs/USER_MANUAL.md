# Managing-Up v2.0 用户手册

> **版本**: 2.0 | **更新日期**: 2026-04-23

---

## 目录

1. [概述](#概述)
2. [Gateway Sessions 会话追踪](#gateway-sessions-会话追踪)
3. [Policy Hook 策略钩子](#policy-hook-策略钩子)
4. [Capability Snapshots 能力快照](#capability-snapshots-能力快照)
5. [Regression Gate 回归门禁](#regression-gate-回归门禁)
6. [Memory Hub 记忆中心](#memory-hub-记忆中心)
7. [Bridge Adapter 桥接适配器](#bridge-adapter-桥接适配器)
8. [前端页面使用指南](#前端页面使用指南)

---

## 概述

Managing-Up v2.0 带来了企业级治理能力升级：

- **治理内核收敛**: 统一 Session、Policy、Gate 协议
- **闭环治理**: 请求 → Policy Check → 路由 → 执行 → 评测 → 发布
- **记忆能力**: 跨会话状态保持和检索
- **快速集成**: Bridge Adapter 支持 Any-to-MCP

---

## Gateway Sessions 会话追踪

### 什么是 Gateway Session

Gateway Session 是每个请求在网关层的唯一标识，记录从请求到响应的完整生命周期。

### 工作流程

```
Agent 请求 → 创建 Session → Policy Check → 路由决策 → 执行 → Session 结束
```

### 查看 Session 历史

1. 访问 `/gateway/sessions`
2. 在搜索框输入 `agent_id` 进行过滤
3. 点击任意 Session 查看详情

### Session 详情包含

- **基本信息**: Session ID、类型、状态、开始时间
- **请求信息**: Agent ID、Correlation ID、Task Intent
- **风险评估**: Risk Level (low/medium/high/critical)
- **Policy 决策**: 是否允许、策略ID、原因

### API 调用

```bash
# 列出所有 Session
curl http://localhost:8080/api/v1/gateway/sessions

# 按 Agent 过滤
curl "http://localhost:8080/api/v1/gateway/sessions?agent_id=openclaw-v1"

# 获取单个 Session
curl http://localhost:8080/api/v1/gateway/sessions/{session_id}
```

---

## Policy Hook 策略钩子

### 什么是 Policy Hook

Policy Hook 在请求执行前进行风险/合规判定，决定请求是否被允许执行。

### 内置风险等级

| 等级 | 说明 | 默认行为 |
|------|------|---------|
| `low` | 查看、搜索等低风险操作 | 自动放行 |
| `medium` | 创建、更新等中等风险 | 日志记录后放行 |
| `high` | 删除、部署等高风险 | 标记需要审批 |
| `critical` | 支付、管理等极高风险 | 阻止执行 |

### 高风险关键词

以下 Task Type 会自动提升风险等级：

- `delete` - 删除操作
- `deploy` - 部署操作
- `payment` - 支付相关
- `admin` - 管理操作
- `user_data` - 用户数据访问

### 自定义 Policy 规则

通过环境变量配置：

```bash
# JSON 格式的策略规则数组
DEFAULT_POLICY_RULES='[
  {"condition": "task_type:delete", "action": "deny", "reason": "Delete operations require approval"},
  {"condition": "task_type:deploy", "action": "require_approval", "reason": "Deploy requires admin approval"},
  {"condition": "risk_level:high", "action": "require_approval", "reason": "High risk tasks need review"}
]'
```

### 规则条件匹配

| 条件格式 | 说明 | 示例 |
|---------|------|------|
| `task_type:xxx` | 精确匹配任务类型 | `task_type:delete` |
| `tag:xxx` | 匹配标签 | `tag:payment` |
| `risk_level:xxx` | 匹配风险等级 | `risk_level:high` |
| `contains(description, xxx)` | 描述包含关键词 | `contains(description, deploy)` |

---

## Capability Snapshots 能力快照

### 什么是 Capability Snapshot

Snapshot 是 Skill 版本的评测结果记录，包含评分、指标、通过/失败状态。

### Snapshot 生命周期

```
版本提交 → 自动触发评测 → 创建 Snapshot → 结果存储 → 可查询
```

### 查看 Snapshot

1. 访问 `/skills/snapshots`
2. 输入 Skill ID
3. 查看该 Skill 所有版本的 Snapshot 列表

### Snapshot 信息说明

| 字段 | 说明 |
|------|------|
| `skill_id` | 关联的 Skill ID |
| `version` | 版本号 |
| `snapshot_type` | 快照类型 (regression_gate 等) |
| `overall_score` | 综合评分 (0-1) |
| `passed` | 是否通过门禁 |
| `metrics` | 详细指标 (accuracy, precision, recall 等) |
| `evaluated_at` | 评测时间 |

### API 调用

```bash
# 获取最新 Snapshot
curl "http://localhost:8080/api/v1/snapshots?skill_id=skill_xxx&version=1.0.0"

# 列出所有 Snapshot
curl "http://localhost:8080/api/v1/snapshots/list?skill_id=skill_xxx"
```

---

## Regression Gate 回归门禁

### 什么是 Regression Gate

Regression Gate 是 Skill promote 前的自动检查，确保只有通过评测的版本才能发布。

### 工作流程

```
开发者提交新版本
       ↓
自动触发评测 (如已配置)
       ↓
创建 Snapshot 记录结果
       ↓
执行 Promote 操作
       ↓
Gate 检查: snapshot.passed == true?
       ↓
    是 → Promote 成功
    否 → Promote 失败，提示原因
```

### 通过 Gate 的条件

1. 存在对应版本的 Snapshot
2. `passed` 字段为 `true`
3. `overall_score` 达到阈值要求

### 手动触发评测

```bash
# 通过前端
访问 /evaluations → 点击"运行评估" → 选择任务和 Skill → 提交

# 通过 API
POST /api/v1/orchestrator/v1/tests/runs
{
  "skill_id": "skill_xxx",
  "version": "1.2.0",
  "runner": {"type": "automated"}
}
```

### 查看 Gate 状态

1. 在 Skill 详情页查看当前版本的 Gate 状态
2. 访问 `/skills/snapshots` 查看历史评测结果
3. 未通过时显示具体分数和失败原因

---

## Memory Hub 记忆中心

### 什么是 Memory Hub

Memory Hub 提供跨会话的记忆能力，支持 Agent 在多轮对话中保持上下文。

### 内存 Scope

| Scope | 生命周期 | 适用场景 |
|-------|---------|---------|
| `execution` | 单次执行 | 步骤间传递中间结果 |
| `session` | 会话期间 | 跨 Agent 调用保持上下文 |
| `agent` | Agent 级别 | Agent 的长期偏好设置 |
| `tenant` | 租户级别 | 组织级共享知识 |

### 存储记忆

```bash
curl -X POST http://localhost:8080/api/v1/memory \
  -H "Content-Type: application/json" \
  -d '{
    "scope": "session",
    "agent_id": "openclaw-v1",
    "session_id": "sess_abc123",
    "key": "user_preferences",
    "value": {"theme": "dark", "language": "zh-CN"},
    "tags": ["preferences", "ui"]
  }'
```

### 检索记忆

```bash
# 按 Session 检索
curl "http://localhost:8080/api/v1/memory?scope=session&session_id=sess_abc123"

# 按 Agent 检索
curl "http://localhost:8080/api/v1/memory?scope=agent&agent_id=openclaw-v1"
```

### 自动注入

Gateway 可配置自动将 Memory Context 注入到请求中：

```bash
ENABLE_MEMORY_INJECTION=true
MEMORY_INJECTION_SCOPE=session
```

---

## Bridge Adapter 桥接适配器

### 什么是 Bridge Adapter

Bridge Adapter 将 REST API 快速转换为 MCP 工具，支持：

- OpenAPI/Swagger 规范导入
- 请求参数映射
- 响应字段过滤/裁剪
- 数组摘要

### 导入 OpenAPI 规范

```bash
curl -X POST http://localhost:8080/api/v1/gateway/bridge/import \
  -H "Content-Type: application/json" \
  -d @my_openapi_spec.json
```

### 响应优化示例

```json
{
  "type": "pick",
  "fields": ["id", "name", "status", "created_at"]
}
```

```json
{
  "type": "truncate",
  "fields": ["description", "content"],
  "max_length": 500
}
```

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

---

## 前端页面使用指南

### /gateway/sessions - Session 历史

**功能**: 查看所有 Gateway Session 记录和 Policy 决策

**使用步骤**:
1. 访问页面
2. 在搜索框输入 Agent ID 进行过滤（可选）
3. 点击任意 Session 卡片查看详情

**显示信息**:
- Session ID、Agent ID、Correlation ID
- 风险等级 (低/中/高/危)
- Policy 决策结果 (允许/拒绝)
- 状态 (活跃/已完成/已取消)
- 创建时间

---

### /skills/snapshots - Snapshot 历史

**功能**: 查看 Skill 版本的评测结果和 Regression Gate 状态

**使用步骤**:
1. 访问页面
2. 输入 Skill ID
3. 查看该 Skill 的所有 Snapshot

**显示信息**:
- Snapshot ID、Skill ID、版本号
- PASSED/FAILED 状态（绿色/红色标签）
- 综合评分 (0-100)
- 评测时间
- 详细指标 (accuracy, precision, recall, f1)

**Check 功能**:
- 可输入 Skill ID 和版本号快速检查 Gate 状态
- 返回 "PASSED" 或 "FAILED (score: xx)"

---

### /evaluations - 评估引擎

**功能**: 管理任务、执行、指标，可运行评估和查看结果

**特性**:
- 顶部统计卡片（sticky，滚动时固定）
- 双列布局：左侧任务执行，右侧任务概览
- 支持搜索和筛选

**使用步骤**:
1. 查看顶部统计：执行数、任务数、指标数、运行中数
2. 查看可用指标标签
3. 在左侧面板搜索/筛选执行记录
4. 在右侧面板搜索/筛选任务定义

**操作按钮**:
- **新建指标**: 创建新的评分指标
- **运行评估**: 选择任务触发评估执行

---

### /mcp-router - MCP 路由仪表盘

**功能**: 查看路由目录、服务器状态、信任评分

**显示信息**:
- 服务器总数、活跃数、平均信任分
- 服务器列表（名称、描述、状态、类型）
- Task Types 支持列表

---

## 故障排除

### Session 未找到

- 检查 `agent_id` 是否正确
- 确认请求已通过 Gateway 处理

### Snapshot 未通过

- 查看 `overall_score` 是否达到阈值
- 检查 `metrics` 中的详细指标
- 参考历史 Snapshot 分析趋势

### Memory 检索为空

- 确认 `scope`、`session_id`、`agent_id` 参数正确
- 检查 Memory 是否已过期（根据 TTL 配置）

### Bridge Adapter 导入失败

- 确认 OpenAPI Spec 为有效 JSON/YAML
- 检查 `servers` 字段配置
- 验证 Endpoint 定义完整

---

## 环境变量参考

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

### Bridge Adapter

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `RESPONSE_MAX_SIZE` | 最大响应大小 | 1048576 |
| `DEFAULT_TRUNCATE_LENGTH` | 默认截断长度 | 4000 |

---

## 相关文档

- [架构设计](docs/architecture.md)
- [API 参考](docs/api-reference.md)
- [部署指南](docs/deployment.md)
- [v2.0 开发计划](docs/plans/2026-04-22-managing-up-v2-evolution-review-and-delivery-plan.md)
