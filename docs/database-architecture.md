# Database Architecture

## Goal

Move the API from in-memory state to PostgreSQL without changing handler contracts.

## Current Design

Handlers depend on the `Repository` interface in:

- `apps/api/internal/server/repository.go`

The current implementation is:

- `apps/api/internal/server/store.go`

This is intentional. It keeps:

- HTTP handlers stable
- tests fast and deterministic
- PostgreSQL integration isolated to a new repository implementation

## Planned PostgreSQL Integration

Recommended package:

```text
apps/api/internal/persistence/postgres
```

Recommended responsibilities:

- open and own the database handle
- map SQL rows into API domain structs
- implement the `Repository` interface
- keep SQL close to repository methods

## Environment Contract

- `DB_DRIVER`
- `DATABASE_URL`

When both are present, the server chooses PostgreSQL at boot.

Current test DSN location:

- `apps/api/.env.example`

## Initial Migration Set

Migration files are in:

- `apps/api/migrations/0001_init.up.sql`
- `apps/api/migrations/0001_init.down.sql`

Tables included:

- `skills`
- `skill_versions`
- `procedure_drafts`
- `executions`
- `approvals`

Operational commands:

- `go run ./cmd/migrate`
- `go run ./cmd/seed`

## Suggested Next Step

The repository now supports:

- memory repository for local development fallback
- PostgreSQL repository when `DB_DRIVER=postgres` and `DATABASE_URL` are configured

Optional local Docker asset:

- `docker-compose.yml`

It exists for portability, but it is not required when using the provided test PostgreSQL instance.
