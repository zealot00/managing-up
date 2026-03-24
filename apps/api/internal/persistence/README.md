# Persistence Plan

The API currently uses an in-memory repository for deterministic local development and tests.

The persistence boundary is now defined by `internal/server/repository.go`.

Next PostgreSQL integration step:

1. add a PostgreSQL repository implementation under `internal/persistence/postgres`
2. wire `config.Database` into server boot
3. use the SQL migrations in `migrations/`
4. switch handlers from memory-backed repository to PostgreSQL-backed repository

Environment variables reserved for that step:

- `DB_DRIVER=postgres`
- `DATABASE_URL=postgres://...`

For the current test setup, use the DSN provided in:

- `apps/api/.env.example`

The repository also contains `docker-compose.yml` as an optional local fallback, but it is not required for your current laptop workflow.

Operational commands:

- `go run ./cmd/migrate`
- `go run ./cmd/seed`
