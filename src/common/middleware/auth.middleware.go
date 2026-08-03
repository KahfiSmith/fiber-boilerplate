package middleware

import (
	"errors"
	"strings"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/config"

	"github.com/gofiber/fiber/v3"
)

func Protected(cfg config.AuthConfig) fiber.Handler {
	tokenService := jwt.NewTokenService(cfg)

	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.HandleError(c, exceptions.Unauthorized("ACCESS_TOKEN_MISSING", "Access token is missing"))
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenService.ValidateAccessToken(tokenString)

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return response.HandleError(c, exceptions.Unauthorized("ACCESS_TOKEN_EXPIRED", "Access token has expired"))
			}
			return response.HandleError(c, exceptions.Unauthorized("ACCESS_TOKEN_INVALID", "Access token is invalid"))
		}

		c.Locals("user_id", claims.Sub)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("session_id", claims.SessionID)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return response.HandleError(c, exceptions.Forbidden("FORBIDDEN", "Forbidden: role required"))
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		return response.HandleError(c, fiber.ErrForbidden)
	}
}
