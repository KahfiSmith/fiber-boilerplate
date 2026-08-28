# Self-Audit Checklist

> **Purpose:** This document is the **source of truth** for "is this repo
> healthy?". Every item has a status field. Re-run on every major refactor,
> release, or quarterly.
>
> **How to use:**
> 1. Run `pnpm verify:all` for automated items (marked `✅`)
> 2. Manually verify items marked `➖`
> 3. Fill the report section at the bottom
> 4. See [agent-instructions.md](agent-instructions.md) for the full workflow
>
> **Status legend:** ✅ pass | ⚠ drift | ❌ broken | ➖ N/A | ⏳ in progress | ⚪ cannot verify
>
> **Automation legend:** ✅ covered by `pnpm verify:audit` or sibling scripts | ➖ manual only

---

## A. Documentation Synchronization (8 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| A1 | All `.md` links inside `docs/` resolve to existing files | `pnpm docs:check` | ✅ |
| A2 | All `src/...` paths referenced in docs exist | `pnpm docs:check` | ✅ |
| A3 | All endpoints in docs match `src/modules/auth/auth.controller.go` | `pnpm docs:check` | ✅ |
| A4 | Every module under `src/modules/<name>/` has `docs/features/<name>.md` | `pnpm docs:check` | ✅ |
| A5 | `CHANGELOG.md` reflects all changes since last release | manual diff against `git log` | ➖ |
| A6 | `AGENTS.md` rules match current `package.json` scripts and `Makefile` | manual review | ➖ |
| A7 | `README.md` quick-start commands actually work | manual run | ➖ |
| A8 | Cross-repo doc links (to sibling frontend) resolve | `pnpm verify:cross-repo` | ✅ |

## B. Feature Completeness (10 items)

Each item: ✅ implemented & synced | ⚠ implemented but different | 🟡 partial | ❌ missing | 🔵 implemented but undocumented

| # | Feature | BE entry point | FE counterpart | Status |
|---|---|---|---|---|
| B1 | Login (email + password) | `POST /auth/login` | `useLogin` + `LoginForm` | ☐ |
| B2 | Register (auto-login) | `POST /auth/register` (returns `AuthResponse`) | `useRegister` | ☐ |
| B3 | Logout (idempotent) | `POST /auth/logout` | `useLogout` | ☐ |
| B4 | Refresh token rotation (atomic Lua) | `POST /auth/refresh` | `SessionProvider` + interceptor | ☐ |
| B5 | Reuse detection → revoke family | Lua script in `refresh.repository.go` | (transparent) | ☐ |
| B6 | Delete account (password confirm) | `DELETE /auth/account` | `useDeleteAccount` | ☐ |
| B7 | Forgot password | `POST /auth/forgot-password` | (FE: planned) | ☐ |
| B8 | Reset password | `POST /auth/reset-password` | (FE: planned) | ☐ |
| B9 | Verify email | `POST /auth/verify-email` | (FE: planned) | ☐ |
| B10 | Google OAuth (browser-navigated) | `GET /auth/google` + callback | `googleAuthUrl()` | ☐ |

## C. Functional Correctness (8 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| C1 | Refresh token rotation is atomic (no double-spend possible) | trace `AtomicRotationLuaScript` in `refresh.repository.go` | ➖ |
| C2 | Reuse detection revokes the entire family | trace Lua + `RevokeFamily` | ➖ |
| C3 | Access token contains expected claims (`sub`, `email`, `role`, `session_id`) | inspect `jwt.AccessTokenClaims` | ➖ |
| C4 | JWT validates issuer + audience | inspect `jwt.ValidateAccessToken` | ➖ |
| C5 | Bcrypt cost is between 10-14 | `config.go:BcryptCost` | ➖ |
| C6 | Cookie `HttpOnly: true` always set | inspect `auth.controller.go:setRefreshCookie` | ➖ |
| C7 | Cookie `SameSite` configurable, default `Lax` | `config.go:CookieSameSite` | ➖ |
| C8 | `GoogleCallback` does not redirect to arbitrary URL (state validates) | trace `oauth.service.go:ConsumeState` | ➖ |

