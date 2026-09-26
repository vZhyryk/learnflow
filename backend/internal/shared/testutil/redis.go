package testutil

import (
	"time"

	redisinfra "learnflow_backend/internal/infrastructure/redis"

	"github.com/redis/go-redis/v9"
)

// UnreachableRedis returns a Redis client pointed at a non-listening address, for tests exercising Redis-failure paths.
func UnreachableRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
}

// UnreachableRedisInstance is UnreachableRedis wrapped in the Instance the app and services are wired with.
func UnreachableRedisInstance() *redisinfra.Instance {
	return &redisinfra.Instance{Client: UnreachableRedis()}
}
