package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type rateLimitRepository struct {
	db *gorm.DB
}

func NewRateLimitRepository(db *gorm.DB) RateLimitRepository {
	return &rateLimitRepository{
		db: db,
	}
}

func (r *rateLimitRepository) Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	type rateLimitState struct {
		Count     int64
		ExpiresAt time.Time
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return 0, 0, fmt.Errorf("rate limit key must not be empty")
	}

	var state rateLimitState
	expiresAt := time.Now().UTC().Add(window)
	query := `
INSERT INTO auth_rate_limits (rate_key, count, expires_at, created_at, updated_at)
VALUES (?, 1, ?, NOW(), NOW())
ON CONFLICT (rate_key) DO UPDATE SET
  count = CASE
    WHEN auth_rate_limits.expires_at <= NOW() THEN 1
    ELSE auth_rate_limits.count + 1
  END,
  expires_at = CASE
    WHEN auth_rate_limits.expires_at <= NOW() THEN EXCLUDED.expires_at
    ELSE auth_rate_limits.expires_at
  END,
  updated_at = NOW()
RETURNING count, expires_at
`
	if err := r.db.WithContext(ctx).Raw(query, key, expiresAt).Scan(&state).Error; err != nil {
		return 0, 0, fmt.Errorf("increment rate limit key: %w", err)
	}

	ttl := time.Until(state.ExpiresAt)
	if ttl <= 0 {
		ttl = window
	}

	return state.Count, ttl, nil
}
