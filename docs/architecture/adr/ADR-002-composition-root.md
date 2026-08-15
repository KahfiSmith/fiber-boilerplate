# ADR-002: Composition Root with Dependency Injection

## Context

The application needs to assemble many pieces — config, database, Redis, token
service, repositories, services, controllers — and register routes. Options:

- Each package constructs its own dependencies (service locator / global
  singletons everywhere). This couples packages to concrete constructors and
  makes testing and swapping implementations hard.
- A framework DI container (e.g. wire, dig). Adds a dependency and indirection
  for a codebase of this size.
- A single composition root that builds concrete objects explicitly and injects
  them into the server wiring.

## Decision

- All object construction happens in **one composition root**:
  `cmd/api/main.go`.
- `cmd/api/main.go` loads config, connects PostgreSQL and Redis, builds
  controllers/services/repositories, and passes them to
  `server.RegisterRoutes` via a `server.Dependencies` struct
  (`src/common/server/server.go`).
- Packages expose constructors (`NewAuthController`, `NewAuthService`,
  `NewTokenService`, ...) but never call them from other packages — only the
  composition root does.
- Globals are limited to connection handles (`database.DB`, `redis.Client`)
  initialized once at startup.

## Consequences

- Dependencies are explicit and traceable from one file.
- Swapping an implementation (e.g. a fake repository in tests) is a one-line
  change at the root.
- No framework DI dependency is required.
- The composition root must stay thin; business logic lives in modules.

## Status

Accepted.
