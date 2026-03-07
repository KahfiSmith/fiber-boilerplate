# Fiber Boilerplate

Simple Fiber (Go) starter structure with clear layering and bootstrap modules:
- viper config loader
- zap logger
- fiber app/server
- gorm postgres connector
- validator initializer

## Structure

```text
.
├── db
│   └── migrations
├── cmd
│   └── api
│       └── main.go
├── pkg
│   ├── configs
│   ├── controllers
│   │   ├── auth.go
│   │   └── health.go
│   ├── dto
│   │   ├── request
│   │   │   └── auth.go
│   │   └── response
│   │       ├── auth.go
│   │       ├── common.go
│   │       └── health.go
│   ├── entities
│   │   ├── auth.go
│   │   ├── health.go
│   │   └── user.go
│   ├── mappers
│   │   ├── auth.go
│   │   └── user.go
│   ├── models
│   │   ├── auth.go
│   │   └── user.go
│   ├── repositories
│   │   ├── auth_otp_repository.go
│   │   ├── auth_session_repository.go
│   │   ├── health_repository.go
│   │   ├── rate_limit_repository.go
│   │   └── user_repository.go
│   ├── server
│   │   ├── app.go
│   │   ├── middleware
│   │   │   └── request.go
│   │   ├── routes.go
│   │   ├── routes
│   │   │   ├── auth.go
│   │   │   ├── health.go
│   │   │   └── register.go
│   │   └── run.go
│   ├── services
│   │   ├── auth_service.go
│   │   └── health_service.go
│   └── utils
│       └── response.go
├── .env.example
├── go.mod
├── go.sum
└── scripts
    ├── migrate.sh
    ├── migrate-down.sh
    ├── migrate-status.sh
    └── swagger-generate.sh
```

## Run

```bash
go run ./cmd/api
```

At startup the app also runs GORM auto-migration for registered DB models.

## Database Migrations

Requirements:
- `psql` must be installed.
- `.env` or environment must point to the target PostgreSQL database.
- Migration scripts only execute SQL files from `db/migrations`.

Check migration status:

```bash
./scripts/migrate-status.sh
```

`status` shows available local migration files, not applied database history.

Run SQL migrations:

```bash
./scripts/migrate.sh
```

Run a specific migration by version:

```bash
./scripts/migrate.sh 000001_create_users_table
```

Rollback the latest `*.down.sql` file:

```bash
./scripts/migrate-down.sh
```

Rollback a specific migration file:

```bash
./scripts/migrate-down.sh 000001_create_users_table
```

## Data Flow

The boilerplate uses one boundary rule across features:

- controller: request DTO -> entity -> response DTO
- service: entity-only business logic
- repository: entity <-> model translation via `pkg/mappers`
- model: persistence-only structs for PostgreSQL/GORM storage

## Swagger

Generate Swagger docs into `docs/swagger`:

```bash
./scripts/swagger-generate.sh
```

Optional overrides:
- `SWAG_MAIN_FILE` to point to a different entry file.
- `SWAG_OUTPUT_DIR` to change the output folder.

## Env

Copy `.env.example` into `.env` and adjust DB values.
You can use `DATABASE_URL` (PostgreSQL URL format) or individual `DB_*` keys.
Important DB pool keys: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`.
Auth keys: `JWT_SECRET`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`, `BCRYPT_COST`, `AUTH_RATE_LIMIT_PER_MINUTE`, `AUTH_OTP_TTL`, `AUTH_OTP_MAX_ATTEMPTS`, `AUTH_DEBUG_EXPOSE_OTP`.
Forgot-password uses the same OTP TTL and attempt settings as login OTP.
Legacy env aliases still supported: `HTTP_ADDR`, `GRACEFUL_SHUTDOWN_MS`, and `AUTH_DEBUG_EXPOSE_TOKENS`.
`FRONTEND_BASE_URL` is currently unused by this backend.

## Routing and Controllers

- Route registration entrypoint: `pkg/server/routes.go`
- Route groups/modules: `pkg/server/routes/*`
- HTTP handlers (controllers): `pkg/controllers/*`
- Server/transport middleware: `pkg/server/middleware/*`

## Data Contracts

- Domain/business objects: `pkg/entities/*`
- Model/entity mappers: `pkg/mappers/*`
- HTTP request DTOs: `pkg/dto/request/*`
- HTTP response DTOs: `pkg/dto/response/*`
- Persistence-only models: `pkg/models/*`
- PostgreSQL persistence covers `users`, auth sessions, OTP challenges, and auth rate limits.

## Documentation

- Prompt patterns: `docs/patterns.md`
- Agent rules: `docs/rules.md`
- Workflow: `docs/workflow.md`
- Architecture: `docs/architecture.md`
- API contract: `docs/api.md`
- Database conventions: `docs/database.md`

## Health Check

```bash
curl http://localhost:3000/api/v1/health
```

Response example:
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "message": "service is healthy",
    "service": "fiber-boilerplate",
    "timestamp": "2026-03-05T10:00:00Z"
  }
}
```

## Auth Quick Test

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"kahfi@example.com","password":"password123"}'
```

Verify OTP:

```bash
curl -X POST http://localhost:3000/api/v1/auth/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"challenge_id":"<challenge_id>","otp":"123456"}'
```

Forgot password:

```bash
curl -X POST http://localhost:3000/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"kahfi@example.com"}'
```

Reset password:

```bash
curl -X POST http://localhost:3000/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"challenge_id":"<challenge_id>","otp":"123456","new_password":"newpassword123"}'
```
