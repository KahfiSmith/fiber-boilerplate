package middleware

import (
	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/config"

	"github.com/gofiber/fiber/v3"
)

func ValidateOrigin(cfg config.AuthConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")
		if origin != "" && origin != cfg.FrontendOrigin {
			return response.HandleError(c, exceptions.Forbidden("FORBIDDEN", "Origin not allowed"))
		}
		return c.Next()
	}
}
