package exceptions

import "github.com/gofiber/fiber/v3"

// HttpError represents a standard HTTP error structure
type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func New(code int, message string) *HttpError {
	return &HttpError{
		Code:    code,
		Message: message,
	}
}

// Common Exceptions
func NotFound(msg string) *HttpError {
	if msg == "" {
		msg = "Resource not found"
	}
	return New(fiber.StatusNotFound, msg)
}

func BadRequest(msg string) *HttpError {
	if msg == "" {
		msg = "Bad request"
	}
	return New(fiber.StatusBadRequest, msg)
}

func Unauthorized(msg string) *HttpError {
	if msg == "" {
		msg = "Unauthorized access"
	}
	return New(fiber.StatusUnauthorized, msg)
}

func Internal(msg string) *HttpError {
	if msg == "" {
		msg = "Internal server error"
	}
	return New(fiber.StatusInternalServerError, msg)
}
