# Deployment & Infrastructure

## Dockerfile

Multi-stage Go build (`golang:1.25.4-alpine` → `alpine:3.22`).

- Build: `CGO_ENABLED=0 go build -o /fiber-boilerplate ./cmd/api`.
- Exposes port **3000**.
- Runs the compiled binary directly.

## docker-compose

Services: `app`, `postgres` (17-alpine), `redis` (7-alpine).

| Service | Port | Notes |
|---|---|---|
| `app` | `3000:3000` | Overrides `APP_PORT=3000` |
| `postgres` | `5432:5432` | Healthcheck via `pg_isready` |
| `redis` | `6379:6379` | Healthcheck via `redis-cli ping`; appendonly |

Volumes: `postgres_data`, `redis_data` (persist across restarts).

### Env in docker-compose

- `DATABASE_URL` points at `postgres:5432/fiber_boilerplate`.
- `REDIS_ADDR=redis:6379` is set (compose uses the service address).
- Note: compose sets `JWT_SECRET`, `AUTH_OTP_TTL`, `AUTH_OTP_MAX_ATTEMPTS`,
  `GENERATE_SWAGGER_ON_MIGRATE`, `AUTH_DEBUG_EXPOSE_OTP` — some of these are
  not read by the current `config.go`. Prefer aligning compose env with
  `.env.example` (`JWT_ACCESS_SECRET`, `REFRESH_TOKEN_HMAC_KEY`, etc.).

## Run modes

| Mode | Command | Backend URL | DB/Redis |
|---|---|---|---|
| Host | `go run ./cmd/api` | `http://localhost:8080` | `127.0.0.1` |
| Docker | `docker compose up --build` | `http://localhost:3000` | service names |

## Frontend relationship

The frontend `NEXT_PUBLIC_BACKEND_API_URL` must point at the active run mode
(`http://localhost:8080` host, or `http://localhost:3000` compose). The backend
`FRONTEND_ORIGIN` must allow the frontend origin (`http://localhost:3000`).

## CI/CD

CI is configured via `.github/workflows/ci.yml` and runs on every push/PR:

1. `go build ./...`
2. `go vet ./...`
3. `go test ./...`
4. Docs validation (`node ./scripts/harness/check-docs.mjs`)
5. Risk classification (`pnpm verify:risk`)

### Cross-repo check

`pnpm verify:cross-repo` validates BE↔FE sync locally (routes, error codes,
doc links) against the sibling frontend repo. It is not run in CI by default;
enable it by checking out the sibling repo in the workflow.

## Not implemented

- Container image registry / production deployment config.
