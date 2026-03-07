package repository

import (
	"context"
	"errors"
	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/mappers"
	"fiber-boilerplate/pkg/models"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSessionNotFound = errors.New("auth session not found")

type AuthSessionRepository interface {
	StoreRefreshSession(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error
	ConsumeRefreshToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	ListUserSessions(ctx context.Context, userID uint) ([]entities.RefreshSession, error)
	RevokeSession(ctx context.Context, userID uint, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID uint) error
}

type authSessionRepository struct {
	db *gorm.DB
}

func NewAuthSessionRepository(db *gorm.DB) AuthSessionRepository {
	return &authSessionRepository{
		db: db,
	}
}

func (r *authSessionRepository) StoreRefreshSession(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error {
	modelSession := mappers.ToRefreshSessionModel(session)
	if err := r.db.WithContext(ctx).Create(&modelSession).Error; err != nil {
		return fmt.Errorf("store refresh session: %w", err)
	}

	return nil
}

func (r *authSessionRepository) ConsumeRefreshToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error) {
	var session models.AuthSession
	now := time.Now().UTC()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refresh_token_hash = ?", strings.TrimSpace(tokenHash)).
			First(&session).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSessionNotFound
			}
			return fmt.Errorf("find refresh session by token hash: %w", err)
		}

		if !session.ExpiresAt.After(now) {
			if err := tx.Delete(&session).Error; err != nil {
				return fmt.Errorf("delete expired refresh session: %w", err)
			}
			return ErrSessionNotFound
		}

		if err := tx.Delete(&session).Error; err != nil {
			return fmt.Errorf("delete consumed refresh session: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return entities.RefreshSession{}, ErrSessionNotFound
		}
		return entities.RefreshSession{}, fmt.Errorf("consume refresh token: %w", err)
	}

	return mappers.ToRefreshSessionEntity(session), nil
}

func (r *authSessionRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	var session models.AuthSession

	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", strings.TrimSpace(tokenHash)).
		First(&session).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("find refresh session for delete: %w", err)
	}

	if err := r.db.WithContext(ctx).Delete(&session).Error; err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (r *authSessionRepository) ListUserSessions(ctx context.Context, userID uint) ([]entities.RefreshSession, error) {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at <= ?", userID, now).
		Delete(&models.AuthSession{}).
		Error; err != nil {
		return nil, fmt.Errorf("cleanup expired user sessions: %w", err)
	}

	var modelSessions []models.AuthSession
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&modelSessions).
		Error; err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}

	sessions := make([]entities.RefreshSession, 0, len(modelSessions))
	for _, session := range modelSessions {
		sessions = append(sessions, mappers.ToRefreshSessionEntity(session))
	}

	return sessions, nil
}

func (r *authSessionRepository) RevokeSession(ctx context.Context, userID uint, sessionID string) error {
	result := r.db.WithContext(ctx).
		Where("session_id = ? AND user_id = ?", strings.TrimSpace(sessionID), userID).
		Delete(&models.AuthSession{})
	if result.Error != nil {
		return fmt.Errorf("revoke session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *authSessionRepository) RevokeAllSessions(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.AuthSession{}).
		Error; err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}

	return nil
}
