package redis

import (
	"context"
	"crypto/tls"
	"errors"
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
	TLS             bool
}

// InitRedis creates and pings a Redis client using the given address, password, and pool settings.
func InitRedis(addr, password string, pool PoolConfig) (*Instance, error) {
	var tlsConfig *tls.Config
	if pool.TLS {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(&redis.Options{
		TLSConfig:       tlsConfig,
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

	return NewInstance(client), nil
}

// Instance wraps the go-redis client and adds the token/user blocklist operations shared by the middleware
// and services. The raw client is reachable only through Raw, so code bypassing the blocklist API is explicit.
type Instance struct {
	client *redis.Client
}

// NewInstance wraps an existing go-redis client.
func NewInstance(client *redis.Client) *Instance {
	return &Instance{client: client}
}

// Raw returns the underlying go-redis client, for queues, scripts and health checks.
func (ri *Instance) Raw() *redis.Client { return ri.client }

// Close closes the underlying client.
func (ri *Instance) Close() error {
	if err := ri.client.Close(); err != nil {
		return fmt.Errorf("redis.Close: %w", err)
	}

	return nil
}

// checkTTL rejects a non-positive ttl, which Redis would store as a key that never expires.
func checkTTL(op string, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("redis.%s: ttl must be positive, got %s", op, ttl)
	}

	return nil
}

// BlockUser marks userID as blocked for ttl, so the middleware rejects that user's already-issued access tokens.
func (ri *Instance) BlockUser(ctx context.Context, userID string, ttl time.Duration) error {
	if err := checkTTL("BlockUser", ttl); err != nil {
		return err
	}

	if err := ri.client.Set(ctx, rediskeys.UserBlocked(userID), "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis.BlockUser: %w", err)
	}

	return nil
}

// UnBlockUser removes the blocked mark of userID; deleting a missing key is not an error.
func (ri *Instance) UnBlockUser(ctx context.Context, userID string) error {
	if err := ri.client.Del(ctx, rediskeys.UserBlocked(userID)).Err(); err != nil {
		return fmt.Errorf("redis.UnBlockUser: %w", err)
	}

	return nil
}

// BlockToken blocklists a single access token by its jti for ttl (the token's remaining lifetime).
func (ri *Instance) BlockToken(ctx context.Context, jti string, ttl time.Duration) error {
	if err := checkTTL("BlockToken", ttl); err != nil {
		return err
	}

	if err := ri.client.SetNX(ctx, rediskeys.JTIBlocked(jti), "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis.BlockToken: %w", err)
	}

	return nil
}

// RevokeUserRole marks userID's role as revoked for ttl, so RequireRole rejects tokens issued with the old role.
func (ri *Instance) RevokeUserRole(ctx context.Context, userID string, ttl time.Duration) error {
	if err := checkTTL("RevokeUserRole", ttl); err != nil {
		return err
	}

	if err := ri.client.Set(ctx, rediskeys.UserRoleRevoked(userID), "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis.RevokeUserRole: %w", err)
	}

	return nil
}

// ClearUserRoleRevoked removes the role-revoked mark of userID; deleting a missing key is not an error.
func (ri *Instance) ClearUserRoleRevoked(ctx context.Context, userID string) error {
	if err := ri.client.Del(ctx, rediskeys.UserRoleRevoked(userID)).Err(); err != nil {
		return fmt.Errorf("redis.ClearUserRoleRevoked: %w", err)
	}

	return nil
}

// IsUserRoleRevoked reports whether userID's role is currently marked as revoked.
func (ri *Instance) IsUserRoleRevoked(ctx context.Context, userID string) (bool, error) {
	exists, err := ri.client.Exists(ctx, rediskeys.UserRoleRevoked(userID)).Result()
	if err != nil {
		return false, fmt.Errorf("redis.IsUserRoleRevoked: %w", err)
	}

	return exists > 0, nil
}

// RevokeUserTokens revokes every access token issued to userID up to now, for ttl (one access-token lifetime is enough:
// later the older tokens have expired anyway). Tokens issued after this call are not affected.
func (ri *Instance) RevokeUserTokens(ctx context.Context, userID string, ttl time.Duration) error {
	if err := checkTTL("RevokeUserTokens", ttl); err != nil {
		return err
	}

	if err := ri.client.Set(ctx, rediskeys.TokensRevokedBefore(userID), time.Now().Unix(), ttl).Err(); err != nil {
		return fmt.Errorf("redis.RevokeUserTokens: %w", err)
	}

	return nil
}

// IsBlocked reports whether the user, the access token (by jti), or every token issued up to issuedAt (see
// RevokeUserTokens) is blocklisted, in one round-trip. A token is revoked when issued at or before the revocation
// second, because the iat claim has one-second precision.
func (ri *Instance) IsBlocked(ctx context.Context, userID, jti string, issuedAt time.Time) (bool, error) {
	pipe := ri.client.Pipeline()
	exists := pipe.Exists(ctx, rediskeys.UserBlocked(userID), rediskeys.JTIBlocked(jti))
	revokedBefore := pipe.Get(ctx, rediskeys.TokensRevokedBefore(userID))

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("redis.IsBlocked: %w", err)
	}

	if exists.Val() > 0 {
		return true, nil
	}

	cutoff, err := revokedBefore.Int64()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis.IsBlocked: parse revoked-before: %w", err)
	}

	return issuedAt.Unix() <= cutoff, nil
}
