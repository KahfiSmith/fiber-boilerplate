package middleware

import (
	"context"
	"fmt"
	"time"

	"fiber-boilerplate/src/common/exceptions"
	redisclient "fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"

	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"
)

var rateLimitLua = goredis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return current
`)

func RateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		if redisclient.Client == nil {
			return c.Next()
		}

		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

		res, err := rateLimitLua.Run(ctx, redisclient.Client, []string{key}, int(window.Seconds())).Result()
		if err != nil {
			return c.Next()
		}

		count, ok := res.(int64)
		if !ok {
			return c.Next()
		}

		if int(count) > maxRequests {
			return response.HandleError(c, exceptions.TooManyRequests("Too many requests. Please try again later."))
		}

		return c.Next()
	}
}
