package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	DBLocalKey        = "db"
	ValidatorLocalKey = "validator"
)

func InjectRequestContext(db *gorm.DB, validate *validator.Validate) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(DBLocalKey, db)
		c.Locals(ValidatorLocalKey, validate)
		return c.Next()
	}
}

func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()

		requestID := c.Get("X-Request-ID")
		log.Info("request",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)

		return err
	}
}
