# Feature: <Feature Name>

**Feature ID:** <lowercase-kebab-case-feature-id>
**Status:** implemented | in-progress | planned
**Owner:** <name or team>
**Risk:** low | medium | high
**Module:** <module under src/modules/, e.g. auth>
**Related docs:** <link to API doc, ADR, or N/A>

Date: YYYY-MM-DD  
Related issue/PR: <link or N/A>

## Objective

Describe the concrete outcome this feature delivers to the user.

## Overview

<1-3 sentences: what the user gets and why it exists.>

## Constraints

- Architecture constraints:
- Product/runtime constraints:
- Out of scope:

## Impact Areas

- API/endpoints: yes | no
- Auth/session/RBAC: yes | no
- DB/migrations: yes | no
- Redis/queues/jobs: yes | no
- Env/config/secrets: yes | no
- Observability/logging: yes | no
- External integrations: yes | no
- CI/release/harness: yes | no

## Core flow

```text
<ASCII diagram or numbered steps:
HTTP request -> middleware -> controller -> service -> repository/redis -> response>
```

## Flow states

<Numbered states the feature moves through, when relevant.>

## Implementation map

| Concern | Files |
|---|---|
| Route | `src/modules/<feature>/...` |
| Controller | `src/modules/<feature>/<feature>.controller.go` |
| Service | `src/modules/<feature>/<feature>.service.go` |
| Repository | `src/modules/<feature>/<feature>.repository.go` |
| DTO | `src/modules/<feature>/dto/...` |
| Types | `src/modules/<feature>/types/...` |
| Tests | `src/modules/<feature>/..._test.go` |

## Endpoints

<Table of endpoints this feature exposes; link to the relevant API doc.>

## Acceptance Criteria

1.
2.
3.

## Implementation Checklist

- [ ] Step 1
- [ ] Step 2
- [ ] Step 3

## Decision Log

- YYYY-MM-DD: <decision> -> <reason>

## Verification

List exact commands and outcomes.

```bash
pnpm verify:all
```

Additional targeted checks when relevant:

```bash
go build ./...
go vet ./...
go test ./...
pnpm verify:cross-repo
```

## Runtime Evidence

Required when static checks do not sufficiently prove behavior.

- Environment:
- Dependencies/services:
- Executed request/flow:
- Relevant logs/request IDs:
- Notes:

## Risks And Mitigations

- Risk:
- Mitigation:

## Completion Notes

Summarize what shipped, what changed, and any important caveats.

## Follow-Ups

- [ ] <open items / known debt>
