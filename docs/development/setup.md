# Developer Setup

## Prerequisites

- Go `1.25.4`
- PostgreSQL (local or Docker)
- Redis (local or Docker)

## Getting started

```bash
go mod download
cp .env.example .env
go run ./cmd/api
```

The app reads `.env` via Viper and fails fast on invalid/missing config.

## Environment

Key groups (see `.env.example` for all values):

| Group | Keys |
|---|---|
| App | `APP_NAME`, `APP_ENV`, `APP_HOST`, `APP_PORT`, timeouts, `APP_BODY_LIMIT_MB`, `APP_PREFORK`, metrics/pprof |
| Logging | `LOG_LEVEL`, `LOG_ENCODING` |
| DB | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_TIMEZONE`, pool settings |
| Redis | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_KEY_PREFIX` |
| JWT | `JWT_ACCESS_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`, `ACCESS_TOKEN_TTL` |
| Refresh | `REFRESH_TOKEN_HMAC_KEY`, `REFRESH_TOKEN_TTL` |
| CORS | `FRONTEND_ORIGIN` |
| Cookie | `COOKIE_NAME`, `COOKIE_PATH`, `COOKIE_SECURE`, `COOKIE_SAME_SITE`, `COOKIE_DOMAIN` |
| Auth | `BCRYPT_COST`, `AUTH_RATE_LIMIT_PER_MINUTE`, `AUTH_DEBUG_EXPOSE_OTP` |

## Run modes

### Host

Set DB/Redis to `127.0.0.1` and run `go run ./cmd/api` (default `APP_PORT=8080`).

### docker-compose

```bash
docker compose up --build
```

Compose overrides `APP_PORT=3000` and uses service names for DB/Redis.

## Verify it works

1. `GET http://localhost:<port>/api/v1/health` returns
   `{"success": true, ...}` with DB/Redis status.
2. Register → login → refresh → me flow works against the frontend
   (`nextjs-boilerplate`).

## Verification harness

The repo ships a tiered verification harness (via `package.json` scripts):

- `pnpm verify:fast` - `go build`, `go vet`, `go test`, and `docs:check`.
- `pnpm verify` - same as `verify:fast`.
- `pnpm verify:risk` - classify change risk by path (low/medium/high).
- `pnpm verify:cross-repo` - validate BE↔FE sync (routes, error codes,
  cross-repo doc links) against the sibling `nextjs-boilerplate` repo.
- `pnpm verify:all` - everything above.

A pre-commit hook (`.githooks/pre-commit`) runs `pnpm verify:fast`
automatically. CI runs build/vet/test, docs check, and risk classification on
every push/PR.
