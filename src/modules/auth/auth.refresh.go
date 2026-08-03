package auth

import (
	"context"
	"fmt"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/response"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/modules/auth/dto"
	"fiber-boilerplate/src/modules/auth/types"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// --- controller ---
func (c *AuthController) Refresh(ctx fiber.Ctx) error {
	oldRefreshToken := ctx.Cookies("refresh_token")
	if oldRefreshToken == "" {
		return response.HandleError(ctx, exceptions.Unauthorized("Missing refresh token"))
	}

	res, newRefreshToken, err := c.service.Refresh(oldRefreshToken)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(c.cfg.RefreshTokenTTL),
		HTTPOnly: true,
		Secure:   true, 
		SameSite: "Strict",
	})

	return response.Success(ctx, fiber.StatusOK, "Token refreshed successfully", res)
}

// --- service ---
func (s *AuthService) Refresh(refreshToken string) (dto.AuthResponse, string, error) {
	claims := &types.JwtPayload{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Invalid refresh token")
	}

	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d:%s", claims.ID, claims.SessionID)
	
	// validasi terhadap redis (menghindari penggunaan token yang sudah di-revoke)
	storedToken, err := redisclient.Client.Get(ctx, key).Result()
	if err != nil || storedToken != refreshToken {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Session expired or invalid")
	}

	// pastikan user masih ada di db
	var user types.User
	if err := database.DB.First(&user, claims.ID).Error; err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("User no longer exists")
	}

	// generate ulang set token (token rotation dengan sessionid yang sama)
	newAccessToken, newRefreshToken, err := s.tokenService.GenerateTokens(user, claims.SessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate tokens: %w", err)
	}

	// update redis dengan refresh token baru
	if err := redisclient.Client.Set(ctx, key, newRefreshToken, s.cfg.RefreshTokenTTL).Err(); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("store new refresh token: %w", err)
	}

	return dto.AuthResponse{
		AccessToken: newAccessToken,
		User:        user,
	}, newRefreshToken, nil
}
