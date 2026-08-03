package exceptions

import "github.com/gofiber/fiber/v3"

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

func TooManyRequests(msg string) *HttpError {
	if msg == "" {
		msg = "Too many requests"
	}
	return New(fiber.StatusTooManyRequests, msg)
}
