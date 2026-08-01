package controllers

import (
	
	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/services"
	dtoResponse "fiber-boilerplate/pkg/dto/response"
	"fiber-boilerplate/internal/pkg/response"

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
// @Success 200 {object} response.APIResponse{data=dtoResponse.HealthStatusResponse}
// @Router /health [get]
func (h *HealthController) Health(c fiber.Ctx) error {
	return response.Success(c, fiber.StatusOK, healthStatusResponse(h.healthService.GetStatus()))
}

func healthStatusResponse(status entities.HealthStatus) dtoResponse.HealthStatusResponse {
	return dtoResponse.HealthStatusResponse{
		Status:    status.Status,
		Message:   status.Message,
		Service:   status.Service,
		Timestamp: status.Timestamp,
	}
}
