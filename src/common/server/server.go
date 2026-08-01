package server

import (
	"fiber-boilerplate/src/modules/health"
	"github.com/gofiber/fiber/v3"
)

type Dependencies struct {
	HealthController *health.HealthController
}

func RegisterRoutes(app *fiber.App, deps Dependencies) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	health.RegisterRoutes(v1, deps.HealthController)
}
