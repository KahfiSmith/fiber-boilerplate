package auth

import (
	"context"
	"fmt"

	"fiber-boilerplate/src/common/exceptions"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/modules/auth/dto"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

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

func (s *AuthService) DeleteAccount(userID uint, req dto.DeleteAccountRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return exceptions.NotFound("User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return exceptions.Unauthorized("Invalid password confirmation")
	}

	if err := s.repo.Delete(userID); err != nil {
		return fmt.Errorf("delete user account: %w", err)
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("refresh_token:%d:*", userID)
	iter := redisclient.Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		redisclient.Client.Del(ctx, iter.Val())
	}

	return nil
}
