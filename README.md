# Fiber Boilerplate

Simple Fiber (Go) starter structure with clear layering and bootstrap modules:
- viper config loader
- zap logger
- fiber app/server
- gorm postgres connector
- validator initializer

## Structure

```text
.
├── cmd
│   └── api
│       └── main.go
├── pkg
│   ├── configs
│   │   ├── config.go
│   │   ├── db.go
│   │   ├── fiber.go
│   │   ├── gorm.go
│   │   ├── validator.go
│   │   └── zap.go
│   ├── controllers
│   │   └── health.go
│   ├── dto
│   │   ├── request
│   │   └── response
│   ├── entities
│   ├── models
│   │   └── health.go
│   ├── repositories
│   │   └── health_repository.go
│   ├── server
│   │   ├── app.go
│   │   ├── dependencies.go
│   │   ├── middleware
│   │   │   └── request.go
│   │   ├── routes.go
│   │   ├── routes
│   │   │   ├── health.go
│   │   │   └── register.go
│   │   └── run.go
│   ├── services
│   │   └── health_service.go
│   └── utils
│       └── response.go
├── .env.example
├── go.mod
└── go.sum
```

## Run

```bash
go run ./cmd/api
```

## Redis (Docker Compose)

Start Redis:

```bash
docker compose up -d redis
```

Check logs:

```bash
docker compose logs -f redis
```

Stop and remove:

```bash
docker compose down
```

## Env

Copy `.env.example` into `.env` and adjust DB values.
You can use `DATABASE_URL` (PostgreSQL URL format) or individual `DB_*` keys.
Important DB pool keys: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`.
Redis keys: `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`.

## Routing and Controllers

- Route registration entrypoint: `pkg/server/routes.go`
- Route groups/modules: `pkg/server/routes/*`
- HTTP handlers (controllers): `pkg/controllers/*`
- Server/transport middleware: `pkg/server/middleware/*`

## Data Contracts

- Domain/business objects: `pkg/entities/*`
- HTTP request DTOs: `pkg/dto/request/*`
- HTTP response DTOs: `pkg/dto/response/*`
- Legacy/shared models currently used by health endpoint: `pkg/models/*`

## Documentation

- Prompt patterns: `docs/patterns.md`
- Agent rules: `docs/rules.md`
- Workflow: `docs/workflow.md`
- Architecture: `docs/architecture.md`
- API contract: `docs/api.md`
- Database conventions: `docs/database.md`

## Health Check

```bash
curl http://localhost:3000/api/v1/health
```

Response example:
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
