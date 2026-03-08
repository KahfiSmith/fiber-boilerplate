package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type healthRepositoryStub struct {
	serviceName string
	now         time.Time
}

func (h healthRepositoryStub) ServiceName() string {
	return h.serviceName
}

func (h healthRepositoryStub) NowUTC() time.Time {
	return h.now
}

func TestHealthServiceGetStatus(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2026, time.March, 7, 13, 45, 50, 0, time.UTC)
	service := NewHealthService(healthRepositoryStub{
		serviceName: "fiber-boilerplate",
		now:         fixedTime,
	})

	status := service.GetStatus()

	assert.Equal(t, "ok", status.Status)
	assert.Equal(t, "service is healthy", status.Message)
	assert.Equal(t, "fiber-boilerplate", status.Service)

	parsedTime, err := time.Parse(time.RFC3339, status.Timestamp)
	require.NoError(t, err)
	assert.True(t, parsedTime.Equal(fixedTime))
}
