# SOP-to-Skill Integration Learnings

## Patterns Used

### Package Structure
- `apps/api/internal/orchestrator/` - New package for orchestrator API
- `types.go` - All request/response structs matching OpenAPI schema
- `service.go` - Mock orchestration logic (stub implementation)
- `server.go` - HTTP handlers delegating to service layer

### Handler Method Naming
- Handler methods must be **exported** (capitalized) to be called from outside the package via method references
- Used `HandleXxx` naming convention for exported HTTP handlers

### Route Registration
- Added orchestrator import to server.go
- Added `orchestrator *orchestrator.Server` field to Server struct
- Initialized in NewWithRepository: `orchestratorServer := orchestrator.NewServer(orchestratorSvc)`
- Routes registered using `srv.orchestrator.HandleXxx` pattern

### Response Patterns
- Used existing `writeJSON`, `writeError` helper patterns from server.go
- Created local helpers in orchestrator/server.go: `decodeJSON`, `writeError`, `writeMethodNotAllowed`
- ISO-8601 UTC time format via `time.RFC3339`

## Issues Encountered

### Unexported Methods
- Initially defined handlers as `handleHealth` (lowercase) which caused compilation errors when called from server.go
- **Fix**: Capitalized all handler methods to make them exported: `HandleHealth`

### Path Extraction
- Had to create helper functions to extract IDs from URL paths since Go's gorilla/mux isn't being used
- Functions: `extractRunID`, `extractSkillIDFromPath`, `extractTestRunID`, etc.

## Mock Implementation Notes
- All service methods return mock data matching the OpenAPI schema
- Real CLI integration will be implemented later
- Build verification: `cd apps/api && go build ./...` passes

## Implementation Details (2026-04-01)

### JWT Authentication (auth.go)
- Created `AuthMiddleware` function that validates Bearer JWT tokens
- Public endpoints (`/v1/healthz`, `/v1/models`) bypass auth check via `isPublicEndpoint()`
- Token validation uses HMAC-SHA256 signing method
- Context key `ContextKeyClaims` stores parsed JWT claims
- Configuration via environment variables: `ORCHESTRATOR_JWT_SECRET`, `ORCHESTRATOR_JWT_ISSUER`, `ORCHESTRATOR_SKIP_AUTH`

### Idempotency (idempotency.go)
- In-memory `IdempotencyStore` with TTL-based cleanup
- `Get(key)` returns cached response if exists and not expired
- `Set(key, response, statusCode)` stores response for later retrieval
- 5-minute cleanup goroutine removes expired entries

### PostgreSQL Repository (repo.go)
- Uses same patterns as existing `repository/postgres/` package
- Tables: `orchestrator_runs`, `orchestrator_artifacts`, `orchestrator_skills`, `orchestrator_skill_versions`, `orchestrator_test_runs`, `orchestrator_policies`, `orchestrator_actions`
- JSONB columns for flexible data (source, options, result, errors, artifacts)
- `InitSchema()` creates tables if not exist

### CLI Integration (service.go)
- `s.cliPath` defaults to "sop-to-skill" but configurable
- `runExtractionAsync()` calls `sop-to-skill extract` asynchronously
- Falls back to mock data if CLI fails or repo unavailable
- `EnhanceExtraction()` calls CLI and parses JSON output

### Type Changes
- `TestRun` struct in types.go now has `json:"-"` tags for internal fields (SkillID, Version, Runner, DatasetRef)
- These fields are needed for internal storage but not serialized in API responses

### Server Wiring (server.go)
- Added `os` import for environment variable access
- Auth middleware wrapped around all orchestrator routes except `/v1/healthz`
- Uses `orchestrator.AuthConfig` with env vars for secret/issuer/skip
