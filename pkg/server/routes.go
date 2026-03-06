package server

import (
	serverRoutes "fiber-boilerplate/pkg/server/routes"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App, deps Dependencies) {
	serverRoutes.Register(app, serverRoutes.Dependencies{
		HealthController: deps.HealthController,
	})
}
