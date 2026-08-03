package auth

import (
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/dto"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AuthController struct {
	service *AuthService
	cfg     config.AuthConfig
}

func NewAuthController(service *AuthService, cfg config.AuthConfig) *AuthController {
	return &AuthController{
		service: service,
		cfg:     cfg,
	}
}

func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req dto.AuthRequest
	
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	res, refreshToken, err := c.service.Login(req)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(c.cfg.RefreshTokenTTL),
		HTTPOnly: true,
		Secure:   true, 
		SameSite: "Strict",
	})

	return response.Success(ctx, fiber.StatusOK, "Login successful", res)
}

func RegisterRoutes(router fiber.Router, controller *AuthController, protected fiber.Handler, rateLimiter fiber.Handler) {
	authGroup := router.Group("/auth")
	authGroup.Post("/login", rateLimiter, controller.Login)
	authGroup.Post("/register", rateLimiter, controller.Register)
	authGroup.Post("/refresh", rateLimiter, controller.Refresh)
	authGroup.Post("/forgot-password", rateLimiter, controller.ForgotPassword)
	authGroup.Post("/reset-password", rateLimiter, controller.ResetPassword)
	authGroup.Post("/verify-email", rateLimiter, controller.VerifyEmail)
	authGroup.Post("/resend-verification", rateLimiter, controller.ResendVerification)
	
	authGroup.Post("/logout", protected, controller.Logout)
	authGroup.Delete("/account", protected, controller.DeleteAccount)
	
	authGroup.Get("/me", protected, func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		email := c.Locals("email")
		role := c.Locals("role")
		isVerified := c.Locals("is_email_verified")
		return response.Success(c, fiber.StatusOK, "User profile retrieved successfully", fiber.Map{
			"id":                userID,
			"email":             email,
			"role":              role,
			"is_email_verified": isVerified,
		})
	})
}
