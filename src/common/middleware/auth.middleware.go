package middleware

import (
	"fmt"

	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/types"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func Protected(cfg config.AuthConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.HandleError(c, fiber.ErrUnauthorized)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &types.JwtPayload{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return response.HandleError(c, fiber.ErrUnauthorized)
	}

		c.Locals("user_id", claims.ID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("is_email_verified", claims.IsEmailVerified)
		c.Locals("session_id", claims.SessionID)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return response.HandleError(c, fiber.ErrUnauthorized)
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return response.HandleError(c, fiber.ErrForbidden)
	}
}
