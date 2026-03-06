# API

Current API contract.

## Base URL
- Local: `http://localhost:3000`
- Base prefix: `/api/v1`

## Endpoint: Health Check
- Method: `GET`
- Path: `/api/v1/health`
- Handler: `pkg/controllers/health.go`

## Success Response
Status code: `200`

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

## Response Envelope
Defined in `pkg/models/health.go`:
- `success` (bool)
- `message` (string, optional)
- `data` (any, optional)
- `error` (any, optional)

## DTO Convention
- Request DTOs should be placed in `pkg/dto/request`.
- Response DTOs should be placed in `pkg/dto/response`.
- Existing health response currently uses `pkg/models` as a transitional model.

## Notes
- Route registration entrypoint: `pkg/server/routes.go` (can delegate to `pkg/server/routes/*` modules).
- Response helper functions: `pkg/utils/response.go`.
