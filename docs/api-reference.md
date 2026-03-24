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

- `GET /healthz`
- `GET /api/v1/meta`
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
