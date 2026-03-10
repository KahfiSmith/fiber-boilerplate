package middleware

import (
	"strings"
	"time"

	"fiber-boilerplate/pkg/server/observability"

	"github.com/gofiber/fiber/v3"
)

func RequestMetrics(metrics *observability.Metrics) fiber.Handler {
	return func(c fiber.Ctx) error {
		path := strings.TrimSpace(c.Path())
		if metrics == nil || path == "/metrics" || strings.HasPrefix(path, "/debug/pprof") {
			return c.Next()
		}

		start := time.Now()
		method := c.Method()
		metrics.IncInflight()
		defer func() {
			status := c.Response().StatusCode()
			if status == 0 {
				status = fiber.StatusOK
			}

			metrics.Observe(method, path, status, time.Since(start))
			metrics.DecInflight()
		}()

		return c.Next()
	}
}
