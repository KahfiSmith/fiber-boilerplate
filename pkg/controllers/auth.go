package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	config "fiber-boilerplate/pkg/configs"
	"fiber-boilerplate/pkg/dto/request"
	"fiber-boilerplate/pkg/dto/response"
	"fiber-boilerplate/pkg/entities"
	serverMiddleware "fiber-boilerplate/pkg/server/middleware"
	"fiber-boilerplate/pkg/services"
	"fiber-boilerplate/pkg/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a user account and returns access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.RegisterRequest true "Register request"
// @Success 201 {object} response.APIResponse{data=response.AuthTokenResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 409 {object} response.APIResponse{error=string}
// @Failure 429 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/register [post]
func (a *AuthController) Register(c fiber.Ctx) error {
	var req request.RegisterRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	result, err := a.authService.Register(c.Context(), entities.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}, requestMeta(c))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid input", nil)
		case errors.Is(err, services.ErrEmailAlreadyUsed):
			return utils.Error(c, fiber.StatusConflict, "email already registered", nil)
		case errors.Is(err, services.ErrRateLimited):
			return utils.Error(c, fiber.StatusTooManyRequests, "too many requests", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "register failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusCreated, authTokenResponse(result))
}

// Login godoc
// @Summary Login with email and password
// @Description Validates credentials and returns an OTP challenge.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.LoginRequest true "Login request"
// @Success 200 {object} response.APIResponse{data=response.OTPChallengeResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 429 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/login [post]
func (a *AuthController) Login(c fiber.Ctx) error {
	var req request.LoginRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	result, err := a.authService.Login(c.Context(), entities.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, requestMeta(c))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid input", nil)
		case errors.Is(err, services.ErrInvalidCredentials):
			return utils.Error(c, fiber.StatusUnauthorized, "invalid credentials", nil)
		case errors.Is(err, services.ErrRateLimited):
			return utils.Error(c, fiber.StatusTooManyRequests, "too many requests", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "login failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusOK, otpChallengeResponse(result))
}

// ForgotPassword godoc
// @Summary Start forgot-password flow
// @Description Generates an OTP challenge for password reset.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} response.APIResponse{data=response.OTPChallengeResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 429 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/forgot-password [post]
func (a *AuthController) ForgotPassword(c fiber.Ctx) error {
	var req request.ForgotPasswordRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	result, err := a.authService.ForgotPassword(c.Context(), entities.ForgotPasswordInput{
		Email: req.Email,
	}, requestMeta(c))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid input", nil)
		case errors.Is(err, services.ErrRateLimited):
			return utils.Error(c, fiber.StatusTooManyRequests, "too many requests", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "forgot password failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusOK, otpChallengeResponse(result))
}

// VerifyOTP godoc
// @Summary Verify OTP challenge
// @Description Verifies an OTP challenge and returns access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.VerifyOTPRequest true "Verify OTP request"
// @Success 200 {object} response.APIResponse{data=response.AuthTokenResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 429 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/otp/verify [post]
func (a *AuthController) VerifyOTP(c fiber.Ctx) error {
	var req request.VerifyOTPRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	result, err := a.authService.VerifyOTP(c.Context(), entities.VerifyOTPInput{
		ChallengeID: req.ChallengeID,
		OTP:         req.OTP,
	}, requestMeta(c))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid input", nil)
		case errors.Is(err, services.ErrInvalidOTP):
			return utils.Error(c, fiber.StatusUnauthorized, "invalid otp", nil)
		case errors.Is(err, services.ErrOTPExpired):
			return utils.Error(c, fiber.StatusUnauthorized, "otp expired", nil)
		case errors.Is(err, services.ErrOTPAttemptsExceeded):
			return utils.Error(c, fiber.StatusTooManyRequests, "otp attempts exceeded", nil)
		case errors.Is(err, services.ErrRateLimited):
			return utils.Error(c, fiber.StatusTooManyRequests, "too many requests", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "verify otp failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusOK, authTokenResponse(result))
}

// ResetPassword godoc
// @Summary Reset password with OTP
// @Description Resets a user's password after OTP verification.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} response.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 429 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/reset-password [post]
func (a *AuthController) ResetPassword(c fiber.Ctx) error {
	var req request.ResetPasswordRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	err := a.authService.ResetPassword(c.Context(), entities.ResetPasswordInput{
		ChallengeID: req.ChallengeID,
		OTP:         req.OTP,
		NewPassword: req.NewPassword,
	}, requestMeta(c))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid input", nil)
		case errors.Is(err, services.ErrInvalidOTP):
			return utils.Error(c, fiber.StatusUnauthorized, "invalid otp", nil)
		case errors.Is(err, services.ErrOTPExpired):
			return utils.Error(c, fiber.StatusUnauthorized, "otp expired", nil)
		case errors.Is(err, services.ErrOTPAttemptsExceeded):
			return utils.Error(c, fiber.StatusTooManyRequests, "otp attempts exceeded", nil)
		case errors.Is(err, services.ErrRateLimited):
			return utils.Error(c, fiber.StatusTooManyRequests, "too many requests", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "reset password failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusOK, map[string]any{
		"message": "password reset success",
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchanges a refresh token for a new access token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.APIResponse{data=response.AuthTokenResponse}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/refresh [post]
func (a *AuthController) Refresh(c fiber.Ctx) error {
	var req request.RefreshTokenRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	result, err := a.authService.Refresh(c.Context(), req.RefreshToken, requestMeta(c))
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid refresh token", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "refresh failed", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, authTokenResponse(result))
}

// Logout godoc
// @Summary Logout a session
// @Description Revokes a refresh token and logs the current client out.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body request.LogoutRequest true "Logout request"
// @Success 200 {object} response.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/logout [post]
func (a *AuthController) Logout(c fiber.Ctx) error {
	var req request.LogoutRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := a.authService.Logout(c.Context(), req.RefreshToken); err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid refresh token", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "logout failed", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, map[string]any{
		"message": "logout success",
	})
}

// Me godoc
// @Summary Get current user
// @Description Returns the profile of the authenticated user.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=response.AuthUserResponse}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/me [get]
func (a *AuthController) Me(c fiber.Ctx) error {
	accessToken, err := bearerToken(c.Get("Authorization"))
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "missing or invalid authorization header", nil)
	}

	user, err := a.authService.Me(c.Context(), accessToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidAccessToken) {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "fetch profile failed", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, authUserResponse(user))
}

// Sessions godoc
// @Summary List active sessions
// @Description Returns all sessions for the authenticated user.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]response.AuthSessionResponse}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/sessions [get]
func (a *AuthController) Sessions(c fiber.Ctx) error {
	accessToken, err := bearerToken(c.Get("Authorization"))
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "missing or invalid authorization header", nil)
	}

	sessions, err := a.authService.ListSessions(c.Context(), accessToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidAccessToken) {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "list sessions failed", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, authSessionResponses(sessions))
}

