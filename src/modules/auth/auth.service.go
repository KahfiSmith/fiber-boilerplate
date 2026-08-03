package auth

import (
	"context"
	"strings"
	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/jwt"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/dto"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         *AuthRepository
	tokenService *jwt.TokenService
	cfg          config.AuthConfig
}

func NewAuthService(repo *AuthRepository, tokenService *jwt.TokenService, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		repo:         repo,
		tokenService: tokenService,
		cfg:          cfg,
	}
}

func (s *AuthService) Login(req dto.AuthRequest) (dto.AuthResponse, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Invalid credentials")
	}

	sessionID := uuid.New().String()
	accessToken, refreshToken, err := s.tokenService.GenerateTokens(*user, sessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate tokens: %w", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d:%s", user.ID, sessionID)
	if err := redisclient.Client.Set(ctx, key, refreshToken, s.cfg.RefreshTokenTTL).Err(); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("store refresh token: %w", err)
	}

	return dto.AuthResponse{
		AccessToken: accessToken,
		User:        *user,
	}, refreshToken, nil
}
