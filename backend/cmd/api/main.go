// Package main is the entrypoint for the LearnFlow API server.
package main

import (
	"flag"
	"fmt"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/cmd/api/router"
	"learnflow_backend/cmd/api/server"
	"learnflow_backend/internal/infrastructure/config"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	"os"
	"strings"
)

func main() {
	traceLevel := logger.LevelFatal

	ENV := env.GetStringEnv("ENVIRONMENT", "production")
	if ENV == "dev" {
		traceLevel = logger.LevelError
	}

	jsonLogger := logger.New(os.Stdout, nil, traceLevel)

	appCfg, err := getAppConfig()
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	appCfg.Env = ENV

	application := &app.App{
		Config: appCfg,
		Logger: jsonLogger,
	}

	r := router.NewRouter(application)
	srv := server.NewServer(r, application)

	if err := srv.Serve(); err != nil {
		jsonLogger.Fatal(err, nil)
	}
}

func getAppConfig() (app.Config, error) {
	cfg := app.Config{}
	flag.Float64Var(&cfg.Limiter.Rps, "limiter-rps", -1, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "limiter-burst", -1, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.Limiter.Enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	dbConfig, err := config.ResolveConfig(cfg.Config, "api")
	if err != nil {
		return cfg, fmt.Errorf("failed to resolve config: %w", err)
	}

	origins := env.GetStringEnv("CORS_TRUSTED_ORIGINS", "http://localhost:8080")
	cfg.Cors.TrustedOrigins = strings.Split(origins, ",")

	cfg.Config = dbConfig

	cfg.Port = env.GetIntEnv("PORT", 8080)

	cfg = handleAPIServiceConfig(cfg)

	return cfg, nil
}

func handleAPIServiceConfig(cfg app.Config) app.Config {
	if cfg.Limiter.Burst == -1 {
		cfg.Limiter.Burst = env.GetIntEnv("LIMITER_BURST", 10)
	}
	if cfg.Limiter.Rps == -1 {
		cfg.Limiter.Rps = env.GetFloat64Env("LIMITER_RPS", 5)
	}

	return cfg
}
