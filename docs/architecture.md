# Architecture

Current backend architecture and dependency direction.

## Layer Map
- `cmd/api/main.go`
  - app entrypoint
  - initializes config/db/redis
  - builds concrete controller/service/repository dependencies and registers routes
- runtime assets at repo root
  - `Dockerfile`
  - `docker-compose.yml`
  - container/host env reference: `.env.example`
- `src/config`
  - config schema loading via `viper` from `.env`
- `src/database`
  - PostgreSQL connection and GORM initialization
- `src/common/redis`
  - Redis client connection
- `src/common/server`
  - route registration and HTTP server dependency injection
- `src/common/middleware`
  - JWT protection and request logging middleware
- `src/common/response` & `src/common/validator` & `src/common/exceptions`
  - shared response formatting, error handling, and request validation helpers
- `src/common/jwt`
  - token generation service for JWT access and refresh tokens
- `src/modules/health`
  - health check module (Controller, Service, Route)
- `src/modules/auth`
  - auth module (Controller, Service, Repository, DTOs, Types)

## Dependency Rules
- `cmd/api` depends on `src/config`, `src/database`, `src/common/*`, and `src/modules/*`.
- `src/common/server` focuses on route registration and receives controllers via injected dependencies.
- Modules in `src/modules/*` encapsulate their own HTTP handlers, services, repositories, and DTOs.
- Controllers parse/validate HTTP input, invoke services, and format HTTP responses.
- Services contain business logic, invoke repositories/token services, and return domain types/errors.
- Repositories handle database persistence via GORM.

## Configuration Ownership
- Config loading and schema reside in `src/config/config.go`.
- Database initialization resides in `src/database/postgres.go`.
- Redis initialization resides in `src/common/redis/redis.go`.

## Runtime Notes
- The app requires PostgreSQL and Redis at startup.
- If the API runs on the host machine, use host addresses such as `127.0.0.1`.
- If the API runs inside Docker Compose, use container service names such as `postgres` and `redis`.
