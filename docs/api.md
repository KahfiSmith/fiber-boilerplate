# API

Current API contract.

## Base URL
- Local: `http://localhost:3000`
- Base prefix: `/api/v1`

## Health Check Endpoint
- `GET /api/v1/health`
- Handler: `src/modules/health/controller/health.controller.go`

## Auth Endpoints
- `POST /api/v1/auth/register` — Register a new user
- `POST /api/v1/auth/login` — Login with email and password (returns access token, sets HttpOnly refresh token cookie)
- `POST /api/v1/auth/refresh` — Refresh access token using cookie
- `POST /api/v1/auth/forgot-password` — Request password reset token
- `POST /api/v1/auth/reset-password` — Submit new password using reset token
- `POST /api/v1/auth/logout` — Logout specific device session (requires Bearer access token)
- `GET /api/v1/auth/me` — Protected endpoint returning current user info (requires Bearer access token)

## Auth Protection & Security Features
- Protected endpoints require `Authorization: Bearer <access_token>`.
- JWT tokens are verified using HS256 algorithm and `JWT_SECRET`.
- Multi-device sessions supported via UUID `session_id` in Redis keys (`refresh_token:<userID>:<sessionID>`).
- Rate limiting active on auth endpoints (default 5 requests/min per IP via Redis).
- Email addresses are normalized (`lowercase` & `trimmed`).
- RBAC middleware (`RequireRole`) available for role-based route restriction (`user`, `admin`).

## Auth Request Contracts
- Register:
  - `name` (string, required, min 2)
  - `email` (string, required, email format)
  - `password` (string, required, min 8)
- Login:
  - `email` (string, required, email format)
  - `password` (string, required, min 8)
- Forgot Password:
  - `email` (string, required, email format)
- Reset Password:
  - `reset_token` (string, required)
  - `new_password` (string, required, min 8)
- Refresh:
  - Reads `refresh_token` from HttpOnly cookie
- Logout:
  - Requires `Authorization: Bearer <access_token>` header
  - Clears `refresh_token` cookie and revokes session in Redis

## Auth Response Contracts
- Register response `data`:
  - `id` (uint)
  - `name` (string)
  - `email` (string)
  - `role` (string)
  - `created_at` (timestamp)
  - `updated_at` (timestamp)
- Login response `data`:
  - `access_token` (string)
  - `user` (`id`, `name`, `email`, `role`, `created_at`, `updated_at`)
- Refresh response `data`:
  - `access_token` (string)
  - `user` (`id`, `name`, `email`, `role`, `created_at`, `updated_at`)
- Forgot Password response `data`:
  - `reset_token` (string, returned only when `AUTH_DEBUG_EXPOSE_OTP=true`)
- Reset Password response:
  - `success`: true
  - `message`: `"Password reset successfully"`
- Logout response:
  - `success`: true
  - `message`: `"Logged out successfully"`
- `GET /api/v1/auth/me` response `data`:
  - `id` (uint)
  - `email` (string)
  - `role` (string)

## Success Status Codes
- Health: `200`
- Register: `201`
- Login: `200`
- Refresh: `200`
- Forgot Password: `200`
- Reset Password: `200`
- Logout: `200`
- Me: `200`

## Response Envelope
Defined in `src/common/response/response.go`:
- `success` (bool)
- `message` (string)
- `data` (any, optional)
- `error` (any, optional)

## Location Reference
- Route registration: `src/common/server/server.go`
- Auth routes: `src/modules/auth/auth.controller.go`
- Health routes: `src/modules/health/health.route.go`
