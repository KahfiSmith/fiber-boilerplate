package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/jwt"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/modules/auth/dto"
	"fiber-boilerplate/src/modules/auth/types"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo         *AuthRepository
	refreshRepo  *RefreshRepository
	tokenService *jwt.TokenService
	cfg          config.AuthConfig
}

func NewAuthService(repo *AuthRepository, refreshRepo *RefreshRepository, tokenService *jwt.TokenService, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		repo:         repo,
		refreshRepo:  refreshRepo,
		tokenService: tokenService,
		cfg:          cfg,
	}
}

func (s *AuthService) RegisterWithSession(req dto.RegisterRequest, ip string, userAgent string) (dto.AuthResponse, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	exists, _ := s.repo.FindByEmail(cleanEmail)
	if exists != nil {
		return dto.AuthResponse{}, "", exceptions.BadRequest("Email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("hash password: %w", err)
	}
	passwordHash := string(hashedPassword)

	user := types.User{
		Name:            strings.TrimSpace(req.Name),
		Email:           cleanEmail,
		PasswordHash:    &passwordHash,
		Role:            "user",
		IsEmailVerified: false,
	}

	if err := s.repo.Create(&user); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("create user: %w", err)
	}

	_, _ = s.createVerificationToken(user.ID)

	sessionID := uuid.New().String()
	familyID := uuid.New().String()

	accessToken, _, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role, sessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashRefreshToken(rawRefreshToken, s.cfg.RefreshTokenHMACKey)

	now := time.Now()
	meta := SessionMetadata{
		SessionID:        sessionID,
		FamilyID:         familyID,
		UserID:           user.ID,
		CurrentTokenHash: tokenHash,
		IssuedAt:         now,
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
		LastUsedAt:       now,
		IPAddress:        ip,
		UserAgent:        userAgent,
	}

	ctx := context.Background()
	if err := s.refreshRepo.CreateSession(ctx, meta, s.cfg.RefreshTokenTTL); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("create redis session: %w", err)
	}

	slog.Info("register_success", slog.Uint64("user_id", uint64(user.ID)), slog.String("session_id", sessionID))

	return dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(s.cfg.AccessTokenTTL.Seconds()),
		User:        user,
	}, rawRefreshToken, nil
}

func (s *AuthService) Login(req dto.AuthRequest, ip string, userAgent string) (dto.AuthResponse, string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil || user == nil {
		slog.Warn("login_failure_user_not_found", slog.String("email", cleanEmail))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_CREDENTIALS", "Invalid credentials")
	}

	if user.PasswordHash == nil {
		slog.Warn("login_failure_no_password", slog.Uint64("user_id", uint64(user.ID)))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_CREDENTIALS", "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		slog.Warn("login_failure_password_mismatch", slog.Uint64("user_id", uint64(user.ID)))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_CREDENTIALS", "Invalid credentials")
	}

	sessionID := uuid.New().String()
	familyID := uuid.New().String()

	accessToken, _, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role, sessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate access token: %w", err)
	}

	rawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashRefreshToken(rawRefreshToken, s.cfg.RefreshTokenHMACKey)

	now := time.Now()
	meta := SessionMetadata{
		SessionID:        sessionID,
		FamilyID:         familyID,
		UserID:           user.ID,
		CurrentTokenHash: tokenHash,
		IssuedAt:         now,
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
		LastUsedAt:       now,
		IPAddress:        ip,
		UserAgent:        userAgent,
	}

	ctx := context.Background()
	if err := s.refreshRepo.CreateSession(ctx, meta, s.cfg.RefreshTokenTTL); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("create redis session: %w", err)
	}

	slog.Info("login_success", slog.Uint64("user_id", uint64(user.ID)), slog.String("session_id", sessionID), slog.String("family_id", familyID))

	return dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(s.cfg.AccessTokenTTL.Seconds()),
		User:        *user,
	}, rawRefreshToken, nil
}

