package router

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
    local count = redis.call("INCR", KEYS[1])
    if count == 1 then
        redis.call("EXPIRE", KEYS[1], ARGV[1])
    end
    return count
`)

func redisRateLimit(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (allowed bool, err error) {
	windowSecs := int(window.Seconds())
	if windowSecs <= 0 {
		windowSecs = 1
	}

	result, err := rateLimitScript.Run(ctx, rdb, []string{key}, windowSecs).Int()
	if err != nil {
		return true, fmt.Errorf("redisRateLimit: %w", err)
	}

	return result <= limit, nil
}
