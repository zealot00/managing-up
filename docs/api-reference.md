# Skill Hub EE API Reference

## Overview

This document defines the MVP HTTP contract for the Go backend and Next.js frontend.

Base URL:

- local: `http://localhost:8080`
- prefix: `/api/v1`

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

## Resource Summary

### Core Endpoints
- `GET /healthz`
- `GET /health`
- `GET /api/v1/meta`

### LLM Gateway (OpenAI/Anthropic Compatible)
- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/messages`

### SOP-to-Skill Orchestrator API
- `GET /v1/healthz`
- `POST /v1/runs`
- `GET /v1/runs/{runId}`
- `GET /v1/runs/{runId}/artifacts`
- `POST /v1/extraction/enhance`
- `POST /v1/extraction/compare`
- `POST /v1/skills`
- `GET /v1/skills/{skillId}/versions`
- `GET /v1/skills/{skillId}/versions/{version}`
- `GET /v1/skills/{skillId}/diff`
- `POST /v1/skills/{skillId}/rollback`
- `POST /v1/skills/{skillId}/promote`
- `POST /v1/tests/runs`
- `GET /v1/tests/runs/{testRunId}`
- `GET /v1/tests/runs/{testRunId}/report`
- `POST /v1/gates/evaluate`
- `GET /v1/policies/{policyId}`

### Skill Registry API
- `GET /api/v1/dashboard`
- `GET /api/v1/procedure-drafts`
- `GET /api/v1/skills`
- `POST /api/v1/skills`
- `GET /api/v1/skills/{skillId}`
- `GET /api/v1/skill-versions`
- `GET /api/v1/executions`
- `POST /api/v1/executions`
- `GET /api/v1/executions/{executionId}`
- `POST /api/v1/executions/{executionId}/approve`
- `GET /api/v1/approvals`

## 1. Health and Meta

### `GET /healthz`

- liveness endpoint

### `GET /api/v1/meta`

- service identity and scope metadata

## 2. Dashboard

### `GET /api/v1/dashboard`

- homepage summary payload

Fields:

- `summary.active_skills`
- `summary.published_versions`
- `summary.running_executions`
- `summary.waiting_approvals`
- `summary.success_rate`
- `summary.avg_duration_seconds`
- `recent_executions`

## 3. Procedure Drafts

### `GET /api/v1/procedure-drafts`

- list parsed SOP drafts

Query params:

- `status` optional

Response item shape:

```json
{
  "id": "draft_001",
  "procedure_key": "runbook_restart_service",
  "title": "Restart Service Runbook",
  "validation_status": "validated",
  "required_tools": ["monitor_api", "orchestrator_api"],
  "source_type": "markdown",
  "created_at": "2026-03-19T10:00:00Z"
}
```

## 4. Skills

### `GET /api/v1/skills`

- list registry entries

Query params:

- `status` optional

### `POST /api/v1/skills`

- create a draft skill metadata record

Request:

```json
{
  "name": "rollback_deployment_skill",
  "owner_team": "platform_team",
  "risk_level": "high"
}
```

Validation:

- `name` required
- `owner_team` required
- `risk_level` in `low | medium | high`

### `GET /api/v1/skills/{skillId}`

- fetch a single skill

## 5. Skill Versions

### `GET /api/v1/skill-versions`

- list immutable skill versions

Query params:

- `skill_id` optional

Response item shape:

```json
{
  "id": "version_001",
  "skill_id": "skill_001",
  "version": "v1",
  "status": "published",
  "change_summary": "Initial restart automation flow.",
  "approval_required": true,
  "created_at": "2026-03-19T10:00:00Z"
}
```

## 6. Executions

### `GET /api/v1/executions`

- list execution runs

Query params:

- `status` optional

### `POST /api/v1/executions`

- create a new execution

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

### `GET /api/v1/executions/{executionId}`

- fetch one execution

### `POST /api/v1/executions/{executionId}/approve`

- resolve an approval checkpoint bound to the execution

Request:

```json
{
  "approver": "ops_manager",
  "decision": "approved",
  "note": "safe to continue"
}
```

Validation:

- `approver` required
- `decision` in `approved | rejected`

## 7. Approvals

### `GET /api/v1/approvals`

- list approval checkpoints

Query params:

- `status` optional

Response item shape:

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

## 8. Error Codes

- `BAD_REQUEST`
- `NOT_FOUND`
- `METHOD_NOT_ALLOWED`
- `UNSUPPORTED_MEDIA_TYPE`
- `INTERNAL_ERROR`

## 11. LLM Gateway

Multi-provider LLM proxy with OpenAI and Anthropic API compatibility.

### Authentication

All endpoints require API key authentication via:
- `Authorization: Bearer <key>` header (OpenAI style)
- `x-api-key: <key>` header (Anthropic style)

Public endpoints (no auth required):
- `GET /health`
- `GET /v1/models`

### `GET /v1/models`

List available models across all providers.

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

OpenAI-compatible chat completions endpoint.

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
  "max_tokens": 1024
}
```

