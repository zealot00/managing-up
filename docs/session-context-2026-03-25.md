# Session Context Save — skill-hub-ee / Agent Harness

**Saved:** 2026-03-25
**Reason:** Session context pollution concern, preparing for fresh session

---

## 1. What Was Built (Session Summary)

### Current Codebase: `skill-hub-ee` / `managing-up`
**Location:** `/Users/zealot/Code/skill-hub-ee`
**Stack:** Go backend (port 8080) + Next.js frontend (port 3000) + PostgreSQL

### Completed This Session

| # | Task | Status |
|---|------|--------|
| 1 | Add YAML spec viewer to skill detail page | ✅ |
| 2 | Add trace viewer to execution detail page | ✅ |
| 3 | Create `/dashboard` page with 6 metrics | ✅ |
| 4 | Root `/` redirect authenticated users to `/dashboard` | ✅ |
| 5 | Add Dashboard link to NavBar | ✅ |
| 6 | Update LoginForm redirect → `/dashboard` | ✅ |
| 7 | Update middleware to protect `/dashboard` | ✅ |
| 8 | Register Calculator + SEH tools to Tool Gateway | ✅ |
| 9 | Wire DBTraceEmitter in main.go (trace persistence) | ✅ |
| 10 | Fix trace.go — engine.TraceRepository accepts server.TraceEvent | ✅ |
| 11 | Add spec_yaml to in-memory store seed data | ✅ |
| 12 | Add seedUsers() + spec_yaml to PostgreSQL seed script | ✅ |
| 13 | Root cause fix: server needs env vars to use PostgreSQL | ✅ |
| 14 | Created `docs/product-spec.md` — Agent Harness product spec | ✅ |

### All TODOs Complete
```
[completed] Register HTTP adapter tools in gateway
[completed] Fix trace store persistence
[completed] Seed spec_yaml data for skill versions
[completed] Add test cases and self-test
[completed] Consolidate features and direction review
[completed] Fix admin login - server needs env vars
```

---

## 2. How to Start the System

### Backend (must use env vars for PostgreSQL)
```bash
cd /Users/zealot/Code/skill-hub-ee/apps/api
DB_DRIVER=postgres DATABASE_URL="postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable" go run cmd/server/main.go
```

### Frontend
```bash
cd /Users/zealot/Code/skill-hub-ee/apps/web
npm run dev
```

### Seed (after schema changes)
```bash
cd /Users/zealot/Code/skill-hub-ee/apps/api
DB_DRIVER=postgres DATABASE_URL="postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable" go run cmd/seed/main.go
```

### Login
- Username: `admin`
- Password: `admin`
- **IMPORTANT:** Server MUST be started with env vars or it uses in-memory store (no users)

---

## 3. Critical Bug Found This Session

**Problem:** Login always failed with "Invalid username or password"

**Root Cause:** `go run cmd/server/main.go` without env vars → uses in-memory `newStore()` → no users seeded

**Fix:** Must use `DB_DRIVER=postgres DATABASE_URL="..."` to connect to PostgreSQL which has seeded users

---

## 4. Key Files Changed This Session

### Backend
```
apps/api/cmd/server/main.go          — wired DBTraceEmitter, registered Calculator/SEH tools
apps/api/internal/engine/trace.go     — TraceRepository now accepts server.TraceEvent
apps/api/internal/server/store.go     — added CreateSkillVersion, seed spec_yaml
apps/api/internal/repository/postgres/
  bootstrap.go                        — added seedUsers() + spec_yaml in seed
  repository.go                       — expanded Repository interface
apps/api/internal/server/repository.go — added TraceRepository support
```

### Frontend
```
apps/web/app/
  page.tsx                            — auth redirect to /dashboard
  dashboard/page.tsx                  — NEW: metrics dashboard
  skills/[id]/page.tsx               — added YAML spec viewer
  executions/[id]/page.tsx           — added trace timeline viewer
apps/web/
  components/NavBar.tsx               — added Dashboard link
  components/LoginForm.tsx            — redirect to /dashboard
  middleware.ts                       — added /dashboard to protected
```

### Documentation
```
docs/product-spec.md                  — NEW: Agent Harness product spec v1.0
```

---

## 5. Product Spec Summary (docs/product-spec.md)

**Product:** Agent Harness — AI Agent 实验操作系统

**6 功能域：**
- D1: Benchmark / Task Management（数据底座）
- D2: Experiment Orchestration（调度核心）
- D3: Agent Execution & Replay（核心差异点）⭐
- D4: Evaluation & Capability Modeling（最有价值）
- D5: Analytics & Observability（洞察产品）
- D6: Governance & CI Integration（企业门槛）

**4 Killer Features：**
- ⭐ Capability Diff Engine
- ⭐ Trajectory Search Visualization
- ⭐ Entropy Budget Panel
- ⭐ Shadow Eval Integration

**6 阶段排期：** Phase 1-6（从 Benchmark Foundation → Advanced Execution）

---

## 6. Architecture Context

### Core Value Chain
```
SOP文档 → Skill Spec YAML → Execution Engine → Tool Gateway → Real HTTP Calls
```

### Current Capabilities
- Execution state machine: pending → running → waiting_approval → succeeded/failed
- Tool Gateway with Calculator registered
- Trace events persisted to PostgreSQL
- JWT cookie auth + CORS
- Dashboard with 6 metrics
- Skill Registry + YAML viewer
- Execution Timeline + trace viewer
- Approval inbox

### MVP定位（docs/mvp-architecture.md）
当前代码库是 **v0.1 — 单一执行路径**，是 Agent Harness 的执行引擎（Domain 3）基础。

---

## 7. Verification Commands

```bash
# Backend tests
cd /Users/zealot/Code/skill-hub-ee/apps/api && go test ./...

# Frontend type check
cd /Users/zealot/Code/skill-hub-ee/apps/web && npx tsc --noEmit

# Login test
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Dashboard
curl -s -b /tmp/cookies.txt http://localhost:8080/api/v1/dashboard

# Skill spec
curl -s -b /tmp/cookies.txt http://localhost:8080/api/v1/skills/skill_001/spec
```

---

## 8. Existing Documentation

| File | Description |
|------|-------------|
| `docs/product-spec.md` | **NEW** Agent Harness product spec v1.0 |
| `docs/todo-plan.md` | Development status tracker |
| `docs/mvp-architecture.md` | MVP architecture design |
| `docs/system-topology-and-service-boundaries.md` | Service boundaries |
| `docs/agent-architecture.md` | Agent interfaces design |
| `docs/api-reference.md` | API reference |
| `docs/backend-operations.md` | Backend operations guide |
| `docs/database-architecture.md` | Database schema |
| `docs/data-model-schema-and-api.md` | Data model |

---

*End of session context save*
