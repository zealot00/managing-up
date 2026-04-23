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

## v2.0 核心功能完成状态

> 更新日期: 2026-04-23

### ✅ Phase 0: 架构收敛与补债 (已完成)

| 功能 | 状态 | 说明 |
|------|------|------|
| Gateway Sessions 数据模型 | ✅ | `mcp_gateway_sessions` 表 + CRUD |
| Skill Capability Snapshots | ✅ | `skill_capability_snapshots` 表 + CRUD |
| Pre-flight Policy Hook | ✅ | `DefaultPolicyChecker` + 规则引擎 |
| GatewaySessionService | ✅ | Session 创建、Policy 决策记录 |
| MCPRouterHandler 集成 Session | ✅ | 请求自动创建 Session |
| Regression Gate (Promote 检查) | ✅ | `OrchestratorService.Promote()` 检查 Snapshot |

### ✅ Phase 1: Unified Governance Kernel (已完成)

| 功能 | 状态 | 说明 |
|------|------|------|
| Gateway Session 创建与追踪 | ✅ | `/api/v1/gateway/sessions` |
| Policy Checker 实现 | ✅ | 支持 task_type/tag/risk_level 条件匹配 |
| Policy 决策记录 | ✅ | Session 关联 PolicyDecision |
| Snapshot API | ✅ | `/api/v1/snapshots`, `/api/v1/snapshots/list` |
| Promote 前 Gate 检查 | ✅ | 未通过 Snapshot 不可 Promote |
| 前端 Session 历史页面 | ✅ | `/gateway/sessions` |
| 前端 Snapshot 页面 | ✅ | `/skills/snapshots` |

### ✅ Phase 2: Memory Hub MVP (已完成)

| 功能 | 状态 | 说明 |
|------|------|------|
| Memory Cells 数据模型 | ✅ | `memory_cells` 表 + CRUD |
| MemoryHubService | ✅ | 多 Scope 支持 (execution/session/agent/tenant) |
| Gateway 内存注入 | ✅ | `BuildMemoryContext` 自动注入 |
| In-Memory Repository | ✅ | `inMemoryMemoryRepo` 实现 |

### ✅ Phase 3: Bridge Adapter MVP (已完成)

| 功能 | 状态 | 说明 |
|------|------|------|
| OpenAPI Importer | ✅ | `OpenAPIImporter` 从 spec 生成模板 |
| Response Optimizer | ✅ | pick/omit/truncate/summarize 规则 |
| Adapter Template | ✅ | 结构化存储 endpoints + mapping |

### 前端页面状态

| 页面 | 路由 | 状态 |
|------|------|------|
| Gateway Sessions | `/gateway/sessions` | ✅ |
| Snapshot History | `/skills/snapshots` | ✅ |
| Evaluations Dashboard | `/evaluations` | ✅ (重新设计) |
| MCP Router Dashboard | `/mcp-router` | ✅ |

---

## Frontend Refactoring — 2026-04-08

### ✅ Completed

#### P0-1: TanStack Query Introduction
- `QueryProvider` client component with `QueryClient`
- `useApiMutation` hook with query invalidation, toast, router refresh
- **14 forms migrated** from imperative state to declarative mutations

#### P0-2: Inline Modal Extraction
- `FormModal` component replaces 248 lines of inline styles across 3 files
- ExecutionsPageClient, SkillsPageClient, ApprovalsPageClient now use `FormModal`

#### P0-3: API Type Safety
- `zod` installed for runtime type validation
- `api.schemas.ts`: 16 Zod schemas for all API types
- `api.validator.ts`: `validateResponse()` utility
- Note: OpenAPI codegen deferred pending backend spec coverage

#### UX-1: Data Control Layer
- Search + filter controls on Evaluations page
- `DataToolbar` component for reusable search/filter bars
- Client-side LoadMore pagination (20/page) on Executions, Tasks, Evaluations

#### UX-2: Form Validation
- `react-hook-form` + `@hookform/resolvers` installed
- Real-time inline field validation on all forms
- `form-schemas.ts` with Zod schemas including JSON field validation
- Spinner component for inline loading states

#### UX-3: Data Formatting
- `date-fns` for relative time ("2 mins ago"), duration ("1m 5s"), percent formatting
- `TruncatedText` component with expand/collapse for long text

