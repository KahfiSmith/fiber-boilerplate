package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

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
	ConnMaxIdleTime time.Duration
}

func loadDBConfig(v *viper.Viper) (DBConfig, error) {
	cfg := DBConfig{
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
		ConnMaxIdleTime: v.GetDuration("DB_CONN_MAX_IDLE_TIME"),
	}

	databaseURL := strings.TrimSpace(v.GetString("DATABASE_URL"))
	if databaseURL == "" {
		return cfg, nil
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return DBConfig{}, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return DBConfig{}, fmt.Errorf("DATABASE_URL scheme must be postgres or postgresql")
	}

	if host := strings.TrimSpace(parsed.Hostname()); host != "" {
		cfg.Host = host
	}

	if port := strings.TrimSpace(parsed.Port()); port != "" {
		portInt, convErr := strconv.Atoi(port)
		if convErr != nil {
			return DBConfig{}, fmt.Errorf("invalid DATABASE_URL port %q: %w", port, convErr)
		}
		cfg.Port = portInt
	}

	if parsed.User != nil {
		if user := strings.TrimSpace(parsed.User.Username()); user != "" {
			cfg.User = user
		}
		if password, ok := parsed.User.Password(); ok {
			cfg.Password = password
		}
	}

	if dbName := strings.TrimSpace(strings.TrimPrefix(parsed.Path, "/")); dbName != "" {
		cfg.Name = dbName
	}

	if sslMode := strings.TrimSpace(parsed.Query().Get("sslmode")); sslMode != "" {
		cfg.SSLMode = sslMode
	}

	timeZone := strings.TrimSpace(parsed.Query().Get("TimeZone"))
	if timeZone == "" {
		timeZone = strings.TrimSpace(parsed.Query().Get("timezone"))
	}
	if timeZone != "" {
		cfg.TimeZone = timeZone
	}

	return cfg, nil
}

func setDBDefaults(v *viper.Viper) {
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
	v.SetDefault("DB_CONN_MAX_IDLE_TIME", "2m")
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
		c.TimeZone,
	)
}

func validateDBConfig(c DBConfig) error {
	if err := requireNonEmpty("DB_HOST", c.Host); err != nil {
		return err
	}
	if err := requirePositiveInt("DB_PORT", c.Port); err != nil {
		return err
	}
	if err := requireNonEmpty("DB_USER", c.User); err != nil {
		return err
	}
	if err := requireNonEmpty("DB_NAME", c.Name); err != nil {
		return err
	}
	if err := requirePositiveInt("DB_MAX_OPEN_CONNS", c.MaxOpenConns); err != nil {
		return err
	}
	if err := requirePositiveInt("DB_MAX_IDLE_CONNS", c.MaxIdleConns); err != nil {
		return err
	}
	if err := requirePositiveDuration("DB_CONN_MAX_LIFETIME", c.ConnMaxLifetime); err != nil {
		return err
	}
	if err := requirePositiveDuration("DB_CONN_MAX_IDLE_TIME", c.ConnMaxIdleTime); err != nil {
		return err
	}

	return nil
}
