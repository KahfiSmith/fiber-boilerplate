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
		reqID, _ := c.Locals("requestid").(string)
		if reqID == "" {
			reqID = c.Get("X-Request-ID")
		}

		attrs := []slog.Attr{
			slog.String("request_id", reqID),
			slog.String("method", method),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("path", path),
		}

		if userID, ok := c.Locals("user_id").(uint); ok && userID != 0 {
			attrs = append(attrs, slog.Uint64("user_id", uint64(userID)))
		}
		if sessionID, ok := c.Locals("session_id").(string); ok && sessionID != "" {
			attrs = append(attrs, slog.String("session_id", sessionID))
		}

		log.LogAttrs(c.Context(), slog.LevelInfo, "http_request", attrs...)
		return err
	}
}
