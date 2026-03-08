package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fiber-boilerplate/pkg/entities"
	"fiber-boilerplate/pkg/mappers"

	"github.com/redis/go-redis/v9"
)

type redisAuthSessionRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisAuthSessionRepository(client *redis.Client, keyPrefix string) AuthSessionRepository {
	return &redisAuthSessionRepository{
		client: client,
		prefix: strings.TrimSpace(keyPrefix),
	}
}

func (r *redisAuthSessionRepository) StoreRefreshSession(ctx context.Context, session entities.RefreshSession, ttl time.Duration) error {
	session = mappers.NormalizeRefreshSession(session)

	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal refresh session: %w", err)
	}

	_, err = r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, r.tokenKey(session.RefreshTokenHash), payload, ttl)
		pipe.Set(ctx, r.sessionKey(session.UserID, session.SessionID), payload, ttl)
		pipe.ZAdd(ctx, r.userSessionsKey(session.UserID), redis.Z{
			Score:  float64(session.CreatedAt.UnixMilli()),
			Member: session.SessionID,
		})
		pipe.Expire(ctx, r.userSessionsKey(session.UserID), ttl)
		return nil
	})
	if err != nil {
		return fmt.Errorf("store refresh session in redis: %w", err)
	}

	return nil
}

func (r *redisAuthSessionRepository) ConsumeRefreshToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error) {
	session, err := r.loadSessionByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return entities.RefreshSession{}, ErrSessionNotFound
		}
		return entities.RefreshSession{}, fmt.Errorf("load refresh session by token hash: %w", err)
	}

	if err := r.deleteSession(ctx, session); err != nil {
		return entities.RefreshSession{}, fmt.Errorf("consume refresh session: %w", err)
	}

	return session, nil
}

func (r *redisAuthSessionRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	session, err := r.loadSessionByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("load refresh session for delete: %w", err)
	}

	if err := r.deleteSession(ctx, session); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (r *redisAuthSessionRepository) HasSession(ctx context.Context, userID uint, sessionID string) (bool, error) {
	exists, err := r.client.Exists(ctx, r.sessionKey(userID, sessionID)).Result()
	if err != nil {
		return false, fmt.Errorf("check redis session existence: %w", err)
	}

	return exists > 0, nil
}

func (r *redisAuthSessionRepository) ListUserSessions(ctx context.Context, userID uint) ([]entities.RefreshSession, error) {
	sessionIDs, err := r.client.ZRevRange(ctx, r.userSessionsKey(userID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("list user session ids from redis: %w", err)
	}

	sessions := make([]entities.RefreshSession, 0, len(sessionIDs))
	staleIDs := make([]string, 0)

	for _, sessionID := range sessionIDs {
		payload, err := r.client.Get(ctx, r.sessionKey(userID, sessionID)).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				staleIDs = append(staleIDs, sessionID)
				continue
			}
			return nil, fmt.Errorf("load user session from redis: %w", err)
		}

		var session entities.RefreshSession
		if err := json.Unmarshal(payload, &session); err != nil {
			return nil, fmt.Errorf("unmarshal user session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if len(staleIDs) > 0 {
		members := make([]any, 0, len(staleIDs))
		for _, sessionID := range staleIDs {
			members = append(members, sessionID)
		}
		if err := r.client.ZRem(ctx, r.userSessionsKey(userID), members...).Err(); err != nil {
			return nil, fmt.Errorf("cleanup stale redis session ids: %w", err)
		}
	}

	return sessions, nil
}

func (r *redisAuthSessionRepository) RevokeSession(ctx context.Context, userID uint, sessionID string) error {
	payload, err := r.client.Get(ctx, r.sessionKey(userID, sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("load redis session for revoke: %w", err)
	}

	var session entities.RefreshSession
	if err := json.Unmarshal(payload, &session); err != nil {
		return fmt.Errorf("unmarshal redis session for revoke: %w", err)
	}

	if err := r.deleteSession(ctx, session); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func (r *redisAuthSessionRepository) RevokeAllSessions(ctx context.Context, userID uint) error {
	sessionIDs, err := r.client.ZRange(ctx, r.userSessionsKey(userID), 0, -1).Result()
	if err != nil {
		return fmt.Errorf("list redis sessions for revoke all: %w", err)
	}

	_, err = r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, sessionID := range sessionIDs {
			payload, err := r.client.Get(ctx, r.sessionKey(userID, sessionID)).Bytes()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					pipe.ZRem(ctx, r.userSessionsKey(userID), sessionID)
					continue
				}
				return fmt.Errorf("load redis session during revoke all: %w", err)
			}

			var session entities.RefreshSession
			if err := json.Unmarshal(payload, &session); err != nil {
				return fmt.Errorf("unmarshal redis session during revoke all: %w", err)
			}

			pipe.Del(ctx, r.tokenKey(session.RefreshTokenHash))
			pipe.Del(ctx, r.sessionKey(userID, sessionID))
			pipe.ZRem(ctx, r.userSessionsKey(userID), sessionID)
		}

		pipe.Del(ctx, r.userSessionsKey(userID))
		return nil
	})
	if err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}

	return nil
}

func (r *redisAuthSessionRepository) loadSessionByToken(ctx context.Context, tokenHash string) (entities.RefreshSession, error) {
	payload, err := r.client.Get(ctx, r.tokenKey(tokenHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return entities.RefreshSession{}, ErrSessionNotFound
		}
		return entities.RefreshSession{}, err
	}

	var session entities.RefreshSession
	if err := json.Unmarshal(payload, &session); err != nil {
		return entities.RefreshSession{}, fmt.Errorf("unmarshal refresh session: %w", err)
	}

	return session, nil
}

func (r *redisAuthSessionRepository) deleteSession(ctx context.Context, session entities.RefreshSession) error {
	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, r.tokenKey(session.RefreshTokenHash))
		pipe.Del(ctx, r.sessionKey(session.UserID, session.SessionID))
		pipe.ZRem(ctx, r.userSessionsKey(session.UserID), session.SessionID)
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete redis session: %w", err)
	}

	return nil
}

func (r *redisAuthSessionRepository) tokenKey(tokenHash string) string {
	return r.prefix + ":auth:session:token:" + strings.TrimSpace(tokenHash)
}

func (r *redisAuthSessionRepository) sessionKey(userID uint, sessionID string) string {
	return fmt.Sprintf("%s:auth:session:user:%d:%s", r.prefix, userID, strings.TrimSpace(sessionID))
}

func (r *redisAuthSessionRepository) userSessionsKey(userID uint) string {
	return fmt.Sprintf("%s:auth:session:user:%d:index", r.prefix, userID)
}
