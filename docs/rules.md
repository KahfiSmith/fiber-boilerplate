# Agent Rules

Repository-specific rules for prompting and execution quality.

## Prompt Rules
- Include concrete file paths whenever possible.
- Include exact error messages instead of paraphrasing.
- State hard constraints explicitly (what must not change).
- Keep one primary objective per prompt.
- Include validation commands (`go test ./...`, `go run ./cmd/api`) when relevant.
- Mention output preference (summary format, file list, next steps).

## Execution Rules
- Keep third-party setup centralized in `pkg/configs`.
- Keep route registration in `pkg/server/routes.go`.
- Preserve layer boundaries:
  - controllers -> dto + services
  - services -> repositories + entities
  - repositories -> models/entities mapping
  - server -> wiring/runtime
- Build concrete objects in composition root (`cmd/api/main.go`) and inject into `server`.
- Add new env keys to `.env.example`.
- Update docs when behavior/setup changes.
- Avoid unnecessary file moves or architectural changes.

## Principle Rules
- DRY:
  - If the same validation/helper appears in multiple files, extract it once.
- SOLID:
  - Do not make `server` construct domain dependencies directly.
  - Keep each package focused on one reason to change.
- KISS:
  - Prefer clear and direct implementations over generic abstractions.
  - Reject unnecessary indirection when one small function is enough.

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
