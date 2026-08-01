package response

import (
	"github.com/gofiber/fiber/v3"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func Success(c fiber.Ctx, statusCode int, data any) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c fiber.Ctx, statusCode int, message string, err any) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}