## D. Business Logic (10 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| D1 | Login with wrong password returns `INVALID_CREDENTIALS` (not user-not-found) | trace `auth.service.go:Login` | ➖ |
| D2 | Login with non-existent user returns same `INVALID_CREDENTIALS` (no enumeration) | trace `auth.service.go:Login` | ➖ |
| D3 | Register rejects duplicate email with clear error | trace `auth.service.go:RegisterWithSession` | ➖ |
| D4 | Refresh rejects reused token by revoking the family | trace Lua script | ➖ |
| D5 | Logout is idempotent (no error if already logged out) | trace `auth.controller.go:Logout` | ➖ |
| D6 | `DeleteAccount` revokes all of the user's sessions (not just the current one) | trace `auth.service.go:DeleteAccount` | ➖ |
| D7 | `ResetPassword` revokes all of the user's sessions | trace `auth.service.go:ResetPassword` | ➖ |
| D8 | `VerifyEmail` is single-use (token deleted on success) | trace `auth.service.go:VerifyEmail` | ➖ |
| D9 | Rate limiter is per-IP, applied to all public auth routes | inspect `server.go:RegisterRoutes` | ➖ |
| D10 | OAuth auto-creates user with `IsEmailVerified=true` | trace `oauth.service.go:HandleCallback` | ➖ |

## E. Validation & Error Handling (10 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| E1 | All request DTOs use `validate` tags | inspect `dto/*.go` | ➖ |
| E2 | Validation errors return 400 with `VALIDATION_ERROR` code | inspect `response.go:HandleError` | ➖ |
| E3 | All 401 responses include a `code` (not empty string) | inspect `response.go` for `*fiber.Error` branch | ➖ |
| E4 | All 4xx/5xx responses have a consistent envelope | inspect `APIResponse` | ➖ |
| E5 | `INVALID_CREDENTIALS` is the only login failure code (no enumeration) | grep `auth.service.go` | ➖ |
| E6 | `REFRESH_TOKEN_REUSED` triggers family revocation | trace `auth.service.go:Refresh` | ➖ |
| E7 | Health check returns DB+Redis status (not just `ok`) | inspect `health.service.go:Check` | ➖ |
| E8 | Error response `message` is safe to expose (no internal stack trace) | inspect `HandleError` | ➖ |
| E9 | Logging on error does not leak user input (email is OK, password is NOT) | grep `slog.*Password\|slog.*Token` | ➖ |
| E10 | `c.Bind().JSON()` returns typed error, not silently ignores | inspect `validator.go:ParseAndValidate` | ➖ |

## F. Authentication & Authorization (8 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| F1 | All `Protected` routes validate Bearer token | inspect `auth.controller.go` route registration | ➖ |
| F2 | `RequireRole` middleware exists and is applied to admin routes | inspect usage in routes | ➖ |
| F3 | Direct API call to `/auth/me` without Bearer returns 401 | curl test | ➖ |
| F4 | Direct API call to `/auth/account` (DELETE) without Bearer returns 401 | curl test | ➖ |
| F5 | Direct API call to `/auth/me` with another user's ID does not return their data | inspect `Me` handler | ➖ |
| F6 | Session metadata is per-user, not shared across users | inspect `SessionMetadata` | ➖ |
| F7 | Refresh token family revocation invalidates ALL related sessions | trace Lua + `RevokeFamily` | ➖ |
| F8 | OAuth state is single-use (cannot be replayed) | trace `oauth.service.go:ConsumeState` | ➖ |

## G. Database & Migrations (8 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| G1 | All migrations are numbered sequentially (no gaps) | inspect `db/migrations/` | ✅ (audit) |
| G2 | Each migration has matching `.up.sql` and `.down.sql` | inspect filenames | ✅ (audit) |
| G3 | GORM `User` struct matches latest migration columns | compare `auth.type.go` with `000006_add_oauth_to_users.up.sql` | ➖ |
| G4 | Unique constraints on `email` are case-insensitive (Postgres `LOWER(email)`) | inspect `000003_add_users_email_lower_index.up.sql` + `auth.repository.go:FindByEmail` | ➖ |
| G5 | Partial unique index on `(oauth_provider, oauth_subject) WHERE oauth_provider IS NOT NULL` is present | inspect `000006_add_oauth_to_users.up.sql` | ➖ |
| G6 | `database.DB.AutoMigrate` at startup matches migrations (no drift) | inspect `cmd/api/main.go` | ➖ |
| G7 | Migrations do not contain destructive changes without `down` script | inspect each migration | ➖ |
| G8 | All foreign keys have `ON DELETE` clause (CASCADE or RESTRICT) | inspect SQL files | ➖ |

