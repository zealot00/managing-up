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

