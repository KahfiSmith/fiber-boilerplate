package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	Fiber FiberConfig
	Log   LogConfig
	DB    DBConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Host            string
	Port            string
	ShutdownTimeout time.Duration
}

type FiberConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	BodyLimitMB  int
	Prefork      bool
}

type LogConfig struct {
	Level    string
	Encoding string
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	TimeZone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
	}

	cfg := Config{
		App: AppConfig{
			Name:            v.GetString("APP_NAME"),
			Env:             v.GetString("APP_ENV"),
			Host:            v.GetString("APP_HOST"),
			Port:            v.GetString("APP_PORT"),
			ShutdownTimeout: v.GetDuration("APP_SHUTDOWN_TIMEOUT"),
		},
		Fiber: FiberConfig{
			ReadTimeout:  v.GetDuration("APP_READ_TIMEOUT"),
			WriteTimeout: v.GetDuration("APP_WRITE_TIMEOUT"),
			BodyLimitMB:  v.GetInt("APP_BODY_LIMIT_MB"),
			Prefork:      v.GetBool("APP_PREFORK"),
		},
		Log: LogConfig{
			Level:    v.GetString("LOG_LEVEL"),
			Encoding: v.GetString("LOG_ENCODING"),
		},
		DB: DBConfig{
			Host:            v.GetString("DB_HOST"),
			Port:            v.GetInt("DB_PORT"),
			User:            v.GetString("DB_USER"),
			Password:        v.GetString("DB_PASSWORD"),
			Name:            v.GetString("DB_NAME"),
			SSLMode:         v.GetString("DB_SSLMODE"),
			TimeZone:        v.GetString("DB_TIMEZONE"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "fiber-boilerplate")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_HOST", "0.0.0.0")
	v.SetDefault("APP_PORT", "3000")
	v.SetDefault("APP_READ_TIMEOUT", "10s")
	v.SetDefault("APP_WRITE_TIMEOUT", "10s")
	v.SetDefault("APP_SHUTDOWN_TIMEOUT", "10s")
	v.SetDefault("APP_BODY_LIMIT_MB", 4)
	v.SetDefault("APP_PREFORK", false)

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_ENCODING", "console")

	v.SetDefault("DB_HOST", "127.0.0.1")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "fiber_boilerplate")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_TIMEZONE", "UTC")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 25)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
}
