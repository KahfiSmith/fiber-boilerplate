package controller

import (
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/modules/health/service"

	"github.com/gofiber/fiber/v3"
)

type HealthController struct {
	service *service.HealthService
}

func NewHealthController(service *service.HealthService) *HealthController {
	return &HealthController{service: service}
}

func (c *HealthController) Check(ctx fiber.Ctx) error {
	status := c.service.Check()
	return response.Success(ctx, fiber.StatusOK, "Health check successful", status)
}

func (c *HealthController) Ready(ctx fiber.Ctx) error {
	if c.service.IsReady() {
		return response.Success(ctx, fiber.StatusOK, "Service is ready", c.service.Check())
	}
	return response.HandleError(ctx, fiber.NewError(fiber.StatusServiceUnavailable, "Service dependencies unavailable"))
}

func (c *HealthController) Live(ctx fiber.Ctx) error {
	return response.Success(ctx, fiber.StatusOK, "Service is alive", map[string]string{"status": "alive"})
}
