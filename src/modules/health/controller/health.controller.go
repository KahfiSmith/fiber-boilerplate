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
