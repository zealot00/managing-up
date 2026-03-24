# Enterprise Operating System for Intelligence

## Data Model, Skill Schema, and Core API Design

### 1. Core Data Model

The MVP only needs a compact relational model. Keep it small and auditable.

#### 1.1 `sop_documents`

Stores raw uploaded procedure sources.

Suggested fields:

- `id` UUID PK
- `source_type` text
- `file_uri` text
- `filename` text
- `uploaded_by` text
- `parse_status` text
- `created_at` timestamptz
- `updated_at` timestamptz

#### 1.2 `procedure_drafts`

Stores structured extraction results derived from SOP documents.

Suggested fields:

- `id` UUID PK
- `sop_document_id` UUID FK
- `procedure_key` text
- `title` text
- `steps_json` jsonb
- `decision_points_json` jsonb
- `required_tools_json` jsonb
- `validation_status` text
- `validated_by` text null
- `created_at` timestamptz
- `updated_at` timestamptz

#### 1.3 `skills`

Logical skill identity.

Suggested fields:

- `id` UUID PK
- `name` text unique
- `owner_team` text
- `risk_level` text
- `status` text
- `current_version_id` UUID null
- `created_at` timestamptz
- `updated_at` timestamptz

#### 1.4 `skill_versions`

Immutable skill versions.

Suggested fields:

- `id` UUID PK
- `skill_id` UUID FK
- `version` text
- `spec_yaml` text
- `spec_json` jsonb
- `approval_policy_json` jsonb
- `change_summary` text
- `created_by` text
- `published_at` timestamptz null
- `created_at` timestamptz

#### 1.5 `executions`

Execution run header.

Suggested fields:

- `id` UUID PK
- `skill_version_id` UUID FK
- `status` text
- `triggered_by` text
- `input_json` jsonb
- `current_step_id` text null
- `started_at` timestamptz null
- `ended_at` timestamptz null
- `duration_ms` bigint null
- `estimated_cost_usd` numeric(12,4) null
- `failure_reason` text null
- `created_at` timestamptz

#### 1.6 `execution_steps`

Execution details for each step.

Suggested fields:

- `id` UUID PK
- `execution_id` UUID FK
- `step_id` text
- `step_type` text
- `status` text
- `attempt_no` int
- `tool_ref` text null
- `input_json` jsonb
- `output_json` jsonb null
- `error_message` text null
- `started_at` timestamptz null
- `ended_at` timestamptz null
- `duration_ms` bigint null

#### 1.7 `approval_checkpoints`

Approval trail for human-in-the-loop actions.

Suggested fields:

- `id` UUID PK
- `execution_id` UUID FK
- `step_id` text
- `status` text
- `approver_group` text
- `requested_at` timestamptz
- `resolved_at` timestamptz null
- `resolved_by` text null
- `resolution_note` text null

#### 1.8 `tool_definitions`

Metadata for allowed tool integrations.

Suggested fields:

- `id` UUID PK
- `tool_key` text unique
- `tool_type` text
- `display_name` text
- `policy_json` jsonb
- `secret_ref` text null
- `owner_team` text
- `status` text
- `created_at` timestamptz

### 2. Minimal Entity Relationships

```text
sop_documents 1---n procedure_drafts
skills 1---n skill_versions
skill_versions 1---n executions
executions 1---n execution_steps
executions 1---n approval_checkpoints
```

### 3. Suggested Status Enums

Keep status values explicit and finite.

#### `parse_status`

- `uploaded`
- `parsing`
- `parsed`
- `failed`

#### `validation_status`

- `draft`
- `validated`
- `rejected`

#### `skill.status`

- `draft`
- `published`
- `deprecated`

#### `execution.status`

- `pending`
- `running`
- `waiting_approval`
- `succeeded`
- `failed`
- `stopped`

#### `execution_step.status`

- `pending`
- `running`
- `succeeded`
- `failed`
- `skipped`
- `waiting_approval`

### 4. Skill YAML Schema for MVP

The Skill DSL should remain intentionally small.

