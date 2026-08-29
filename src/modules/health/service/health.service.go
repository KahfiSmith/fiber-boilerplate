package service

import (
	"context"
	"fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/database"
)

type HealthService struct {
	appName string
}

func NewHealthService(appName string) *HealthService {
	return &HealthService{appName: appName}
}

func (s *HealthService) Check() map[string]interface{} {
	dbStatus := "down"
	if database.DB != nil {
		sqlDB, err := database.DB.DB()
		if err == nil && sqlDB.Ping() == nil {
			dbStatus = "up"
		}
	}

	redisStatus := "down"
	if redis.Client != nil {
		if err := redis.Client.Ping(context.Background()).Err(); err == nil {
			redisStatus = "up"
		}
	}

	overall := "ok"
	if dbStatus != "up" || redisStatus != "up" {
		overall = "degraded"
	}

	return map[string]interface{}{
		"app":      s.appName,
		"status":   overall,
		"database": dbStatus,
		"redis":    redisStatus,
	}
}

func (s *HealthService) IsReady() bool {
	res := s.Check()
	return res["database"] == "up" && res["redis"] == "up"
}
