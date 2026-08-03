package auth

import (
	"fmt"
	"strings"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/validator"
	"fiber-boilerplate/src/modules/auth/dto"
	"fiber-boilerplate/src/modules/auth/types"
	"fiber-boilerplate/src/common/exceptions"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

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

func (s *AuthService) Register(req dto.RegisterRequest) (types.User, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	exists, _ := s.repo.FindByEmail(cleanEmail)
	if exists != nil {
		return types.User{}, "", exceptions.BadRequest("Email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		return types.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	user := types.User{
		Name:            strings.TrimSpace(req.Name),
		Email:           cleanEmail,
		PasswordHash:    string(hashedPassword),
		Role:            "user",
		IsEmailVerified: false,
	}

	if err := s.repo.Create(&user); err != nil {
		return types.User{}, "", fmt.Errorf("create user: %w", err)
	}

	verifyToken, _ := s.createVerificationToken(user.ID)

	return user, verifyToken, nil
}
