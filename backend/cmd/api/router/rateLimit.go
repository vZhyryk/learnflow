package router

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// rateLimitScript implements a token bucket algorithm in Redis.
// KEYS[1] = bucket key
// ARGV[1] = now (unix nanoseconds)
// ARGV[2] = refill rate (tokens per second, float)
// ARGV[3] = burst capacity (max tokens)
// ARGV[4] = TTL in seconds
var rateLimitScript = redis.NewScript(`
	local key    = KEYS[1]
	local now    = tonumber(ARGV[1])
	local rate   = tonumber(ARGV[2])
	local burst  = tonumber(ARGV[3])
	local ttl    = tonumber(ARGV[4])

	local data   = redis.call("HMGET", key, "tokens", "last_refill")
	local tokens = tonumber(data[1]) or burst
	local last   = tonumber(data[2]) or now

	local elapsed = math.max(0, now - last)
	tokens = math.min(burst, tokens + elapsed * rate / 1e9)

	local allowed = 0
	if tokens >= 1 then
		tokens  = tokens - 1
		allowed = 1
	end

	redis.call("HSET", key, "tokens", tokens, "last_refill", now)
	redis.call("EXPIRE", key, ttl)
	return allowed
`)

func redisRateLimit(ctx context.Context, rdb *redis.Client, key string, rps float64, burst int, window time.Duration) (bool, error) {
	now := time.Now().UnixNano()
	ttl := max(int(window.Seconds())*2, 1)

	result, err := rateLimitScript.Run(ctx, rdb, []string{key}, now, rps, burst, ttl).Int()
	if err != nil {
		return true, fmt.Errorf("redisRateLimit: %w", err)
	}

	return result == 1, nil
}
