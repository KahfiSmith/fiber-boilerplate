package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/mappers"
	repository "fiber-boilerplate/pkg/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryStub struct {
	createFn         func(ctx context.Context, user *entities.User) error
	findByEmailFn    func(ctx context.Context, email string) (*entities.User, error)
	findByIDFn       func(ctx context.Context, id uint) (*entities.User, error)
	updatePasswordFn func(ctx context.Context, id uint, passwordHash string) error
}

func (s *userRepositoryStub) Create(ctx context.Context, user *entities.User) error {
	if s.createFn != nil {
		return s.createFn(ctx, user)
	}
	return nil
}

func (s *userRepositoryStub) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if s.findByEmailFn != nil {
		return s.findByEmailFn(ctx, email)
	}
	return nil, repository.ErrUserNotFound
}

func (s *userRepositoryStub) FindByID(ctx context.Context, id uint) (*entities.User, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, repository.ErrUserNotFound
}

func (s *userRepositoryStub) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	if s.updatePasswordFn != nil {
		return s.updatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}

type authSessionRepositoryStub struct {
	storeRefreshSessionFn func(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error
	consumeRefreshTokenFn func(ctx context.Context, tokenHash string) (entities.RefreshSession, error)
	deleteRefreshTokenFn  func(ctx context.Context, tokenHash string) error
	hasSessionFn         func(ctx context.Context, userID uint, sessionID string) (bool, error)
	listUserSessionsFn    func(ctx context.Context, userID uint) ([]entities.RefreshSession, error)
	revokeSessionFn       func(ctx context.Context, userID uint, sessionID string) error
	revokeAllSessionsFn   func(ctx context.Context, userID uint) error
}

func (s *authSessionRepositoryStub) StoreRefreshSession(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error {
	if s.storeRefreshSessionFn != nil {
		return s.storeRefreshSessionFn(ctx, session, ttl)
	}
	return nil
}

func (s *authSessionRepositoryStub) ConsumeRefreshToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error) {
	if s.consumeRefreshTokenFn != nil {
		return s.consumeRefreshTokenFn(ctx, tokenHash)
	}
	return entities.RefreshSession{}, repository.ErrSessionNotFound
}

func (s *authSessionRepositoryStub) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	if s.deleteRefreshTokenFn != nil {
		return s.deleteRefreshTokenFn(ctx, tokenHash)
	}
	return nil
}

func (s *authSessionRepositoryStub) HasSession(ctx context.Context, userID uint, sessionID string) (bool, error) {
	if s.hasSessionFn != nil {
		return s.hasSessionFn(ctx, userID, sessionID)
	}
	return true, nil
}

func (s *authSessionRepositoryStub) ListUserSessions(ctx context.Context, userID uint) ([]entities.RefreshSession, error) {
	if s.listUserSessionsFn != nil {
		return s.listUserSessionsFn(ctx, userID)
	}
	return nil, nil
}

func (s *authSessionRepositoryStub) RevokeSession(ctx context.Context, userID uint, sessionID string) error {
	if s.revokeSessionFn != nil {
		return s.revokeSessionFn(ctx, userID, sessionID)
	}
	return nil
}

func (s *authSessionRepositoryStub) RevokeAllSessions(ctx context.Context, userID uint) error {
	if s.revokeAllSessionsFn != nil {
		return s.revokeAllSessionsFn(ctx, userID)
	}
	return nil
}

type otpRepositoryStub struct {
	storeChallengeFn  func(ctx context.Context, challenge entities.OTPChallenge, ttl time.Duration) error
	verifyChallengeFn func(ctx context.Context, challengeID string, codeHash string, purpose string) (uint, error)
}

func (s *otpRepositoryStub) StoreChallenge(ctx context.Context, challenge entities.OTPChallenge, ttl time.Duration) error {
	if s.storeChallengeFn != nil {
		return s.storeChallengeFn(ctx, challenge, ttl)
	}
	return nil
}

func (s *otpRepositoryStub) VerifyChallenge(ctx context.Context, challengeID string, codeHash string, purpose string) (uint, error) {
	if s.verifyChallengeFn != nil {
		return s.verifyChallengeFn(ctx, challengeID, codeHash, purpose)
	}
	return 0, repository.ErrOTPChallengeNotFound
}

type rateLimitRepositoryStub struct {
	hitFn func(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error)
}

func (s *rateLimitRepositoryStub) Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if s.hitFn != nil {
		return s.hitFn(ctx, key, window)
	}
	return 1, window, nil
}

