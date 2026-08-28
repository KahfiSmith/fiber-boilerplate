# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- Restructured documentation into nested domain directories under `docs/` (api, architecture, conventions, database, development, infrastructure).
- Aligned documentation with the actual codebase (endpoints, error codes, env keys, folder structure, run modes).
- Documented the cross-repo relationship with the `nextjs-boilerplate` frontend.
- **Register endpoint now auto-logs in the user**: returns the same `dto.AuthResponse` envelope as `/auth/login` (access_token + refresh cookie), matching the FE contract. `Register` service method renamed to `RegisterWithSession`.
- Rate limiter now uses a Lua script for atomic `INCR` + conditional `EXPIRE`, preventing the previous race where a key could lose its TTL.
- `HandleError` now maps `*fiber.Error` to a proper code string (`BAD_REQUEST`/`UNAUTHORIZED`/`FORBIDDEN`/`NOT_FOUND`/`CONFLICT`/`TOO_MANY_REQUESTS`/`INTERNAL_SERVER_ERROR`) so the FE always receives a non-empty `code` field.
- Replaced `log.Printf` calls in `auth.service.go`, `database/postgres.go`, and `common/redis/redis.go` with structured `slog` logger (honors `LOG_LEVEL`/`LOG_ENCODING` from `.env`).
- `config.Load` now supports `REDIS_ADDR=host:port` as a fallback when `REDIS_HOST`/`REDIS_PORT` are not set, fixing startup crash when copying `.env.example` as-is.
- CORS no longer advertises the unused `X-CSRF-Token` header (no CSRF middleware implemented).

### Removed
- Dropped `db/migrations/000002_create_auth_tables.*` (tables `auth_sessions`, `otp_challenges`, `auth_rate_limits` were never read by the Go code — sessions/OTP live in Redis). Added `db/migrations/000007_drop_unused_auth_tables.*` to clean up existing dev DBs.
- Unused `framer-motion` dependency removed from `nextjs-boilerplate/package.json` (and `AGENTS.md`).

### Added
- Verification harness: `package.json` with `verify:fast`/`verify`/`verify:all` scripts, risk classification, and cross-repo sync check.
- Pre-commit hook (`.githooks/pre-commit`) running `pnpm verify:fast`.
- GitHub Actions CI (`.github/workflows/ci.yml`): build, vet, test, docs check, risk classification.
- Feature documentation gate: new modules must be documented in `docs/features/<module>.md` (template enforced by `docs:check`).
- API collection + environment template under `docs/openapi/` for testing the API (Postman v2.1 format).
- Removed swagger artifacts (`docs/swagger.json`, `docs/swagger.yaml`, `scripts/swagger-generate.sh`) — the API collection under `docs/openapi/` replaces them as the importable contract.
- Added `ROADMAP.md` and `docs/product/overview.md`.
- Google SSO (OIDC, backend-first): `GET /auth/google` + `/auth/google/callback`, auto-create OAuth users, `password_hash` nullable + `oauth_provider`/`oauth_subject` columns (migration 000006). Disabled by default (`GOOGLE_ENABLED=false`).
- Structured `slog` logger in `src/common/logger/logger.go` (console or JSON output, level filtering).
