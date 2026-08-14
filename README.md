# Fiber Boilerplate

Backend starter built with Go, Fiber v3, PostgreSQL, and Redis. Serves the
`nextjs-boilerplate` frontend.

## Development Setup

- **Backend (host):** `http://localhost:8080` (`APP_PORT=8080`)
- **Backend (docker-compose):** `http://localhost:3000` (compose overrides port)
- **Frontend:** `http://localhost:3000` (`FRONTEND_ORIGIN=http://localhost:3000`)

## Tech Stack

- Go `1.25.4`
- Fiber v3 (`github.com/gofiber/fiber/v3`)
- GORM + PostgreSQL (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- Redis (`github.com/redis/go-redis/v9`) for session storage and rate limiting
- Viper (`github.com/spf13/viper`) for configuration
- `go-playground/validator` for request validation
- JWT (`github.com/golang-jwt/jwt/v5`) for short-lived access tokens

## Security Architecture

- **Access Token:** Short-lived JWT (15m default), returned via JSON response, used via `Authorization: Bearer <access_token>` header. Strictly validated for HS256 algorithm, issuer, audience, and expiration.
- **Refresh Token:** Cryptographically secure 32-byte opaque random token. Sent ONLY via HttpOnly cookie (`refresh_token` in dev, `__Secure-refresh_token` in prod).
- **HMAC Storage:** Raw refresh token is never stored. Only HMAC-SHA256 hash is kept in Redis.
- **Atomic Rotation & Reuse Detection:** Atomic refresh token rotation backed by Redis Lua Scripts. Reuse of invalid/old tokens triggers revocation of the entire token family (`REFRESH_TOKEN_REUSED`).
- **Idempotent Logout:** Logout revokes the session family and clears the cookie even if the access token has expired.
- **CORS & Origin Check:** Restricts requests to `FRONTEND_ORIGIN` with credentials enabled.

## API Surface

- `GET /api/v1/health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/resend-verification`
- `DELETE /api/v1/auth/account` (Protected)
- `GET /api/v1/auth/me` (Protected)

## Verification Commands

Format code:
```bash
gofmt -w .
```

Run tests:
```bash
go test ./...
```

Run vet:
```bash
go vet ./...
```

Full verification harness (via `package.json` scripts):
```bash
pnpm verify:fast   # go build + vet + test + docs:check
pnpm verify:all    # verify:fast + risk classification + BE↔FE cross-repo sync check
```

A pre-commit hook runs `pnpm verify:fast` automatically on commit. GitHub
Actions CI runs build, vet, test, docs check, and risk classification on every
push/PR.

## Documentation

- [Documentation index](docs/README.md) - Architecture, API, database, conventions, development, infrastructure.
- Cross-repo contract with the frontend: see [API overview](docs/api/overview.md).
