package server

import (
	controller "fiber-boilerplate/pkg/controllers"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, healthController *controller.HealthController) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", healthController.Health)
}
