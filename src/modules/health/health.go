package health

import (
	"time"
	"fiber-boilerplate/src/common/response"
	"github.com/gofiber/fiber/v3"
)

// --- Types ---
type HealthStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// --- Controller ---
type HealthController struct {
	service *HealthService
}

func NewHealthController(service *HealthService) *HealthController {
	return &HealthController{service: service}
}

func (c *HealthController) Check(ctx fiber.Ctx) error {
	status := c.service.GetStatus()
	return response.Success(ctx, fiber.StatusOK, status)
}

// --- Service ---
type HealthService struct {
	serviceName string
}

func NewHealthService(serviceName string) *HealthService {
	return &HealthService{serviceName: serviceName}
}

func (s *HealthService) GetStatus() HealthStatus {
	// Simplify ping check to avoid circular dependencies with database layer initially
	return HealthStatus{
		Status:    "ok",
		Message:   "service is healthy",
		Service:   s.serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// --- Module Routes ---
func RegisterRoutes(router fiber.Router, controller *HealthController) {
	router.Get("/health", controller.Check)
}
