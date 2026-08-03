package auth

import (
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/dto"

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

func (c *AuthController) Register(ctx fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	user, token, err := c.service.Register(req)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	resData := fiber.Map{
		"user": user,
	}
	if c.cfg.DebugExposeOTP {
		resData["verification_token"] = token
	}

	return response.Success(ctx, fiber.StatusCreated, "User registered successfully", resData)
}

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

func (c *AuthController) ForgotPassword(ctx fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	resetToken, err := c.service.ForgotPassword(req)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	resData := fiber.Map{}
	if c.cfg.DebugExposeOTP {
		resData["reset_token"] = resetToken
	}

	return response.Success(ctx, fiber.StatusOK, "If email is registered, a password reset link/token has been generated", resData)
}

func (c *AuthController) ResetPassword(ctx fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	if err := c.service.ResetPassword(req); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.StatusOK, "Password reset successfully", nil)
}

func (c *AuthController) VerifyEmail(ctx fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	if err := c.service.VerifyEmail(req); err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.StatusOK, "Email verified successfully", nil)
}

func (c *AuthController) ResendVerification(ctx fiber.Ctx) error {
	var req dto.ResendVerificationRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	token, err := c.service.ResendVerification(req)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	resData := fiber.Map{}
	if c.cfg.DebugExposeOTP {
		resData["verification_token"] = token
	}

	return response.Success(ctx, fiber.StatusOK, "Verification token generated successfully", resData)
}

func (c *AuthController) Logout(ctx fiber.Ctx) error {
	userID, okUser := ctx.Locals("user_id").(uint)
	sessionID, okSess := ctx.Locals("session_id").(string)
	if !okUser || !okSess {
		c.clearCookie(ctx)
		return response.Success(ctx, fiber.StatusOK, "Logged out (cookie cleared)", nil)
	}

	if err := c.service.Logout(userID, sessionID); err != nil {
		return response.HandleError(ctx, err)
	}

	c.clearCookie(ctx)
	return response.Success(ctx, fiber.StatusOK, "Logged out successfully", nil)
}

func (c *AuthController) DeleteAccount(ctx fiber.Ctx) error {
	userID, okUser := ctx.Locals("user_id").(uint)
	if !okUser {
		return response.HandleError(ctx, exceptions.Unauthorized("Unauthorized"))
	}

	var req dto.DeleteAccountRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	if err := c.service.DeleteAccount(userID, req); err != nil {
		return response.HandleError(ctx, err)
	}

	c.clearCookie(ctx)
	return response.Success(ctx, fiber.StatusOK, "Account deleted successfully", nil)
}

func (c *AuthController) clearCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
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