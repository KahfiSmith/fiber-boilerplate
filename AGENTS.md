# AGENTS.md

## Purpose
This file defines how Codex should work in this repository so responses are consistent, practical, and production-minded.

## Source of Truth (Read First)
- Architecture: `docs/architecture/overview.md`
- API reference: `docs/api/overview.md` (endpoints: `docs/api/authentication.md`)
- Database notes: `docs/database/schema.md`
- Repository rules: `docs/README.md`
- Coding standards: `docs/conventions/coding.md`
- Implementation patterns: `docs/conventions/validation.md`
- Cross-repo: the frontend repo (`nextjs-boilerplate`) consumes this API; keep endpoint paths and error codes in sync with frontend `src/lib/api/endpoints.ts`.

## Core Behavior
- Be concise, technical, and action-oriented.
- Prefer implementing directly instead of only explaining.
- If requirements are unclear but low risk, make a reasonable assumption and proceed.
- If requirements are unclear and high risk (data loss/security/schema), ask a short clarification first.

## Engineering Principles
- DRY: avoid duplicated validation/bootstrap logic; extract shared helpers.
- SOLID: keep each module focused and inject dependencies from `cmd/api/main.go`.
- KISS: prefer straightforward implementations and minimal abstraction.

## Standard Workflow Per Prompt
1. Understand request and success criteria.
2. Scan related files quickly (`rg`, `sed`, `ls`).
3. Identify root cause/design gap before editing.
4. Apply minimal but complete changes.
5. Verify with local checks (`go test ./...`, `go vet ./...`, `go run ./cmd/api`) when available.
6. Return a short summary:
   - what changed
   - files touched
   - what to run next (if verification not possible in environment)

## Project Architecture Rules
- Keep config setup in `src/config`.
- Keep database setup in `src/database`.
- Keep shared infrastructure/utilities in `src/common/*` (`jwt`, `middleware`, `redis`, `response`, `server`, `validator`, `exceptions`).
- Keep route registration in `src/common/server/server.go`.
- Keep feature modules under `src/modules/<feature>` containing:
  - `<feature>.controller.go` — HTTP handlers (parse request, call service, return response)
  - `<feature>.service.go` — business logic
  - `<feature>.repository.go` — database operations
  - `dto/` — request and response structures
  - `types/` — domain models and types

## Coding Conventions
- Use small focused functions.
- Return wrapped errors with context (`fmt.Errorf("context: %w", err)`).
- Keep config validation strict and fail fast on startup.

## Change Boundaries
- Do not introduce new architectural layers unless requested.
- Do not move files/folders unless needed for the task.
- Do not remove existing behavior without mentioning it in the final summary.

## Quality Checklist Before Final Response
- Imports are valid and consistent with folder/package names.
- No stale references to moved packages/files.
- New env keys added to `.env.example` when needed.
- README updated when behavior/setup changes.

## Response Pattern
- Start with result first.
- Then list concrete file changes.
- End with exact commands the user should run locally to verify.
