# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [2026-05-18]

### Added

#### Provider Fallback 灾备降级

- **FallbackChain 路由**: 多级 Fallback Chain 配置，主 Provider 失败自动切换到备用 Provider
- **熔断器集成**: Circuit Breaker 状态自动影响 Fallback 路由选择
- **Fallback Chain 管理 API**: `/api/v1/admin/fallback-chains` CRUD 端点，auth 中间件保护
- **热更新**: 创建/更新/删除 Fallback Chain 后自动热加载到 Router，无需重启
- **DB 持久化**: `llm_fallback_chains` + `llm_fallback_targets` 数据库表
- **启动加载**: 进程重启后从 DB 自动加载 Fallback Chain（合并 env 配置）
- **前端管理页面**: `/fallback-chains` 支持 Chain 的创建、编辑、删除、启停
- **GATEWAY_FALLBACK_CHAINS 环境变量**: JSON 格式快速配置

#### MCP Proxy 代理网关

- **MCP Proxy**: 完整的 MCP 协议代理，拦截 Agent 请求实现认证和权限过滤
- **认证集成**: 复用 Gateway API Key 认证，Agent 使用同一 Key 访问 LLM 和 MCP
- **工具权限过滤**: 仅返回用户有权限访问的 MCP Server 的工具列表
- **工具命名空间**: `{server_name}__{tool_name}` 格式避免不同 MCP Server 工具名冲突
- **MCP Server 工具发现**: `GET /api/v1/mcp-servers/{id}/tools` 端点
- **MCP Server 健康检查**: `GET /api/v1/mcp-servers/{id}/health` 单个和 `/api/v1/mcp-servers/health` 全部
- **MCP 流式调用**: `POST /api/v1/mcp/invoke/stream` SSE 流式工具调用
- **MCP Resources**: 资源列表、读取、模板、订阅端点
- **MCP Prompts**: Prompt 列表和获取端点
- **MCP 权限管理**: `POST /api/v1/mcp/permissions` 授权、`GET /api/v1/mcp/permissions/list` 查询、`DELETE /api/v1/mcp/permissions/{id}` 撤销

#### 用户设置与偏好

- **个人资料 API**: `GET /api/v1/user/profile`、`PUT /api/v1/user/password`
- **偏好设置 API**: `GET/PUT /api/v1/user/preferences`，支持语言和侧边栏折叠状态
- **user_preferences 表**: migration 0028，`ON CONFLICT` upsert
- **前端 Profile 页面**: `/profile` 修改密码
- **前端 Preferences 页面**: `/preferences` 语言和侧边栏偏好
- **UserDropdown 组件**: 展开方向修复 + 暗色主题 + i18n

#### UUID 主键迁移

- **migration 0027**: `skills.id` 和 `mcp_servers.id` 从 TEXT 转为 UUID
- **id_migration_map 表**: 保留 TEXT → UUID 映射，供 Go 代码查询旧 ID
- **Go 代码适配**: 所有 Skill ID 生成改用 `uuid.New().String()`
- **Bootstrap seeder**: 使用 `uuid.NewSHA1` 确定性 UUID 替代硬编码文本 ID
- **幂等迁移**: 所有 migration 文件支持安全重执行

### Fixed

- **P1**: Skill ID 生成兼容 UUID 列 — `skill_<timestamp>` 等文本 ID 改为 `uuid.New().String()`，避免 migration 0027 后的 invalid UUID input 错误
- **P1**: Fallback Chain admin 路由缺少认证 — 所有 `/api/v1/admin/fallback-chains` 端点包裹 `authMW.RequireAuth()`
- **P2**: 重启后 DB Fallback Chain 被忽略 — 启动时从 DB 加载 Chain 并合并 env 配置
- **Fallback Chain 测试**: 修复 `ServeHTTP` 未定义问题，改用 mux 路由测试

### Changed

- **README.md**: 更新核心功能表、架构图、API 端点、前端页面、功能状态表
- **USER_MANUAL.md**: v2.0 → v2.1，新增 Fallback/Proxy/User Preferences 等章节，全面重写

## [2026-04-24]

### Added

#### v2.0 P3 Features

**Sweep Engine (Hyperparameter Matrix)**
- Backend: `SweepHandler` with full CRUD operations
- PostgreSQL persistence for `sweep_configs` and `sweep_runs` tables
- API endpoints: `GET/POST /api/v1/sweeps`, `DELETE /api/v1/sweeps/delete/{id}`, `GET /api/v1/sweeps/matrix/{id}`
- Frontend: Form-based sweep creation UI with Model × Temperature × MaxTokens × Prompt matrix
- Progress tracking and results visualization

**MCP Server Permission Binding**
- Backend: `MCPInvokeHandler` and `GrantMCPHandler` for permission-based MCP server access
- Permission check before invoke (user/api_key → MCP server)
- API endpoints: `POST /api/v1/mcp/invoke`, `POST /api/v1/mcp/grant`, `GET /api/v1/mcp/permissions`
- Frontend: Permission grant form and MCP invoke interface

**PolicyVersion UI**
- Backend: `PoliciesHandler` wired to `/api/v1/policies`
- Form-based policy creation with rules editor
- Frontend: `PoliciesPageClient` with list/create/expand functionality

**Unit Test Coverage**
- 64 handler tests covering P0, P1, and P3 features
- Test files: `sweeps_test.go`, `policies_test.go`, `mcp_permissions_test.go`, `snapshots_test.go`, `mcp_servers_test.go`
- All tests passing

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
