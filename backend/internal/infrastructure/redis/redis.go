package redis

import (
	"context"
	"fmt"
	"learnflow_backend/internal/shared/rediskeys"
	"time"

	"github.com/redis/go-redis/v9"
)

// PoolConfig holds Redis connection-pool tuning parameters. Field names mirror
// go-redis's redis.Options, not pgxpool's — this is a thin wrapper, not a cross-infra abstraction.
type PoolConfig struct {
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	ConnMaxLifetime time.Duration
}

// InitRedis creates and pings a Redis client using the given address, password, and pool settings.
func InitRedis(addr, password string, pool PoolConfig) (*Instance, error) {
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

	return &Instance{client}, nil
}

// Instance wraps the go-redis client (still reachable through the embedded Client, e.g. for queues and scripts)
// and adds the token/user blocklist operations shared by the middleware and services.
type Instance struct {
	*redis.Client
}

// BlockUser marks userID as blocked for ttl, so the middleware rejects that user's already-issued access tokens.
func (ri *Instance) BlockUser(ctx context.Context, userID string, ttl time.Duration) error {
	if err := ri.Set(ctx, rediskeys.UserBlocked(userID), "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis.BlockUser: %w", err)
	}

	return nil
}

// UnBlockUser removes the blocked mark of userID; deleting a missing key is not an error.
func (ri *Instance) UnBlockUser(ctx context.Context, userID string) error {
	if err := ri.Del(ctx, rediskeys.UserBlocked(userID)).Err(); err != nil {
		return fmt.Errorf("redis.UnBlockUser: %w", err)
	}

	return nil
}

// BlockToken blocklists a single access token by its jti for ttl (the token's remaining lifetime).
func (ri *Instance) BlockToken(ctx context.Context, jti string, ttl time.Duration) error {
	if err := ri.SetNX(ctx, rediskeys.JTIBlocked(jti), "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis.BlockToken: %w", err)
	}

	return nil
}

// IsUserBlocked reports whether userID is currently marked as blocked.
func (ri *Instance) IsUserBlocked(ctx context.Context, userID string) (bool, error) {
	exists, err := ri.Exists(ctx, rediskeys.UserBlocked(userID)).Result()
	if err != nil {
		return false, fmt.Errorf("redis.IsUserBlocked: %w", err)
	}

	return exists > 0, nil
}

// IsTokenBlocked reports whether the access token with the given jti is blocklisted.
func (ri *Instance) IsTokenBlocked(ctx context.Context, jti string) (bool, error) {
	exists, err := ri.Exists(ctx, rediskeys.JTIBlocked(jti)).Result()
	if err != nil {
		return false, fmt.Errorf("redis.IsTokenBlocked: %w", err)
	}

	return exists > 0, nil
}
