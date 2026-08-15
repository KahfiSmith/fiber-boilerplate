# Documentation Directory

These documents distinguish implemented behavior from planned work. Accepted
architecture and conventions live in the current sources of truth; planned
capabilities are called out explicitly rather than described as if they exist.

This is the backend (Go Fiber) repository. The companion frontend repository is
`nextjs-boilerplate` (Next.js App Router); cross-repo contracts are documented
in [API](api/overview.md).

## Current sources of truth

- [Architecture](architecture/overview.md) - System design, folder structure, run modes.
- [ADRs](architecture/adr/README.md) - Architecture decision records.
- [API](api/overview.md) - Backend contract, envelope, error codes, frontend relationship.
- [Authentication](api/authentication.md) - Endpoints, token flow, cookie handling.
- [Conventions](conventions/coding.md) - Go style, module layout, error wrapping.
- [Validation](conventions/validation.md) - Request validation, response envelope, exceptions.
- [Database](database/schema.md) - PostgreSQL + Redis persistence, migrations.
- [Features](features/README.md) - Feature modules (authentication, health).
- [Development](development/setup.md) - Setup, environment, commands, verification harness.
- [API collections](openapi/collection.json) - Importable API collection + environment for testing.
- [Infrastructure](infrastructure/deployment.md) - Docker, docker-compose, ports, CI/CD.

## Planned, not yet implemented

The following are intentionally not described as implemented:

- Task-state management (`task:begin`/`task:verify`/`task:handoff`) - not yet added.
- Additional modules beyond `auth` and `health`.

When a planned capability is shipped, promote only its durable decisions into
the current sources of truth.
