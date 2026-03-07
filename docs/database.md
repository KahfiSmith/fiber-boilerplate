# Database

Database setup and conventions.

## Current Driver
- `gorm` with PostgreSQL driver (`gorm.io/driver/postgres`)
- Bootstrap location: `pkg/configs/gorm.go`
- GORM auto-migration entrypoint: `config.AutoMigrate`
- SQL migrations directory: `db/migrations`
- Primary migration entrypoints:
  - `scripts/migrate.sh`
  - `scripts/migrate-status.sh`
  - `scripts/migrate-down.sh`

## Configuration Source
- Loaded by `viper` in `pkg/configs/config.go` + `pkg/configs/db.go`
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

## Connection Behavior
- DSN built by `DBConfig.DSN()`
- If `DATABASE_URL` is present, it is parsed first and mapped to DB config fields.
- Pool configuration:
  - `SetMaxOpenConns`
  - `SetMaxIdleConns`
  - `SetConnMaxLifetime`
  - `SetConnMaxIdleTime`
- Startup health check: `sqlDB.Ping()`
- Startup auto-migrate: registered models are migrated via `config.AutoMigrate`
- Registered auth tables: `auth_sessions`, `otp_challenges`, `auth_rate_limits`

## Persistence Split
- PostgreSQL stores `users`, auth sessions, OTP challenges, and auth rate-limit counters.

## Validation Rules
Startup fails fast if required DB config is invalid:
- empty host/user/db name
- non-positive port or pool values
- non-positive lifetime/idle durations

## Future DB Workflow
When adding new DB features:
1. Define/adjust env key + default in `pkg/configs/db.go`.
2. Add validation rule in `validateDBConfig`.
3. Wire runtime usage in `pkg/configs/gorm.go` or repository layer.
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
