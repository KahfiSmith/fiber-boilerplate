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

func (c *AuthController) setRefreshCookie(ctx fiber.Ctx, token string, expiresAt time.Time) {
	ctx.Cookie(&fiber.Cookie{
		Name:     c.cfg.CookieName,
		Value:    token,
		Path:     c.cfg.CookiePath,
		Domain:   c.cfg.CookieDomain,
		Expires:  expiresAt,
		MaxAge:   int(c.cfg.RefreshTokenTTL.Seconds()),
		Secure:   c.cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: c.cfg.CookieSameSite,
	})
}

func (c *AuthController) clearRefreshCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     c.cfg.CookieName,
		Value:    "",
		Path:     c.cfg.CookiePath,
		Domain:   c.cfg.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   c.cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: c.cfg.CookieSameSite,
	})
}

func (c *AuthController) Login(ctx fiber.Ctx) error {
	var req dto.AuthRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	res, refreshToken, err := c.service.Login(req, ip, userAgent)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	c.setRefreshCookie(ctx, refreshToken, time.Now().Add(c.cfg.RefreshTokenTTL))

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
	refreshToken := ctx.Cookies(c.cfg.CookieName)
	if refreshToken == "" {
		return response.HandleError(ctx, exceptions.Unauthorized("REFRESH_TOKEN_MISSING", "Refresh token is missing"))
	}

	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	res, newRefreshToken, err := c.service.Refresh(refreshToken, ip, userAgent)
	if err != nil {
		c.clearRefreshCookie(ctx)
		return response.HandleError(ctx, err)
	}

	c.setRefreshCookie(ctx, newRefreshToken, time.Now().Add(c.cfg.RefreshTokenTTL))

	return response.Success(ctx, fiber.StatusOK, "Token refreshed successfully", res)
}

func (c *AuthController) Logout(ctx fiber.Ctx) error {
	refreshToken := ctx.Cookies(c.cfg.CookieName)

	if refreshToken != "" {
		_ = c.service.Logout(refreshToken)
	}

	c.clearRefreshCookie(ctx)

	return response.Success(ctx, fiber.StatusOK, "Logout successful", nil)
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

func (c *AuthController) DeleteAccount(ctx fiber.Ctx) error {
	userID, okUser := ctx.Locals("user_id").(uint)
	if !okUser {
		return response.HandleError(ctx, exceptions.Unauthorized("ACCESS_TOKEN_INVALID", "Unauthorized"))
	}

	var req dto.DeleteAccountRequest
	if err := validator.ParseAndValidate(ctx, &req); err != nil {
		return response.HandleError(ctx, err)
	}

	if err := c.service.DeleteAccount(userID, req); err != nil {
		return response.HandleError(ctx, err)
	}

	c.clearRefreshCookie(ctx)
	return response.Success(ctx, fiber.StatusOK, "Account deleted successfully", nil)
}

func RegisterRoutes(router fiber.Router, controller *AuthController, protected fiber.Handler, rateLimiter fiber.Handler, originValidator fiber.Handler) {
	authGroup := router.Group("/auth")
	authGroup.Post("/login", originValidator, rateLimiter, controller.Login)
	authGroup.Post("/register", originValidator, rateLimiter, controller.Register)
	authGroup.Post("/refresh", originValidator, rateLimiter, controller.Refresh)
	authGroup.Post("/logout", originValidator, rateLimiter, controller.Logout) // Idempotent, doesn't force protected access-token middleware
	authGroup.Post("/forgot-password", originValidator, rateLimiter, controller.ForgotPassword)
	authGroup.Post("/reset-password", originValidator, rateLimiter, controller.ResetPassword)
	authGroup.Post("/verify-email", originValidator, rateLimiter, controller.VerifyEmail)
	authGroup.Post("/resend-verification", originValidator, rateLimiter, controller.ResendVerification)

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
