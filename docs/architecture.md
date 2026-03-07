# Architecture

Current backend architecture and dependency direction.

## Layer Map
- `cmd/api/main.go`
  - app entrypoint
  - initializes config/logger/db/validator
  - builds concrete controller/service/repository dependencies and injects into server
- `pkg/configs`
  - third-party library setup (`viper`, `zap`, `gorm`, `fiber`, `validator`)
  - config schema and validation
- `pkg/server`
  - app wiring, middleware, route registration, runtime start/shutdown
- `pkg/controllers`
  - HTTP handlers only
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
  - `zap.go`
  - `validator.go`
