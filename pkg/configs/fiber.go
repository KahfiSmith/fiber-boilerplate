package config

import (
	serverMiddleware "fiber-boilerplate/pkg/server/middleware"

	"github.com/gofiber/fiber/v3"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewFiberApp(cfg Config) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.Fiber.ReadTimeout,
		WriteTimeout: cfg.Fiber.WriteTimeout,
		BodyLimit:    cfg.Fiber.BodyLimitMB * 1024 * 1024,
	})
}

func NewFiberListenConfig(cfg Config) fiber.ListenConfig {
	return fiber.ListenConfig{
		EnablePrefork: cfg.Fiber.Prefork,
	}
}

func ApplyFiberMiddlewares(app *fiber.App, log *zap.Logger, db *gorm.DB, validate *validator.Validate) {
	app.Use(requestid.New())
	app.Use(recovermw.New())
	app.Use(serverMiddleware.InjectRequestContext(db, validate))
	app.Use(serverMiddleware.RequestLogger(log))
}
