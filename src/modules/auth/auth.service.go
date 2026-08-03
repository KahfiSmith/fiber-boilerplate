package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/jwt"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/modules/auth/dto"
	"fiber-boilerplate/src/modules/auth/types"

	golangjwt "github.com/golang-jwt/jwt/v5"
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

func (s *AuthService) Refresh(refreshToken string) (dto.AuthResponse, string, error) {
	claims := &types.JwtPayload{}
	token, err := golangjwt.ParseWithClaims(refreshToken, claims, func(token *golangjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Invalid refresh token")
	}

	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d:%s", claims.ID, claims.SessionID)

	storedToken, err := redisclient.Client.Get(ctx, key).Result()
	if err != nil || storedToken != refreshToken {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("Session expired or invalid")
	}

	var user types.User
	if err := database.DB.First(&user, claims.ID).Error; err != nil {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("User no longer exists")
	}

	newAccessToken, newRefreshToken, err := s.tokenService.GenerateTokens(user, claims.SessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate tokens: %w", err)
	}

	if err := redisclient.Client.Set(ctx, key, newRefreshToken, s.cfg.RefreshTokenTTL).Err(); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("store new refresh token: %w", err)
	}

	return dto.AuthResponse{
		AccessToken: newAccessToken,
		User:        user,
	}, newRefreshToken, nil
}

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) (string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil {
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

	redisclient.Client.Del(ctx, key)

	return nil
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

func (s *AuthService) Logout(userID uint, sessionID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%d:%s", userID, sessionID)

	if err := redisclient.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
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