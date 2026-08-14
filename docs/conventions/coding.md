# Coding Standards

## General

- Keep code simple and focused.
- Prefer explicit names over abbreviations.
- Avoid premature abstraction.
- Keep functions small and cohesive.

## Engineering principles

- **DRY** - consolidate repeated logic into shared helpers (`src/common/*`).
- **SOLID** - single responsibility per layer; dependency inversion via
  composition root (`cmd/api/main.go`).
- **KISS** - prefer straightforward flow over framework-heavy abstraction.

## Module-driven layout

Each feature lives under `src/modules/<feature>`:

- `<feature>.controller.go` - HTTP handlers (parse, validate, call service, respond).
- `<feature>.service.go` - business logic.
- `<feature>.repository.go` - database operations (GORM).
- `dto/` - request/response structures.
- `types/` - domain models and types.

## Error handling

- Return wrapped errors with context: `fmt.Errorf("context: %w", err)`.
- Custom HTTP errors use `src/common/exceptions`.
- Global error handling uses `src/common/response/response.go`
  (`HandleError` / `GlobalErrorHandler`).

## Dependency injection

- Build concrete controllers/services/repositories in `cmd/api/main.go`.
- Inject them via `server.Dependencies` into `RegisterRoutes`
  (`src/common/server/server.go`).

## Logging

- Logger middleware in `src/common/middleware/logger.middleware.go`.
- Avoid logging sensitive values (passwords, tokens, raw secrets).
- Security events (login success/failure, token reuse, account deletion) are
  logged via `log.Printf` in the auth service.

## Testing

- Table-driven tests in `_test.go` files next to the code
  (`src/modules/auth/auth_test.go`, `src/common/jwt/jwt_test.go`).
- Run with `go test ./...`.

## Documentation

- Keep `README.md`, `docs/*`, and `AGENTS.md` aligned with the codebase.
- Cross-repo: the frontend repo depends on the API contract documented in
  `docs/api/`.
