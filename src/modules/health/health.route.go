package health

import (
	"fiber-boilerplate/src/modules/health/controller"
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(router fiber.Router, c *controller.HealthController) {
	healthGroup := router.Group("/health")
	healthGroup.Get("", c.Check)
}