func defaultAuthSettings() AuthSettings {
	return AuthSettings{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		BcryptCost:      bcrypt.MinCost,
		RateLimitPerMin: 5,
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		DebugExposeOTP:  false,
	}
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var createdUser entities.User
	var storedSession entities.RefreshSession

	users := &userRepositoryStub{
		findByEmailFn: func(_ context.Context, email string) (*entities.User, error) {
			assert.Equal(t, "user@example.com", email)
			return nil, repository.ErrUserNotFound
		},
		createFn: func(_ context.Context, user *entities.User) error {
			createdUser = *user
			user.ID = 42
			return nil
		},
	}
	sessions := &authSessionRepositoryStub{
		storeRefreshSessionFn: func(_ context.Context, session entities.RefreshSession, ttl time.Duration) error {
			storedSession = session
			assert.Equal(t, 24*time.Hour, ttl)
			return nil
		},
	}

	svc := NewAuthService(defaultAuthSettings(), users, sessions, &otpRepositoryStub{}, &rateLimitRepositoryStub{})
	authSvc := svc.(*authService)

	result, err := svc.Register(ctx, entities.RegisterInput{
		Name:     "  Kahfi  ",
		Email:    "  USER@Example.com ",
		Password: "Secret123",
	}, entities.AuthClientMeta{
		IPAddress: "127.0.0.1",
		UserAgent: "unit-test",
	})

	require.NoError(t, err)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, int64(900), result.ExpiresInSec)
	assert.Equal(t, uint(42), result.User.ID)
	assert.Equal(t, "Kahfi", result.User.Name)
	assert.Equal(t, "user@example.com", result.User.Email)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEmpty(t, result.SessionID)

	assert.Equal(t, "Kahfi", createdUser.Name)
	assert.Equal(t, "user@example.com", createdUser.Email)
	assert.NotEqual(t, "Secret123", createdUser.PasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(createdUser.PasswordHash), []byte("Secret123")))

	assert.Equal(t, uint(42), storedSession.UserID)
	assert.Equal(t, "127.0.0.1", storedSession.IPAddress)
	assert.Equal(t, "unit-test", storedSession.UserAgent)
	assert.Equal(t, result.SessionID, storedSession.SessionID)

	principal, err := authSvc.parseAccessToken(result.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, uint(42), principal.UserID)
	assert.Equal(t, result.SessionID, principal.SessionID)
}

func TestAuthServiceLoginUserNotFoundReturnsInvalidCredentials(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{
		findByEmailFn: func(_ context.Context, email string) (*entities.User, error) {
			assert.Equal(t, "missing@example.com", email)
			return nil, repository.ErrUserNotFound
		},
	}, &authSessionRepositoryStub{}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})

	_, err := svc.Login(context.Background(), entities.LoginInput{
		Email:    " Missing@Example.com ",
		Password: "Secret123",
	}, entities.AuthClientMeta{IPAddress: "127.0.0.1"})

	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthServiceForgotPasswordUnknownUserReturnsOpaqueChallenge(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{
		findByEmailFn: func(context.Context, string) (*entities.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}, &authSessionRepositoryStub{}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})

	result, err := svc.ForgotPassword(context.Background(), entities.ForgotPasswordInput{
		Email: "nobody@example.com",
	}, entities.AuthClientMeta{IPAddress: "127.0.0.1"})

	require.NoError(t, err)
	assert.NotEmpty(t, result.ChallengeID)
	assert.Equal(t, int64(300), result.ExpiresInSec)
	assert.Empty(t, result.OTP)
}

func TestAuthServiceVerifyOTPMapsInvalidCode(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{}, &authSessionRepositoryStub{}, &otpRepositoryStub{
		verifyChallengeFn: func(_ context.Context, challengeID string, codeHash string, purpose string) (uint, error) {
			assert.Equal(t, "challenge-1", challengeID)
			assert.Equal(t, otpChallengePurposeLogin, purpose)
			assert.Equal(t, mappers.HashToken("123456"), codeHash)
			return 0, repository.ErrOTPInvalidCode
		},
	}, &rateLimitRepositoryStub{})

	_, err := svc.VerifyOTP(context.Background(), entities.VerifyOTPInput{
		ChallengeID: "challenge-1",
		OTP:         "123456",
	}, entities.AuthClientMeta{IPAddress: "127.0.0.1"})

	require.ErrorIs(t, err, ErrInvalidOTP)
}

func TestAuthServiceResetPasswordRevokesSessions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var updatedPasswordHash string
	var revokedUserID uint

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{
		updatePasswordFn: func(_ context.Context, id uint, passwordHash string) error {
			assert.Equal(t, uint(7), id)
			updatedPasswordHash = passwordHash
			return nil
		},
	}, &authSessionRepositoryStub{
		revokeAllSessionsFn: func(_ context.Context, userID uint) error {
			revokedUserID = userID
			return nil
		},
	}, &otpRepositoryStub{
		verifyChallengeFn: func(_ context.Context, challengeID string, codeHash string, purpose string) (uint, error) {
			assert.Equal(t, "challenge-reset", challengeID)
			assert.Equal(t, otpChallengePurposeForgotPassword, purpose)
			assert.Equal(t, mappers.HashToken("654321"), codeHash)
			return 7, nil
		},
	}, &rateLimitRepositoryStub{})

	err := svc.ResetPassword(ctx, entities.ResetPasswordInput{
		ChallengeID: "challenge-reset",
		OTP:         "654321",
		NewPassword: "NewSecret123",
	}, entities.AuthClientMeta{IPAddress: "127.0.0.1"})

	require.NoError(t, err)
	assert.Equal(t, uint(7), revokedUserID)
	assert.NotEmpty(t, updatedPasswordHash)
	assert.NotEqual(t, "NewSecret123", updatedPasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(updatedPasswordHash), []byte("NewSecret123")))
}

