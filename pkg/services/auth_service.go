package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fiber-boilerplate/pkg/mappers"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"fiber-boilerplate/pkg/entities"
	repository "fiber-boilerplate/pkg/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailAlreadyUsed    = errors.New("email is already registered")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrRateLimited         = errors.New("rate limited")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrOTPExpired          = errors.New("otp expired")
	ErrOTPAttemptsExceeded = errors.New("otp attempts exceeded")
	ErrSessionNotFound     = errors.New("session not found")
)

const (
	otpChallengePurposeLogin          = "login"
	otpChallengePurposeForgotPassword = "forgot_password"
)

type AuthService interface {
	Register(ctx context.Context, input entities.RegisterInput, meta entities.AuthClientMeta) (entities.AuthTokens, error)
	Login(ctx context.Context, input entities.LoginInput, meta entities.AuthClientMeta) (entities.OTPChallengeResult, error)
	ForgotPassword(ctx context.Context, input entities.ForgotPasswordInput, meta entities.AuthClientMeta) (entities.OTPChallengeResult, error)
	VerifyOTP(ctx context.Context, input entities.VerifyOTPInput, meta entities.AuthClientMeta) (entities.AuthTokens, error)
	ResetPassword(ctx context.Context, input entities.ResetPasswordInput, meta entities.AuthClientMeta) error
	Refresh(ctx context.Context, refreshToken string, meta entities.AuthClientMeta) (entities.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, accessToken string) (entities.AuthUser, error)
	ListSessions(ctx context.Context, accessToken string) ([]entities.AuthSession, error)
	RevokeSession(ctx context.Context, accessToken string, sessionID string) error
	RevokeAllSessions(ctx context.Context, accessToken string) error
}

type AuthSettings struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	BcryptCost      int
	RateLimitPerMin int
	OTPTTL          time.Duration
	OTPMaxAttempts  int
	DebugExposeOTP  bool
}

type authService struct {
	cfg         AuthSettings
	users       repository.UserRepository
	authSession repository.AuthSessionRepository
	otpRepo     repository.OTPRepository
	rateLimiter repository.RateLimitRepository
}

