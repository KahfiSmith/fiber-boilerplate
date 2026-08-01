package routes

import (
	"fiber-boilerplate/internal/core/server/observability"

	"github.com/gofiber/fiber/v3"
	pprofmw "github.com/gofiber/fiber/v3/middleware/pprof"
)

func registerObservabilityRoutes(app *fiber.App, metrics *observability.Metrics, enablePprof bool) {
	if metrics != nil {
		app.Get("/metrics", metrics.Handle)
	}

	if enablePprof {
		app.Use(pprofmw.New())
	}
}
