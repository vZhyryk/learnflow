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
	"os"
	"strings"
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

	dbInstance, err := db.InitDatabase(appCfg.Database.DSN, appCfg.Database.MaxIdleTime, appCfg.Database.MaxLifetime, appCfg.Database.MaxOpenConns, appCfg.Database.MaxIdleConns)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := &app.App{
		Config: appCfg,
		Logger: jsonLogger,
		DB:     dbInstance,
		Ctx:    ctx,
		Cancel: cancel,
	}

	r := router.NewRouter(application)
	srv := server.NewServer(r, application)

	if err := srv.Serve(); err != nil {
		jsonLogger.Fatal(err, nil)
	}
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

	cfg.Cors.TrustedOrigins = getCorsTrustedOrigins()

	maxOpenConns, maxIdleConns, maxIdleTime, maxLifetime := getDatabaseConfig()

	cfg.Database.DSN = dsn
	cfg.Database.MaxIdleTime = maxIdleTime
	cfg.Database.MaxOpenConns = maxOpenConns
	cfg.Database.MaxIdleConns = maxIdleConns
	cfg.Database.MaxLifetime = maxLifetime

	cfg.Port = env.GetIntEnv("PORT", 8080)

	return cfg, nil
}

func getCorsTrustedOrigins() []string {
	origins := env.GetStringEnv("CORS_TRUSTED_ORIGINS", "http://localhost:3000")
	parts := strings.Split(origins, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func getDatabaseConfig() (maxOpenConns, maxIdleConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25)
	maxIdleConns = env.GetIntEnv("DB_IDLE_CONNECTION_LIMIT", 15)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "15m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "30m")
	return maxOpenConns, maxIdleConns, maxIdleTime, maxLifetime
}
