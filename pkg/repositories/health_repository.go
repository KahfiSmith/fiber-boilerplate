package repository

import "time"

type HealthRepository interface {
	ServiceName() string
	NowUTC() time.Time
}

type healthRepository struct {
	appName string
}

func NewHealthRepository(appName string) HealthRepository {
	return &healthRepository{
		appName: appName,
	}
}

func (h *healthRepository) ServiceName() string {
	return h.appName
}

func (h *healthRepository) NowUTC() time.Time {
	return time.Now().UTC()
}
