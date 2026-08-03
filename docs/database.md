# Database

Database setup and conventions.

## Current Driver
- `gorm` with PostgreSQL driver (`gorm.io/driver/postgres`)
- Database connection entrypoint: `src/database/postgres.go`
- Redis connection entrypoint: `src/common/redis/redis.go`
- Auto-migration on startup: `database.DB.AutoMigrate(&types.User{})` in `cmd/api/main.go`
- SQL migrations directory: `db/migrations`
- Migration scripts:
  - `scripts/migrate.sh`
  - `scripts/migrate-status.sh`
  - `scripts/migrate-down.sh`

## Configuration Source
- Loaded by `viper` in `src/config/config.go`
- Environment values in `.env` (reference in `.env.example`)

## DB Environment Keys
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`
- `DB_TIMEZONE`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_CONN_MAX_LIFETIME`
- `DB_CONN_MAX_IDLE_TIME`

## Redis Environment Keys
- `REDIS_ADDR` (or `REDIS_HOST` + `REDIS_PORT`)
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_KEY_PREFIX`

## Persistence Split
- PostgreSQL stores `users`.
- Redis stores refresh tokens (`refresh_token:<userID>`).

## Docker Networking
- If the API runs on the host and dependencies run in Docker, use host addresses:
  - `DB_HOST=127.0.0.1`
  - `REDIS_HOST=127.0.0.1`
- If the API runs inside Docker Compose, use container service names:
  - `DB_HOST=postgres`
  - `REDIS_HOST=redis`
