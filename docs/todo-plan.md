# Development Status

All Priority 1-6 items from the original plan have been **completed**.

## Completed Items

### ✅ Priority 1: Core Write Workflows

- [x] `POST /api/v1/skills` — Create skill with validation
- [x] `POST /api/v1/executions` — Trigger execution (persisted)
- [x] `POST /api/v1/executions/{id}/approve` — Approve/reject with audit fields
- [x] Frontend forms: CreateSkillForm, TriggerExecutionForm, ApprovalForm

### ✅ Priority 2: Persistence Hardening

- [x] Repository methods with explicit errors (not silent empty results)
- [x] Service/domain layer between handlers and repositories
- [x] Database CHECK constraints for status enums
- [x] Migration for audit columns (created_by, updated_at, approved_by, resolution_note)

### ✅ Priority 3: API Quality

- [x] Normalized error handling across all handlers
- [x] Request validation (JSON, Content-Type, required fields)
- [x] Pagination for list endpoints (`?limit=20&offset=0`)
- [x] Structured logging (log/slog) for write operations
- [x] Unique request IDs in response envelopes

### ✅ Priority 4: Frontend Console Expansion

- [x] Detail page: `/skills/[id]`
- [x] Detail page: `/executions/[id]`
- [x] Action panels: approval forms, execution triggers
- [x] Skeleton loading states for all data views
- [x] Error boundary (error.tsx)
- [x] Loading fallback (loading.tsx)

### ✅ Priority 5: Database Operations

- [x] Makefile with migrate, seed, serve, db-reset
- [x] Migration tracking (sqlx migrate)
- [x] Rollback documented (make migrate-down)

### ✅ Priority 6: Testing

- [x] Repository integration tests (10 tests)
- [x] Handler tests for create/approve (26 tests)
- [x] Generator tests (15 tests)
- [x] Runtime tests (15 tests)
- [x] Service tests (13 tests)
- [x] LLM client tests (4 tests)
- **Total: 83 tests**

## Additional Completed Items

Beyond the original plan:

- [x] **Execution Engine** — State machine (pending → running → waiting_approval → succeeded/failed)
- [x] **Skill Generator** — LLM-powered SOP → YAML conversion
- [x] **LLM Provider Integration** — 10 providers (OpenAI, Anthropic, Google, Azure, Ollama, Minimax, Zhipu AI, DeepSeek, Baidu, Alibaba)
- [x] **Agent SDKs** — Python and TypeScript SDKs
- [x] **OpenAPI Spec** — Agent-friendly API specification
- [x] **Tool Gateway** — Mock HTTP adapter for MVP
- [x] **Background Worker** — 2s polling for pending executions

## Agent Harness Phase 1 — 2026-03-25 → 重构

> 2026-03-26 优先级调整：砍掉 Task Builder + Skill Sandbox + Shadow Eval
> 聚焦：Task Registry + Local CLI Runner + Trajectory Capture + Basic Dashboard

### 🔴 P0 — 立即开始（阻断验收）

| Item | Domain | Notes | Status |
|------|--------|-------|--------|
| Task Registry | D1 | 管理已有数据集（砍掉自动生成） | ✅ Phase 1 已完成 |
| Local CLI Runner | D2 | 单 Agent 评测工具 | ✅ Phase 1 已完成 |
| Trajectory Capture | D3 | 录制 Agent 运行轨迹 | ✅ Phase 1 已完成 |
| Basic Dashboard | D5 | 看评测结果 | ✅ Phase 1 已完成 |

### 🟡 本周 Sprint

| Item | Notes | Status |
|------|-------|--------|
| 3 个内部 Agent 项目 pilot | 验证录制轨迹是否有用 + 开发者愿不愿意用 CLI | ⬜ 进行中 |
| 确定 Metric 优先级 | 哪些 metric 是内部开发者真正关心的 | ⬜ 待确认 |

### 🟠 本月 — 数据模型确定（技术债源头）

| Item | Notes | Status |
|------|-------|--------|
| Trajectory Schema 定义 | 轨迹数据结构怎么存 | ⬜ 待定 |
| 存储选型 | Postgres？TimescaleDB？ClickHouse？ | ⬜ 待定 |

### 🟢 砍掉（已移除）

| Item | 原因 |
|------|------|
| Task Builder (from-trace) | 复杂度高，优先用现有数据集 |
| Skill Sandbox | Phase 6，后续再定 |
| Shadow Eval | Phase 6，后续再定 |
| Distributed Runner | Phase 2 |
| Replay Bus | Phase 3 |

---

## Phase 2-6 Roadmap

| Phase | Focus | Key Items |
|-------|-------|-----------|
| Phase 2 | Pilot + Data Model | 3 Agent 项目验证 + Trajectory Schema + 存储选型 |
| Phase 3 | Capability Modeling | Capability Graph, Trajectory Search, Failure Mining |
| Phase 4 | Distributed Execution | Worker Pool, Budget Scheduler, Deterministic Mode |
| Phase 5 | Governance & CI | CI Gate API, GitHub Integration, RBAC |
| Phase 6 | Observability Deepening | Radar Dashboard, Entropy Budget, Cost Intelligence |

---

## PDCA 迭代模式

当前 Sprint 以 **2 周**为周期滚动：

```
Sprint 1 (已完): P0 items → Task Registry + CLI Runner + Trajectory Capture + Dashboard
    ↓
Sprint 2 (本周): 3 个内部 Agent 项目 pilot → 验证轨迹价值 + CLI 接受度 + Metric 优先级
    ↓
Sprint 3 (本月): Trajectory Schema 定义 + 存储选型（P0 技术债）
    ↓
Check: pilot 结果 + 数据模型评审
    ↓
Act: 调整优先级，下一轮 Sprint
```

---

## 原有 Remaining Items（保留）

| Item | Priority | Notes |
|------|----------|-------|
| Real Tool Gateway implementation | High | Phase 3 后再实现，当前 mock 够用 |
| Frontend smoke tests | Medium | Playwright/Cypress |
| PostgreSQL-backed CI | Medium | GitHub Actions workflow |
| Idempotent seed data | Low | Environment-specific seeds |

## Metrics

| Metric | Value |
|--------|-------|
| Go Tests | 83 passing |
| Frontend Pages | 9 routes (+2: /tasks/from-trace, /dashboard/capabilities) |
| API Endpoints | 16 (+5: tasks/from-trace, capabilities, capabilities/{name}, capabilities/{name}/diff, experiments variants) |
| Metric Types | 5 (exact_match, semantic_similarity, embedding_similarity, judge_model, statistical) |
| Internal Packages | stats (shared statistical functions) |
| LLM Providers | 10 |
| SDK Languages | 2 (Python, TypeScript) |
| Phase 1 完成 | 6/6 items ✅ |
