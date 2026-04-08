# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [2026-04-08]

### Added

#### Frontend P0 Refactoring (TanStack Query)

- **TanStack Query integration**: Added `QueryClientProvider` in `providers.tsx` with `QueryProvider` client component
- **`useApiMutation` hook**: Declarative mutations with automatic query invalidation, toast notifications, and router refresh
- **14 forms migrated** from imperative `useState + router.refresh()` to `useMutation`:
  - TaskManagerClient, CreateTaskForm, EditTaskForm, TaskCardWithActions
  - ExecutionsPageClient, TriggerExecutionForm
  - SkillsPageClient, CreateSkillForm
  - ApprovalsPageClient, ApprovalListCard, ApprovalForm
  - RunEvaluationForm, CreateSkillVersionForm, CreateExperimentForm, ExperimentCardWithActions
  - CreateMetricForm, SEHManager, CreateDatasetForm

#### Frontend UX Improvements

- **Zod + React Hook Form**: Real-time inline field validation on all forms (`CreateTaskForm`, `TriggerExecutionForm`, `CreateSkillForm`, `CreateExperimentForm`, `CreateMetricForm`, `CreateDatasetForm`)
- **`form-schemas.ts`**: Zod schemas for all form types with JSON field validation
- **Skeleton loading**: All list pages (Executions, Tasks, Skills, Approvals) now show `ListSkeleton`/`CardGridSkeleton` on first load
- **`keepPreviousData`**: Placeholder data prevents flicker on filter changes (smooth opacity transitions)
- **DataToolbar**: Search + filter controls on Evaluations page with status/difficulty filters
- **LoadMore pagination**: Client-side pagination (20 items/page) on Executions, Tasks, Evaluations pages
- **Data formatters**: `date-fns` utilities for relative time, duration, text truncation (`formatRelativeTime`, `formatDurationMs`, `formatPercent`, `TruncatedText` component)
- **Bulk actions**:
  - `BulkActionBar` component with slide-up animation
  - `SelectableCard` wrapper with checkbox selection
  - Tasks: bulk delete with confirmation
  - Approvals: bulk approve/reject

#### Frontend Components

- **`FormModal`**: Reusable centered modal replacing 248 lines of inline modal styles across 3 files
- **`BulkActionBar`**: Fixed bottom action bar for batch operations
- **`SelectableCard`**: Checkbox-selectable card wrapper
- **`DataToolbar`**: Search + filter bar component
- **`LoadMore`**: Pagination trigger button
- **`TruncatedText`**: Expandable text with "Show more/less"
- **`Spinner`**: Inline loading spinner for buttons

#### API Type Safety

- **`zod` runtime validation**: `api.schemas.ts` with 16 Zod schemas for all API types
- **`api.validator.ts`**: `validateResponse()` utility for catching backend/frontend type drift
- **Note**: OpenAPI codegen deferred (backend needs full OpenAPI spec coverage)

### Fixed

- **Sidebar dropdown scrollbar**: `sidebar-nav` changed to `overflow-y: visible` to prevent scrollbar when expanding children menus
- **Sidebar SEH/Gateway navigation**: Clicking parent module (SEH, Gateway) now navigates to parent page AND expands children (removed `e.preventDefault()`)
- **Skeleton flicker prevention**: `placeholderData` keeps old data visible (opacity 0.5) while fetching new data

### Added
- Redis configuration support (`REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`)
- `EMBEDDING_BASE_URL` environment variable for configurable embedding API endpoint

### Changed
- **Gateway**: Improved multi-provider routing with fallback failure switching
- **Gateway**: API Key authentication now uses AES-GCM encryption for secure storage
- **Gateway**: Pricing lookup optimized to O(1) using lowercase map
- **Gateway**: Redis-based distributed rate limiting with circuit breaker
- **MCP**: Two-phase registration (validation separated from network I/O) to avoid blocking other MCP operations
- **MCP**: Proper context lifecycle management to prevent goroutine leaks

### Fixed
- **Worker**: Fixed duplicate execution storm - now uses `sync.Map` for deduplication
- **Worker**: Added bounded semaphore (max 50 concurrent) to prevent overload
- **Router**: Removed unnecessary `currentIndex` state from FallbackRouter
- **Embedding**: Return error on non-200 responses instead of silent nil
- **Stdio validation**: Properly defer cancel() to release timer resources
- **Redis rate limiter**: Added expiration (PX) to `:reset` key to prevent memory leak

## [2026-04-07]

### Added
- Redis-based circuit breaker with exponential backoff
- Redis-based distributed rate limiter
- Redis-based budget checker with atomic check and decrement
- Token budget middleware for gateway endpoints
- Gateway configuration for scanner buffer sizes
- `GATEWAY_MAX_TOKEN_ESTIMATE` for configurable token limit

### Fixed
- Streaming truncation issues
- Usage tracking improvements

## [2026-04-03]

### Added
- PostgreSQL CRUD for provider keys and user budgets
- Gateway provider key management API

### Changed
- Scanner buffer size now configurable via environment variables

## [2026-04-01]

### Added
- LLM Gateway with OpenAI/Anthropic compatible endpoints
- Support for 10 LLM providers (OpenAI, Anthropic, Google, Azure, Ollama, Minimax, Zhipu AI, DeepSeek, Baidu, Alibaba)
- API Key authentication and usage tracking
- Cost tracking per provider/model

## [2026-03-31]

### Added
- MCP Server management API (CRUD + approve workflow)
- MCP Server validation (command whitelist, shell metacharacter validation, CRLF header injection prevention)
- Stdio and HTTP/SSE transport support

## [2026-03-25]

### Added
- Experiment tracking with A/B comparison
- Regression detection
- Trace replay functionality

## [2026-03-20]

### Added
- Skill registry with version control
- SOP to Skill generator
- Execution engine with state machine and checkpoints
- Approval gate for human-in-the-loop

### Changed
- PostgreSQL persistence with migrations
