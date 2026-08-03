package service

import (
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/common/redis"
	"context"
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

	return map[string]interface{}{
		"app":      s.appName,
		"status":   "ok",
		"database": dbStatus,
		"redis":    redisStatus,
	}
}
