package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisRateLimitHitScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`)

type redisRateLimitRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisRateLimitRepository(client *redis.Client, keyPrefix string) RateLimitRepository {
	return &redisRateLimitRepository{
		client: client,
		prefix: strings.TrimSpace(keyPrefix),
	}
}

func (r *redisRateLimitRepository) Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return 0, 0, fmt.Errorf("rate limit key must not be empty")
	}

	result, err := redisRateLimitHitScript.Run(ctx, r.client, []string{r.key(key)}, window.Milliseconds()).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("increment redis rate limit key: %w", err)
	}

	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return 0, 0, fmt.Errorf("unexpected redis rate limit result: %T", result)
	}

	count, ok := values[0].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected redis rate limit count type: %T", values[0])
	}

	ttlMillis, ok := values[1].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("unexpected redis rate limit ttl type: %T", values[1])
	}

	ttl := time.Duration(ttlMillis) * time.Millisecond
	if ttl <= 0 {
		ttl = window
	}

	return count, ttl, nil
}

func (r *redisRateLimitRepository) key(key string) string {
	return r.prefix + ":rate:" + strings.TrimSpace(key)
}
