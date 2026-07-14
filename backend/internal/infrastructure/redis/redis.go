package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PoolConfig holds Redis connection-pool tuning parameters.
// Field names follow go-redis's redis.Options naming (PoolSize, MinIdleConns,
// ConnMaxLifetime) rather than pgxpool's (MaxConns, MinConns, MaxConnLifetime) —
// this struct is a thin, directly-mapped wrapper over redis.Options, not a
// cross-infrastructure abstraction, so it intentionally mirrors its own client
// library instead of the PostgreSQL pool naming in internal/infrastructure/db.
type PoolConfig struct {
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	ConnMaxLifetime time.Duration
}

// InitRedis creates and pings a Redis client using the given address, password, and pool settings.
func InitRedis(addr, password string, pool PoolConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        password,
		DB:              0,
		PoolSize:        pool.PoolSize,
		MinIdleConns:    pool.MinIdleConns,
		MaxRetries:      pool.MaxRetries,
		DialTimeout:     3 * time.Second,
		ReadTimeout:     3 * time.Second,
		ConnMaxLifetime: pool.ConnMaxLifetime,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: failed to ping: %w", err)
	}

	return client, nil
}
