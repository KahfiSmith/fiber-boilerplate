package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisclient "fiber-boilerplate/src/common/redis"

	"github.com/redis/go-redis/v9"
)

type SessionMetadata struct {
	SessionID        string     `json:"session_id"`
	FamilyID         string     `json:"family_id"`
	UserID           uint       `json:"user_id"`
	CurrentTokenHash string     `json:"current_token_hash"`
	IssuedAt         time.Time  `json:"issued_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	LastUsedAt       time.Time  `json:"last_used_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	RevocationReason string     `json:"revocation_reason,omitempty"`
	IPAddress        string     `json:"ip_address"`
	UserAgent        string     `json:"user_agent"`
}

type RefreshRepository struct{}

func NewRefreshRepository() *RefreshRepository {
	return &RefreshRepository{}
}

func (r *RefreshRepository) CreateSession(ctx context.Context, meta SessionMetadata, ttl time.Duration) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	sessionKey := fmt.Sprintf("auth:session:%s", meta.SessionID)
	activeKey := fmt.Sprintf("auth:refresh:active:%s", meta.CurrentTokenHash)
	familyKey := fmt.Sprintf("auth:family:%s", meta.FamilyID)

	pipe := redisclient.Client.TxPipeline()
	pipe.Set(ctx, sessionKey, data, ttl)
	pipe.Set(ctx, activeKey, meta.SessionID, ttl)
	pipe.SAdd(ctx, familyKey, meta.CurrentTokenHash)
	pipe.Expire(ctx, familyKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

type RotationResult struct {
	Status      string 
	SessionData string
}

const AtomicRotationLuaScript = `
local token_hash = KEYS[1]
local new_token_hash = KEYS[2]
local ttl_seconds = tonumber(ARGV[1])
local now_iso = ARGV[2]

local active_key = "auth:refresh:active:" .. token_hash
local used_key = "auth:refresh:used:" .. token_hash

local session_id = redis.call("GET", active_key)

if not session_id then
    local used_session_id = redis.call("GET", used_key)
    if used_session_id then
        return {"USED", used_session_id}
    else
        return {"NOT_FOUND", ""}
    end
end

local session_key = "auth:session:" .. session_id
local session_data_str = redis.call("GET", session_key)

if not session_data_str then
    return {"NOT_FOUND", ""}
end

local session_data = cjson.decode(session_data_str)

if session_data["revoked_at"] then
    return {"REVOKED", session_data_str}
end

-- Update session metadata
session_data["current_token_hash"] = new_token_hash
session_data["last_used_at"] = now_iso

local updated_session_str = cjson.encode(session_data)
local new_active_key = "auth:refresh:active:" .. new_token_hash
local family_key = "auth:family:" .. session_data["family_id"]

-- Move token from active to used
redis.call("DEL", active_key)
redis.call("SET", used_key, session_id, "EX", ttl_seconds)
redis.call("SET", new_active_key, session_id, "EX", ttl_seconds)
redis.call("SET", session_key, updated_session_str, "EX", ttl_seconds)
redis.call("SADD", family_key, new_token_hash)
redis.call("EXPIRE", family_key, ttl_seconds)

return {"SUCCESS", updated_session_str}
`

func (r *RefreshRepository) RotateToken(ctx context.Context, tokenHash string, newTokenHash string, ttl time.Duration) (*RotationResult, error) {
	nowStr := time.Now().Format(time.RFC3339)
	ttlSec := int64(ttl.Seconds())

	res, err := redisclient.Client.Eval(ctx, AtomicRotationLuaScript, []string{tokenHash, newTokenHash}, ttlSec, nowStr).Result()
	if err != nil {
		return nil, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return nil, fmt.Errorf("invalid lua response format")
	}

	status, _ := arr[0].(string)
	sessionData, _ := arr[1].(string)

	return &RotationResult{
		Status:      status,
		SessionData: sessionData,
	}, nil
}

func (r *RefreshRepository) RevokeFamily(ctx context.Context, familyID string, reason string) error {
	familyKey := fmt.Sprintf("auth:family:%s", familyID)

	hashes, err := redisclient.Client.SMembers(ctx, familyKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := redisclient.Client.TxPipeline()

	for _, h := range hashes {
		activeKey := fmt.Sprintf("auth:refresh:active:%s", h)
		usedKey := fmt.Sprintf("auth:refresh:used:%s", h)

		sessionID, _ := redisclient.Client.Get(ctx, activeKey).Result()
		if sessionID == "" {
			sessionID, _ = redisclient.Client.Get(ctx, usedKey).Result()
		}

		if sessionID != "" {
			sessionKey := fmt.Sprintf("auth:session:%s", sessionID)
			sData, err := redisclient.Client.Get(ctx, sessionKey).Result()
			if err == nil && sData != "" {
				var meta SessionMetadata
				if err := json.Unmarshal([]byte(sData), &meta); err == nil {
					now := time.Now()
					meta.RevokedAt = &now
					meta.RevocationReason = reason
					updated, _ := json.Marshal(meta)
					pipe.Set(ctx, sessionKey, updated, time.Hour*24*30)
				}
			}
		}

		pipe.Del(ctx, activeKey)
		pipe.Del(ctx, usedKey)
	}

	pipe.Del(ctx, familyKey)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RefreshRepository) RevokeSessionByTokenHash(ctx context.Context, tokenHash string, reason string) error {
	activeKey := fmt.Sprintf("auth:refresh:active:%s", tokenHash)
	usedKey := fmt.Sprintf("auth:refresh:used:%s", tokenHash)

	sessionID, err := redisclient.Client.Get(ctx, activeKey).Result()
	if err != nil || sessionID == "" {
		sessionID, _ = redisclient.Client.Get(ctx, usedKey).Result()
	}

	if sessionID == "" {
		return nil
	}

	sessionKey := fmt.Sprintf("auth:session:%s", sessionID)
	sData, err := redisclient.Client.Get(ctx, sessionKey).Result()
	if err == nil && sData != "" {
		var meta SessionMetadata
		if err := json.Unmarshal([]byte(sData), &meta); err == nil {
			return r.RevokeFamily(ctx, meta.FamilyID, reason)
		}
	}

	return nil
}

func (r *RefreshRepository) GetSession(ctx context.Context, sessionID string) (*SessionMetadata, error) {
	sessionKey := fmt.Sprintf("auth:session:%s", sessionID)
	data, err := redisclient.Client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var meta SessionMetadata
	if err := json.Unmarshal([]byte(data), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
