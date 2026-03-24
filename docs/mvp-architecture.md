# Enterprise Operating System for Intelligence

## MVP Architecture Design

### 1. Product Positioning

This MVP should be built as a governed SOP execution system, not as a general-purpose autonomous agent platform.

Its job is to transform enterprise SOPs into executable Skills that can be:

- reviewed by humans
- versioned and approved
- executed safely
- observed end to end
- measured for cost and ROI

### 2. MVP Design Principles

The architecture should optimize for control and trust before intelligence depth.

Key principles:

- Human validation is mandatory between SOP parsing and executable skill release.
- Runtime only executes published skill versions.
- All external tool access goes through a controlled gateway.
- High-risk steps support pause, approval, override, and stop.
- Execution is deterministic and spec-driven.
- AI is used for ingestion assistance, not for runtime autonomy.

### 3. Core MVP Modules

#### 3.1 SOP Ingestion Service

Purpose:

- ingest SOP documents from PDF, Word, Markdown, Confluence export, or spreadsheets
- parse content into a structured procedure draft
- extract steps, conditional hints, and candidate tools

Responsibilities:

- file intake
- document parsing
- LLM-assisted step extraction
- structured draft generation

Non-goals:

- direct execution
- version publishing

Suggested implementation:

- Go service for MVP stack consistency
- split this module later only if document parsing needs a separate specialized stack

#### 3.2 Skill Composer

Purpose:

- convert a validated procedure draft into a Skill Spec
- allow human editing before publish

Responsibilities:

- visual step editing
- YAML editing
- test-run entry point
- approval policy selection

Non-goals:

- runtime orchestration

#### 3.3 Skill Registry

Purpose:

- act as the source of truth for skill definitions and versions

Responsibilities:

- skill metadata management
- version storage
- publish and deprecate operations
- ownership and risk metadata
- linking skills to execution history

#### 3.4 Runtime Service

Purpose:

- execute a published Skill Spec deterministically

Responsibilities:

- step orchestration
- retries and timeout handling
- state transitions
- approval pause/resume
- structured logging
- failure handling

Supported execution model:

- sequential steps
- basic conditional branching
- approval checkpoints
- rollback trigger hooks

Non-goals:

- planning
- dynamic tool discovery
- memory-based reasoning

Suggested implementation:

- Go service

#### 3.5 Tool Gateway

Purpose:

- provide a single secure interface for all executable tools

Supported MVP tool types:

- HTTP API
- Shell command
- SQL query

Responsibilities:

- secret injection
- permission enforcement
- command allowlisting
- output normalization
- tool-level audit logging

#### 3.6 Ops Console

Purpose:

- provide the human control plane

Views:

- SOP review
- Skill editor
- Skill registry
- execution timeline
- approval inbox
- observability dashboard

### 4. Control Plane and Data Plane

The MVP should be split into control-plane and data-plane responsibilities.

#### Control Plane

- Ops Console
- SOP Ingestion Service
- Skill Composer
- Skill Registry

This layer manages creation, review, approval, and release of executable skills.

#### Data Plane

- Runtime Service
- Tool Gateway
- Worker Sandbox

This layer manages actual execution and tool interaction.

### 5. Recommended System Topology

```text
User / Operator
   |
Ops Console
   |
   +--> SOP Ingestion Service --> Object Storage
   |             |
   |             +--> Parser / LLM Extraction
   |
   +--> Skill Composer --> Skill Registry --> PostgreSQL
   |
   +--> Approval API
                   |
                   v
             Runtime Service <--> Redis
                   |
                   v
               Tool Gateway
           /        |         \
        HTTP      Shell       SQL
                    |
              Worker Sandbox
```

### 6. Runtime Execution Flow

The minimum execution lifecycle should be:

1. Operator uploads SOP
2. Ingestion service creates a `procedure_draft`
3. Human validates and edits the draft
4. Composer converts it to `skill_version`
5. Registry publishes a version
6. Runtime creates an `execution_run`
7. Runtime invokes tools through Tool Gateway
8. Approval checkpoints pause and resume execution as needed
9. Execution result and metrics are stored for observability

### 7. Why This Architecture Fits the MVP

This structure keeps the MVP small while preserving enterprise credibility:

- SOP parsing errors cannot directly trigger automation
- runtime remains deterministic and testable
- tool access stays auditable
- approvals are explicit
- observability is built into the execution path

### 8. Recommended Build Order

The architecture should be implemented in business-closure order, not module order.

#### Phase 1

- Skill Spec format
- Runtime Service
- Tool Gateway
- Execution timeline

Goal:

- prove one real SOP can run safely end to end

#### Phase 2

- Skill Registry
- approval flow
- execution dashboard

Goal:

- prove controlled operational use

#### Phase 3

- SOP Ingestion Service
- semi-automated draft generation
- operator validation workflow

Goal:

- prove SOP-to-skill conversion efficiency

### 9. Recommended Technology Stack

To keep implementation and operations focused, use a two-stack model:

- Backend: Go
- Frontend: Next.js + TypeScript

Recommended backend ownership:

- `runtime-service`: Go
- `registry-service`: Go
- `tool-gateway`: Go
- `ingestion-service`: Go for MVP consistency
- `composer-service`: Go APIs behind the web console

Recommended frontend ownership:

- `ops-console`: Next.js App Router + TypeScript
- server-rendered execution detail pages
- typed clients for backend APIs

### 10. Repository Layout Recommendation

Use a frontend-backend separated structure inside one repository:

```text
apps/
  web/   -> Next.js + TypeScript
  api/   -> Go HTTP API and service composition root
docs/    -> product and technical design
```

### 11. Final MVP Judgment

The MVP should be positioned as an enterprise operating system for governed execution, not a generalized AI agent runtime.

Its success depends on:

- a stable Skill Spec
- a trustworthy runtime
- auditable tool execution
- strong human control points
- measurable operational ROI
