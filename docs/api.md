# API Documentation

Current API contract and authentication specifications.

## Base URL
- Local Backend: `http://localhost:8080`
- Frontend Origin: `http://localhost:3000`
- Base Prefix: `/api/v1`

## Health Check Endpoint
- `GET /api/v1/health`
- Handler: `src/modules/health/controller/health.controller.go`

## Auth Endpoints
- `POST /api/v1/auth/login` — Login with email & password. Returns access token in JSON body and sets HttpOnly cookie for refresh token.
- `POST /api/v1/auth/refresh` — Atomic refresh token rotation using HttpOnly cookie (no request body).
- `POST /api/v1/auth/logout` — Revokes session/family in Redis and clears HttpOnly cookie. Idempotent.
- `POST /api/v1/auth/register` — Register a new user.
- `POST /api/v1/auth/forgot-password` — Request password reset token.
- `POST /api/v1/auth/reset-password` — Submit new password using reset token.
- `POST /api/v1/auth/verify-email` — Verify email address.
- `POST /api/v1/auth/resend-verification` — Resend email verification token.
- `DELETE /api/v1/auth/account` — Delete user account (requires Bearer access token).
- `GET /api/v1/auth/me` — Protected endpoint returning current user profile (requires Bearer access token).

## Auth Protection & Security Features
- **Access Token:** Short-lived JWT (default 15m) using `HS256`. Sent in response JSON, used via `Authorization: Bearer <access_token>` header. Contains claims: `sub`, `iss`, `aud`, `iat`, `nbf`, `exp`, `jti`, `session_id`, `role`, `token_type: "access"`.
- **Refresh Token:** Cryptographically secure 32-byte opaque random token encoded in `base64.RawURLEncoding`. Sent ONLY via HttpOnly Cookie (`refresh_token` in dev, `__Secure-refresh_token` in prod).
- **No Raw Token Storage:** Raw refresh token is NEVER stored in database or Redis. Only HMAC-SHA256 hash (`REFRESH_TOKEN_HMAC_KEY`) is stored.
- **Atomic Rotation & Reuse Detection:** Executed atomically in Redis using Lua Script. If a previously used refresh token is presented, all refresh tokens in the family are revoked (`REFRESH_TOKEN_REUSED`).
- **Origin Validation & CORS:** Strict CORS matching `FRONTEND_ORIGIN` (`http://localhost:3000`).

## Token Lifetimes & Configuration
- **Access Token TTL:** 15m (`ACCESS_TOKEN_TTL`)
- **Refresh Token TTL:** 168h / 7 days (`REFRESH_TOKEN_TTL`)
- **Cookie Name:** `refresh_token` (dev), `__Secure-refresh_token` (prod)
- **Cookie Path:** `/api/v1/auth`
- **Cookie SameSite:** `Lax`

## Response Contracts

### Success Envelope
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "<jwt-access-token>",
    "expires_in": 900,
    "user": {
      "id": 1,
      "name": "User Name",
      "email": "user@example.com",
      "role": "user"
    }
  }
}
```

### Error Response Codes
- `ACCESS_TOKEN_MISSING` (401)
- `ACCESS_TOKEN_INVALID` (401)
- `ACCESS_TOKEN_EXPIRED` (401)
- `REFRESH_TOKEN_MISSING` (401)
- `REFRESH_TOKEN_INVALID` (401)
- `REFRESH_TOKEN_EXPIRED` (401)
- `REFRESH_TOKEN_REUSED` (401)
- `SESSION_REVOKED` (401)
- `INVALID_CREDENTIALS` (401)
- `FORBIDDEN` (403)
