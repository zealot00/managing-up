# Skill Hub EE API Reference

## Overview

Base URL: `http://localhost:8080`

Response envelope:

```json
{
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_123"
  }
}
```

Error envelope:

```json
{
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  },
  "meta": {
    "request_id": "req_123"
  }
}
```

---

## 1. 认证 Authentication

### `POST /api/v1/auth/login`

用户登录。

Request:
```json
{
  "username": "admin",
  "password": "admin"
}
```

Response (200):
```json
{
  "data": {
    "user": {
      "id": "user_admin",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

Sets authentication cookie on success.

### `POST /api/v1/auth/logout`

用户登出。清除认证 cookie。

### `GET /api/v1/auth/me`

获取当前用户信息。需要认证。

Response (200):
```json
{
  "data": {
    "user": {
      "id": "user_admin",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

---

## 2. 健康检查 Health

### `GET /healthz`

Liveness probe.

Response:
```json
{
  "status": "ok"
}
```

### `GET /api/v1/meta`

Service metadata.

Response:
```json
{
  "data": {
    "service": "managing-up-api",
    "runtime": "go",
    "scope": ["registry", "execution", "approval"]
  }
}
```

---

## 3. Tips (登录页名言)

### `GET /api/v1/tip`

获取随机名言。无需认证。每次随机返回一条 active 的 tip。

Response (200):
```json
{
  "data": {
    "id": "tip_001",
    "content": "Talk is cheap. Show me the code.",
    "author": "Linus Torvalds",
    "category": "quote",
    "is_active": true,
    "created_at": "2026-03-19T10:00:00Z",
    "updated_at": "2026-03-19T10:00:00Z"
  }
}
```

Tip 数据存储在 `tips` 表中，可通过 SQL 直接管理：

```sql
-- 添加新 tip
INSERT INTO tips (id, content, author, category)
VALUES ('tip_new', 'Your quote here', 'Author', 'quote');

-- 禁用某条 tip
UPDATE tips SET is_active = false WHERE id = 'tip_005';
```

---

## 4. Dashboard

### `GET /api/v1/dashboard`

首页统计摘要。

Response (200):
```json
{
  "data": {
    "summary": {
      "active_skills": 3,
      "published_versions": 2,
      "running_executions": 1,
      "waiting_approvals": 1,
      "success_rate": 0.75,
      "avg_duration_seconds": 120
    },
    "recent_executions": [...]
  }
}
```

---

## 5. Skills

### `GET /api/v1/skills`

列出技能。

Query params: `status` (可选)

### `POST /api/v1/skills`

创建技能。

Request:
```json
{
  "name": "rollback_deployment_skill",
  "owner_team": "platform_team",
  "risk_level": "high"
}
```

### `GET /api/v1/skills/{id}`

技能详情。

### `GET /api/v1/skills/{id}/spec`

下载 YAML spec 文件。

### `GET /api/v1/skill-versions`

列出技能版本。

Query params: `skill_id` (可选)

---

## 6. Executions

### `GET /api/v1/executions`

列出执行记录。

Query params: `status` (可选)

### `POST /api/v1/executions`

触发执行。

Request:
```json
{
  "skill_id": "skill_001",
  "triggered_by": "platform_operator",
  "input": {
    "server_id": "srv-001"
  }
}
```

### `GET /api/v1/executions/{id}`

执行详情。

### `POST /api/v1/executions/{id}/approve`

审批/拒绝。

Request:
```json
{
  "approver": "ops_manager",
  "decision": "approved",
  "note": "safe to continue"
}
```

---

## 7. Approvals

### `GET /api/v1/approvals`

列出审批。

Query params: `status` (可选)

Response item:
```json
{
  "id": "approval_001",
  "execution_id": "exec_002",
  "skill_name": "collect_logs_skill",
  "step_id": "approval_before_export",
  "status": "waiting",
  "approver_group": "ops_manager",
  "requested_at": "2026-03-19T10:00:00Z"
}
```

---

## 8. Tasks

### `GET /api/v1/tasks`

列出任务。

Query params: `skill_id`, `difficulty` (可选)

### `POST /api/v1/tasks`

创建任务。

### `GET /api/v1/tasks/{id}`

任务详情。

### `PUT /api/v1/tasks/{id}`

更新任务。

### `DELETE /api/v1/tasks/{id}`

删除任务。

### `GET /api/v1/metrics`

列出指标定义。

### `GET /api/v1/task-executions`

任务执行列表。

### `GET /api/v1/task-executions/{id}`

任务执行详情。

---

## 9. Experiments

### `GET /api/v1/experiments`

列出实验。

### `POST /api/v1/experiments`

创建实验。

### `GET /api/v1/experiments/{id}`

实验结果。

### `GET /api/v1/experiments/{id}/compare?compare_with={other_id}`

对比两个实验。

### `POST /api/v1/check-regression`

回归检测。

Request:
```json
{
  "baseline_experiment_id": "exp_001",
  "candidate_experiment_id": "exp_002"
}
```

---

## 10. Replay

### `GET /api/v1/replay-snapshots`

回放快照列表。

Query params: `execution_id` (可选)

### `GET /api/v1/replay-snapshots/{id}`

快照详情。

---

## 11. LLM Gateway

Multi-provider LLM proxy with OpenAI and Anthropic API compatibility.

### Authentication

所有端点需要 API Key 认证：
- `Authorization: Bearer <key>` (OpenAI 风格)
- `x-api-key: <key>` (Anthropic 风格)

公开端点（无需认证）：
- `GET /health`
- `GET /v1/models`

### `GET /v1/models`

列出可用模型。

Response:
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1774954597,
      "owned_by": "openai",
      "provider": "openai"
    }
  ]
}
```

### `POST /v1/chat/completions`

OpenAI 兼容接口。支持流式响应。

Headers:
- `Authorization: Bearer <key>` required
- `Content-Type: application/json`

Request:
```json
{
  "model": "gpt-4o-mini",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

Response (non-streaming):
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1774954597,
  "model": "gpt-4o-mini",
  "choices": [
    {
      "index": 0,
      "message": {"role": "assistant", "content": "Hi there!"},
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 8,
    "total_tokens": 18
  }
}
```

Response (streaming, `stream: true`):
```
data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1774954597,"model":"gpt-4o-mini","choices":[{"index":0,"delta":{"content":"Hi"},"finish_reason":null}]}

