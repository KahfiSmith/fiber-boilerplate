# Agent Prompt Patterns

Use this structure when prompting Codex for consistent outputs.

## High-Quality Prompt Template
```text
Goal:
- What you want built/fixed.

Context:
- Related files/folders.
- Current behavior and expected behavior.

Constraints:
- Architecture rules, package location, style requirements.
- "Do not change X" boundaries.

Validation:
- Commands to run (go test, go run, lint).
- Expected success criteria.

Output:
- What summary format you want (files changed, reason, next steps).
```

## Skill-Like Prompt Modes
Treat each prompt as one of these modes.

## `bugfix` Mode
Use when you see compile/runtime errors.

Required in prompt:
- Exact error text
- File path and line (if available)
- Repro command

Prompt example:
```text
Mode: bugfix
Fix compile error in pkg/server/app.go.
Error: undefined: requestid.FromContext
Constraints: keep routes in pkg/server.
Validation: go test ./... and go run ./cmd/api.
```

## `feature` Mode
Use when adding new endpoint/module behavior.

Required in prompt:
- Input/output contract
- Affected layer (controller/service/repository)
- Backward compatibility expectations

Prompt example:
```text
Mode: feature
Add GET /api/v1/version endpoint.
Request/response contracts must use pkg/dto/request and pkg/dto/response.
Place route registration in pkg/server/routes.go.
```

## `refactor` Mode
Use when restructuring code without behavior changes.

Required in prompt:
- What should remain unchanged
- What should be moved/renamed
- Scope limit

Prompt example:
```text
Mode: refactor
Move DB config helpers from config.go to db.go.
No behavior changes.
Keep public function names stable.
```

## `review` Mode
Use for bug/risk-focused code review.

Required in prompt:
- Branch or changed files
- Risk focus (security/performance/regression)

Prompt example:
```text
Mode: review
Review changed files for runtime regressions and missing tests.
Focus on config validation and server startup path.
```

## Prompt Rules
- See `docs/rules.md`.

## Anti-Patterns
- See `docs/rules.md`.
