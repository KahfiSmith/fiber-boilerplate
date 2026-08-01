package routes

import (
	controller "fiber-boilerplate/pkg/controllers"
	"fiber-boilerplate/internal/core/server/observability"
	"fiber-boilerplate/internal/domain/health"

	"github.com/gofiber/fiber/v3"
)

type Dependencies struct {
	HealthController *health.HealthController
	AuthController   *controller.AuthController
	Metrics          *observability.Metrics
	EnablePprof      bool
}

func Register(app *fiber.App, deps Dependencies) {
	registerObservabilityRoutes(app, deps.Metrics, deps.EnablePprof)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	health.RegisterRoutes(v1, deps.HealthController)
	registerAuthRoutes(v1, deps.AuthController)
}
