package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	Fiber FiberConfig
	Log   LogConfig
	DB    DBConfig
	Auth  AuthConfig
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

func Load() (Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) && !errors.Is(err, os.ErrNotExist) {
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
	}

	dbCfg, err := loadDBConfig(v)
	if err != nil {
		return Config{}, err
	}
	cfg.DB = dbCfg

	authCfg, err := loadAuthConfig(v)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth = authCfg

	if err := applyLegacyAppEnvOverrides(v, &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
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
	setDBDefaults(v)
	setAuthDefaults(v)
}

func (c AppConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c Config) Validate() error {
	if err := requireNonEmpty("APP_NAME", c.App.Name); err != nil {
		return err
	}
	if err := requireNonEmpty("APP_HOST", c.App.Host); err != nil {
		return err
	}
	if err := requireNonEmpty("APP_PORT", c.App.Port); err != nil {
		return err
	}
	if err := requirePositiveDuration("APP_SHUTDOWN_TIMEOUT", c.App.ShutdownTimeout); err != nil {
		return err
	}

	if err := requirePositiveDuration("APP_READ_TIMEOUT", c.Fiber.ReadTimeout); err != nil {
		return err
	}
	if err := requirePositiveDuration("APP_WRITE_TIMEOUT", c.Fiber.WriteTimeout); err != nil {
		return err
	}
	if err := requirePositiveInt("APP_BODY_LIMIT_MB", c.Fiber.BodyLimitMB); err != nil {
		return err
	}

	encoding := strings.ToLower(strings.TrimSpace(c.Log.Encoding))
	if encoding != "json" && encoding != "console" {
		return fmt.Errorf("LOG_ENCODING must be one of: json, console")
	}
	if err := validateDBConfig(c.DB); err != nil {
		return err
	}

	return validateAuthConfig(c.Auth)
}

func applyLegacyAppEnvOverrides(v *viper.Viper, cfg *Config) error {
	httpAddr := strings.TrimSpace(v.GetString("HTTP_ADDR"))
	if httpAddr != "" {
		host, port, err := parseHTTPAddr(httpAddr)
		if err != nil {
			return fmt.Errorf("invalid HTTP_ADDR: %w", err)
		}
		if host != "" {
			cfg.App.Host = host
		}
		cfg.App.Port = port
	}

	gracefulShutdownMS := v.GetInt("GRACEFUL_SHUTDOWN_MS")
	if gracefulShutdownMS > 0 {
		cfg.App.ShutdownTimeout = time.Duration(gracefulShutdownMS) * time.Millisecond
	}

	return nil
}

func parseHTTPAddr(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", errors.New("must not be empty")
	}

	if strings.HasPrefix(value, ":") {
		port := strings.TrimPrefix(value, ":")
		if err := validatePort(port); err != nil {
			return "", "", err
		}
		return "", port, nil
	}

	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return "", "", fmt.Errorf("must be in host:port or :port format")
	}
	if strings.TrimSpace(host) == "" {
		return "", "", errors.New("host must not be empty")
	}
	if err := validatePort(port); err != nil {
		return "", "", err
	}

	return strings.TrimSpace(host), port, nil
}

func validatePort(value string) error {
	portInt, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return errors.New("port must be numeric")
	}
	if portInt < 1 || portInt > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}
