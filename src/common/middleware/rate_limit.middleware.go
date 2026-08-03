package middleware

import (
	"context"
	"fmt"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	"fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"

	"github.com/gofiber/fiber/v3"
)

// ratelimiter returns a middleware that limits requests per ip using redis sliding window / counter
func RateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		if redis.Client == nil {
			return c.Next()
		}

		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

		count, err := redis.Client.Incr(ctx, key).Result()
		if err != nil {
			return c.Next() // fail-open if redis error
		}

		if count == 1 {
			redis.Client.Expire(ctx, key, window)
		}

		if int(count) > maxRequests {
			return response.HandleError(c, exceptions.TooManyRequests("Too many requests. Please try again later."))
		}

		return c.Next()
	}
}
