package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/common/middleware"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/config"

	"github.com/gofiber/fiber/v3"
)

func TestProtectedMiddleware(t *testing.T) {
	cfg := config.AuthConfig{
		JWTAccessSecret: "this-is-a-test-secret-at-least-32-chars-long!",
		JWTIssuer:       "fiber-boilerplate",
		JWTAudience:     "nextjs-boilerplate",
		AccessTokenTTL:  15 * time.Minute,
	}
	tokenService := jwt.NewTokenService(cfg)

	app := fiber.New(fiber.Config{
		ErrorHandler: response.GlobalErrorHandler,
	})

	app.Get("/test-protected", middleware.Protected(cfg), func(c fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

	t.Run("Missing Authorization header returns 401 ACCESS_TOKEN_MISSING", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-protected", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}

		var body response.APIResponse
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body.Code != "ACCESS_TOKEN_MISSING" {
			t.Errorf("expected ACCESS_TOKEN_MISSING, got %s", body.Code)
		}
	})

	t.Run("Valid token succeeds", func(t *testing.T) {
		token, _, err := tokenService.GenerateAccessToken(42, "user@example.com", "user", "session-123")
		if err != nil {
			t.Fatalf("GenerateAccessToken error: %v", err)
		}

		req := httptest.NewRequest("GET", "/test-protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}
