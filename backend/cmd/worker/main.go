package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	lredis "learnflow_backend/internal/infrastructure/redis"

	"github.com/redis/go-redis/v9"
)

func main() {
	traceLevel := logger.LevelFatal

	environment := env.GetStringEnv("ENVIRONMENT", "production")
	if environment == "dev" {
		traceLevel = logger.LevelError
	}

	jsonLogger := logger.New(os.Stdout, nil, traceLevel)

	appCfg, err := getAppConfig(environment)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	dbInstance, err := db.InitDatabase(appCfg.Database.DSN, appCfg.Database.MaxIdleTime, appCfg.Database.MaxLifetime, int32(appCfg.Database.MaxOpenConns)) //nolint:gosec // bounded by runtime config, cannot overflow int32
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	defer dbInstance.Close()

	redisClient, err := getRedis()
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	defer func() {
		if err := redisClient.Close(); err != nil {
			jsonLogger.Fatal(err, nil)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := App{
		Config:    appCfg,
		Logger:    jsonLogger,
		Outbox:    events.NewOutboxWriter(dbInstance),
		Ctx:       ctx,
		Cancel:    cancel,
		Publisher: events.NewRedisPublisher(redisClient),
	}

	// graceful shutdown при SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()
	app.Wg.Wait()
}

func getRedis() (*redis.Client, error) {
	return lredis.InitRedis(env.GetStringEnv("REDIS_ADDR", "redis:6379"), env.GetStringEnv("REDIS_PASSWORD", ""))
}

func getAppConfig(environment string) (Config, error) {
	cfg := Config{}
	cfg.Env = environment

	dsn, err := db.BuildDSNFromEnv()
	if err != nil {
		return cfg, fmt.Errorf("failed to resolve database DSN: %w", err)
	}

	maxOpenConns, maxIdleTime, maxLifetime := getDatabaseConfig()

	cfg.Database.DSN = dsn
	cfg.Database.MaxIdleTime = maxIdleTime
	cfg.Database.MaxOpenConns = maxOpenConns
	cfg.Database.MaxLifetime = maxLifetime

	return cfg, nil
}

func getDatabaseConfig() (maxOpenConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = max(env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25), runtime.NumCPU()*4)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "30m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "1h")
	return maxOpenConns, maxIdleTime, maxLifetime
}
