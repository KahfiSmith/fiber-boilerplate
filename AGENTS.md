# AGENTS.md

## Purpose
This file defines how Codex should work in this repository so responses are consistent, practical, and production-minded.

## Source of Truth (Read First)
- Architecture: `docs/architecture.md`
- API reference: `docs/api.md`
- Database notes: `docs/database.md`
- Repository rules: `docs/rules.md`
- Coding standards: `docs/coding-standards.md`
- Implementation patterns: `docs/patterns.md`

## Core Behavior
- Be concise, technical, and action-oriented.
- Prefer implementing directly instead of only explaining.
- If requirements are unclear but low risk, make a reasonable assumption and proceed.
- If requirements are unclear and high risk (data loss/security/schema), ask a short clarification first.

## Engineering Principles
- DRY: avoid duplicated validation/bootstrap logic; extract shared helpers.
- SOLID: keep each package focused and inject dependencies from `cmd/api/main.go`.
- KISS: prefer straightforward implementations and minimal abstraction.

## Standard Workflow Per Prompt
1. Understand request and success criteria.
2. Scan related files quickly (`rg`, `sed`, `ls`).
3. Identify root cause/design gap before editing.
4. Apply minimal but complete changes.
5. Verify with local checks (`go test ./...`, `go run ./cmd/api`) when available.
6. Return a short summary:
   - what changed
   - files touched
   - what to run next (if verification not possible in environment)

## Project Architecture Rules
- Keep third-party setup in `pkg/configs`:
  - `viper` config loading
  - `zap` logger setup
  - `gorm` DB setup
  - `fiber` app/middleware/listen setup
  - `validator` initialization
- Keep HTTP server composition in `pkg/server`.
- Keep route registration in `pkg/server/routes.go`.
- Keep business logic in `pkg/services`.
- Keep data access contracts/implementations in `pkg/repositories`.
- Keep handlers/controllers thin (`pkg/controllers`): parse request, call service, return response.

## Coding Conventions
- Use small focused functions.
- Avoid global mutable state unless necessary.
- Return wrapped errors with context (`fmt.Errorf("context: %w", err)`).
- Keep config validation strict and fail fast on startup.
- Prefer explicit config fields over magic constants in runtime code.

## Change Boundaries
- Do not introduce new architectural layers unless requested.
- Do not move files/folders unless needed for the task.
- Do not remove existing behavior without mentioning it in the final summary.

## Quality Checklist Before Final Response
- Imports are valid and consistent with folder/package names.
- No stale references to moved packages/files.
- New env keys added to `.env.example` when needed.
- README updated when behavior/setup changes.
- If tests/build cannot run in this environment, state that clearly.

## Response Pattern
- Start with result first.
- Then list concrete file changes.
- End with exact commands the user should run locally to verify.
