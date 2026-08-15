# Folder Structure & Ownership

## Directory tree

```
.
├── cmd/
│   └── api/
│       └── main.go            # Composition root: config, DB, Redis, DI, routes
├── src/
│   ├── config/
│   │   └── config.go          # Viper config schema + validation
│   ├── database/
│   │   └── postgres.go        # GORM connect + pool settings
│   ├── common/
│   │   ├── exceptions/        # HttpError + constructors (NotFound, BadRequest, ...)
│   │   ├── jwt/               # TokenService (HS256 access tokens)
│   │   ├── middleware/        # Protected, RequireRole, RateLimiter, ValidateOrigin, Logger
│   │   ├── redis/             # Redis client connect
│   │   ├── response/          # APIResponse, Success, HandleError
│   │   ├── server/            # RegisterRoutes, Dependencies wiring
│   │   └── validator/         # ParseAndValidate + field messages
│   └── modules/
│       ├── auth/
│       │   ├── auth.controller.go
│       │   ├── auth.service.go
│       │   ├── auth.repository.go
│       │   ├── refresh.repository.go   # Redis sessions, Lua rotation, family revoke
│       │   ├── refresh_token.go        # generate + HMAC hash
│       │   ├── auth_test.go
│       │   ├── dto/auth.dto.go
│       │   └── types/auth.type.go
│       └── health/
│           ├── health.route.go
│           ├── controller/health.controller.go
│           └── service/health.service.go  # DB + Redis status check
├── db/
│   └── migrations/            # SQL migration .up/.down files
├── scripts/                   # migrate.sh, migrate-status.sh, migrate-down.sh
├── docs/                      # This documentation
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
└── .env.example
```

## Ownership

| Path | Responsibility |
|---|---|
| `cmd/api/main.go` | Wire everything; register routes |
| `src/config` | Env config schema, validation, `DSN()` |
| `src/database` | GORM/Postgres connection and pool |
| `src/common/redis` | Redis connection |
| `src/common/server` | Route registration + dependency injection |
| `src/common/middleware` | Auth, role, rate limit, origin, logger |
| `src/common/jwt` | Access token service |
| `src/common/response` | Response envelope + error handler |
| `src/common/exceptions` | Typed HTTP errors |
| `src/common/validator` | Request binding + validation |
| `src/modules/auth` | Auth feature (controller/service/repo/dto/types) |
| `src/modules/health` | Health check feature |
| `db/migrations` | SQL migrations |
| `scripts` | Migration scripts |
