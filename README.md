# managing-up

Enterprise Operating System for Intelligence — AI Agent Skill Governance Platform.

## Overview

managing-up is a platform for managing, distributing, and governing AI Agent skills (SOPs) in enterprise environments. It provides:

- **Skill Registry** — Version-controlled skill specifications that AI Agents can discover and download
- **Execution Engine** — Stateful execution with approval checkpoints for high-risk operations
- **Skill Generator** — LLM-powered conversion of SOP documents into executable skill specs
- **Agent SDKs** — Python and TypeScript SDKs for external AI Agents (OpenClaw, OpenCode, Codex, etc.)

## Architecture

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
│  ┌──────────────────────────────────────────┐             │
│  │         PostgreSQL + Tool Gateway          │             │
│  └──────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Start Backend (In-Memory Mode)

```bash
cd apps/api
go run cmd/server/main.go
# Server running at http://localhost:8080
```

### 2. Start Frontend

```bash
cd apps/web
npm install
npm run dev
# Frontend at http://localhost:3000
```

### 3. With PostgreSQL

```bash
cd apps/api
make migrate
make seed
make serve-pg DATABASE_URL="postgres://localhost:5432/skillhub?sslmode=disable"
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/skills` | List published skills |
| POST | `/api/v1/skills` | Create skill |
| GET | `/api/v1/skills/{id}` | Get skill details |
| GET | `/api/v1/skills/{id}/spec` | Download skill YAML spec |
| POST | `/api/v1/executions` | Trigger execution |
| GET | `/api/v1/executions/{id}` | Get execution status |
| POST | `/api/v1/executions/{id}/approve` | Approve/reject |
| POST | `/api/v1/generate-skill` | Generate skill from SOP |
| POST | `/api/v1/agents` | Register agent |
| GET | `/api/v1/approvals` | List approvals |
| GET | `/api/v1/dashboard` | Dashboard metrics |

## LLM Providers

Supports 10 LLM providers for skill generation:

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

Configure via environment variables:

```bash
LLM_PROVIDER=ollama
LLM_MODEL=llama3
LLM_API_KEY=           # Not required for Ollama
LLM_BASE_URL=http://localhost:11434
```

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

## Project Structure

```
managing-up/
├── apps/
│   ├── api/
│   │   ├── cmd/server/          # HTTP server
│   │   ├── internal/
│   │   │   ├── server/          # HTTP handlers
│   │   │   ├── service/          # Domain logic
│   │   │   ├── runtime/         # Execution engine
│   │   │   ├── generator/       # LLM skill generator
│   │   │   ├── llm/             # LLM provider clients
│   │   │   └── persistence/      # PostgreSQL repository
│   │   ├── migrations/          # SQL migrations
│   │   └── openapi/            # OpenAPI spec
│   └── web/
│       └── app/                 # Next.js frontend
├── sdk/
│   ├── python/                 # Python SDK
│   └── typescript/            # TypeScript SDK
└── docs/                      # Architecture docs
```

## Features

| Feature | Status |
|---------|--------|
| Skill CRUD | ✅ |
| Execution Engine (State Machine) | ✅ |
| Approval Checkpoints | ✅ |
| Skill Generator (SOP → YAML) | ✅ |
| LLM Provider Integration (10 providers) | ✅ |
| Agent SDK (Python, TypeScript) | ✅ |
| OpenAPI Spec | ✅ |
| Skeleton Loading States | ✅ |
| Unit Tests (83 tests) | ✅ |

## Testing

```bash
# Backend tests
cd apps/api && go test ./...

# Frontend build
cd apps/web && npm run build
```

## Makefile Commands

```bash
make serve          # Start server (in-memory)
make serve-pg       # Start server (PostgreSQL)
make migrate         # Run migrations
make migrate-down   # Revert last migration
make seed           # Seed test data
make db-reset       # Reset database
make build          # Build binary
make test           # Run tests
```
