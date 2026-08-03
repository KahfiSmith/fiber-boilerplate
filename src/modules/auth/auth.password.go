package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/modules/auth/dto"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- controller ---

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

// --- service ---

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) (string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil {
		// return empty token without error to avoid email enumeration
		return "", nil
	}

	resetToken := uuid.New().String()
	ctx := context.Background()
	key := fmt.Sprintf("reset_token:%s", resetToken)
	ttl := 15 * time.Minute

	if err := redisclient.Client.Set(ctx, key, user.ID, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	return resetToken, nil
}

func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	ctx := context.Background()
	key := fmt.Sprintf("reset_token:%s", req.ResetToken)

	val, err := redisclient.Client.Get(ctx, key).Result()
	if err != nil || val == "" {
		return exceptions.BadRequest("Invalid or expired reset token")
	}

	userID64, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return exceptions.BadRequest("Invalid reset token data")
	}
	userID := uint(userID64)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.repo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// delete reset token so it cannot be reused
	redisclient.Client.Del(ctx, key)

	return nil
}
