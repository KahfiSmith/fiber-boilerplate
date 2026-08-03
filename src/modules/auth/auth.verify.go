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
)

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

func (s *AuthService) VerifyEmail(req dto.VerifyEmailRequest) error {
	ctx := context.Background()
	key := fmt.Sprintf("verify_email_token:%s", req.Token)

	val, err := redisclient.Client.Get(ctx, key).Result()
	if err != nil || val == "" {
		return exceptions.BadRequest("Invalid or expired verification token")
	}

	userID64, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return exceptions.BadRequest("Invalid verification token data")
	}
	userID := uint(userID64)

	if err := s.repo.MarkEmailAsVerified(userID); err != nil {
		return fmt.Errorf("mark email as verified: %w", err)
	}

	redisclient.Client.Del(ctx, key)

	return nil
}

func (s *AuthService) ResendVerification(req dto.ResendVerificationRequest) (string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil {
		return "", exceptions.BadRequest("User not found")
	}

	if user.IsEmailVerified {
		return "", exceptions.BadRequest("Email is already verified")
	}

	return s.createVerificationToken(user.ID)
}

func (s *AuthService) createVerificationToken(userID uint) (string, error) {
	token := uuid.New().String()
	ctx := context.Background()
	key := fmt.Sprintf("verify_email_token:%s", token)
	ttl := 24 * time.Hour

	if err := redisclient.Client.Set(ctx, key, userID, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to store verification token: %w", err)
	}

	return token, nil
}
