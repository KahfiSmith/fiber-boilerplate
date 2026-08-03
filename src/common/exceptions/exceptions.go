package exceptions

import "github.com/gofiber/fiber/v3"

type HttpError struct {
	Code    int    `json:"-"`
	CodeStr string `json:"code,omitempty"`
	Message string `json:"message"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func New(code int, codeStr string, message string) *HttpError {
	return &HttpError{
		Code:    code,
		CodeStr: codeStr,
		Message: message,
	}
}

func NotFound(msg string) *HttpError {
	if msg == "" {
		msg = "Resource not found"
	}
	return New(fiber.StatusNotFound, "NOT_FOUND", msg)
}

func BadRequest(msg string) *HttpError {
	if msg == "" {
		msg = "Bad request"
	}
	return New(fiber.StatusBadRequest, "BAD_REQUEST", msg)
}

func Unauthorized(codeStr string, msg string) *HttpError {
	if msg == "" {
		msg = "Unauthorized access"
	}
	if codeStr == "" {
		codeStr = "UNAUTHORIZED"
	}
	return New(fiber.StatusUnauthorized, codeStr, msg)
}

func Forbidden(codeStr string, msg string) *HttpError {
	if msg == "" {
		msg = "Forbidden"
	}
	if codeStr == "" {
		codeStr = "FORBIDDEN"
	}
	return New(fiber.StatusForbidden, codeStr, msg)
}

func Internal(msg string) *HttpError {
	if msg == "" {
		msg = "Internal server error"
	}
	return New(fiber.StatusInternalServerError, "INTERNAL_SERVER_ERROR", msg)
}

func TooManyRequests(msg string) *HttpError {
	if msg == "" {
		msg = "Too many requests"
	}
	return New(fiber.StatusTooManyRequests, "TOO_MANY_REQUESTS", msg)
}
