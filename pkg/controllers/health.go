package controllers

import (
	"fiber-boilerplate/pkg/dto/response"
	"fiber-boilerplate/pkg/entities"
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

// Health godoc
// @Summary Health check
// @Description Returns service health information.
// @Tags Health
// @Produce json
// @Success 200 {object} response.APIResponse{data=response.HealthStatusResponse}
// @Router /health [get]
func (h *HealthController) Health(c fiber.Ctx) error {
	return utils.Success(c, fiber.StatusOK, healthStatusResponse(h.healthService.GetStatus()))
}

func healthStatusResponse(status entities.HealthStatus) response.HealthStatusResponse {
	return response.HealthStatusResponse{
		Status:    status.Status,
		Message:   status.Message,
		Service:   status.Service,
		Timestamp: status.Timestamp,
	}
}
