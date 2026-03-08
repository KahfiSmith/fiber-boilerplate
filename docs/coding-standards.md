# Coding Standards

Coding conventions for this repository.

## General
- Keep code simple and focused.
- Prefer explicit names over abbreviations.
- Avoid premature abstraction.
- Keep functions small and cohesive.

## Principal Engineer Expectations
- Prefer strong defaults and explicit tradeoffs over hidden behavior.
- Make operationally important behavior easy to discover in code and docs.
- Do not keep features that lack a clear product, security, or operational reason.
- When a feature is intentionally kept, document why it exists.

## Engineering Principles (DRY, SOLID, KISS)
- DRY:
  - Consolidate repeated logic into shared helpers.
  - Keep env defaults/validation centralized in `pkg/configs`.
- SOLID:
  - Single Responsibility: keep wiring in `server`, business logic in `services`.
  - Dependency Inversion: build concrete dependencies in `cmd/api/main.go`, inject into `server`.
  - Interface Segregation: keep interfaces small and purpose-specific.
- KISS:
  - Prefer straightforward flow over framework-heavy abstraction.
  - Add layers only when complexity justifies them.

## Error Handling
- Return wrapped errors with context:
  - `fmt.Errorf("context: %w", err)`
- Do not swallow errors silently.
- Fail fast on startup/configuration problems.

## Package and Layer Boundaries
- Keep library bootstrap/setup in `pkg/configs`.
- Keep app/runtime wiring in `pkg/server`.
- Keep HTTP handlers thin in `pkg/controllers`.
- Keep request parsing/validation helpers close to the transport layer instead of `pkg/configs`.
- Keep server middleware in `pkg/server/middleware`.
- Put HTTP request contracts in `pkg/dto/request`.
- Put HTTP response contracts in `pkg/dto/response`.
- Put domain objects in `pkg/entities`.
- Put reusable model/entity conversion logic in `pkg/mappers`.
- Put business logic in `pkg/services`.
- Put data access contracts/implementations in `pkg/repositories`.
- In `pkg/repositories`, prefer one contract file plus storage-specific implementation files instead of mixing both in one large file.

## Configuration
- Add defaults and validation for every new env key.
- Keep `.env.example` updated when config changes.
- Keep DB-related config logic in `pkg/configs/db.go`.

## API and Responses
- Keep response envelope consistent with the shared response DTO (`dto/response.APIResponse`).
- Prefer new API contracts under `pkg/dto/request` and `pkg/dto/response`.
- Prefer utility response helpers in `pkg/utils/response.go`.
- Preserve backward compatibility unless requested.

## Logging
- Use structured logs via zap fields.
- Avoid logging sensitive values (passwords, tokens, raw secrets).

## Database
- Use pool settings from config.
- Keep DSN creation centralized in DB config helper.
- Validate connectivity on startup path.

## Documentation
- Update docs when architecture, config, or behavior changes.
- When runtime setup changes, document both host-based and container-networked env values if they differ.
- Keep `README.md`, `docs/*`, and `tools/agent/*` guidance aligned with the current repo reality.
