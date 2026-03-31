## Agent Harness Phase 1 Issues

### 2026-03-26: CreateTaskRequest missing Gold/Scoring fields

**Problem:** `server.CreateTaskRequest` (types.go:301) was missing `Gold` and `Scoring` fields even though `service.CreateTaskRequest` (task.go:81) had them. The `toServiceCreateTaskRequest` mapping function was also not mapping these fields.

**Files modified:**
- `apps/api/internal/server/types.go`: Added `Gold GoldConfig` and `Scoring ScoringConfig` fields with `omitempty` JSON tags
- `apps/api/internal/server/server.go`: Updated `toServiceCreateTaskRequest` to map Gold and Scoring (with proper type conversion since server.GoldConfig and service.GoldConfig are different Go types)

**Note:** The server types (GoldConfig/ScoringConfig) and service types have the same structure but are distinct Go types - proper conversion is needed in the mapping function.

### 2026-03-26: PostgreSQL DSN using unreachable IP

**Problem:** `.env` had `DATABASE_URL=postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable` but 172.20.0.16 is Docker's internal network IP and is not reachable from the host machine.

**Resolution:** docker-compose.yml exposes postgres on port 5432 of the host. Changed `.env` to use `localhost:5432` instead of `172.20.0.16:5432`.

**Additional note:** The server does NOT gracefully fall back to in-memory mode when postgres is unavailable - it calls `log.Fatalf` in main.go:36. If postgres is not running, the server will exit. To use in-memory mode, ensure `DB_DRIVER` or `DATABASE_URL` is empty in `.env`.

### 2026-03-26: In-memory store stubs for experiments

**Problem:** Multiple methods in `store.go` returned stub/empty values:
- `CreateTask` returned `Task{}, nil` (all-zero Task)
- `CreateExperiment` returned `Experiment{}, nil`
- `GetExperiment` returned `Experiment{}, false`
- `ListExperiments` returned `[]Experiment{}`
- `CreateExperimentRun` returned `ExperimentRun{}, nil`
- `UpdateExperimentRun` returned `nil`
- `ListExperimentRuns` returned `[]ExperimentRun{}`

**Resolution:** 
- Added `experiments map[string]Experiment` and `experimentRuns map[string]ExperimentRun` to store struct
- Initialized maps in `NewStore()`
- Implemented all methods properly with mutex locks

### 2026-03-26: newStore() not accessible from main.go

**Problem:** main.go (package main) tried to call `newStore()` from the server package, but `newStore` was unexported (lowercase).

**Resolution:** Renamed `newStore()` to `NewStore()` in store.go (exported), updated internal caller in `server.New()`, and updated main.go to use `server.NewStore()`.

### 2026-03-26: Postgres fallback didn't set up experimentSvc

**Problem:** When postgres failed and server fell back to `server.New(cfg)`, `experimentSvc` was set to `nil` in the server, making experiment runs non-functional.

**Resolution:** Updated fallback path in main.go to create full stack with in-memory store: `server.NewStore()` + LLM client + agent + evaluationRunner + taskRunnerAdapter + experimentSvc.