```yaml
name: restart_service_skill
version: v1
risk_level: medium
description: Restart a service after pre-checks and approval.

inputs:
  - name: server_id
    type: string
    required: true
    description: Target server identifier

steps:
  - id: check_cpu
    type: tool
    tool_ref: monitor_api.check_cpu
    with:
      server_id: "{{ inputs.server_id }}"
    timeout_seconds: 30

  - id: approval_before_restart
    type: approval
    required: true
    approver_group: sre_oncall
    message: Confirm restart is safe to proceed

  - id: restart_service
    type: tool
    tool_ref: orchestrator.restart_service
    with:
      server_id: "{{ inputs.server_id }}"
    retry_policy:
      max_attempts: 2
      backoff_seconds: 10

  - id: verify_health
    type: tool
    tool_ref: monitor_api.verify_health
    with:
      server_id: "{{ inputs.server_id }}"
    timeout_seconds: 120

on_failure:
  action: mark_failed
```

#### Allowed top-level fields

- `name`
- `version`
- `risk_level`
- `description`
- `inputs`
- `steps`
- `on_failure`

#### Allowed step types

- `tool`
- `approval`
- `condition`

MVP should not support:

- loops
- dynamic step generation
- autonomous planning
- arbitrary script blocks in the spec

### 5. Tool Invocation Envelope

The runtime should call the Tool Gateway with a normalized payload.

```json
{
  "execution_id": "exec_123",
  "step_id": "restart_service",
  "tool_ref": "orchestrator.restart_service",
  "input": {
    "server_id": "srv-001"
  },
  "timeout_seconds": 120,
  "attempt_no": 1
}
```

The Tool Gateway should return:

```json
{
  "status": "succeeded",
  "output": {
    "job_id": "job_789"
  },
  "started_at": "2026-03-19T10:00:00Z",
  "ended_at": "2026-03-19T10:00:05Z"
}
```

### 6. Core API Design

The MVP API surface should favor clear state transitions over generic CRUD.

#### SOP ingestion

`POST /api/sop-documents`

- uploads a document and creates `sop_document`

`POST /api/sop-documents/{id}/parse`

- triggers parsing and draft extraction

`GET /api/procedure-drafts/{id}`

- returns structured draft for human review

`POST /api/procedure-drafts/{id}/validate`

- marks the draft as human-validated

#### Skill authoring and registry

`POST /api/skills`

- creates a logical skill record

`POST /api/skills/{id}/versions`

- creates a new draft version

`POST /api/skill-versions/{id}/publish`

- publishes an immutable executable version

`GET /api/skills/{id}`

- returns skill metadata and current version pointer

`GET /api/skills/{id}/versions`

- returns version history

#### Execution

`POST /api/executions`

- starts a run from a published skill version

Request example:

```json
{
  "skill_version_id": "sv_001",
  "input": {
    "server_id": "srv-001"
  }
}
```

`GET /api/executions/{id}`

- returns run summary and current state

`GET /api/executions/{id}/steps`

- returns step-level timeline

`POST /api/executions/{id}/approve`

- resolves the active approval checkpoint

`POST /api/executions/{id}/resume`

- resumes a paused run after operator intervention

`POST /api/executions/{id}/stop`

- triggers emergency stop

#### Tool administration

`POST /api/tools`

- registers an approved tool definition

`GET /api/tools`

- lists available tools and policies

### 7. API Response Shape Recommendation

Use explicit envelopes to simplify frontend state handling.

```json
{
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_123"
  }
}
```

On failures:

```json
{
  "data": null,
  "error": {
    "code": "APPROVAL_REQUIRED",
    "message": "Execution is waiting for approval."
  },
  "meta": {
    "request_id": "req_124"
  }
}
```

### 8. Security Constraints for MVP

These constraints should be encoded in the platform from day one.

- skill specs must not contain raw secrets
- shell tools must use predefined command templates
- SQL tools must use scoped credentials and read-only modes where possible
- all approval and override actions must be attributed to a human identity
- every tool invocation must be auditable by execution and step id

### 9. Recommended Next Step After MVP

Once this model is stable, the next evolution should be:

- reusable skill templates
- richer conditional expressions
- policy engine integration
- stronger rollback primitives

That should happen only after the base runtime, registry, and observability model are stable.

### 10. Stack Mapping for This Repository

The current repository should implement the design with:

- `apps/api`: Go backend
- `apps/web`: Next.js + TypeScript frontend

Suggested ownership mapping:

- API routes, orchestration, registry, approvals: Go
- operator console, dashboards, editor, timeline views: Next.js + TypeScript
