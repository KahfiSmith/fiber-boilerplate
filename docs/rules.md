# Agent Rules

Repository-specific rules for prompting and execution quality.

## Prompt Rules
- Include concrete file paths whenever possible.
- Include exact error messages instead of paraphrasing.
- State hard constraints explicitly (what must not change).
- Keep one primary objective per prompt.
- Include validation commands (`go test ./...`, `go run ./cmd/api`) when relevant.
- Include runtime mode when relevant (`host`, `docker compose`, or mixed host/container setup).
- State documentation expectations explicitly when behavior, workflow, or repo conventions change.
- Mention output preference (summary format, file list, next steps).

## Execution Rules
- Keep configuration setup in `src/config`.
- Keep route registration in `src/common/server/server.go`.
- Preserve module-driven structure under `src/modules/<feature>`:
  - `<feature>.controller.go` handles HTTP input/output.
  - `<feature>.service.go` handles business logic.
  - `<feature>.repository.go` handles database access.
  - `dto/` contains request/response structures.
  - `types/` contains model definitions.
- Build concrete objects in composition root (`cmd/api/main.go`) and inject into `server`.
- Add new env keys to `.env.example` and `.env`.
- Update `README.md` and `docs/*` when behavior, setup, or workflow changes.
- Avoid unnecessary file moves or architectural changes.

## Safety Rules
- Fail fast on invalid config.
- Wrap returned errors with context.
- Avoid destructive changes unless explicitly requested.
- Prefer minimal, reversible edits.

## Anti-Patterns
- "Fix all my backend" without context.
- Combining unrelated tasks in one prompt.
- Refactor + feature + migration in one request without boundaries.
- Ignoring verification or not reporting verification limits.
