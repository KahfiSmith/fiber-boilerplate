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
│   │   └── config.go
│   ├── controllers
│   │   └── health.go
│   ├── database
│   │   └── postgres.go
│   ├── logger
│   │   └── logger.go
│   ├── models
│   │   └── health.go
│   ├── repositories
│   │   └── health_repository.go
│   ├── routes
│   │   └── routes.go
│   └── server
│       ├── app.go
│       └── run.go
│   ├── services
│   │   └── health_service.go
│   └── utils
│       └── response.go
│   └── validation
│       └── validator.go
├── .env.example
├── go.mod
└── go.sum
```

## Run

```bash
go run ./cmd/api
```

## Env

Copy `.env.example` into `.env` and adjust DB values.

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
