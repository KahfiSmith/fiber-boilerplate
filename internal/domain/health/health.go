package health

import (
	"time"
	"fiber-boilerplate/internal/pkg/response"
	"github.com/gofiber/fiber/v3"
)

// --- Domain / Entity ---
type HealthStatus struct {
	Status    string
	Message   string
	Service   string
	Timestamp string
}

// --- DTO ---
type HealthStatusResponse struct {
	Status    string `json:"status" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Service   string `json:"service" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
}

// --- Repository Contract ---
type HealthRepository interface {
	Ping() bool
	GetServiceName() string
}

// --- Repository Implementation ---
type healthRepository struct {
	serviceName string
}

func NewHealthRepository(serviceName string) HealthRepository {
	return &healthRepository{
		serviceName: serviceName,
	}
}

func (r *healthRepository) Ping() bool {
	return true
}

func (r *healthRepository) GetServiceName() string {
	return r.serviceName
}

// --- Service Contract ---
type HealthService interface {
	GetStatus() HealthStatus
}

// --- Service Implementation ---
type healthService struct {
	repo HealthRepository
}

func NewHealthService(repo HealthRepository) HealthService {
	return &healthService{repo: repo}
}

func (s *healthService) GetStatus() HealthStatus {
	isDbUp := s.repo.Ping()
	status := "down"
	if isDbUp {
		status = "ok"
	}

	return HealthStatus{
		Status:    status,
		Message:   "service is healthy",
		Service:   s.repo.GetServiceName(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// --- Controller ---
type HealthController struct {
	healthService HealthService
}

func NewHealthController(healthService HealthService) *HealthController {
	return &HealthController{
		healthService: healthService,
	}
}

func (h *HealthController) Health(c fiber.Ctx) error {
	status := h.healthService.GetStatus()
	return response.Success(c, fiber.StatusOK, HealthStatusResponse{
		Status:    status.Status,
		Message:   status.Message,
		Service:   status.Service,
		Timestamp: status.Timestamp,
	})
}

// --- Routes ---
func RegisterRoutes(v1 fiber.Router, healthController *HealthController) {
	v1.Get("/health", healthController.Health)
}