Response:
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

Model prefix routing:
- `openai:gpt-4o` or just `gpt-4o` → OpenAI
- `anthropic:claude-sonnet-4` or `claude-sonnet-4` → Anthropic
- `gemini:gemini-2.0-flash` → Google
- `deepseek:deepseek-chat` → DeepSeek

### `POST /v1/messages`

Anthropic-compatible messages endpoint.

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
- `unauthorized` - Missing or invalid API key
- `invalid_request` - Malformed request body
- `invalid_model` - Unknown model identifier
- `generation_failed` - LLM provider error

## 12. SOP-to-Skill Orchestrator API

CLI orchestration API for remote enhancement, skill versioning, test orchestration, and gate evaluation.

Base path: `/v1`

### Authentication

Uses Bearer JWT token authentication.

### Common Headers

- `Idempotency-Key` - Optional, for write operations
- `Content-Type: application/json`

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

Run statuses: `queued`, `running`, `succeeded`, `failed`, `canceled`
Run stages: `extraction`, `generation`, `validation`, `versioning`, `testing`, `completed`

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

Artifact kinds: `skill_md`, `full_skill_md`, `schema_json`, `manifest_yaml`, `constraints_dir`, `framework_bundle`

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

### Skill Version Management

#### `POST /v1/skills`

Create skill metadata.

Request:
```json
{
  "skillId": "skill_xxx",
  "name": "My Skill",
  "owner": "team-x",
  "tags": ["production", "critical"]
}
```

Response (201 Created):
```json
{
  "skillId": "skill_xxx",
  "name": "My Skill",
  "owner": "team-x",
  "tags": ["production", "critical"],
  "createdAt": "2026-03-31T10:00:00Z"
}
```

#### `POST /v1/skills/{skillId}/versions`

Create new skill version.

Request:
```json
{
  "version": "1.1.0",
  "sourceHash": "sha256:abc123",
  "schemaHash": "sha256:def456",
  "artifacts": [
    {"kind": "skill_md", "uri": "s3://bucket/v1.1.0/skill.md"}
  ],
  "runId": "run_xxx"
}
```

#### `GET /v1/skills/{skillId}/versions`

List all versions.

Response:
```json
{
  "skillId": "skill_xxx",
  "versions": [...]
}
```

#### `GET /v1/skills/{skillId}/diff?from=v1.0.0&to=v1.1.0`

Compare two versions.

### Async Operations

#### `POST /v1/skills/{skillId}/rollback`

Rollback to previous version.

Request:
```json
{
  "targetVersion": "1.0.0",
  "reason": "Critical bug in v1.1.0"
}
```

Response (202 Accepted):
```json
{
  "actionId": "action_xxx",
  "status": "accepted",
  "acceptedAt": "2026-03-31T10:00:00Z"
}
```

#### `POST /v1/skills/{skillId}/promote`

Promote version to environment.

Request:
```json
{
  "version": "1.1.0",
  "channel": "staging"
}
```

Response (202 Accepted):
```json
{
  "actionId": "action_xxx",
  "status": "accepted",
  "acceptedAt": "2026-03-31T10:00:00Z"
}
```

Channels: `dev`, `staging`, `prod`

### Test Orchestration

#### `POST /v1/tests/runs`

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

#### `GET /v1/tests/runs/{testRunId}`

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

#### `GET /v1/tests/runs/{testRunId}/report`

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

### Gate Evaluation

#### `POST /v1/gates/evaluate`

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

### Policies

#### `GET /v1/policies/{policyId}`

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

### Error Response Format

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Skill not found",
    "details": {},
    "requestId": "req_xxx"
  }
}
```

## 9. Frontend Integration Notes

Frontend pages should map to these endpoints:

- `/` -> `dashboard`
- `/skills` -> `skills` + `skill-versions`
- `/executions` -> `executions`
- `/approvals` -> `approvals`

## 10. Persistence Note

These HTTP contracts are designed to remain stable when the backend switches from the in-memory repository to PostgreSQL.

Migration files live in:

- `apps/api/migrations/0001_init.up.sql`
- `apps/api/migrations/0001_init.down.sql`