## H. Security (10 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| H1 | `JWT_ACCESS_SECRET` is not the placeholder `replace-with-...` in startup | run app with default env | ✅ (audit) |
| H2 | `REFRESH_TOKEN_HMAC_KEY` is not the placeholder | run app with default env | ✅ (audit) |
| H3 | Placeholder secrets are rejected in production (`APP_ENV=production`) | run with `APP_ENV=production` and default env | ➖ |
| H4 | `COOKIE_SECURE=true` in production config (default: `false` for dev) | inspect `.env.example` | ➖ |
| H5 | CORS `AllowOrigins` is not `"*"` (must be specific) | inspect `main.go` CORS config | ✅ (audit) |
| H6 | `ValidateOrigin` middleware rejects requests with wrong Origin | inspect `origin.middleware.go` | ➖ |
| H7 | Bcrypt used for password hashing (not plain SHA, not MD5) | grep `crypto/sha1\|md5` in `auth.service.go` | ✅ (audit) |
| H8 | No `log.Printf` in production code (must use slog) | grep `log\.(Printf\|Println\|Fatal)` | ✅ (audit) |
| H9 | No `panic(` outside main or test files | grep `panic(` | ✅ (audit) |
| H10 | Refresh token is HMAC-hashed before storing in Redis (not stored in plain) | inspect `refresh_token.go:hashRefreshToken` | ➖ |

## I. Code Quality (8 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| I1 | All files in `src/` use kebab-case or `_test.go` | inspect filenames | ✅ (audit) |
| I2 | All exported functions have a doc comment (best-effort) | grep `^func [A-Z]` and check for `// ` above | ✅ (audit) |
| I3 | `src/common/` does not import `src/modules/` (dependency direction) | grep imports | ✅ (audit) |
| I4 | `src/database/` does not import `src/modules/` | grep imports | ✅ (audit) |
| I5 | `src/config/` does not import `src/modules/` or `src/database/` | grep imports | ✅ (audit) |
| I6 | No dead code: exported functions that have no callers | parse + grep | ✅ (audit) |
| I7 | No `//nolint` without justification comment | grep `//nolint` | ✅ (audit) |
| I8 | No `_ = err` (silent error swallow) without comment | grep `_ = ` | ✅ (audit) |

## J. Performance (5 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| J1 | No N+1 in repository methods | inspect each `*Repository` method | ➖ |
| J2 | GORM connection pool sized appropriately | inspect `postgres.go:Connect` | ➖ |
| J3 | Redis client uses connection pool (not per-call) | inspect `redis.go:Connect` | ➖ |
| J4 | Lua script for refresh rotation is small (< 1KB) | inspect `refresh.repository.go:AtomicRotationLuaScript` | ➖ |
| J5 | Logger output goes to stdout (not file) for container-friendly logging | inspect `logger.go` | ➖ |

## K. Testing (5 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| K1 | `go test ./...` passes | `go test ./...` | ✅ (CI) |
| K2 | `go test -race ./...` passes | `go test -race ./...` | ➖ |
| K3 | Auth login → refresh → reuse → logout flow has integration test | inspect `auth_test.go` | ➖ |
| K4 | JWT generation + validation has unit test | inspect `jwt_test.go` | ➖ |
| K5 | Each module has at least one test file | inspect `*_test.go` | ✅ (audit) |

## L. Operational (5 items)

| # | Check | How to verify | Automation |
|---|---|---|---|
| L1 | `go.mod` and `go.sum` are committed and up to date | `git status` | ➖ |
| L2 | CI runs `go test`, `go vet`, `pnpm verify:audit` on every PR | inspect `.github/workflows/ci.yml` | ➖ |
| L3 | `.env.example` documents all required env vars | inspect | ✅ (audit) |
| L4 | No secrets committed (only `.env.example` is in git) | `git ls-files \| grep -E '^\.env$'` | ➖ |
| L5 | Graceful shutdown handles SIGTERM (in-flight requests complete) | inspect `main.go` | ➖ |

---

## Audit Summary (fill in)

**Last audit:** YYYY-MM-DD
**Auditor:** <name or "agent">
**Repo version:** <git sha>
**Items total:** 95
**Items passed (✅):** ?
**Items with drift (⚠):** ?
**Items broken (❌):** ?
**Items in progress (⏳):** ?
**Items N/A (➖):** ?
**Items cannot verify (⚪):** ?

**Overall verdict:** 🟢 / 🟡 / 🟠 / 🔴

**Top 3 priorities:**
1. ...
2. ...
3. ...

**Link to full report:** `reports/YYYY-MM-DD-audit.md`

---

## How to extend this checklist

When you add a new check:

1. Add a row to the appropriate section.
2. Mark `Automation: ✅` if a script enforces it, `➖` if manual.
3. If automated, also add the check to
   [`automated-checks.md`](automated-checks.md).
4. Commit with `docs(audit): add <category> check for <description>`.

When you find a new bug:
- Add a row to the relevant section with status ❌
- Open an issue / PR with the fix
- Update status to ✅ when shipped
