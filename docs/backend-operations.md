# Backend Operations

## Test Database

Use the shared PostgreSQL test instance via:

- `apps/api/.env.example`

DSN:

- `postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable`

## Run Migrations

From `apps/api`:

```bash
export DB_DRIVER=postgres
export DATABASE_URL='postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable'
go run ./cmd/migrate
```

## Seed Data

From `apps/api`:

```bash
export DB_DRIVER=postgres
export DATABASE_URL='postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable'
go run ./cmd/seed
```

## Start API Against PostgreSQL

From `apps/api`:

```bash
export DB_DRIVER=postgres
export DATABASE_URL='postgresql://postgres:pass@172.20.0.16:5432/auditer?sslmode=disable'
go run ./cmd/server
```

## Smoke Check

After the API is running, validate key endpoints:

```bash
curl http://127.0.0.1:8080/api/v1/dashboard
curl http://127.0.0.1:8080/api/v1/skills
curl http://127.0.0.1:8080/api/v1/approvals
```

## Notes

- `docker-compose.yml` exists only as an optional fallback
- current preferred workflow is the shared test PostgreSQL instance
- handlers still fall back to in-memory mode when database env vars are absent
