package routes

import (
	controller "fiber-boilerplate/pkg/controllers"

	"github.com/gofiber/fiber/v3"
)

func registerHealthRoutes(v1 fiber.Router, healthController *controller.HealthController) {
	v1.Get("/health", healthController.Health)
}
