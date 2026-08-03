package jwt_test

import (
	"testing"
	"time"

	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/config"
)

func TestJWTValidation(t *testing.T) {
	cfg := config.AuthConfig{
		JWTAccessSecret: "test-jwt-access-secret-32byteslong!",
		JWTIssuer:       "fiber-boilerplate",
		JWTAudience:     "nextjs-boilerplate",
		AccessTokenTTL:  1 * time.Second,
	}

	tokenService := jwt.NewTokenService(cfg)

	tokenStr, expiresAt, err := tokenService.GenerateAccessToken(1, "test@example.com", "user", "sess-123")
	if err != nil || tokenStr == "" {
		t.Fatalf("failed to generate access token: %v", err)
	}
	if expiresAt.Before(time.Now()) {
		t.Fatalf("expiresAt should be in the future")
	}

	claims, err := tokenService.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.Sub != 1 || claims.SessionID != "sess-123" || claims.Role != "user" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	_, err = tokenService.ValidateAccessToken("invalid.jwt.token")
	if err != jwt.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}

	time.Sleep(1200 * time.Millisecond)
	_, err = tokenService.ValidateAccessToken(tokenStr)
	if err != jwt.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
