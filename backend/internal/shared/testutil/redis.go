package testutil

import "github.com/redis/go-redis/v9"

// UnreachableRedis returns a Redis client pointed at a non-listening address, for tests exercising Redis-failure paths.
func UnreachableRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
}
