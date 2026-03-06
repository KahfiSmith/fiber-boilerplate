package server

import (
	"fmt"

	config "fiber-boilerplate/pkg/configs"

	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func New(cfg config.Config, log *zap.Logger, db *gorm.DB, validate *validator.Validate, deps Dependencies) (*fiber.App, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid server dependencies: %w", err)
	}

	app := config.NewFiberApp(cfg)
	config.ApplyFiberMiddlewares(app, log, db, validate)

	RegisterRoutes(app, deps)

	return app, nil
}