type accessTokenClaims struct {
	Email     string `json:"email"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type accessPrincipal struct {
	UserID    uint
	SessionID string
}

func NewAuthService(
	cfg AuthSettings,
	users repository.UserRepository,
	authSession repository.AuthSessionRepository,
	otpRepo repository.OTPRepository,
	rateLimiter repository.RateLimitRepository,
) AuthService {
	return &authService{
		cfg:         cfg,
		users:       users,
		authSession: authSession,
		otpRepo:     otpRepo,
		rateLimiter: rateLimiter,
	}
}

func (s *authService) Register(ctx context.Context, input entities.RegisterInput, meta entities.AuthClientMeta) (entities.AuthTokens, error) {
	email := normalizeEmail(input.Email)
	name := sanitizeName(input.Name)
	if email == "" || name == "" || strings.TrimSpace(input.Password) == "" {
		return entities.AuthTokens{}, ErrInvalidInput
	}

	if err := s.enforceRateLimit(ctx, "register", rateIdentifier(meta, email)); err != nil {
		return entities.AuthTokens{}, err
	}

	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return entities.AuthTokens{}, ErrEmailAlreadyUsed
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return entities.AuthTokens{}, fmt.Errorf("find user for register: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.cfg.BcryptCost)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("hash password: %w", err)
	}

	user := &entities.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(passwordHash),
	}
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return entities.AuthTokens{}, ErrEmailAlreadyUsed
		}
		return entities.AuthTokens{}, fmt.Errorf("create user: %w", err)
	}

	result, err := s.issueTokens(ctx, user, meta)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("issue register tokens: %w", err)
	}

	return result, nil
}

func (s *authService) Login(ctx context.Context, input entities.LoginInput, meta entities.AuthClientMeta) (entities.OTPChallengeResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" || strings.TrimSpace(input.Password) == "" {
		return entities.OTPChallengeResult{}, ErrInvalidInput
	}

	if err := s.enforceRateLimit(ctx, "login", rateIdentifier(meta, email)); err != nil {
		return entities.OTPChallengeResult{}, err
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return entities.OTPChallengeResult{}, ErrInvalidCredentials
		}
		return entities.OTPChallengeResult{}, fmt.Errorf("find user for login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return entities.OTPChallengeResult{}, ErrInvalidCredentials
	}

	return s.issueOTPChallenge(ctx, user.ID, otpChallengePurposeLogin)
}

func (s *authService) ForgotPassword(ctx context.Context, input entities.ForgotPasswordInput, meta entities.AuthClientMeta) (entities.OTPChallengeResult, error) {
	email := normalizeEmail(input.Email)
	if email == "" {
		return entities.OTPChallengeResult{}, ErrInvalidInput
	}

	if err := s.enforceRateLimit(ctx, "forgot_password", rateIdentifier(meta, email)); err != nil {
		return entities.OTPChallengeResult{}, err
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return entities.OTPChallengeResult{}, fmt.Errorf("find user for forgot password: %w", err)
	}
	if err == nil {
		result, err := s.issueOTPChallenge(ctx, user.ID, otpChallengePurposeForgotPassword)
		if err != nil {
			return entities.OTPChallengeResult{}, fmt.Errorf("issue forgot password challenge: %w", err)
		}
		return result, nil
	}

	return s.newOTPChallengeResult()
}

func (s *authService) VerifyOTP(ctx context.Context, input entities.VerifyOTPInput, meta entities.AuthClientMeta) (entities.AuthTokens, error) {
	challengeID := strings.TrimSpace(input.ChallengeID)
	otpCode := strings.TrimSpace(input.OTP)
	if challengeID == "" || otpCode == "" {
		return entities.AuthTokens{}, ErrInvalidInput
	}

	if err := s.enforceRateLimit(ctx, "otp_verify", rateIdentifier(meta, challengeID)); err != nil {
		return entities.AuthTokens{}, err
	}

	userID, err := s.otpRepo.VerifyChallenge(ctx, challengeID, mappers.HashToken(otpCode), otpChallengePurposeLogin)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOTPChallengeNotFound):
			return entities.AuthTokens{}, ErrOTPExpired
		case errors.Is(err, repository.ErrOTPInvalidCode):
			return entities.AuthTokens{}, ErrInvalidOTP
		case errors.Is(err, repository.ErrOTPTooManyAttempts):
			return entities.AuthTokens{}, ErrOTPAttemptsExceeded
		default:
			return entities.AuthTokens{}, fmt.Errorf("verify otp challenge: %w", err)
		}
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return entities.AuthTokens{}, ErrInvalidCredentials
		}
		return entities.AuthTokens{}, fmt.Errorf("find user for verified otp: %w", err)
	}

	result, err := s.issueTokens(ctx, user, meta)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("issue tokens after otp verification: %w", err)
	}

	return result, nil
}

func (s *authService) ResetPassword(ctx context.Context, input entities.ResetPasswordInput, meta entities.AuthClientMeta) error {
	challengeID := strings.TrimSpace(input.ChallengeID)
	otpCode := strings.TrimSpace(input.OTP)
	newPassword := strings.TrimSpace(input.NewPassword)
	if challengeID == "" || otpCode == "" || newPassword == "" {
		return ErrInvalidInput
	}

	if err := s.enforceRateLimit(ctx, "reset_password", rateIdentifier(meta, challengeID)); err != nil {
		return err
	}

	userID, err := s.otpRepo.VerifyChallenge(ctx, challengeID, mappers.HashToken(otpCode), otpChallengePurposeForgotPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOTPChallengeNotFound):
			return ErrOTPExpired
		case errors.Is(err, repository.ErrOTPInvalidCode):
			return ErrInvalidOTP
		case errors.Is(err, repository.ErrOTPTooManyAttempts):
			return ErrOTPAttemptsExceeded
		default:
			return fmt.Errorf("verify reset password challenge: %w", err)
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash reset password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return fmt.Errorf("update password after reset: %w", err)
	}
	if err := s.authSession.RevokeAllSessions(ctx, userID); err != nil {
		return fmt.Errorf("revoke sessions after password reset: %w", err)
	}

	return nil
}

func (s *authService) issueOTPChallenge(ctx context.Context, userID uint, purpose string) (entities.OTPChallengeResult, error) {
	challengeID, err := generateOpaqueToken(18)
	if err != nil {
		return entities.OTPChallengeResult{}, fmt.Errorf("generate otp challenge id: %w", err)
	}
	otpCode, err := generateOTPCode()
	if err != nil {
		return entities.OTPChallengeResult{}, fmt.Errorf("generate otp code: %w", err)
	}

	now := time.Now().UTC()
	challenge := entities.OTPChallenge{
		ChallengeID: challengeID,
		Purpose:     purpose,
		UserID:      userID,
		CodeHash:    mappers.HashToken(otpCode),
		Attempts:    0,
		MaxAttempts: s.cfg.OTPMaxAttempts,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.cfg.OTPTTL),
	}
	if err := s.otpRepo.StoreChallenge(ctx, challenge, s.cfg.OTPTTL); err != nil {
		return entities.OTPChallengeResult{}, fmt.Errorf("store otp challenge: %w", err)
	}

	return s.newOTPChallengeResultWithCode(challengeID, otpCode), nil
}

func (s *authService) newOTPChallengeResult() (entities.OTPChallengeResult, error) {
	challengeID, err := generateOpaqueToken(18)
	if err != nil {
		return entities.OTPChallengeResult{}, fmt.Errorf("generate otp challenge id: %w", err)
	}
	otpCode := ""
	if s.cfg.DebugExposeOTP {
		var err error
		otpCode, err = generateOTPCode()
		if err != nil {
			return entities.OTPChallengeResult{}, fmt.Errorf("generate otp code: %w", err)
		}
	}

	return s.newOTPChallengeResultWithCode(challengeID, otpCode), nil
}

func (s *authService) newOTPChallengeResultWithCode(challengeID string, otpCode string) entities.OTPChallengeResult {
	result := entities.OTPChallengeResult{
		ChallengeID:  challengeID,
		ExpiresInSec: int64(s.cfg.OTPTTL.Seconds()),
	}
	if s.cfg.DebugExposeOTP {
		result.OTP = otpCode
	}

	return result
}

func (s *authService) Refresh(ctx context.Context, refreshToken string, meta entities.AuthClientMeta) (entities.AuthTokens, error) {
	trimmed := strings.TrimSpace(refreshToken)
	if trimmed == "" {
		return entities.AuthTokens{}, ErrInvalidRefreshToken
	}

	session, err := s.authSession.ConsumeRefreshToken(ctx, mappers.HashToken(trimmed))
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return entities.AuthTokens{}, ErrInvalidRefreshToken
		}
		return entities.AuthTokens{}, fmt.Errorf("consume refresh token: %w", err)
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return entities.AuthTokens{}, ErrInvalidRefreshToken
		}
		return entities.AuthTokens{}, fmt.Errorf("find user for refresh: %w", err)
	}

	if strings.TrimSpace(meta.IPAddress) == "" {
		meta.IPAddress = session.IPAddress
	}
	if strings.TrimSpace(meta.UserAgent) == "" {
		meta.UserAgent = session.UserAgent
	}

	result, err := s.issueTokens(ctx, user, meta)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("issue refresh tokens: %w", err)
	}

	return result, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	trimmed := strings.TrimSpace(refreshToken)
	if trimmed == "" {
		return ErrInvalidRefreshToken
	}

	err := s.authSession.DeleteRefreshToken(ctx, mappers.HashToken(trimmed))
	if err != nil && !errors.Is(err, repository.ErrSessionNotFound) {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (s *authService) Me(ctx context.Context, accessToken string) (entities.AuthUser, error) {
	principal, err := s.authenticateAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, ErrInvalidAccessToken) {
			return entities.AuthUser{}, ErrInvalidAccessToken
		}
		return entities.AuthUser{}, fmt.Errorf("authenticate access token for me: %w", err)
	}

	user, err := s.users.FindByID(ctx, principal.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return entities.AuthUser{}, ErrInvalidAccessToken
		}
		return entities.AuthUser{}, fmt.Errorf("find user for me: %w", err)
	}

	return entities.AuthUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *authService) ListSessions(ctx context.Context, accessToken string) ([]entities.AuthSession, error) {
	principal, err := s.authenticateAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, ErrInvalidAccessToken) {
			return nil, ErrInvalidAccessToken
		}
		return nil, fmt.Errorf("authenticate access token for list sessions: %w", err)
	}

	sessions, err := s.authSession.ListUserSessions(ctx, principal.UserID)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}

	result := make([]entities.AuthSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, mappers.ToAuthSessionEntity(session, principal.SessionID))
	}

	return result, nil
}

func (s *authService) RevokeSession(ctx context.Context, accessToken string, sessionID string) error {
	principal, err := s.authenticateAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, ErrInvalidAccessToken) {
			return ErrInvalidAccessToken
		}
		return fmt.Errorf("authenticate access token for revoke session: %w", err)
	}

	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ErrInvalidInput
	}

	if err := s.authSession.RevokeSession(ctx, principal.UserID, sessionID); err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func (s *authService) RevokeAllSessions(ctx context.Context, accessToken string) error {
	principal, err := s.authenticateAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, ErrInvalidAccessToken) {
			return ErrInvalidAccessToken
		}
		return fmt.Errorf("authenticate access token for revoke all sessions: %w", err)
	}

	if err := s.authSession.RevokeAllSessions(ctx, principal.UserID); err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}

	return nil
}

func (s *authService) issueTokens(ctx context.Context, user *entities.User, meta entities.AuthClientMeta) (entities.AuthTokens, error) {
	now := time.Now().UTC()

	sessionID, err := generateOpaqueToken(18)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("generate session id: %w", err)
	}

	accessToken, expiresInSec, err := s.issueAccessToken(user, sessionID)
	if err != nil {
		return entities.AuthTokens{}, err
	}

	refreshToken, err := generateOpaqueToken(48)
	if err != nil {
		return entities.AuthTokens{}, fmt.Errorf("generate refresh token: %w", err)
	}

	session := entities.RefreshSession{
		SessionID:        sessionID,
		UserID:           user.ID,
		RefreshTokenHash: mappers.HashToken(refreshToken),
		UserAgent:        meta.UserAgent,
		IPAddress:        meta.IPAddress,
		CreatedAt:        now,
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
	}
	if err := s.authSession.StoreRefreshSession(ctx, session, s.cfg.RefreshTokenTTL); err != nil {
		return entities.AuthTokens{}, fmt.Errorf("store refresh session: %w", err)
	}

	return entities.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresInSec: expiresInSec,
		SessionID:    sessionID,
		User: entities.AuthUser{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *authService) issueAccessToken(user *entities.User, sessionID string) (string, int64, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)

	claims := accessTokenClaims{
		Email:     user.Email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return signed, int64(s.cfg.AccessTokenTTL.Seconds()), nil
}

func (s *authService) parseAccessToken(accessToken string) (accessPrincipal, error) {
	claims := &accessTokenClaims{}
	token, err := jwt.ParseWithClaims(
		strings.TrimSpace(accessToken),
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidAccessToken
			}
			return []byte(s.cfg.JWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil || !token.Valid {
		return accessPrincipal{}, ErrInvalidAccessToken
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return accessPrincipal{}, ErrInvalidAccessToken
	}
	if strings.TrimSpace(claims.SessionID) == "" {
		return accessPrincipal{}, ErrInvalidAccessToken
	}

	return accessPrincipal{
		UserID:    uint(userID),
		SessionID: claims.SessionID,
	}, nil
}

func (s *authService) authenticateAccessToken(ctx context.Context, accessToken string) (accessPrincipal, error) {
	principal, err := s.parseAccessToken(accessToken)
	if err != nil {
		return accessPrincipal{}, err
	}

	hasSession, err := s.authSession.HasSession(ctx, principal.UserID, principal.SessionID)
	if err != nil {
		return accessPrincipal{}, fmt.Errorf("validate access token session: %w", err)
	}
	if !hasSession {
		return accessPrincipal{}, ErrInvalidAccessToken
	}

	return principal, nil
}

func (s *authService) enforceRateLimit(ctx context.Context, action string, identifier string) error {
	key := "auth:rate:" + strings.TrimSpace(action) + ":" + mappers.HashToken(strings.TrimSpace(identifier))
	count, _, err := s.rateLimiter.Hit(ctx, key, time.Minute)
	if err != nil {
		return fmt.Errorf("apply rate limit: %w", err)
	}

	if count > int64(s.cfg.RateLimitPerMin) {
		return ErrRateLimited
	}

	return nil
}

func rateIdentifier(meta entities.AuthClientMeta, value string) string {
	ip := strings.TrimSpace(meta.IPAddress)
	if ip == "" {
		ip = "unknown"
	}

	normalizedValue := strings.TrimSpace(strings.ToLower(value))
	return ip + "|" + normalizedValue
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sanitizeName(name string) string {
	return strings.TrimSpace(name)
}

func generateOpaqueToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", errors.New("byteLength must be > 0")
	}

	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate random otp number: %w", err)
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}
