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
	"net/url"
	"os"
	"strings"
	"time"
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

	if appCfg.Auth.JWTSecret == "" {
		jsonLogger.Fatal(fmt.Errorf("JWT_SECRET required"), nil)
	}

	if len(appCfg.Auth.JWTSecret) < 32 {
		jsonLogger.Fatal(fmt.Errorf("JWT_SECRET must be at least 32 bytes, got %d", len(appCfg.Auth.JWTSecret)), nil)
	}

	dbInstance, err := db.InitDatabase(appCfg.Database.DSN, appCfg.Database.MaxIdleTime, appCfg.Database.MaxLifetime, int32(appCfg.Database.MaxOpenConns))
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	defer dbInstance.Close()

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

	cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.EmailVerificationTokenTTL, cfg.Auth.PasswordResetTokenTTL = getTokensData()
	return cfg, nil
}

func getCorsTrustedOrigins() ([]string, error) {
	origins := env.GetStringEnv("CORS_TRUSTED_ORIGINS", "http://localhost:3000")
	parts := strings.Split(origins, ",")
	var valid []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u, err := url.Parse(p)
		if err != nil || u.Host == "" {
			return nil, fmt.Errorf("invalid CORS origin: %q (expected https://host)", p)
		}
		valid = append(valid, p)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("CORS_TRUSTED_ORIGINS is empty — at least one origin is required")
	}
	return valid, nil
}

func getDatabaseConfig() (maxOpenConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "15m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "30m")
	return maxOpenConns, maxIdleTime, maxLifetime
}

func getTokensData() (jwtSecret string, accessTokenTTL, refreshTokenTTL, emailVerificationTokenTTL, passwordResetTokenTTL time.Duration) {
	jwtSecret = env.GetStringEnv("JWT_SECRET", "")

	accessTokenTTL = env.GetDurationEnv("ACCESS_TOKEN_TTL", 5*time.Minute)
	refreshTokenTTL = env.GetDurationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	emailVerificationTokenTTL = env.GetDurationEnv("EMAIL_VERIFICATION_TOKEN_TTL", time.Hour)
	passwordResetTokenTTL = env.GetDurationEnv("PASSWORD_RESET_TOKEN_TTL", time.Hour)

	return jwtSecret, accessTokenTTL, refreshTokenTTL, emailVerificationTokenTTL, passwordResetTokenTTL
}
