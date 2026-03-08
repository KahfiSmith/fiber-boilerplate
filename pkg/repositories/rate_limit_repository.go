package repository

import (
	"context"
	"time"
)

type RateLimitRepository interface {
	Hit(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error)
}
