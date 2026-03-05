package services

import (
	"fiber-boilerplate/pkg/models"
	repository "fiber-boilerplate/pkg/repositories"
)

type HealthService interface {
	GetStatus() models.HealthStatus
}

type healthService struct {
	healthRepository repository.HealthRepository
}

func NewHealthService(healthRepository repository.HealthRepository) HealthService {
	return &healthService{
		healthRepository: healthRepository,
	}
}

func (h *healthService) GetStatus() models.HealthStatus {
	return models.HealthStatus{
		Status:    "ok",
		Message:   "service is healthy",
		Service:   h.healthRepository.ServiceName(),
		Timestamp: h.healthRepository.NowUTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
