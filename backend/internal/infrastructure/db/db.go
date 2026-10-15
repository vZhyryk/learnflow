package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDatabase opens a PostgreSQL connection and configures pool settings. Param names
// follow pgxpool's own Config naming (see redis.PoolConfig for the Redis equivalent).
func InitDatabase(dsn, maxIdleTime, maxLifetime string, maxOpenConns, minOpenConns int32) (*pgxpool.Pool, error) {
	config, err := parseConfigs(dsn, maxIdleTime, maxLifetime, maxOpenConns, minOpenConns)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db: failed to ping: %w", err)
	}

	return pool, nil
}

func parseConfigs(dsn, maxIdleTime, maxLifetime string, maxOpenConns, minOpenConns int32) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db: failed to parse config: %w", err)
	}

	// 100 caps one instance at PostgreSQL's default max_connections, leaving headroom
	// for other clients so misconfiguration can't exhaust the server's budget.
	if maxOpenConns <= 0 || maxOpenConns > 100 {
		return nil, fmt.Errorf("db: maxOpenConns must be between 1 and 100, got %d", maxOpenConns)
	}

	idleDuration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, fmt.Errorf("db: failed to parse max idle time: %w", err)
	}
	if idleDuration <= 0 {
		return nil, fmt.Errorf("db: max idle time must be positive, got %s", maxIdleTime)
	}

	lifetime, err := time.ParseDuration(maxLifetime)
	if err != nil {
		return nil, fmt.Errorf("db: failed to parse max lifetime: %w", err)
	}

	if lifetime <= 0 {
		return nil, fmt.Errorf("db: max lifetime must be positive, got %s", maxLifetime)
	}

	if minOpenConns >= maxOpenConns {
		return nil, fmt.Errorf("db: minOpenConns (%d) must be less than maxOpenConns (%d)", minOpenConns, maxOpenConns)
	}
	config.MaxConnLifetime = lifetime
	// MaxConnIdleTime should stay below the upstream proxy/DB idle timeout (e.g. AWS RDS
	// closes idle connections after 5m by default) so pgxpool retires connections on its
	// own schedule instead of racing a server-side disconnect. MaxConns sizing guideline:
	// (peak concurrent handlers/workers per instance) + small headroom, capped well under
	// postgres max_connections (see the <=100 guard below) to leave room for other clients.
	config.MaxConnIdleTime = idleDuration
	config.MaxConns = maxOpenConns
	config.MinConns = minOpenConns
	config.HealthCheckPeriod = 1 * time.Minute

	return config, nil
}
