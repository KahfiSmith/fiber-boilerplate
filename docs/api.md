# API

Current API contract.

## Base URL
- Local: `http://localhost:3000`
- Base prefix: `/api/v1`

## Endpoint: Health Check
- Method: `GET`
- Path: `/api/v1/health`
- Handler: `pkg/controllers/health.go`

## Auth Endpoints
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/otp/verify`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `GET /api/v1/auth/sessions`
- `POST /api/v1/auth/sessions/revoke`
- `POST /api/v1/auth/sessions/revoke-all`
- Handler implementation: `pkg/controllers/auth.go`

## Auth Request Contracts
- Register:
  - `name` (string, required)
  - `email` (string, required, email format)
  - `password` (string, required, min 8)
- Login:
  - `email` (string, required, email format)
  - `password` (string, required)
- Verify OTP:
  - `challenge_id` (string, required)
  - `otp` (string, required, 6 digits)
- Forgot password:
  - `email` (string, required, email format)
- Reset password:
  - `challenge_id` (string, required)
  - `otp` (string, required, 6 digits)
  - `new_password` (string, required, min 8)
- Refresh:
  - `refresh_token` (string, required)
- Logout:
  - `refresh_token` (string, required)
- Revoke session:
  - `session_id` (string, required)

## Auth Response Contracts
- Register response `data`:
  - `access_token`
  - `refresh_token`
  - `token_type` (`Bearer`)
  - `expires_in_sec`
  - `session_id`
  - `user` (`id`, `name`, `email`)
- Verify OTP response `data`:
  - `access_token`
  - `refresh_token`
  - `token_type` (`Bearer`)
  - `expires_in_sec`
  - `session_id`
  - `user` (`id`, `name`, `email`)
- Refresh response `data`:
  - `access_token`
  - `refresh_token`
  - `token_type` (`Bearer`)
  - `expires_in_sec`
  - `session_id`
  - `user` (`id`, `name`, `email`)
- Login response `data`:
  - `challenge_id`
  - `expires_in_sec`
  - `otp` (only when `AUTH_DEBUG_EXPOSE_OTP=true`, or legacy `AUTH_DEBUG_EXPOSE_TOKENS=true`)
- Forgot password response `data`:
  - `challenge_id`
  - `expires_in_sec`
  - `otp` (only when `AUTH_DEBUG_EXPOSE_OTP=true`, or legacy `AUTH_DEBUG_EXPOSE_TOKENS=true`)
- `GET /api/v1/auth/me` response `data`:
  - `id`
  - `name`
  - `email`
- `GET /api/v1/auth/sessions` response `data` item:
  - `session_id`
  - `user_agent`
  - `ip_address`
  - `created_at`
  - `expires_at`
  - `current`

## Success Status Codes
- Health: `200`
- Register: `201`
- Login: `200`
- Forgot password: `200`
- Verify OTP: `200`
- Reset password: `200`
- Refresh: `200`
- Logout: `200`
- Me: `200`
- Sessions: `200`
- Revoke session: `200`
- Revoke all sessions: `200`

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "message": "service is healthy",
    "service": "fiber-boilerplate",
    "timestamp": "2026-03-05T10:00:00Z"
  }
}
```

## Response Envelope
Defined in `pkg/dto/response/common.go`:
- `success` (bool)
- `message` (string, optional)
- `data` (any, optional)
- `error` (any, optional)

## DTO Convention
- Request DTOs should be placed in `pkg/dto/request`.
- Response DTOs should be placed in `pkg/dto/response`.
- Controllers should map entities into response DTOs before returning JSON.

## Notes
- Route registration entrypoint: `pkg/server/routes.go` (can delegate to `pkg/server/routes/*` modules).
- Response helper functions: `pkg/utils/response.go`.
- Auth request/response DTOs live in `pkg/dto/request/auth.go` and `pkg/dto/response/auth.go`.
