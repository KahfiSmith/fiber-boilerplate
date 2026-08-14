# Database

Persistence split between PostgreSQL (users) and Redis (sessions/tokens).

## PostgreSQL (GORM)

- Connection: `src/database/postgres.go`, DSN built by `src/config/config.go`.
- Auto-migration on startup: `database.DB.AutoMigrate(&types.User{})` in
  `cmd/api/main.go`.
- `users` table fields (`src/modules/auth/types/auth.type.go`):
  `id`, `name`, `email` (unique), `password_hash` (hidden from JSON), `role`,
  `is_email_verified`, `created_at`, `updated_at`.

### Migrations

- Directory: `db/migrations/` (`.up.sql` / `.down.sql`).
- Scripts: `scripts/migrate.sh`, `scripts/migrate-status.sh`,
  `scripts/migrate-down.sh`.

## Redis

- Connection: `src/common/redis/redis.go` (global `Client`).
- Stores refresh sessions, reset/verification tokens, and rate limits.

| Key | TTL | Purpose |
|---|---|---|
| `auth:session:<session_id>` | refresh TTL | Session metadata JSON |
| `auth:refresh:active:<hash>` | refresh TTL | Active token → session |
| `auth:refresh:used:<hash>` | refresh TTL | Used token → session |
| `auth:family:<family_id>` | refresh TTL | Token family members |
| `reset_token:<token>` | 15m | Password reset |
| `verify_email_token:<token>` | 24h | Email verification |
| `rate_limit:<ip>` | window | Per-IP rate limiting |

Rotation and reuse detection are atomic via a Lua script
(`AtomicRotationLuaScript` in `src/modules/auth/refresh.repository.go`).

## Environment keys

| Group | Keys |
|---|---|
| DB | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `DB_TIMEZONE`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME` |
| Redis | `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_KEY_PREFIX` |

> Note: `config.go` reads `REDIS_HOST`/`REDIS_PORT`. The `.env.example` also
> lists a `REDIS_ADDR` key which is not consumed by the current config —
> prefer `REDIS_HOST`/`REDIS_PORT`.

## Docker networking

- Host mode: `DB_HOST=127.0.0.1`, `REDIS_HOST=127.0.0.1`.
- docker-compose mode: `DB_HOST=postgres`, `REDIS_HOST=redis`.
