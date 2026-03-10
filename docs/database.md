# Database

Database setup and conventions.

## Current Driver
- `gorm` with PostgreSQL driver (`gorm.io/driver/postgres`)
- Bootstrap location: `pkg/configs/gorm.go`
- Redis client bootstrap location: `pkg/configs/redis.go`
- GORM auto-migration entrypoint: `config.AutoMigrate`
- SQL migrations directory: `db/migrations`
- Primary migration entrypoints:
  - `scripts/migrate.sh`
  - `scripts/migrate-status.sh`
  - `scripts/migrate-down.sh`

## Configuration Source
- Loaded by `viper` in `pkg/configs/config.go`, `pkg/configs/db.go`, and `pkg/configs/redis.go`
- Environment defaults are in `.env.example`
- Runtime fallback defaults in `setDBDefaults` should stay aligned with `.env.example` and should not contain personal or environment-specific secrets.

## DB Environment Keys
- `DATABASE_URL` (optional; when set, it overrides host/user/password/name/sslmode/timezone from `DB_*`)
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
- `REDIS_ADDR`
- `REDIS_USERNAME`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_KEY_PREFIX`

## Connection Behavior
- DSN built by `DBConfig.DSN()`
- If `DATABASE_URL` is present, it is parsed first and mapped to DB config fields.
- Pool configuration:
  - `SetMaxOpenConns`
  - `SetMaxIdleConns`
  - `SetConnMaxLifetime`
  - `SetConnMaxIdleTime`
- Startup health check: `sqlDB.Ping()`
- Redis startup health check: `client.Ping()`
- Startup auto-migrate: registered models are migrated via `config.AutoMigrate`
- Registered PostgreSQL auth tables: `otp_challenges`
- Email lookups remain case-insensitive through `LOWER(email)` queries and a matching SQL index in `db/migrations/000003_add_users_email_lower_index.up.sql`

## Runtime Requirement
- PostgreSQL is required at startup.
- Redis is required at startup.
- The application exits fast if either dependency is unavailable.

## Persistence Split
- PostgreSQL stores `users` and OTP challenges.
- Redis stores refresh sessions and auth rate-limit counters.
- Protected access-token requests also rely on live session presence in the session store.

## Docker Networking
- If the API runs on the host and dependencies run in Docker, use host addresses:
  - `DB_HOST=127.0.0.1`
  - `REDIS_ADDR=127.0.0.1:6379`
- If the API runs inside Docker Compose, use container service names:
  - `DATABASE_URL=postgres://postgres:postgres@postgres:5432/fiber_boilerplate?sslmode=disable&TimeZone=UTC`
  - `REDIS_ADDR=redis:6379`
- Reference container-friendly values live in `.env.docker.example`.

## Validation Rules
Startup fails fast if required DB config is invalid:
- empty host/user/db name
- non-positive port or pool values
- non-positive lifetime/idle durations

## Future DB Workflow
When adding new DB features:
1. Define/adjust env key + default in `pkg/configs/db.go`.
2. Add validation rule in `validateDBConfig` or `validateRedisConfig`.
3. Wire runtime usage in `pkg/configs/gorm.go`, `pkg/configs/redis.go`, or repository layer.
4. Update `.env.example` and docs.

## Migration Workflow
- Create SQL files under `db/migrations` using `*.up.sql` and `*.down.sql`.
- Check migration status with `./scripts/migrate-status.sh`.
- `status` only lists local migration files; it does not track applied database history.
- Apply migrations with `./scripts/migrate.sh`.
- Apply one specific migration with `./scripts/migrate.sh <version>`.
- Roll back the latest migration with `./scripts/migrate-down.sh`.
- Roll back one specific migration with `./scripts/migrate-down.sh <version>`.
- `up` and `down` runs also regenerate Swagger docs unless `GENERATE_SWAGGER_ON_MIGRATE=false`.
- Migration scripts push SQL files directly to the configured database; they do not maintain a `schema_migrations` table.
- Keep GORM auto-migrate for registered application models during startup.
- Existing SQL migrations may still create legacy `auth_sessions` and `auth_rate_limits` tables, but refresh sessions and rate limiting now run from Redis.
- The bundled `docker-compose.yml` is for local runtime dependencies and does not replace the SQL migration scripts.
