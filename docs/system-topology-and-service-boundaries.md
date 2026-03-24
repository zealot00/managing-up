# Enterprise Operating System for Intelligence

## MVP Topology and Service Boundaries

### 1. Service Map

The MVP should be decomposed into six primary services and three infrastructure layers.

#### Primary Services

1. `ops-console`
2. `ingestion-service`
3. `composer-service`
4. `registry-service`
5. `runtime-service`
6. `tool-gateway`

#### Shared Infrastructure

1. `postgres`
2. `redis`
3. `object-storage`

#### Security and Ops Components

1. `secrets-manager`
2. `worker-sandbox`
3. `otel/log-store`

### 2. Boundary Definition

#### `ops-console`

Owns:

- UI for SOP upload
- procedure review
- skill editing
- version publishing
- execution monitoring
- approval operations

Does not own:

- orchestration logic
- tool execution
- document parsing

#### `ingestion-service`

Owns:

- file intake metadata
- document parsing pipeline
- extraction prompts
- procedure draft generation

Inputs:

- raw SOP document

Outputs:

- `procedure_draft`

Does not own:

- skill publishing
- execution

#### `composer-service`

Owns:

- draft-to-skill transformation
- editor validation rules
- spec preview
- dry-run validation hooks

Inputs:

- `procedure_draft`

Outputs:

- draft `skill_version`

Does not own:

- registry state authority
- runtime scheduling

#### `registry-service`

Owns:

- skill metadata
- version graph
- ownership
- publish status
- risk classification
- approval policy binding

Inputs:

- skill version draft

Outputs:

- published immutable skill version

Does not own:

- document parsing
- step execution

#### `runtime-service`

Owns:

- execution state machine
- orchestration of steps
- retry and timeout policy
- approval wait state
- event emission

Inputs:

- published skill version
- execution inputs

Outputs:

- `execution_run`
- `execution_step_run`
- audit events

Does not own:

- free-form planning
- direct secrets lookup by operators
- uncontrolled tool invocation

#### `tool-gateway`

Owns:

- tool adapter abstraction
- secret retrieval
- policy enforcement
- shell sandbox contract
- normalized result envelopes

Inputs:

- step invocation request from runtime

Outputs:

- structured tool result
- adapter logs

Does not own:

- workflow orchestration
- UI concerns

### 3. Runtime Interaction Diagram

```text
                  +------------------+
                  |   Ops Console    |
                  +---------+--------+
                            |
                            v
                  +------------------+
                  | Registry Service |
                  +---------+--------+
                            |
                            v
                  +------------------+
                  | Runtime Service  |
                  +----+--------+----+
                       |        |
                       |        +----------------------+
                       |                               |
                       v                               v
             +------------------+            +------------------+
             |      Redis       |            | Approval Events  |
             +------------------+            +------------------+
                       |
                       v
                  +------------------+
                  |  Tool Gateway    |
                  +----+-------+-----+
                       |       | 
                       |       |
                       v       v
                  +-------+ +------+
                  | HTTP  | | SQL  |
                  +-------+ +------+
                       |
                       v
                 +-------------+
                 | Shell Sand. |
                 +-------------+
```

### 4. Control and Trust Boundaries

The MVP should enforce four hard boundaries.

#### Boundary A: Parsing vs Execution

- parsed SOP output is never executable by default
- a human must validate before skill creation

#### Boundary B: Draft vs Published Skill

- runtime executes only immutable published versions
- draft versions cannot be scheduled

#### Boundary C: Runtime vs Tool Access

- runtime cannot bypass Tool Gateway
- secrets and command templates remain outside runtime spec

#### Boundary D: Automation vs Human Control

- approval-required steps suspend execution
- emergency stop can terminate an active run
- manual override actions are separately audited

### 5. Deployment Recommendation

For MVP, keep deployment simple:

- deploy each service as a container
- run Postgres, Redis, and object storage as managed services if available
- use Docker-based shell sandbox workers
- keep frontend and backend as separated deployable applications
- avoid Kubernetes-native job orchestration unless pilot scale requires it

### 5.1 Frontend and Backend Separation

The implementation should use a clear application split:

- `apps/web` hosts the Next.js + TypeScript control plane
- `apps/api` hosts the Go API and service entrypoint

The frontend should consume backend APIs over HTTP rather than sharing runtime concerns. This keeps responsibilities clean:

- web handles operator workflows and presentation
- api handles orchestration, policy, and execution state

### 6. Suggested Internal Event Types

The runtime and UI should exchange normalized events.

- `execution.created`
- `execution.started`
- `execution.step.started`
- `execution.step.succeeded`
- `execution.step.failed`
- `execution.waiting_approval`
- `execution.approved`
- `execution.resumed`
- `execution.stopped`
- `execution.succeeded`
- `execution.failed`

### 7. MVP Risks by Boundary

#### Weak parsing boundary

Risk:

- incorrect SOP extraction becomes unsafe automation

Mitigation:

- mandatory human review
- explicit draft status

#### Weak runtime boundary

Risk:

- runtime gains too much dynamic intelligence and becomes hard to govern

Mitigation:

- deterministic spec only
- no planning features in MVP

#### Weak tool boundary

Risk:

- shell and SQL become an enterprise escape hatch

Mitigation:

- allowlists
- scoped credentials
- execution sandbox

### 8. MVP Success Condition

This topology is sufficient if it can support:

- at least 3 real SOPs as published skills
- reliable execution with approval checkpoints
- execution history with auditability
- basic observability for cost, duration, and failure analysis
