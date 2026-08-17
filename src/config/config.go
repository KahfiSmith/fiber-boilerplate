package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	HTTP  HTTPConfig
	DB    DBConfig
	Redis RedisConfig
	Auth  AuthConfig
	OAuth OAuthConfig
	Fiber FiberConfig
}

type AppConfig struct {
	Name string `mapstructure:"name" validate:"required"`
	Env  string `mapstructure:"env" validate:"required,oneof=development production test"`
}

type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port" validate:"required"`
}

type DBConfig struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            int           `mapstructure:"port" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password" validate:"required"`
	Name            string        `mapstructure:"name" validate:"required"`
	SSLMode         string        `mapstructure:"sslmode" validate:"required"`
	Timezone        string        `mapstructure:"timezone" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" validate:"required,min=1m"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"required,min=1m"`
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode, db.Timezone)
}

type RedisConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         int           `mapstructure:"port" validate:"required"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout" validate:"required,min=1s"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
	KeyPrefix    string        `mapstructure:"key_prefix" validate:"required"`
}

type AuthConfig struct {
	JWTAccessSecret     string        `mapstructure:"jwt_access_secret" validate:"required,min=32"`
	JWTIssuer           string        `mapstructure:"jwt_issuer" validate:"required"`
	JWTAudience         string        `mapstructure:"jwt_audience" validate:"required"`
	AccessTokenTTL      time.Duration `mapstructure:"access_token_ttl" validate:"required,min=1m"`
	RefreshTokenHMACKey string        `mapstructure:"refresh_token_hmac_key" validate:"required,min=32"`
	RefreshTokenTTL     time.Duration `mapstructure:"refresh_token_ttl" validate:"required,min=24h"`
	FrontendOrigin      string        `mapstructure:"frontend_origin" validate:"required"`
	CookieName          string        `mapstructure:"cookie_name" validate:"required"`
	CookiePath          string        `mapstructure:"cookie_path" validate:"required"`
	CookieSecure        bool          `mapstructure:"cookie_secure"`
	CookieSameSite      string        `mapstructure:"cookie_same_site" validate:"required"`
	CookieDomain        string        `mapstructure:"cookie_domain"`
	BcryptCost          int           `mapstructure:"bcrypt_cost" validate:"required,min=10,max=14"`
	RateLimitPerMin     int           `mapstructure:"rate_limit_per_min" validate:"required,min=1"`
	DebugExposeOTP      bool          `mapstructure:"debug_expose_otp"`
}

type FiberConfig struct {
	ReadTimeout   time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout  time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
	IdleTimeout   time.Duration `mapstructure:"idle_timeout" validate:"required,min=1s"`
	BodyLimit     int           `mapstructure:"body_limit" validate:"required,min=1"`
	EnableMetrics bool          `mapstructure:"enable_metrics"`
	EnablePprof   bool          `mapstructure:"enable_pprof"`
}

// OAuthConfig holds Google OIDC settings. Nothing is required so the feature
// can be disabled (GOOGLE_ENABLED=false) without breaking startup.
type OAuthConfig struct {
	GoogleEnabled      bool   `mapstructure:"google_enabled"`
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	GoogleRedirectURL  string `mapstructure:"google_redirect_url"`
	GoogleDiscoveryURL string `mapstructure:"google_discovery_url"`
}

func Load() (Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