#### UX-4: Bulk Actions
- `BulkActionBar` (fixed bottom, slide-up animation)
- `SelectableCard` wrapper with checkbox selection
- Tasks: bulk delete
- Approvals: bulk approve/reject

#### UX-5: Loading States
- `ListSkeleton`/`CardGridSkeleton` wired to all list pages
- `keepPreviousData: true` prevents flicker on filter changes
- Smooth opacity transition (0.5) while fetching new data

### 📊 Frontend Components Added (2026-04-08)

| Component | Purpose |
|----------|---------|
| `providers/QueryProvider.tsx` | TanStack Query client |
| `lib/use-mutations.ts` | Mutation hook with invalidation + toast |
| `lib/form-schemas.ts` | Zod schemas for forms |
| `lib/api.schemas.ts` | Zod schemas for API types |
| `lib/api.validator.ts` | Runtime validation utility |
| `lib/format.ts` | Date, duration, text formatters |
| `hooks/use-formatters.ts` | Formatter hook |
| `components/ui/FormModal.tsx` | Reusable centered modal |
| `components/ui/BulkActionBar.tsx` | Batch action bar |
| `components/ui/SelectableCard.tsx` | Selectable card wrapper |
| `components/ui/DataToolbar.tsx` | Search + filter bar |
| `components/ui/LoadMore.tsx` | Pagination trigger |
| `components/ui/TruncatedText.tsx` | Expandable text |
| `components/ui/Spinner.tsx` | Loading spinner |

### Files Changed (2026-04-08)

```
frontend/
  app/providers.tsx                    # Added QueryProvider
  app/components/providers/           # QueryProvider
  app/lib/                            # use-mutations, form-schemas, api.schemas, api.validator, format
  app/hooks/                          # use-formatters
  app/components/ui/                   # FormModal, BulkActionBar, SelectableCard, DataToolbar, LoadMore, TruncatedText, Spinner
  app/components/TaskManagerClient.tsx  # useMutation + bulk actions + pagination
  app/components/ExecutionsPageClient.tsx # useMutation + pagination + FormModal
  app/components/SkillsPageClient.tsx    # useMutation + FormModal
  app/components/ApprovalsPageClient.tsx  # useMutation + bulk approve/reject + FormModal
  app/components/EvaluationManager.tsx    # search + filter + pagination + formatting
  app/components/CreateTaskForm.tsx        # react-hook-form + Zod
  app/components/EditTaskForm.tsx          # react-hook-form + Zod
  app/components/TriggerExecutionForm.tsx  # react-hook-form + Zod
  app/components/CreateSkillForm.tsx      # react-hook-form + Zod
  app/components/CreateExperimentForm.tsx # react-hook-form + Zod
  app/components/CreateMetricForm.tsx     # react-hook-form + Zod
  app/components/CreateDatasetForm.tsx     # react-hook-form + Zod
  components/Sidebar.tsx                  # Fixed dropdown expand + parent navigation
  app/globals.css                         # BulkActionBar, SelectableCard, sidebar fixes
```

### Commits (2026-04-08)

```
6978a31 fix(web): sidebar-nav uses overflow-y visible to prevent scrollbar on expand
7329b77 fix(web): sidebar dropdown items now navigate to parent page first, then expand
fcc8b94 feat(web): add pagination + formatting + bulk approve actions
b0bb8f3 feat(web): add bulk actions with selection to Tasks page (UX-4)
554013d feat(web): add search and filter controls to EvaluationManager (UX-1)
f3956f7 feat(web): add date-fns formatters for relative time, duration, text truncation (UX-3)
cd243d1 feat(web): wire up skeletons + keepPreviousData for smooth loading states (UX-5)
7b0283e feat(web): add react-hook-form + Zod inline validation to forms (UX-2)
7f83a65 feat(web): add zod for runtime API type validation (P0-3)
189ca1e feat(web): extract FormModal component, replace inline modals (P0-2)
b72b348 feat(web): migrate remaining forms to useMutation (P0-1)
8995a75 feat(web): migrate ApprovalsPageClient and ApprovalForm to useMutation (P0-1)
fa55af2 feat(web): migrate SkillsPageClient and CreateSkillForm to useMutation (P0-1)
f22a936 feat(web): migrate ExecutionsPageClient and TriggerExecutionForm to useMutation (P0-1)
2fe11d2 feat(web): migrate TaskManager forms to useMutation (P0-1)
66fe144 feat(web): add TanStack Query with QueryClientProvider
```
