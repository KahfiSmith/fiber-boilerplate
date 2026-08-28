package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		log.Info("http_request",
			slog.String("method", method),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("path", path),
		)
		return err
	}
}
