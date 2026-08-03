package server

import (
	"time"
	"fiber-boilerplate/src/modules/health"
	healthControllerPkg "fiber-boilerplate/src/modules/health/controller"
	"fiber-boilerplate/src/modules/auth"
	"fiber-boilerplate/src/common/middleware"
	"fiber-boilerplate/src/config"
	"github.com/gofiber/fiber/v3"
)

type Dependencies struct {
	HealthController *healthControllerPkg.HealthController
	AuthController   *auth.AuthController
	Config           config.Config
}

func RegisterRoutes(app *fiber.App, deps Dependencies) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	protected := middleware.Protected(deps.Config.Auth)
	rateLimiter := middleware.RateLimiter(deps.Config.Auth.RateLimitPerMin, time.Minute)

	health.RegisterRoutes(v1, deps.HealthController)
	auth.RegisterRoutes(v1, deps.AuthController, protected, rateLimiter)
}
