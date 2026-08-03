package auth_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/common/middleware"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/modules/auth"
	"fiber-boilerplate/src/modules/auth/types"

	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) (*fiber.App, config.AuthConfig) {
	cfg := config.AuthConfig{
		JWTAccessSecret:     "test-jwt-access-secret-32byteslong!",
		JWTIssuer:           "fiber-boilerplate",
		JWTAudience:         "nextjs-boilerplate",
		AccessTokenTTL:      2 * time.Second,
		RefreshTokenHMACKey: "test-refresh-hmac-key-32byteslong!",
		RefreshTokenTTL:     10 * time.Minute,
		FrontendOrigin:      "http://localhost:3000",
		CookieName:          "refresh_token",
		CookiePath:          "/api/v1/auth",
		CookieSecure:        false,
		CookieSameSite:      "Lax",
		BcryptCost:          10,
		RateLimitPerMin:     100,
	}

	dbCfg := config.DBConfig{
		Host:            "127.0.0.1",
		Port:            5432,
		User:            "postgres",
		Password:        "kahfismith",
		Name:            "boilerplate",
		SSLMode:         "disable",
		Timezone:        "UTC",
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	var err error
	database.DB, err = gorm.Open(postgres.Open(dbCfg.DSN()), &gorm.Config{})
	if err != nil || database.DB.Exec("SELECT 1").Error != nil {
		t.Skip("PostgreSQL database is not available for integration tests")
		return nil, config.AuthConfig{}
	}

	redisclient.Client = goredis.NewClient(&goredis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := redisclient.Client.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis is not available for integration tests")
		return nil, config.AuthConfig{}
	}

	database.DB.AutoMigrate(&types.User{})

	app := fiber.New(fiber.Config{
		ErrorHandler: response.GlobalErrorHandler,
	})

	tokenService := jwt.NewTokenService(cfg)
	authRepo := auth.NewAuthRepository()
	refreshRepo := auth.NewRefreshRepository()
	authService := auth.NewAuthService(authRepo, refreshRepo, tokenService, cfg)
	authController := auth.NewAuthController(authService, cfg)

	protected := middleware.Protected(cfg)
	originValidator := middleware.ValidateOrigin(cfg)
	rateLimiter := middleware.RateLimiter(cfg.RateLimitPerMin, time.Minute)

	v1 := app.Group("/api/v1")
	auth.RegisterRoutes(v1, authController, protected, rateLimiter, originValidator)

	return app, cfg
}

func createTestUser(t *testing.T, email, password string) types.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	user := types.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
	}
	database.DB.Where("LOWER(email) = LOWER(?)", email).Delete(&types.User{})
	database.DB.Create(&user)
	return user
}

func TestAuthFlow(t *testing.T) {
	app, cfg := setupTestApp(t)
	_ = createTestUser(t, "testauth@example.com", "Password123!")

	t.Run("Foreign Origin Rejected", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"testauth@example.com","password":"Password123!"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://malicious-site.com")
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Login Failure Wrong Password", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"testauth@example.com","password":"WrongPassword"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", cfg.FrontendOrigin)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	var accessToken string
	var refreshCookie *http.Cookie

	t.Run("Login Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"testauth@example.com","password":"Password123!"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", cfg.FrontendOrigin)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var res map[string]interface{}
		json.Unmarshal(body, &res)

		data := res["data"].(map[string]interface{})
		accessToken = data["access_token"].(string)

		if _, exists := data["refresh_token"]; exists {
			t.Fatalf("refresh_token must NOT be in JSON body")
		}

		cookies := resp.Cookies()
		for _, c := range cookies {
			if c.Name == cfg.CookieName {
				refreshCookie = c
				break
			}
		}
		if refreshCookie == nil {
			t.Fatalf("refresh cookie missing")
		}
		if !refreshCookie.HttpOnly || refreshCookie.Path != cfg.CookiePath {
			t.Fatalf("cookie attributes invalid: HttpOnly=%v Path=%s", refreshCookie.HttpOnly, refreshCookie.Path)
		}

		ctx := context.Background()
		rawInRedis, _ := redisclient.Client.Get(ctx, refreshCookie.Value).Result()
		if rawInRedis != "" {
			t.Fatalf("raw refresh token must NOT be stored in redis")
		}
	})

	t.Run("Access Protected Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Access Token Expired", func(t *testing.T) {
		time.Sleep(2100 * time.Millisecond) 
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "ACCESS_TOKEN_EXPIRED") {
			t.Fatalf("expected ACCESS_TOKEN_EXPIRED, got %s", string(body))
		}
	})

	var newAccessToken string
	var newRefreshCookie *http.Cookie

	t.Run("Refresh Token Rotation Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.Header.Set("Origin", cfg.FrontendOrigin)
		req.AddCookie(refreshCookie)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d body=%s", resp.StatusCode, string(body))
		}

		body, _ := io.ReadAll(resp.Body)
		var res map[string]interface{}
		json.Unmarshal(body, &res)

		data := res["data"].(map[string]interface{})
		newAccessToken = data["access_token"].(string)

		for _, c := range resp.Cookies() {
			if c.Name == cfg.CookieName {
				newRefreshCookie = c
				break
			}
		}
		if newRefreshCookie == nil || newRefreshCookie.Value == refreshCookie.Value {
			t.Fatalf("new refresh token must be rotated")
		}
	})

	t.Run("Refresh Token Reuse Triggers Revocation", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.Header.Set("Origin", cfg.FrontendOrigin)
		req.AddCookie(refreshCookie)
		resp, _ := app.Test(req)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "REFRESH_TOKEN_REUSED") {
			t.Fatalf("expected REFRESH_TOKEN_REUSED code, got %s", string(body))
		}

		req2 := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req2.Header.Set("Origin", cfg.FrontendOrigin)
		req2.AddCookie(newRefreshCookie)
		resp2, _ := app.Test(req2)
		if resp2.StatusCode != http.StatusUnauthorized {
			t.Fatalf("new token family should be revoked, got status %d", resp2.StatusCode)
		}
	})

	t.Run("Logout Without Cookie or Expired Token Success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		req.Header.Set("Origin", cfg.FrontendOrigin)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	_ = newAccessToken
}
