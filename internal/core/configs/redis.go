package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Addr      string
	Username  string
	Password  string
	DB        int
	KeyPrefix string
}

func loadRedisConfig(v *viper.Viper) (RedisConfig, error) {
	return RedisConfig{
		Addr:      v.GetString("REDIS_ADDR"),
		Username:  v.GetString("REDIS_USERNAME"),
		Password:  v.GetString("REDIS_PASSWORD"),
		DB:        v.GetInt("REDIS_DB"),
		KeyPrefix: v.GetString("REDIS_KEY_PREFIX"),
	}, nil
}

func setRedisDefaults(v *viper.Viper) {
	v.SetDefault("REDIS_ADDR", "127.0.0.1:6379")
	v.SetDefault("REDIS_USERNAME", "")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_KEY_PREFIX", "fiber-boilerplate")
}

func validateRedisConfig(c RedisConfig) error {
	if err := requireNonEmpty("REDIS_ADDR", c.Addr); err != nil {
		return err
	}
	if c.DB < 0 {
		return fmt.Errorf("REDIS_DB must be 0 or greater")
	}
	if err := requireNonEmpty("REDIS_KEY_PREFIX", c.KeyPrefix); err != nil {
		return err
	}

	return nil
}

func NewRedisClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     strings.TrimSpace(cfg.Redis.Addr),
		Username: strings.TrimSpace(cfg.Redis.Username),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func CloseRedisClient(client *redis.Client) error {
	if client == nil {
		return nil
	}

	return client.Close()
}
