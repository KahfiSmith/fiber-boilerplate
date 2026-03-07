package entities

import "time"

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type ForgotPasswordInput struct {
	Email string
}

type VerifyOTPInput struct {
	ChallengeID string
	OTP         string
}

type ResetPasswordInput struct {
	ChallengeID string
	OTP         string
	NewPassword string
}

type AuthClientMeta struct {
	IPAddress string
	UserAgent string
}

type AuthUser struct {
	ID    uint
	Name  string
	Email string
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresInSec int64
	SessionID    string
	User         AuthUser
}

type OTPChallengeResult struct {
	ChallengeID  string
	ExpiresInSec int64
	OTP          string
}

type AuthSession struct {
	SessionID string
	UserAgent string
	IPAddress string
	CreatedAt time.Time
	ExpiresAt time.Time
	Current   bool
}

type RefreshSession struct {
	SessionID        string
	UserID           uint
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	CreatedAt        time.Time
	ExpiresAt        time.Time
}

type OTPChallenge struct {
	ChallengeID string
	Purpose     string
	UserID      uint
	CodeHash    string
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	ExpiresAt   time.Time
}
