# Architecture Decision Records (ADRs)

Directory housing architectural choices and technical designs.

- [ADR-001: HMAC-Hashed Refresh Tokens in Redis with Atomic Rotation](./ADR-001-hmac-refresh-token-redis.md) - why raw refresh tokens are never stored and rotation is atomic.
- [ADR-002: Composition Root with Dependency Injection](./ADR-002-composition-root.md) - why all wiring lives in `cmd/api/main.go`.

New ADRs should record a context, a decision, a consequence, and a status.
