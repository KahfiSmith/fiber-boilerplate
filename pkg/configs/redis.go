package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func loadRedisConfig(v *viper.Viper) (RedisConfig, error) {
	return RedisConfig{
		Addr:     v.GetString("REDIS_ADDR"),
		Password: v.GetString("REDIS_PASSWORD"),
		DB:       v.GetInt("REDIS_DB"),
	}, nil
}

func setRedisDefaults(v *viper.Viper) {
	v.SetDefault("REDIS_ADDR", "127.0.0.1:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
}

func validateRedisConfig(c RedisConfig) error {
	if err := requireNonEmpty("REDIS_ADDR", c.Addr); err != nil {
		return err
	}
	if strings.ContainsAny(c.Addr, " \t\r\n") {
		return fmt.Errorf("REDIS_ADDR must not contain whitespace")
	}
	if c.DB < 0 {
		return fmt.Errorf("REDIS_DB must be >= 0")
	}

	return nil
}
