package utils

import (
	"fiber-boilerplate/pkg/models"

	"github.com/gofiber/fiber/v3"
)

func Success(c fiber.Ctx, statusCode int, data any) error {
	return c.Status(statusCode).JSON(models.APIResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c fiber.Ctx, statusCode int, message string, err any) error {
	return c.Status(statusCode).JSON(models.APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}
