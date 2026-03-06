# Workflow

Standard workflow Codex should follow in this repo.

## 1. Intake
- Confirm objective and success criteria.
- Identify mode: `bugfix`, `feature`, `refactor`, `review`.

## 2. Discovery
- Inspect relevant files with fast search (`rg`).
- Identify root cause and impacted layers.
- Avoid broad edits before confirming scope.

## 3. Plan
- Define minimal file set to change.
- Keep architecture boundaries intact:
  - setup in `pkg/configs`
  - wiring in `pkg/server`
  - request/response contracts in `pkg/dto`
  - domain objects in `pkg/entities`
  - business logic in `pkg/services`
  - data access in `pkg/repositories`

## 4. Implement
- Make focused edits.
- Keep behavior changes intentional.
- Add env keys to `.env.example` if needed.

## 5. Verify
- Preferred checks:
  - `go test ./...`
  - `go run ./cmd/api`
- If tooling unavailable, state limitation clearly.

## 6. Report
- Outcome first.
- File-by-file summary.
- Commands user should run locally.

## Done Checklist
- Imports and package names consistent.
- No stale references after move/rename.
- Docs updated for behavior/setup changes.
- API response shape preserved unless requested.
