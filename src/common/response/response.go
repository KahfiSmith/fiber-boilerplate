package response

import (
	"errors"
	"fiber-boilerplate/src/common/exceptions"
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

// HandleError gracefully handles standard errors and custom HttpErrors
func HandleError(c fiber.Ctx, err error) error {
	var httpErr *exceptions.HttpError
	
	if errors.As(err, &httpErr) {
		return c.Status(httpErr.Code).JSON(APIResponse{
			Success: false,
			Message: httpErr.Message,
		})
	}

	// Fallback for unhandled/native errors
	return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
		Success: false,
		Message: "Internal server error",
		Error:   err.Error(), // Note: In production, hide actual error strings
	})
}
