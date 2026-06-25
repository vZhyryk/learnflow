// Package main is the entrypoint for the LearnFlow API server.
package main

import (
	"context"
	"flag"
	"fmt"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/cmd/api/router"
	"learnflow_backend/cmd/api/server"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	lredis "learnflow_backend/internal/infrastructure/redis"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

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
		if closeErr := redisClient.Close(); closeErr != nil {
			jsonLogger.Fatal(closeErr, nil)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := &app.App{
		Config: appCfg,
		Logger: jsonLogger,
		DB:     dbInstance,
		Ctx:    ctx,
		Cancel: cancel,
		Redis:  redisClient,
	}

	r, err := router.NewRouter(application)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	readHeaderTimeout, readTimeout, writeTimeout, idleTimeout, requestTimeout, err := getServerTimeout()
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}
	appCfg.Timeouts.ReadHeaderTimeout = readHeaderTimeout
	appCfg.Timeouts.ReadTimeout = readTimeout
	appCfg.Timeouts.WriteTimeout = writeTimeout
	appCfg.Timeouts.IdleTimeout = idleTimeout
	appCfg.Timeouts.RequestTimeout = requestTimeout

	srv := server.NewServer(r, application)

	if err := srv.Serve(); err != nil {
		jsonLogger.Fatal(err, nil)
	}
}

func getServerTimeout() (readHeaderTimeout, readTimeout, writeTimeout, idleTimeout, requestTimeout time.Duration, err error) {
	durations := []struct {
		name     string
		val      *time.Duration
		envKey   string
		fallback time.Duration
	}{
		{"READ_HEADER_TIMEOUT", &readHeaderTimeout, "READ_HEADER_TIMEOUT", 5 * time.Second},
		{"READ_TIMEOUT", &readTimeout, "READ_TIMEOUT", 10 * time.Second},
		{"WRITE_TIMEOUT", &writeTimeout, "WRITE_TIMEOUT", 30 * time.Second},
		{"IDLE_TIMEOUT", &idleTimeout, "IDLE_TIMEOUT", 60 * time.Second},
		{"REQUEST_TIMEOUT", &requestTimeout, "REQUEST_TIMEOUT", 30 * time.Second},
	}
	for _, d := range durations {
		*d.val = env.GetDurationEnv(d.envKey, d.fallback)
		if *d.val <= 0 {
			return 0, 0, 0, 0, 0, fmt.Errorf("%s must be positive, got %v", d.name, *d.val)
		}
	}
	return
}

func getAppConfig(environment string) (app.Config, error) {
	cfg := app.Config{}
	cfg.Env = environment

	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", -1, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", -1, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	visited := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { visited[f.Name] = true })

	if !visited["limiter-rps"] {
		cfg.Limiter.Rps = env.GetFloat64Env("LIMITER_RPS", 5)
	}

	if !visited["limiter-burst"] {
		cfg.Limiter.Burst = env.GetIntEnv("LIMITER_BURST", 10)
	}

	dsn, err := db.BuildDSNFromEnv()
	if err != nil {
		return cfg, fmt.Errorf("failed to resolve database DSN: %w", err)
	}

	cfg.Cors.TrustedOrigins, err = getCorsTrustedOrigins()
	if err != nil {
		return cfg, fmt.Errorf("CORS config: %w", err)
	}

	maxOpenConns, maxIdleTime, maxLifetime := getDatabaseConfig()

	cfg.Database.DSN = dsn
	cfg.Database.MaxIdleTime = maxIdleTime
	cfg.Database.MaxOpenConns = maxOpenConns
	cfg.Database.MaxLifetime = maxLifetime

	cfg.Port = env.GetIntEnv("PORT", 8080)

	cfg.Secret.JWTSecret = env.GetStringEnv("JWT_SECRET", "")

	if cfg.Secret.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET required: %w", err)
	}

	if len(cfg.Secret.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET must be at least 32 bytes, got %d", len(cfg.Secret.JWTSecret))
	}

	cfg.Secret.JWTIssuer = env.GetStringEnv("JWT_ISSUER", "")
	if cfg.Secret.JWTIssuer == "" {
		return cfg, fmt.Errorf("JWT_ISSUER required: %w", err)
	}
	cfg.Secret.JWTAudience = env.GetStringEnv("JWT_AUDIENCE", "")
	if cfg.Secret.JWTAudience == "" {
		return cfg, fmt.Errorf("JWT_AUDIENCE required: %w", err)
	}

	return cfg, nil
}

func getCorsTrustedOrigins() (map[string]struct{}, error) {
	origins := env.GetStringEnv("CORS_TRUSTED_ORIGINS", "http://localhost:3000")
	parts := strings.Split(origins, ",")
	valid := make(map[string]struct{})
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		u, err := url.Parse(p)
		if err != nil || u.Host == "" {
			return nil, fmt.Errorf("invalid CORS origin: %q (expected https://host)", p)
		}

		valid[p] = struct{}{}
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("CORS_TRUSTED_ORIGINS is empty — at least one origin is required")
	}
	return valid, nil
}

func getDatabaseConfig() (maxOpenConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = max(env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25), runtime.NumCPU()*4)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "30m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "1h")
	return maxOpenConns, maxIdleTime, maxLifetime
}

func getRedis() (*redis.Client, error) {
	return lredis.InitRedis(env.GetStringEnv("REDIS_ADDR", "redis:6379"), env.GetStringEnv("REDIS_PASSWORD", ""))
}
