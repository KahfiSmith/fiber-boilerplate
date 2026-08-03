# Agent Prompt Patterns

Use this structure when prompting Codex for consistent outputs.

## High-Quality Prompt Template
```text
Goal:
- What you want built/fixed.

Context:
- Related files/folders under `src/` or `cmd/`.
- Current behavior and expected behavior.
- Runtime mode when relevant (`host`, `docker compose`, or both).
- Documentation surfaces that must stay aligned (`README.md`, `docs/*`).

Constraints:
- Module-driven architecture rules (`src/modules/<feature>`).
- "Do not change X" boundaries.

Validation:
- Commands to run (`go test ./...`, `go run ./cmd/api`, `go vet ./...`).
- Expected success criteria.

Output:
- Summary format (files changed, reason, next steps).
```

## Prompt Modes

### `bugfix` Mode
Use when fixing compile or runtime errors.

Prompt example:
```text
Mode: bugfix
Fix type assertion error in src/modules/auth/auth.logout.go.
Constraints: preserve JWT secret verification in middleware.
Validation: go vet ./... and go build ./cmd/api.
```

### `feature` Mode
Use when adding new endpoint/module behavior.

Prompt example:
```text
Mode: feature
Add GET /api/v1/health endpoint under src/modules/health.
Register route in src/common/server/server.go.
```

### `refactor` Mode
Use when restructuring code without behavior changes.

Prompt example:
```text
Mode: refactor
Extract JWT logic to src/common/jwt/jwt.go.
No behavior changes.
Keep public function names stable.
```
