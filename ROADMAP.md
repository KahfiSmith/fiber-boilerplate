# Roadmap

- Status: Working draft (planning artifact — not implemented behavior)
- Date: 2026-08-17
- Purpose: single place listing what exists and what is planned for this
  backend, so feature work has a build order. Implemented features are
  documented in `docs/features/`; this file tracks intent.

## Conventions

- Status legend: **done** = implemented; **in-progress** = actively being built;
  **planned** = to build; **candidate** = considered, not committed.
- When a feature ships, mark it **done** and make sure it has a
  `docs/features/<feature>.md` (the `docs:check` gate requires it).

## Done

- [x] Authentication & user session — register, login, refresh (atomic
      rotation + reuse detection), logout, forgot-password, reset-password,
      verify-email, resend-verification, delete-account, me — see
      `docs/features/auth.md`
- [x] Health check (`GET /api/v1/health`, DB + Redis status) — see
      `docs/features/health.md`

## In progress

(none)

## Planned

1. Change password endpoint (`POST /api/v1/auth/change-password`,
   authenticated) — password confirmation + rotate session.
2. Session list + revoke (`GET /api/v1/me/sessions`,
   `POST /api/v1/me/sessions/{sessionId}/revoke`) — reuse the existing
   `auth:family:*` Redis structure.
3. Email delivery — currently verification/reset tokens are only returned in
   responses when `AUTH_DEBUG_EXPOSE_OTP=true`; wire a real mailer (SMTP or
   provider).
4. Production cookie hardening — `COOKIE_SECURE=true`,
   `__Secure-refresh_token` name, explicit `COOKIE_DOMAIN` in production env.

## Candidate (not committed)

- RBAC beyond the existing `RequireRole` helper (role-based route guards).
- Password-less login (OIDC / magic link).
- Account lockout / brute-force protection beyond the per-IP rate limiter.
- Audit log for security events.

## Open questions

- Product vision is not finalized (see `docs/product/overview.md`).
- Whether email tokens should be short-lived OTPs (numeric) vs current UUIDs.

## Follow-ups

- After each planned item ships: update this file's status, add the feature
  doc from `docs/features/_TEMPLATE.md`, and run `pnpm verify:all`.
