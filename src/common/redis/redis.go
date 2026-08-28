package redis

import (
	"context"
	"fiber-boilerplate/src/config"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Connect(cfg config.RedisConfig) {
	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx := context.Background()
	if err := Client.Ping(ctx).Err(); err != nil {
		slog.Error("redis_connect_failed", slog.Any("error", err))
		return
	}

	slog.Info("redis_connected", slog.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
}
