# Architecture

Current backend architecture and dependency direction.

## Layer Map
- `cmd/api/main.go`
  - app entrypoint
  - initializes config/logger/db/validator
  - builds concrete controller/service/repository dependencies and injects into server
- runtime assets at repo root
  - `Dockerfile`
  - `docker-compose.yml`
  - optional container env reference: `.env.docker.example`
- `pkg/configs`
  - third-party library setup (`viper`, `zap`, `gorm`, `redis`, `fiber`, `validator`)
  - config schema and validation
- `pkg/server`
  - app wiring, middleware, route registration, runtime start/shutdown
- `pkg/controllers`
  - HTTP handlers only
  - request parsing, request validation, auth-header extraction, and response presentation helpers
- `pkg/server/middleware`
  - server/transport middleware helpers
- `pkg/dto/request`
  - HTTP request contracts
- `pkg/dto/response`
  - HTTP response contracts, including the shared API envelope
- `pkg/entities`
  - domain/business objects
- `pkg/mappers`
  - transformations between `models` and `entities`
- `pkg/services`
  - business logic
  - should operate on `entities`, not persistence models
- `pkg/repositories`
  - data source abstraction
  - responsible for translating `entities <-> models` and persisting `models`
  - keep repository contracts in `*_repository.go`
  - keep storage-specific implementations in files like `*_gorm_repository.go` and `*_redis_repository.go`
- `pkg/models`
  - persistence-only models (GORM/database-facing structs)
- `pkg/utils`
  - shared helper functions (response formatting)

## Dependency Rules
- `cmd` may depend on all `pkg/*`.
- `server` should focus on HTTP wiring and receive controllers via injected dependencies.
- `controllers` depend on services, utils, and DTOs.
- `controllers` translate `request DTO -> entity` before service calls and `entity -> response DTO` before returning.
- `server/middleware` depends on HTTP/framework concerns only.
- `services` depend on repositories and entities; they should not return persistence models.
- `services` may validate session-backed auth state through repository contracts; token parsing stays in service code, storage lookups stay in repositories.
- `repositories` should not depend on controller/server and should translate `models <-> entities` through `pkg/mappers`.
- `dto` should not contain business logic.
- `configs` should not depend on business/domain code.

## Route Ownership
- Route registration entrypoint stays in `pkg/server/routes.go`.
- Route group modules may be split under `pkg/server/routes/` (e.g. `health.go`, `auth.go`) and called by the entrypoint.
- Endpoint handlers are implemented in `pkg/controllers`.

## Configuration Ownership
- All library bootstrap stays in `pkg/configs`:
  - `config.go`
  - `db.go`
  - `auth.go`
  - `fiber.go`
  - `gorm.go`
  - `redis.go`
  - `zap.go`
  - `validator.go`
- `pkg/configs` should initialize the validator, but request-body validation helpers belong in the transport/controller layer.

## Runtime Notes
- The app currently requires both PostgreSQL and Redis at startup.
- If the API runs on the host machine, use host addresses such as `127.0.0.1`.
- If the API runs inside Docker Compose, use service names such as `postgres` and `redis`.
