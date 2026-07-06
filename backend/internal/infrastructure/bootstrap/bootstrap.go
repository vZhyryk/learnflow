// Package bootstrap holds entrypoint setup shared between cmd/api and cmd/worker.
package bootstrap

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	lredis "learnflow_backend/internal/infrastructure/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// NewLogger builds the process logger for a cmd/ entrypoint, defaulting to error-level
// tracing in the "dev" environment and fatal-only otherwise.
func NewLogger(environment string) *logger.Logger {
	traceLevel := logger.LevelFatal
	if environment == "dev" {
		traceLevel = logger.LevelError
	}
	return logger.New(os.Stdout, nil, traceLevel)
}

// DatabaseConfig holds PostgreSQL connection settings, shared by every entrypoint's Config.
type DatabaseConfig struct {
	DSN          string // built from DB_* vars, see db.BuildDSNFromEnv
	MaxIdleTime  string // DB_MAX_IDLE_TIME
	MaxOpenConns int    // DB_OPEN_CONNECTION_LIMIT
	MinOpenConns int    // DB_MIN_CONNECTION_LIMIT
	MaxLifetime  string // DB_MAX_LIFETIME
}

// GetDatabaseConfig reads PostgreSQL connection-pool tuning parameters from the environment.
func GetDatabaseConfig() (maxOpenConns, minOpenConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = max(env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25), runtime.NumCPU()*4)
	minOpenConns = env.GetIntEnv("DB_MIN_CONNECTION_LIMIT", 2)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "30m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "1h")
	return maxOpenConns, minOpenConns, maxIdleTime, maxLifetime
}

// LoadDatabaseConfig resolves the DSN from the environment and combines it with pool
// tuning parameters into a single DatabaseConfig.
func LoadDatabaseConfig() (DatabaseConfig, error) {
	dsn, err := db.BuildDSNFromEnv()
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("failed to resolve database DSN: %w", err)
	}
	maxOpenConns, minOpenConns, maxIdleTime, maxLifetime := GetDatabaseConfig()
	return DatabaseConfig{
		DSN:          dsn,
		MaxIdleTime:  maxIdleTime,
		MaxOpenConns: maxOpenConns,
		MinOpenConns: minOpenConns,
		MaxLifetime:  maxLifetime,
	}, nil
}

// GetRedis builds a Redis client from environment-configured pool settings and pings it.
func GetRedis() (*redis.Client, error) {
	pool := lredis.PoolConfig{
		PoolSize:        env.GetIntEnv("REDIS_POOL_SIZE", 10),
		MinIdleConns:    env.GetIntEnv("REDIS_MIN_IDLE_CONNS", 2),
		MaxRetries:      env.GetIntEnv("REDIS_MAX_RETRIES", 3),
		ConnMaxLifetime: env.GetDurationEnv("REDIS_CONN_MAX_LIFETIME", 5*time.Minute),
	}
	return lredis.InitRedis(env.GetStringEnv("REDIS_ADDR", "redis:6379"), env.GetStringEnv("REDIS_PASSWORD", ""), pool)
}

// MustInitInfra initializes the database pool and Redis client, exiting the process via
// jsonLogger.Fatal (which never returns) if either fails. It returns a single cleanup
// func that closes both connections in the reverse order of acquisition; callers should
// `defer cleanup()` immediately.
func MustInitInfra(dbCfg DatabaseConfig, jsonLogger *logger.Logger) (*pgxpool.Pool, *redis.Client, func()) {
	dbInstance, err := db.InitDatabase(dbCfg.DSN, dbCfg.MaxIdleTime, dbCfg.MaxLifetime, int32(dbCfg.MaxOpenConns), int32(dbCfg.MinOpenConns)) //nolint:gosec // bounded by runtime config, cannot overflow int32
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	redisClient, err := GetRedis()
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	cleanup := func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			jsonLogger.Fatal(closeErr, nil)
		}
		dbInstance.Close()
	}

	return dbInstance, redisClient, cleanup
}
