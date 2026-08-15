# Feature: Health Check

## Overview

A readiness endpoint that reports whether the API, PostgreSQL, and Redis are
reachable. Used for docker-compose healthchecks and manual verification.

## Core flow

```text
GET /api/v1/health
  -> health.route.go
     -> controller/health.controller.go
        -> service/health.service.go (pings DB + Redis)
           -> response.Success
```

## Implementation map

| Concern | Files |
|---|---|
| Route | `src/modules/health/health.route.go` |
| Controller | `src/modules/health/controller/health.controller.go` |
| Service | `src/modules/health/service/health.service.go` |

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/health` | App + DB + Redis status (`{"app", "status", "database", "redis"}`) |

## Not yet implemented

- Deep dependency checks (e.g. running migrations status, latency).
