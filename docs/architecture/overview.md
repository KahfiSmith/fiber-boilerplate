# Architecture Overview

Go Fiber v3 backend for the `nextjs-boilerplate` frontend, backed by
PostgreSQL (GORM) and Redis.

## Current system

```text
browser (frontend: nextjs-boilerplate)
  -> Next.js App Router
     -> Axios clients (authClient / apiClient)
        -> Go Fiber backend (this repo)
           -> src/common/* (jwt, middleware, response, validator, redis)
              -> src/modules/auth (controller -> service -> repository)
                 -> PostgreSQL (users) / Redis (sessions, tokens)
```

## Layers

| Layer | Path | Responsibility |
|---|---|---|
| Composition root | `cmd/api/main.go` | Loads config, connects DB/Redis, builds dependencies, registers routes |
| Config | `src/config/config.go` | Viper loads `.env`; strict validation; `DSN()` |
| Database | `src/database/postgres.go` | GORM connect + pool settings (global `DB`) |
| Redis | `src/common/redis/redis.go` | Redis client connect (global `Client`) |
| Server | `src/common/server/server.go` | Route groups `/api/v1`, middleware wiring |
| Middleware | `src/common/middleware/` | `Protected`, `RequireRole`, `RateLimiter`, `ValidateOrigin`, `Logger` |
| JWT | `src/common/jwt/jwt.go` | HS256 access token generate/validate |
| Response | `src/common/response/response.go` | `APIResponse`, `Success`, `HandleError` |
| Exceptions | `src/common/exceptions/exceptions.go` | Typed HTTP errors |
| Validator | `src/common/validator/validator.go` | Bind + struct validation |
| Modules | `src/modules/auth`, `src/modules/health` | Feature modules (controller/service/repository) |

## Dependency direction

```text
cmd/api -> src/config | src/database | src/common/* | src/modules/*
src/modules/<feature> -> src/common/* | src/config | src/database
```

- Modules encapsulate their own HTTP handlers, services, repositories, DTOs,
  and types.
- Controllers parse/validate HTTP input, invoke services, format responses.
- Services contain business logic and orchestrate repositories/token service.
- Repositories handle persistence (GORM for Postgres, Redis for sessions).

## Run modes

| Mode | Backend URL | Notes |
|---|---|---|
| Host | `http://localhost:8080` | `APP_PORT=8080` default; `.env` with host DB/Redis (`127.0.0.1`) |
| docker-compose | `http://localhost:3000` | `docker-compose.yml` overrides `APP_PORT=3000`; container service names |

The frontend `NEXT_PUBLIC_BACKEND_API_URL` must match the running mode.

## Config ownership

- Config schema and validation: `src/config/config.go`.
- Database init: `src/database/postgres.go`.
- Redis init: `src/common/redis/redis.go`.
- Fail fast on invalid config at startup.

## Not implemented

- CI/CD pipeline.
- Automated docs validation.
- Additional modules beyond `auth` and `health`.
