package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func requireNonEmpty(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", key)
	}

	return nil
}

func requirePositiveInt(key string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0", key)
	}

	return nil
}

func requirePositiveDuration(key string, value time.Duration) error {
	if value <= 0 {
		return fmt.Errorf("%s must be > 0", key)
	}

	return nil
}
