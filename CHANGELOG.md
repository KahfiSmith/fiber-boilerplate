# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **Config validation now actually runs.** `config.Load` now calls
  `validator.New().Struct(&cfg)` after `viper.Unmarshal`, so the
  `validate:"required,min=32"` tags on `AuthConfig.JWTAccessSecret` and
  `RefreshTokenHMACKey` are no longer inert. Deploys with missing or
  short secrets now fail at startup instead of silently producing
  insecure tokens.
- **Placeholder secrets rejected in production.** When `APP_ENV=production`,
  starting the server with `JWT_ACCESS_SECRET=replace-with-strong-random-secret`
  (the `.env.example` default) now hard-fails. Same for
  `REFRESH_TOKEN_HMAC_KEY`. Development mode still accepts placeholders.
- **`DeleteAccount` and `ResetPassword` now revoke all user sessions.**
  Added `RefreshRepository.RevokeAllByUserID` (uses Redis `SCAN` to find
  matching sessions, then revokes each family). Previously, after
  deleting an account or resetting a password, all other sessions
  remained valid until their 7-day refresh TTL — a security hole.
  Failures here are logged at `WARN` but do not fail the operation
  (destructive action already succeeded).
- **Silent error swallows are now explicit.** Added justification comments
  to `_ = c.service.Logout(refreshToken)` in `auth.controller.go` and
  `_ = s.refreshRepo.RevokeSessionByTokenHash(...)` in `auth.service.go`.
- **Removed dead code.** `types.JwtPayload` struct was never referenced —
  removed from `src/modules/auth/types/auth.type.go`.
- **GORM/migration drift fixed.** `User.Name` GORM tag now says
  `varchar(120)` to match migration `000001_create_users_table.up.sql`.

### Added
- Health service test (`src/modules/health/service/health_service_test.go`)
  covering the structure and basic behavior of `Check()`.

### Added (audit framework)
- **Self-audit framework** under `docs/auditing/`:
  - `SELF_AUDIT.md` — 95-item checklist grouped by area (docs sync, features, security, code quality, performance, testing, ops).
  - `agent-instructions.md` — step-by-step instructions for an AI agent running a deep audit.
  - `checklist-template.md` — mandatory report template.
  - `automated-checks.md` — inventory of automated checks and how to extend.
  - `overview.md` — index of the auditing framework.
- New scripts in `scripts/harness/`:
  - `check-self-audit.mjs` — `pnpm verify:audit`: 30+ machine-checkable items (naming, Go hygiene, env consistency, architecture rules, security patterns, migration drift, dead code, test coverage).
  - `audit-report.mjs` — `pnpm audit:report`: aggregates `go build/vet/test` + `docs:check` + `verify:risk` + `verify:cross-repo` + `verify:audit` into `.audit/report-YYYY-MM-DD.{json,md}`.
- New `package.json` scripts: `docs:check` (now wired into `verify:fast`), `verify:audit`, `audit:report`.
- `verify:all` now includes `verify:audit` as a blocking step.

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
