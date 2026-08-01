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
- `internal/core/configs`
  - third-party library setup (`viper`, `zap`, `gorm`, `redis`, `fiber`, `validator`)
  - config schema and validation
- `internal/core/server`
  - app wiring, route registration, runtime start/shutdown
- `internal/pkg/middleware`
  - server/transport middleware helpers
- `internal/pkg/response`
  - shared helper functions (response formatting, HTTP parsing)
- `internal/domain/*` (e.g. `health`)
  - fully encapsulated domain logic (Module-Driven / Package-Oriented)
  - contains its own Controller, Service, Repository, DTOs, and Entity
- `pkg/controllers` (Legacy / Layer-Driven)
  - HTTP handlers for features not yet refactored to `internal/domain`
- `pkg/services` (Legacy / Layer-Driven)
  - business logic
- `pkg/repositories` (Legacy / Layer-Driven)
  - data source abstraction
- `pkg/entities`, `pkg/models`, `pkg/mappers`, `pkg/dto` (Legacy / Layer-Driven)
  - data models and transformations for `pkg/` layers

## Dependency Rules
- `cmd` may depend on `internal/*` and `pkg/*`.
- `internal/core/server` should focus on HTTP wiring and receive controllers via injected dependencies.
- `internal/domain/*` components should not depend on `pkg/controllers` or `pkg/services` from other domains (isolate by domain).
- `pkg/controllers` depend on services, utils, and DTOs.
- `pkg/controllers` translate `request DTO -> entity` before service calls and `entity -> response DTO` before returning.
- `pkg/services` depend on repositories and entities; they should not return persistence models.
- `pkg/repositories` should not depend on controller/server and should translate `models <-> entities` through `pkg/mappers`.

## Configuration Ownership
- All library bootstrap stays in `internal/core/configs`:
  - `config.go`
  - `db.go`
  - `auth.go`
  - `fiber.go`
  - `gorm.go`
  - `redis.go`
  - `zap.go`
  - `validator.go`

## Runtime Notes
- The app currently requires both PostgreSQL and Redis at startup.
- If the API runs on the host machine, use host addresses such as `127.0.0.1`.
- If the API runs inside Docker Compose, use service names such as `postgres` and `redis`.