func (s *AuthService) Refresh(rawRefreshToken string, ip string, userAgent string) (dto.AuthResponse, string, error) {
	tokenHash := hashRefreshToken(rawRefreshToken, s.cfg.RefreshTokenHMACKey)

	newRawRefreshToken, err := generateRefreshToken()
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate new refresh token: %w", err)
	}

	newTokenHash := hashRefreshToken(newRawRefreshToken, s.cfg.RefreshTokenHMACKey)

	ctx := context.Background()
	rotRes, err := s.refreshRepo.RotateToken(ctx, tokenHash, newTokenHash, s.cfg.RefreshTokenTTL)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("rotate token: %w", err)
	}

	if rotRes.Status == "USED" {
		slog.Error("refresh_token_reuse_detected", slog.String("token_hash", tokenHash))
		_ = s.refreshRepo.RevokeSessionByTokenHash(ctx, tokenHash, "refresh_token_reuse")
		return dto.AuthResponse{}, "", exceptions.Unauthorized("REFRESH_TOKEN_REUSED", "Session is no longer valid. Please log in again.")
	}

	if rotRes.Status == "NOT_FOUND" {
		return dto.AuthResponse{}, "", exceptions.Unauthorized("REFRESH_TOKEN_INVALID", "Invalid refresh token")
	}

	var meta SessionMetadata
	if err := json.Unmarshal([]byte(rotRes.SessionData), &meta); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("unmarshal session data: %w", err)
	}

	if rotRes.Status == "REVOKED" || meta.RevokedAt != nil {
		slog.Warn("refresh_token_revoked_session", slog.String("session_id", meta.SessionID))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("SESSION_REVOKED", "Session has been revoked")
	}

	if time.Now().After(meta.ExpiresAt) {
		slog.Warn("refresh_token_expired", slog.String("session_id", meta.SessionID))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("REFRESH_TOKEN_EXPIRED", "Refresh token expired")
	}

	user, err := s.repo.FindByID(meta.UserID)
	if err != nil || user == nil {
		slog.Warn("refresh_user_not_found", slog.String("session_id", meta.SessionID))
		return dto.AuthResponse{}, "", exceptions.Unauthorized("INVALID_CREDENTIALS", "User no longer exists")
	}

	newAccessToken, _, err := s.tokenService.GenerateAccessToken(user.ID, user.Email, user.Role, meta.SessionID)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("generate access token: %w", err)
	}

	slog.Info("refresh_success", slog.Uint64("user_id", uint64(user.ID)), slog.String("session_id", meta.SessionID))

	return dto.AuthResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   int(s.cfg.AccessTokenTTL.Seconds()),
		User:        *user,
	}, newRawRefreshToken, nil
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := hashRefreshToken(rawRefreshToken, s.cfg.RefreshTokenHMACKey)
	ctx := context.Background()

	err := s.refreshRepo.RevokeSessionByTokenHash(ctx, tokenHash, "user_logout")
	slog.Info("logout", slog.String("token_hash", tokenHash))
	return err
}

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) (string, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(cleanEmail)
	if err != nil || user == nil {
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
	slog.Info("password_reset_success", slog.Uint64("user_id", uint64(userID)))

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
	if err != nil || user == nil {
		return "", exceptions.BadRequest("User not found")
	}

	if user.IsEmailVerified {
		return "", exceptions.BadRequest("Email is already verified")
	}

	return s.createVerificationToken(user.ID)
}

func (s *AuthService) GetCurrentUser(userID uint) (*types.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil || user == nil {
		return nil, exceptions.NotFound("User not found")
	}

	return user, nil
}

func (s *AuthService) DeleteAccount(userID uint, req dto.DeleteAccountRequest) error {
	user, err := s.repo.FindByID(userID)
	if err != nil || user == nil {
		return exceptions.NotFound("User not found")
	}

	if user.PasswordHash == nil {
		return exceptions.Unauthorized("INVALID_CREDENTIALS", "Invalid password confirmation")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return exceptions.Unauthorized("INVALID_CREDENTIALS", "Invalid password confirmation")
	}

	if err := s.repo.Delete(userID); err != nil {
		return fmt.Errorf("delete user account: %w", err)
	}

	slog.Info("account_deleted", slog.Uint64("user_id", uint64(userID)))

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
