# Feature: Authentication & User Session

## Overview

User authentication including registration, login, session persistence via
HttpOnly refresh cookies, password reset, email verification, and account
deletion.

## Core flow

```text
HTTP request
  -> auth.controller.go (parse + validate DTO, origin + rate-limit middleware)
     -> auth.service.go (business logic)
        -> auth.repository.go (GORM: users) / refresh.repository.go (Redis)
           -> jwt.TokenService (HS256 access token)
              -> response.Success / response.HandleError
```

## Flow states

1. Register: create user (bcrypt hash, `role: user`, email unverified); returns
   a verification token (when `AUTH_DEBUG_EXPOSE_OTP=true`).
2. Login: validate credentials, create a session family in Redis, set the
   HttpOnly refresh cookie, return the access token.
3. Refresh: atomic rotation via Redis Lua script; reuse detection revokes the
   family (`REFRESH_TOKEN_REUSED`).
4. Logout: revoke session family + clear cookie (idempotent).
5. Delete account: password confirmation, delete user, clear cookie.

## Implementation map

| Concern | Files |
|---|---|
| Controller | `src/modules/auth/auth.controller.go` |
| Service | `src/modules/auth/auth.service.go` |
| Repository (users) | `src/modules/auth/auth.repository.go` |
| Repository (sessions) | `src/modules/auth/refresh.repository.go` |
| Token helpers | `src/modules/auth/refresh_token.go` |
| DTO | `src/modules/auth/dto/auth.dto.go` |
| Types | `src/modules/auth/types/auth.type.go` |
| Tests | `src/modules/auth/auth_test.go` |

## Endpoints

| Method | Path | Protected |
|---|---|---|
| `POST` | `/api/v1/auth/register` | no |
| `POST` | `/api/v1/auth/login` | no |
| `POST` | `/api/v1/auth/refresh` | no |
| `POST` | `/api/v1/auth/logout` | no |
| `POST` | `/api/v1/auth/forgot-password` | no |
| `POST` | `/api/v1/auth/reset-password` | no |
| `POST` | `/api/v1/auth/verify-email` | no |
| `POST` | `/api/v1/auth/resend-verification` | no |
| `DELETE` | `/api/v1/auth/account` | yes |
| `GET` | `/api/v1/auth/me` | yes |

Full details in [Authentication API](../api/authentication.md).

## Not yet implemented

- Email delivery (tokens are stored/returned only; no SMTP integration).
- Role-based route protection via `RequireRole` (helper exists, not wired).
