package server

import (
	config "fiber-boilerplate/pkg/configs"
	controller "fiber-boilerplate/pkg/controllers"
	repository "fiber-boilerplate/pkg/repositories"
	"fiber-boilerplate/pkg/services"

	"github.com/gofiber/fiber/v3"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func New(cfg config.Config, log *zap.Logger, db *gorm.DB, validate *validator.Validate) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.Fiber.ReadTimeout,
		WriteTimeout: cfg.Fiber.WriteTimeout,
		BodyLimit:    cfg.Fiber.BodyLimitMB * 1024 * 1024,
	})

	app.Use(requestid.New())
	app.Use(recovermw.New())
	app.Use(func(c fiber.Ctx) error {
		c.Locals("db", db)
		c.Locals("validator", validate)
		return c.Next()
	})
	app.Use(func(c fiber.Ctx) error {
		requestID := requestid.FromContext(c)
		log.Info("request",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)
		return c.Next()
	})

	healthRepo := repository.NewHealthRepository(cfg.App.Name)
	healthService := services.NewHealthService(healthRepo)
	healthController := controller.NewHealthController(healthService)

	RegisterRoutes(app, healthController)

	return app
}
