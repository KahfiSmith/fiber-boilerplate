package controllers

import (
	"fiber-boilerplate/pkg/services"
	"fiber-boilerplate/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type HealthController struct {
	healthService services.HealthService
}

func NewHealthController(healthService services.HealthService) *HealthController {
	return &HealthController{
		healthService: healthService,
	}
}

func (h *HealthController) Health(c fiber.Ctx) error {
	return utils.Success(c, fiber.StatusOK, h.healthService.GetStatus())
}
