package controllers

import (
	"net/http/httptest"
	"testing"

	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/services"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type healthServiceStub struct {
	status entities.HealthStatus
}

func (h healthServiceStub) GetStatus() entities.HealthStatus {
	return h.status
}

var _ services.HealthService = healthServiceStub{}

func TestHealthControllerHealth(t *testing.T) {
	t.Parallel()

	controller := NewHealthController(healthServiceStub{
		status: entities.HealthStatus{
			Status:    "ok",
			Message:   "service is healthy",
			Service:   "fiber-boilerplate",
			Timestamp: "2026-03-08T10:00:00Z",
		},
	})

	app := fiber.New()
	app.Get("/health", controller.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body := readBody(t, resp)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, body, `"success":true`)
	assert.Contains(t, body, `"service":"fiber-boilerplate"`)
	assert.Contains(t, body, `"timestamp":"2026-03-08T10:00:00Z"`)
}
