package routes

import (
	controller "fiber-boilerplate/pkg/controllers"

	"github.com/gofiber/fiber/v3"
)

type Dependencies struct {
	HealthController *controller.HealthController
	AuthController   *controller.AuthController
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	registerHealthRoutes(v1, deps.HealthController)
	registerAuthRoutes(v1, deps.AuthController)
}