func TestAuthServiceRefreshUsesSessionMetadataWhenMetaMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var consumedTokenHash string
	var storedSession entities.RefreshSession

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{
		findByIDFn: func(_ context.Context, id uint) (*entities.User, error) {
			assert.Equal(t, uint(99), id)
			return &entities.User{
				ID:    99,
				Name:  "Kahfi",
				Email: "kahfi@example.com",
			}, nil
		},
	}, &authSessionRepositoryStub{
		consumeRefreshTokenFn: func(_ context.Context, tokenHash string) (entities.RefreshSession, error) {
			consumedTokenHash = tokenHash
			return entities.RefreshSession{
				SessionID: "old-session",
				UserID:    99,
				UserAgent: "saved-agent",
				IPAddress: "10.0.0.5",
			}, nil
		},
		storeRefreshSessionFn: func(_ context.Context, session entities.RefreshSession, ttl time.Duration) error {
			storedSession = session
			return nil
		},
	}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})

	result, err := svc.Refresh(ctx, " refresh-token ", entities.AuthClientMeta{})

	require.NoError(t, err)
	assert.Equal(t, mappers.HashToken("refresh-token"), consumedTokenHash)
	assert.Equal(t, "saved-agent", storedSession.UserAgent)
	assert.Equal(t, "10.0.0.5", storedSession.IPAddress)
	assert.Equal(t, uint(99), result.User.ID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestAuthServiceListSessionsMarksCurrentSession(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{}, &authSessionRepositoryStub{
		hasSessionFn: func(_ context.Context, userID uint, sessionID string) (bool, error) {
			assert.Equal(t, uint(55), userID)
			assert.Equal(t, "current-session", sessionID)
			return true, nil
		},
		listUserSessionsFn: func(_ context.Context, userID uint) ([]entities.RefreshSession, error) {
			assert.Equal(t, uint(55), userID)
			now := time.Now().UTC()
			return []entities.RefreshSession{
				{SessionID: "current-session", UserID: 55, CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
				{SessionID: "other-session", UserID: 55, CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
			}, nil
		},
	}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})
	authSvc := svc.(*authService)

	token, _, err := authSvc.issueAccessToken(&entities.User{
		ID:    55,
		Email: "kahfi@example.com",
	}, "current-session")
	require.NoError(t, err)

	sessions, err := svc.ListSessions(context.Background(), token)
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	assert.True(t, sessions[0].Current)
	assert.False(t, sessions[1].Current)
}

func TestAuthServiceMeReturnsInvalidAccessTokenWhenSessionMissing(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{}, &authSessionRepositoryStub{
		hasSessionFn: func(_ context.Context, userID uint, sessionID string) (bool, error) {
			assert.Equal(t, uint(11), userID)
			assert.Equal(t, "revoked-session", sessionID)
			return false, nil
		},
	}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})
	authSvc := svc.(*authService)

	token, _, err := authSvc.issueAccessToken(&entities.User{
		ID:    11,
		Email: "kahfi@example.com",
	}, "revoked-session")
	require.NoError(t, err)

	_, err = svc.Me(context.Background(), token)
	require.ErrorIs(t, err, ErrInvalidAccessToken)
}

func TestAuthServiceMeReturnsWrappedErrorWhenSessionValidationFails(t *testing.T) {
	t.Parallel()

	svc := NewAuthService(defaultAuthSettings(), &userRepositoryStub{}, &authSessionRepositoryStub{
		hasSessionFn: func(_ context.Context, userID uint, sessionID string) (bool, error) {
			assert.Equal(t, uint(12), userID)
			assert.Equal(t, "session-check-error", sessionID)
			return false, errors.New("redis unavailable")
		},
	}, &otpRepositoryStub{}, &rateLimitRepositoryStub{})
	authSvc := svc.(*authService)

	token, _, err := authSvc.issueAccessToken(&entities.User{
		ID:    12,
		Email: "kahfi@example.com",
	}, "session-check-error")
	require.NoError(t, err)

	_, err = svc.Me(context.Background(), token)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrInvalidAccessToken)
	assert.ErrorContains(t, err, "authenticate access token for me")
}

func TestRateIdentifierFallsBackToUnknownAndNormalizesValue(t *testing.T) {
	t.Parallel()

	identifier := rateIdentifier(entities.AuthClientMeta{}, " USER@example.com ")
	assert.Equal(t, "unknown|user@example.com", identifier)
}
