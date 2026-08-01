package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"go.uber.org/zap"
)

const (
	ValidatorLocalKey = "validator"
)

func InjectRequestContext(validate *validator.Validate) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(ValidatorLocalKey, validate)
		return c.Next()
	}
}

func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()

		requestID := requestid.FromContext(c)
		if requestID == "" {
			requestID = c.Get("X-Request-ID")
		}
		log.Info("request",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)

		return err
	}
}
