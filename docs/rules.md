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
- Keep third-party setup centralized in `pkg/configs`.
- Keep route registration in `pkg/server/routes.go`.
- Preserve layer boundaries:
  - controllers -> dto + services
  - services -> repositories + entities
  - repositories -> models/entities mapping
  - server -> wiring/runtime
- Build concrete objects in composition root (`cmd/api/main.go`) and inject into `server`.
- Add new env keys to `.env.example`.
- Update `README.md`, `docs/*`, and `tools/agent/*` docs/comments when behavior/setup/workflow changes.
- Avoid unnecessary file moves or architectural changes.

## Principal Engineer Posture
- Optimize for correctness, leverage, and future maintenance rather than short-term convenience.
- Prefer deleting ambiguity over adding cleverness.
- Challenge weak assumptions when they create security, operational, or API-design risk.
- Preserve optionality: prefer changes that are easy to revert, extend, or verify.
- Keep public API surface intentional; if a feature exists, document the product or operational reason for it.

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
