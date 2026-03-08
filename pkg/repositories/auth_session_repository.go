package repository

import (
	"context"
	"errors"
	"fiber-boilerplate/pkg/entities"
	"time"
)

var ErrSessionNotFound = errors.New("auth session not found")

type AuthSessionRepository interface {
	StoreRefreshSession(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error
	ConsumeRefreshToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	HasSession(ctx context.Context, userID uint, sessionID string) (bool, error)
	ListUserSessions(ctx context.Context, userID uint) ([]entities.RefreshSession, error)
	RevokeSession(ctx context.Context, userID uint, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID uint) error
}
