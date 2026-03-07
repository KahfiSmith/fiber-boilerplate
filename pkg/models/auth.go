package models

import "time"

type AuthSession struct {
	SessionID        string    `gorm:"primaryKey;size:64" json:"session_id"`
	UserID           uint      `gorm:"not null;index" json:"user_id"`
	RefreshTokenHash string    `gorm:"size:64;not null;uniqueIndex" json:"refresh_token_hash"`
	UserAgent        string    `gorm:"size:512;not null;default:''" json:"user_agent"`
	IPAddress        string    `gorm:"size:128;not null;default:''" json:"ip_address"`
	CreatedAt        time.Time `gorm:"not null" json:"created_at"`
	ExpiresAt        time.Time `gorm:"not null;index" json:"expires_at"`
}

type OTPChallenge struct {
	ChallengeID string    `gorm:"primaryKey;size:64" json:"challenge_id"`
	Purpose     string    `gorm:"size:64;not null" json:"purpose"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	CodeHash    string    `gorm:"size:64;not null" json:"code_hash"`
	Attempts    int       `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts int       `gorm:"not null" json:"max_attempts"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
}

type AuthRateLimit struct {
	RateKey   string    `gorm:"primaryKey;size:255" json:"rate_key"`
	Count     int64     `gorm:"not null" json:"count"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}

func (OTPChallenge) TableName() string {
	return "otp_challenges"
}

func (AuthRateLimit) TableName() string {
	return "auth_rate_limits"
}
