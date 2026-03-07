package mappers

import (
	"crypto/sha256"
	"encoding/hex"
	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/models"
	"strings"
)

func ToOTPChallengeModel(challenge entities.OTPChallenge) models.OTPChallenge {
	return models.OTPChallenge{
		ChallengeID: challenge.ChallengeID,
		Purpose:     challenge.Purpose,
		UserID:      challenge.UserID,
		CodeHash:    challenge.CodeHash,
		Attempts:    challenge.Attempts,
		MaxAttempts: challenge.MaxAttempts,
		CreatedAt:   challenge.CreatedAt,
		ExpiresAt:   challenge.ExpiresAt,
	}
}

func ToOTPChallengeEntity(challenge models.OTPChallenge) entities.OTPChallenge {
	return entities.OTPChallenge{
		ChallengeID: challenge.ChallengeID,
		Purpose:     challenge.Purpose,
		UserID:      challenge.UserID,
		CodeHash:    challenge.CodeHash,
		Attempts:    challenge.Attempts,
		MaxAttempts: challenge.MaxAttempts,
		CreatedAt:   challenge.CreatedAt,
		ExpiresAt:   challenge.ExpiresAt,
	}
}

func ToRefreshSessionModel(session entities.RefreshSession) models.AuthSession {
	return models.AuthSession{
		SessionID:        session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		UserAgent:        sanitizeUserAgent(session.UserAgent),
		IPAddress:        sanitizeIPAddress(session.IPAddress),
		CreatedAt:        session.CreatedAt,
		ExpiresAt:        session.ExpiresAt,
	}
}

func ToRefreshSessionEntity(session models.AuthSession) entities.RefreshSession {
	return entities.RefreshSession{
		SessionID:        session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		UserAgent:        session.UserAgent,
		IPAddress:        session.IPAddress,
		CreatedAt:        session.CreatedAt,
		ExpiresAt:        session.ExpiresAt,
	}
}

func ToAuthSessionEntity(session entities.RefreshSession, currentSessionID string) entities.AuthSession {
	return entities.AuthSession{
		SessionID: session.SessionID,
		UserAgent: session.UserAgent,
		IPAddress: session.IPAddress,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
		Current:   session.SessionID == currentSessionID,
	}
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sanitizeUserAgent(userAgent string) string {
	userAgent = strings.TrimSpace(userAgent)
	if len(userAgent) > 512 {
		return userAgent[:512]
	}
	return userAgent
}

func sanitizeIPAddress(ipAddress string) string {
	ipAddress = strings.TrimSpace(ipAddress)
	if len(ipAddress) > 128 {
		return ipAddress[:128]
	}
	return ipAddress
}
