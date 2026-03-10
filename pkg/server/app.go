package server

import (
	"errors"
	"fmt"

	controller "fiber-boilerplate/pkg/controllers"
	config "fiber-boilerplate/pkg/configs"
	serverMiddleware "fiber-boilerplate/pkg/server/middleware"
	"fiber-boilerplate/pkg/server/observability"

	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Dependencies struct {
	HealthController *controller.HealthController
	AuthController   *controller.AuthController
}

func (d Dependencies) Validate() error {
	if d.HealthController == nil {
		return errors.New("server dependency HealthController is required")
	}
	if d.AuthController == nil {
		return errors.New("server dependency AuthController is required")
	}

	return nil
}

func New(cfg config.Config, log *zap.Logger, validate *validator.Validate, deps Dependencies) (*fiber.App, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid server dependencies: %w", err)
	}

	app := config.NewFiberApp(cfg)
	config.ApplyFiberMiddlewares(app, log, validate)

	var metricsCollector *observability.Metrics
	if cfg.Fiber.EnableMetrics {
		metricsCollector = observability.NewMetrics(cfg.App.Name)
		app.Use(serverMiddleware.RequestMetrics(metricsCollector))
	}

	RegisterRoutes(app, deps, ObservabilityDependencies{
		Metrics:     metricsCollector,
		EnablePprof: cfg.Fiber.EnablePprof,
	})

	return app, nil
}