data: {"id":"chatcmpl-xxx","object":"chat.completion.chunk","created":1774954597,"model":"gpt-4o-mini","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

Model prefix routing:
- `openai:gpt-4o` 或 `gpt-4o` → OpenAI
- `anthropic:claude-sonnet-4` 或 `claude-sonnet-4` → Anthropic
- `gemini:gemini-2.0-flash` → Google
- `deepseek:deepseek-chat` → DeepSeek

### `POST /v1/messages`

Anthropic 兼容接口。

Headers:
- `x-api-key: <key>` required
- `anthropic-version: 2023-06-01`
- `Content-Type: application/json`

Request:
```json
{
  "model": "claude-sonnet-4-20250514",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024
}
```

### `POST /v1/embeddings`

Embeddings 接口。

Request:
```json
{
  "model": "text-embedding-3-small",
  "input": ["Hello world"]
}
```

### Gateway 管理 API

#### `GET /api/v1/gateway/keys`

列出当前用户的 Gateway API Keys。

Response:
```json
{
  "data": {
    "items": [
      {
        "id": "key_xxx",
        "user_id": "user_admin",
        "name": "default",
        "key_prefix": "sk_live",
        "created_at": "2026-04-01T10:00:00Z",
        "last_used_at": "2026-04-02T10:00:00Z",
        "revoked_at": null
      }
    ]
  }
}
```

#### `POST /api/v1/gateway/keys`

创建 Gateway API Key。

Request:
```json
{
  "name": "my-key"
}
```

Response:
```json
{
  "data": {
    "key": "sk_live_xxx",
    "key_meta": {
      "id": "key_xxx",
      "name": "my-key",
      "key_prefix": "sk_live"
    }
  }
}
```

#### `DELETE /api/v1/gateway/keys/{id}`

撤销 API Key。

#### `GET /api/v1/gateway/usage`

使用统计（按 Provider/Model 聚合）。

Query params:
- `from` - 起始日期 (ISO 8601)
- `to` - 截止日期 (ISO 8601)
- `user_id` - 指定用户 (admin)

Response item:
```json
{
  "user_id": "user_admin",
  "username": "admin",
  "provider": "openai",
  "model": "gpt-4o",
  "request_count": 100,
  "prompt_tokens": 50000,
  "completion_tokens": 100000,
  "total_tokens": 150000,
  "total_cost": 2.25
}
```

#### `GET /api/v1/gateway/usage/users`

用户级别使用统计（admin）。

Response item:
```json
{
  "user_id": "user_admin",
  "username": "admin",
  "request_count": 200,
  "prompt_tokens": 100000,
  "completion_tokens": 200000,
  "total_tokens": 300000,
  "total_cost": 4.50
}
```

### Gateway 功能

| 功能 | 说明 |
|------|------|
| **多提供商路由** | 根据模型名自动路由到对应提供商 |
| **流式响应** | SSE 格式实时返回生成内容 |
| **使用统计** | 按 Provider/Model/User 聚合 token 和费用 |
| **费用追踪** | 基于模型定价自动计算成本 |
| **限流** | 每 Key 每分钟请求限制 (默认 60) |
| **重试机制** | 指数退避自动重试 (最多 3 次) |

### Error Responses

```json
{
  "error": {
    "code": "unauthorized",
    "message": "API key is required"
  }
}
```

Error codes:
- `unauthorized` - 缺少或无效 API Key
- `invalid_request` - 请求格式错误
- `invalid_model` - 未知模型标识
- `generation_failed` - LLM 提供商错误
- `stream_failed` - 流式响应失败
- `rate_limit_exceeded` - 超过请求限制

---

## 12. SOP-to-Skill Orchestrator API

CLI orchestration API for remote enhancement, skill versioning, test orchestration, and gate evaluation.

Base path: `/v1`

### Authentication

Uses Bearer JWT token authentication.

### `GET /v1/healthz`

Health check endpoint.

Response:
```json
{
  "status": "ok",
  "service": "sop-skill-orchestrator",
  "version": "1.0.0",
  "time": "2026-03-31T10:00:00Z"
}
```

### `POST /v1/runs`

Create an async orchestration run.

Request:
```json
{
  "skillName": "my-skill",
  "source": {
    "type": "inline_text",
    "content": "# SOP Content..."
  },
  "options": {
    "framework": "all",
    "extraction": {
      "language": "auto",
      "confidenceThreshold": 0.7
    }
  }
}
```

Response (202 Accepted):
```json
{
  "runId": "run_xxx",
  "status": "queued",
  "createdAt": "2026-03-31T10:00:00Z",
  "links": {
    "self": "/v1/runs/run_xxx"
  }
}
```

### `GET /v1/runs/{runId}`

Get run detail.

Response:
```json
{
  "runId": "run_xxx",
  "status": "succeeded",
  "stage": "completed",
  "skillName": "my-skill",
  "createdAt": "2026-03-31T10:00:00Z",
  "updatedAt": "2026-03-31T10:05:00Z",
  "result": {
    "skillId": "skill_xxx",
    "version": "1.0.0",
    "artifacts": [
      {"kind": "skill_md", "uri": "s3://bucket/skill.md"}
    ]
  },
  "errors": []
}
```

### `GET /v1/runs/{runId}/artifacts`

List run artifacts.

Response:
```json
{
  "runId": "run_xxx",
  "artifacts": [
    {"kind": "skill_md", "uri": "s3://bucket/skill.md"},
    {"kind": "schema_json", "uri": "s3://bucket/schema.json"},
    {"kind": "manifest_yaml", "uri": "s3://bucket/manifest.yaml"}
  ]
}
```

### `POST /v1/extraction/enhance`

Enhanced extraction from raw SOP text.

Request:
```json
{
  "source": {
    "type": "inline_text",
    "content": "# SOP Content..."
  },
  "options": {
    "language": "auto",
    "confidenceThreshold": 0.7
  }
}
```

Response:
```json
{
  "constraints": [
    {
      "id": "c1",
      "level": "MUST",
      "description": "User must be authenticated",
      "roles": ["user"],
      "confidence": 0.95
    }
  ],
  "decisions": [...],
  "roles": [...],
  "boundaries": [...],
  "modelInfo": {
    "provider": "openai",
    "model": "gpt-4o",
    "latencyMs": 1500
  }
}
```

### `POST /v1/extraction/compare`

Compare local and enhanced extraction.

Request:
```json
{
  "local": {...},
  "remote": {...}
}
```

Response:
```json
{
  "summary": {
    "constraintDelta": 2,
    "decisionDelta": -1,
    "roleDelta": 0
  },
  "diffs": [
    {"type": "constraint", "detail": "Added: ..."}
  ]
}
```

### `POST /v1/tests/runs`

Create test run.

Request:
```json
{
  "skillId": "skill_xxx",
  "version": "1.1.0",
  "runner": {
    "type": "cli",
    "command": "seh",
    "args": ["test", "--skill", "skill_xxx"]
  },
  "datasetRef": "s3://datasets/test-cases.csv"
}
```

Response (202 Accepted):
```json
{
  "testRunId": "test_xxx",
  "status": "queued"
}
```

### `GET /v1/tests/runs/{testRunId}`

Get test run status.

Response:
```json
{
  "testRunId": "test_xxx",
  "status": "succeeded",
  "createdAt": "2026-03-31T10:00:00Z",
  "updatedAt": "2026-03-31T10:05:00Z",
  "exitCode": 0
}
```

### `GET /v1/tests/runs/{testRunId}/report`

Get test report.

Response:
```json
{
  "testRunId": "test_xxx",
  "passed": true,
  "metrics": {
    "totalCases": 50,
    "passRate": 0.96,
    "regressions": 1
  },
  "failures": []
}
```

### `POST /v1/gates/evaluate`

Evaluate promotion gate.

Request:
```json
{
  "skillId": "skill_xxx",
  "version": "1.1.0",
  "policyId": "policy_001",
  "testRunId": "test_xxx"
}
```

Response:
```json
{
  "passed": true,
  "policyId": "policy_001",
  "reasons": ["All gates passed"],
  "decisionAt": "2026-03-31T10:00:00Z"
}
```

### `GET /v1/policies/{policyId}`

Get policy definition.

Response:
```json
{
  "policyId": "policy_001",
  "name": "Standard Promotion Policy",
  "rules": [
    {"metric": "test_pass_rate", "op": "gte", "value": 0.9},
    {"metric": "regressions", "op": "lte", "value": 0}
  ]
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | 请求格式错误 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `METHOD_NOT_ALLOWED` | 405 | 请求方法不支持 |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | Content-Type 不支持 |
| `UNAUTHORIZED` | 401 | 未认证或认证失败 |
| `INVALID_CREDENTIALS` | 401 | 用户名或密码错误 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
