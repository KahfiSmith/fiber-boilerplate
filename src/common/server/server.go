package server

import (
	"fiber-boilerplate/src/modules/health"
	"fiber-boilerplate/src/modules/auth"
	"github.com/gofiber/fiber/v3"
)

type Dependencies struct {
	HealthController *health.HealthController
	AuthController   *auth.AuthController
}

func RegisterRoutes(app *fiber.App, deps Dependencies) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	health.RegisterRoutes(v1, deps.HealthController)
	auth.RegisterRoutes(v1, deps.AuthController)
}
