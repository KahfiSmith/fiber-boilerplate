# Workflow

Standard workflow Codex should follow in this repo.

## 1. Intake
- Confirm objective and success criteria.
- Identify mode: `bugfix`, `feature`, `refactor`, `review`.
- Identify whether the task also changes public API shape, runtime workflow, or repo conventions.

## 2. Discovery
- Inspect relevant files with fast search (`rg`).
- Identify root cause and impacted layers.
- Avoid broad edits before confirming scope.
- Check whether the runtime assumption is host-based, container-based, or mixed.
- Check which documentation surfaces will become stale if the change lands.

## 3. Plan
- Define minimal file set to change.
- Keep architecture boundaries intact:
  - setup in `pkg/configs`
  - wiring in `pkg/server`
  - request/response contracts in `pkg/dto`
  - domain objects in `pkg/entities`
  - reusable model/entity translation in `pkg/mappers`
  - business logic in `pkg/services`
  - data access in `pkg/repositories`

## 4. Implement
- Make focused edits.
- Keep behavior changes intentional.
- Add env keys to `.env.example` if needed.
- If keeping a non-obvious feature or API, sharpen the reason in docs instead of leaving it implicit.

## 5. Verify
- Preferred checks:
  - `go test ./...`
  - `go run ./cmd/api`
- When Docker assets are part of the task, also verify with:
  - `docker compose up --build`
- If tooling unavailable, state limitation clearly.

## 6. Report
- Outcome first.
- File-by-file summary.
- Commands user should run locally.
- Mention documentation updates explicitly when they were part of the task.

## Done Checklist
- Imports and package names consistent.
- No stale references after move/rename.
- Docs updated for behavior/setup changes.
- `README.md`, `docs/*`, and `tools/agent/*` are aligned when the task affects them.
- API response shape preserved unless requested.
