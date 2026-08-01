package server

import (
	"fiber-boilerplate/internal/core/server/observability"
	serverRoutes "fiber-boilerplate/internal/core/server/routes"

	"github.com/gofiber/fiber/v3"
)

type ObservabilityDependencies struct {
	Metrics     *observability.Metrics
	EnablePprof bool
}

func RegisterRoutes(app *fiber.App, deps Dependencies, observabilityDeps ObservabilityDependencies) {
	serverRoutes.Register(app, serverRoutes.Dependencies{
		HealthController: deps.HealthController,
		AuthController:   deps.AuthController,
		Metrics:          observabilityDeps.Metrics,
		EnablePprof:      observabilityDeps.EnablePprof,
	})
}