// RevokeSession godoc
// @Summary Revoke a session
// @Description Revokes one session belonging to the authenticated user.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body request.RevokeSessionRequest true "Revoke session request"
// @Success 200 {object} response.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} response.APIResponse{error=string}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 404 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/sessions/revoke [post]
func (a *AuthController) RevokeSession(c fiber.Ctx) error {
	accessToken, err := bearerToken(c.Get("Authorization"))
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "missing or invalid authorization header", nil)
	}

	var req request.RevokeSessionRequest
	if err := parseAndValidate(c, &req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := a.authService.RevokeSession(c.Context(), accessToken, req.SessionID); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidAccessToken):
			return utils.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		case errors.Is(err, services.ErrInvalidInput):
			return utils.Error(c, fiber.StatusBadRequest, "invalid session id", nil)
		case errors.Is(err, services.ErrSessionNotFound):
			return utils.Error(c, fiber.StatusNotFound, "session not found", nil)
		default:
			return utils.Error(c, fiber.StatusInternalServerError, "revoke session failed", err.Error())
		}
	}

	return utils.Success(c, fiber.StatusOK, map[string]any{
		"message": "session revoked",
	})
}

// RevokeAllSessions godoc
// @Summary Revoke all sessions
// @Description Revokes all sessions for the authenticated user.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=map[string]interface{}}
// @Failure 401 {object} response.APIResponse{error=string}
// @Failure 500 {object} response.APIResponse{error=string}
// @Router /auth/sessions/revoke-all [post]
func (a *AuthController) RevokeAllSessions(c fiber.Ctx) error {
	accessToken, err := bearerToken(c.Get("Authorization"))
	if err != nil {
		return utils.Error(c, fiber.StatusUnauthorized, "missing or invalid authorization header", nil)
	}

	if err := a.authService.RevokeAllSessions(c.Context(), accessToken); err != nil {
		if errors.Is(err, services.ErrInvalidAccessToken) {
			return utils.Error(c, fiber.StatusUnauthorized, "invalid access token", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "revoke all sessions failed", err.Error())
	}

	return utils.Success(c, fiber.StatusOK, map[string]any{
		"message": "all sessions revoked",
	})
}

func parseAndValidate(c fiber.Ctx, payload any) error {
	body := c.Body()
	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("parse body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	validateAny := c.Locals(serverMiddleware.ValidatorLocalKey)
	validate, ok := validateAny.(*validator.Validate)
	if !ok || validate == nil {
		return errors.New("validator is not available in request context")
	}

	return config.ValidateStruct(validate, payload)
}

func requestMeta(c fiber.Ctx) entities.AuthClientMeta {
	return entities.AuthClientMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}
}

func authTokenResponse(result entities.AuthTokens) response.AuthTokenResponse {
	return response.AuthTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresInSec: result.ExpiresInSec,
		SessionID:    result.SessionID,
		User:         authUserResponse(result.User),
	}
}

func otpChallengeResponse(result entities.OTPChallengeResult) response.OTPChallengeResponse {
	return response.OTPChallengeResponse{
		ChallengeID:  result.ChallengeID,
		ExpiresInSec: result.ExpiresInSec,
		OTP:          result.OTP,
	}
}

func authUserResponse(user entities.AuthUser) response.AuthUserResponse {
	return response.AuthUserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func authSessionResponses(sessions []entities.AuthSession) []response.AuthSessionResponse {
	resp := make([]response.AuthSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		resp = append(resp, response.AuthSessionResponse{
			SessionID: session.SessionID,
			UserAgent: session.UserAgent,
			IPAddress: session.IPAddress,
			CreatedAt: session.CreatedAt.Format(timeLayoutRFC3339),
			ExpiresAt: session.ExpiresAt.Format(timeLayoutRFC3339),
			Current:   session.Current,
		})
	}

	return resp
}

func bearerToken(authHeader string) (string, error) {
	trimmed := strings.TrimSpace(authHeader)
	if trimmed == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header must be Bearer token")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("bearer token is empty")
	}

	return token, nil
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
