# Validation & Response Conventions

## Request validation

`src/common/validator/validator.go` provides `ParseAndValidate`:

1. Binds the JSON body into the DTO.
2. Runs `go-playground/validator` struct validation.
3. Returns a friendly `BAD_REQUEST` message for the first failing field
   (e.g. "email must be a valid email address").

DTOs declare validation tags, e.g.:

```go
type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

## Response envelope

`src/common/response/response.go` defines the shared envelope:

```go
type APIResponse struct {
    Success bool        `json:"success"`
    Code    string      `json:"code,omitempty"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   interface{} `json:"error,omitempty"`
}
```

- Success: `response.Success(ctx, status, message, data)`.
- Failure: `response.HandleError(ctx, err)` maps errors to status + code.

## Exceptions

`src/common/exceptions/exceptions.go` provides typed HTTP errors:

| Constructor | Code | HTTP |
|---|---|---|
| `NotFound` | `NOT_FOUND` | 404 |
| `BadRequest` | `BAD_REQUEST` | 400 |
| `Unauthorized` | custom (e.g. `INVALID_CREDENTIALS`) | 401 |
| `Forbidden` | `FORBIDDEN` | 403 |
| `Internal` | `INTERNAL_SERVER_ERROR` | 500 |
| `TooManyRequests` | `TOO_MANY_REQUESTS` | 429 |

`HandleError` also maps `fiber.Error` and `validator.ValidationErrors`
(`VALIDATION_ERROR`) to appropriate responses.
