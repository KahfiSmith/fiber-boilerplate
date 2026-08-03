# Coding Standards

Coding conventions for this repository.

## General
- Keep code simple and focused.
- Prefer explicit names over abbreviations.
- Avoid premature abstraction.
- Keep functions small and cohesive.

## Engineering Principles (DRY, SOLID, KISS)
- DRY:
  - Consolidate repeated logic into shared helpers (`src/common/*`).
- SOLID:
  - Single Responsibility: keep wiring in `src/common/server`, business logic in services under `src/modules/*`.
  - Dependency Inversion: build concrete dependencies in `cmd/api/main.go`, inject into server routes.
- KISS:
  - Prefer straightforward flow over framework-heavy abstraction.
  - Keep modules localized under `src/modules/<feature>`.

## Error Handling
- Return wrapped errors with context: `fmt.Errorf("context: %w", err)`.
- Custom HTTP exceptions use `src/common/exceptions`.
- Global error handling uses `src/common/response/response.go`.

## Package and Layer Boundaries
- Config schema in `src/config`.
- Database init in `src/database`.
- Shared components in `src/common/*` (`jwt`, `middleware`, `redis`, `response`, `server`, `validator`, `exceptions`).
- Modules in `src/modules/<feature>` containing:
  - `<feature>.controller.go`
  - `<feature>.service.go`
  - `<feature>.repository.go`
  - `dto/`
  - `types/`

## API and Responses
- Response format: `response.APIResponse` (`success`, `data`, `error`).
- Success responses use `response.Success()`.
- Error responses use `response.HandleError()`.

## Logging
- Logger middleware in `src/common/middleware/logger.middleware.go`.
- Avoid logging sensitive values (passwords, tokens, raw secrets).

## Documentation
- Keep `README.md`, `docs/*`, and `AGENTS.md` aligned with the current codebase structure.
