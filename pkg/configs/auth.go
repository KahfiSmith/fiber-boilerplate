package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int
	RateLimitPerMin int
	OTPTTL          time.Duration
	OTPMaxAttempts  int
	DebugExposeOTP  bool
}

func loadAuthConfig(v *viper.Viper) (AuthConfig, error) {
	debugExposeOTP := v.GetBool("AUTH_DEBUG_EXPOSE_OTP")
	if !v.IsSet("AUTH_DEBUG_EXPOSE_OTP") && v.IsSet("AUTH_DEBUG_EXPOSE_TOKENS") {
		debugExposeOTP = v.GetBool("AUTH_DEBUG_EXPOSE_TOKENS")
	}

	return AuthConfig{
		JWTSecret:       v.GetString("JWT_SECRET"),
		AccessTokenTTL:  v.GetDuration("ACCESS_TOKEN_TTL"),
		RefreshTokenTTL: v.GetDuration("REFRESH_TOKEN_TTL"),
		BcryptCost:      v.GetInt("BCRYPT_COST"),
		RateLimitPerMin: v.GetInt("AUTH_RATE_LIMIT_PER_MINUTE"),
		OTPTTL:          v.GetDuration("AUTH_OTP_TTL"),
		OTPMaxAttempts:  v.GetInt("AUTH_OTP_MAX_ATTEMPTS"),
		DebugExposeOTP:  debugExposeOTP,
	}, nil
}

func setAuthDefaults(v *viper.Viper) {
	v.SetDefault("JWT_SECRET", "")
	v.SetDefault("ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("REFRESH_TOKEN_TTL", "168h")
	v.SetDefault("BCRYPT_COST", 12)
	v.SetDefault("AUTH_RATE_LIMIT_PER_MINUTE", 5)
	v.SetDefault("AUTH_OTP_TTL", "5m")
	v.SetDefault("AUTH_OTP_MAX_ATTEMPTS", 5)
	v.SetDefault("AUTH_DEBUG_EXPOSE_OTP", false)
}

func validateAuthConfig(c AuthConfig) error {
	if err := requireNonEmpty("JWT_SECRET", c.JWTSecret); err != nil {
		return err
	}
	if strings.ContainsAny(c.JWTSecret, " \t\r\n") {
		return fmt.Errorf("JWT_SECRET must not contain whitespace")
	}
	if err := requirePositiveDuration("ACCESS_TOKEN_TTL", c.AccessTokenTTL); err != nil {
		return err
	}
	if err := requirePositiveDuration("REFRESH_TOKEN_TTL", c.RefreshTokenTTL); err != nil {
		return err
	}
	if c.BcryptCost < bcrypt.MinCost || c.BcryptCost > bcrypt.MaxCost {
		return fmt.Errorf("BCRYPT_COST must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	if err := requirePositiveInt("AUTH_RATE_LIMIT_PER_MINUTE", c.RateLimitPerMin); err != nil {
		return err
	}
	if err := requirePositiveDuration("AUTH_OTP_TTL", c.OTPTTL); err != nil {
		return err
	}
	if err := requirePositiveInt("AUTH_OTP_MAX_ATTEMPTS", c.OTPMaxAttempts); err != nil {
		return err
	}

	return nil
}

func ValidateStruct(validate *validator.Validate, payload any) error {
	if err := validate.Struct(payload); err != nil {
		return fmt.Errorf("validate request: %w", err)
	}

	return nil
}
