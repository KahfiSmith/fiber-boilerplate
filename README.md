# Fiber Boilerplate

Backend starter built with Go, Fiber v3, PostgreSQL, and Redis.

## Tech Stack

- Go `1.25.4`
- Fiber v3 (`github.com/gofiber/fiber/v3`)
- GORM + PostgreSQL (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- Redis (`github.com/redis/go-redis/v9`) for refresh token storage
- Viper (`github.com/spf13/viper`) for configuration
- `go-playground/validator` for request validation
- JWT (`github.com/golang-jwt/jwt/v5`) for access & refresh tokens

## Features

- Health check endpoint (`GET /api/v1/health`)
- Email/password registration with email normalization (`POST /api/v1/auth/register`)
- Login with multi-device session support (`POST /api/v1/auth/login`)
- Token refresh with session rotation (`POST /api/v1/auth/refresh`)
- Forgot password & reset password flow (`POST /api/v1/auth/forgot-password`, `POST /api/v1/auth/reset-password`)
- Logout per device session (`POST /api/v1/auth/logout`)
- Protected user profile endpoint (`GET /api/v1/auth/me`)
- Role-based Access Control (RBAC) middleware (`RequireRole`)
- Redis-backed rate limiting on auth endpoints

## API Surface

- `GET /api/v1/health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/logout` (Protected: requires `Authorization: Bearer <access_token>`)
- `GET /api/v1/auth/me` (Protected: requires `Authorization: Bearer <access_token>`)

Detailed contracts live in `docs/api.md`.

## Project Structure

```text
.
├── cmd/
│   └── api/
│       └── main.go          # Application composition root
├── db/
│   └── migrations/          # SQL migrations
├── docs/                    # Architecture and API documentation
└── src/
    ├── common/              # Shared infrastructure & utilities
    │   ├── exceptions/      # Custom HTTP errors
    │   ├── jwt/             # JWT token generation
    │   ├── middleware/      # Logger and JWT auth middleware
    │   ├── redis/           # Redis client setup
    │   ├── response/        # Global response envelope & error handling
    │   ├── server/          # Route registration
    │   └── validator/       # Request validation helper
    ├── config/              # Viper configuration schema & loader
    ├── database/            # GORM PostgreSQL connection
    └── modules/             # Encapsulated feature modules
        ├── auth/            # Auth module (controller, service, repository, DTO, types)
        └── health/          # Health check module
```

## Prerequisites

- Go `1.25.4`
- PostgreSQL
- Redis

## Installation

### Host-based setup

1. Copy the environment file:

```bash
cp .env.example .env
```

2. Update PostgreSQL and Redis credentials in `.env`.

3. Run migrations:

```bash
./scripts/migrate.sh
```

4. Start the API:

```bash
go run ./cmd/api
```

### Docker Compose setup

Run PostgreSQL, Redis, and the API together:

```bash
docker compose up --build
```

## Usage Examples

### Health Check

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
curl -i -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kahfi@example.com","password":"Secret123"}'
```

Response includes `access_token` in body and `refresh_token` in `Set-Cookie` header.

### Protected Endpoint

```bash
curl http://localhost:3000/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

### Refresh Token

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/refresh \
  -H "Cookie: refresh_token=<refresh_token>"
```

### Logout

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>" \
  -H "Cookie: refresh_token=<refresh_token>"
```

## Common Commands

Run the API:

```bash
go run ./cmd/api
```

Run checks:

```bash
go vet ./...
```

Run tests:

```bash
go test ./...
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
- registered GORM models auto-migrate on startup via `database.DB.AutoMigrate(&types.User{})` in `cmd/api/main.go`

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
