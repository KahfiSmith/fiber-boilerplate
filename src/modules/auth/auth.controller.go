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
	service      *AuthService
	oauthService *OAuthService
	cfg          config.AuthConfig
}

func NewAuthController(service *AuthService, oauthService *OAuthService, cfg config.AuthConfig) *AuthController {
	return &AuthController{
		service:      service,
		oauthService: oauthService,
		cfg:          cfg,
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

	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	res, refreshToken, err := c.service.RegisterWithSession(req, ip, userAgent)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	c.setRefreshCookie(ctx, refreshToken, time.Now().Add(c.cfg.RefreshTokenTTL))

	return response.Success(ctx, fiber.StatusCreated, "User registered successfully", res)
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
		// Logout is idempotent: ignore backend errors and clear cookie anyway.
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

func (c *AuthController) Me(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return response.HandleError(ctx, exceptions.Unauthorized("ACCESS_TOKEN_INVALID", "Unauthorized"))
	}

	user, err := c.service.GetCurrentUser(userID)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return response.Success(ctx, fiber.StatusOK, "User profile retrieved successfully", user)
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

func (c *AuthController) GoogleLogin(ctx fiber.Ctx) error {
	if c.oauthService == nil || !c.oauthService.Enabled() {
		return response.HandleError(ctx, exceptions.BadRequest("Google OAuth is not enabled"))
	}

	state, err := c.oauthService.NewState(ctx.Context())
	if err != nil {
		return response.HandleError(ctx, err)
	}

	authURL, err := c.oauthService.AuthURL(state)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	return ctx.Redirect().To(authURL)
}

func (c *AuthController) GoogleCallback(ctx fiber.Ctx) error {
	if c.oauthService == nil || !c.oauthService.Enabled() {
		return response.HandleError(ctx, exceptions.BadRequest("Google OAuth is not enabled"))
	}

	code := ctx.Query("code")
	state := ctx.Query("state")
	if err := c.oauthService.ConsumeState(ctx.Context(), state); err != nil {
		return response.HandleError(ctx, err)
	}

	_, refreshToken, err := c.oauthService.HandleCallback(ctx.Context(), code)
	if err != nil {
		return response.HandleError(ctx, err)
	}

	c.setRefreshCookie(ctx, refreshToken, time.Now().Add(c.cfg.RefreshTokenTTL))

	// Redirect back to the frontend; SessionProvider will bootstrap via /refresh.
	return ctx.Redirect().To(c.cfg.FrontendOrigin)
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

	// Google OAuth (browser redirect flow; no body, no Bearer).
	authGroup.Get("/google", originValidator, controller.GoogleLogin)
	authGroup.Get("/google/callback", originValidator, controller.GoogleCallback)

	authGroup.Delete("/account", protected, controller.DeleteAccount)
	authGroup.Get("/me", protected, controller.Me)
}
