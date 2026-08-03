package auth

import (
	"context"
	"fmt"
	"time"

	"fiber-boilerplate/src/common/response"
	redisclient "fiber-boilerplate/src/common/redis"

	"github.com/gofiber/fiber/v3"
)

// --- controller ---
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

func (c *AuthController) clearCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // expired ke belakang
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
}

// --- service ---
func (s *AuthService) Logout(userID uint, sessionID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d:%s", userID, sessionID)
	
	// hapus token dari redis
	if err := redisclient.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}
