# Fiber Boilerplate

Backend starter built with Go, Fiber, PostgreSQL, and Redis. This repository is structured for small-to-medium API services that need clear layering, auth flows, operational visibility, and a predictable bootstrap path.

## Tech Stack

- Go `1.25.4`
- Fiber v3 for HTTP routing and middleware
- GORM + PostgreSQL for relational persistence
- Redis for refresh-session storage and rate limiting
- Viper for configuration loading
- Zap for structured logging
- `go-playground/validator` for request validation
- Swagger generation into `docs/swagger.json` and `docs/swagger.yaml`

## Features

- Health check endpoint at `GET /api/v1/health`
- Email/password registration
- Login with OTP challenge verification
- Forgot-password and password reset flow with OTP
- JWT access tokens
- Server-side refresh sessions stored in Redis
- Session management endpoints for device/session visibility and revocation
- Redis-backed auth rate limiting
- Env-gated `/metrics` and `/debug/pprof/*` observability endpoints
- SQL migration scripts plus GORM auto-migration for registered models

## API Surface

- `GET /api/v1/health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/otp/verify`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `GET /api/v1/auth/sessions`
- `POST /api/v1/auth/sessions/revoke`
- `POST /api/v1/auth/sessions/revoke-all`

Detailed contract notes live in `docs/api.md`.

## Project Structure

```text
.
├── cmd/api
├── db/migrations
├── docs
├── pkg/configs
├── pkg/controllers
├── pkg/dto
├── pkg/entities
├── pkg/mappers
├── pkg/models
├── pkg/repositories
├── pkg/server
├── pkg/services
├── pkg/utils
└── scripts
```

Architecture rule of thumb:

- `controllers` parse HTTP input and return HTTP responses
- `services` contain business logic
- `repositories` handle persistence and model/entity translation
- `server` wires routes, middleware, and runtime startup
- `configs` owns third-party bootstrap and env config

## Prerequisites

- Go `1.25.4`
- PostgreSQL
- Redis
- `psql` if you want to run SQL migrations
- `swag` if `GENERATE_SWAGGER_ON_MIGRATE=true`

## Installation

### Option 1: Host-based setup

1. Copy the env file:

```bash
cp .env.example .env
```

2. Update the values in `.env` for your local PostgreSQL and Redis instances.

3. Run migrations:

```bash
./scripts/migrate.sh
```

4. Start the API:

```bash
go run ./cmd/api
```

### Option 2: Docker Compose setup

The included `docker-compose.yml` starts:

- API on `http://localhost:3000`
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

Run everything:

```bash
docker compose up --build
```

The containerized API uses:

- `DATABASE_URL=postgres://postgres:postgres@postgres:5432/fiber_boilerplate?sslmode=disable&TimeZone=UTC`
- `REDIS_ADDR=redis:6379`

If you run the API on your host while PostgreSQL and Redis run in Docker, use host addresses instead:

- `DB_HOST=127.0.0.1`
- `REDIS_ADDR=127.0.0.1:6379`

## Usage Guide

### Health check

```bash
curl http://localhost:3000/api/v1/health
```

### Register

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Kahfi","email":"kahfi@example.com","password":"Secret123"}'
```

### Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kahfi@example.com","password":"Secret123"}'
```

### Verify OTP

```bash
curl -X POST http://localhost:3000/api/v1/auth/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"challenge_id":"<challenge_id>","otp":"123456"}'
```

### Session-backed protected request

```bash
curl http://localhost:3000/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

### Forgot password

```bash
curl -X POST http://localhost:3000/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"kahfi@example.com"}'
```

### Reset password

```bash
curl -X POST http://localhost:3000/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"challenge_id":"<challenge_id>","otp":"123456","new_password":"NewSecret123"}'
```

## Configuration

Copy `.env.example` to `.env` and adjust values for your environment.

Important groups:

- App: `APP_HOST`, `APP_PORT`, `APP_READ_TIMEOUT`, `APP_WRITE_TIMEOUT`, `APP_SHUTDOWN_TIMEOUT`, `APP_BODY_LIMIT_MB`, `APP_PREFORK`
- Observability: `APP_ENABLE_METRICS`, `APP_ENABLE_PPROF`
- Logging: `LOG_LEVEL`, `LOG_ENCODING`
- Database: `DATABASE_URL` or `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_TIMEZONE`
- DB pool tuning: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`
- Redis: `REDIS_ADDR`, `REDIS_USERNAME`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_KEY_PREFIX`
- Auth: `JWT_SECRET`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`, `BCRYPT_COST`, `AUTH_RATE_LIMIT_PER_MINUTE`, `AUTH_OTP_TTL`, `AUTH_OTP_MAX_ATTEMPTS`, `AUTH_DEBUG_EXPOSE_OTP`

Notes:

- `AUTH_DEBUG_EXPOSE_OTP=true` is useful for local development only
- `APP_ENABLE_METRICS=true` exposes `GET /metrics`
- `APP_ENABLE_PPROF=true` exposes `/debug/pprof/*`
- keep `APP_ENABLE_PPROF=false` unless the endpoint is protected by trusted network controls

## Common Commands

Run the API:

```bash
go run ./cmd/api
```

Run tests:

```bash
go test ./...
```

Check available migrations:

```bash
./scripts/migrate-status.sh
```

Run all SQL migrations:

```bash
./scripts/migrate.sh
```

Run one migration:

```bash
./scripts/migrate.sh 000003_add_users_email_lower_index
```

Rollback the latest migration:

```bash
./scripts/migrate-down.sh
```

Generate Swagger:

```bash
./scripts/swagger-generate.sh
```

## Observability

Enable metrics:

```bash
APP_ENABLE_METRICS=true go run ./cmd/api
curl http://localhost:3000/metrics
```

Enable pprof:

```bash
APP_ENABLE_PPROF=true go run ./cmd/api
go tool pprof http://localhost:3000/debug/pprof/profile
```

Use metrics for ongoing monitoring such as request count, request latency, in-flight traffic, goroutines, and memory usage.

Use `pprof` for deep debugging when the process is already slow or memory-heavy.

## Database and Migration Notes

- SQL files live in `db/migrations`
- migration scripts execute files directly from the folder and do not maintain a `schema_migrations` table
- registered GORM models still auto-migrate on startup
- users are looked up case-insensitively, and a matching PostgreSQL index is included in `000003_add_users_email_lower_index`

## Session Management Design

This repo intentionally keeps session-management APIs:

- `GET /api/v1/auth/sessions`
- `POST /api/v1/auth/sessions/revoke`
- `POST /api/v1/auth/sessions/revoke-all`

Reason:

- refresh tokens are stored as server-side sessions
- one refresh token maps to one login/device context
- users can inspect active sessions and revoke compromised devices
- protected access-token requests also validate live session presence

If you want a smaller auth surface for an MVP, remove those endpoints deliberately rather than treating them as leftovers.

## Documentation

- Architecture: `docs/architecture.md`
- API reference: `docs/api.md`
- Database notes: `docs/database.md`
- Repository rules: `docs/rules.md`
- Coding standards: `docs/coding-standards.md`
- Implementation patterns: `docs/patterns.md`
- Workflow notes: `docs/workflow.md`

## Production Notes

- keep `JWT_SECRET` strong and private
- do not expose `APP_ENABLE_PPROF=true` to the public internet
- review Redis persistence strategy before using this as-is for high-scale production auth
- treat `AUTH_DEBUG_EXPOSE_OTP=true` as a development-only setting
