// Package main is the entrypoint for the LearnFlow API server.
package main

import (
	"context"
	"flag"
	"fmt"
	"learnflow_backend/cmd/api/app"
	"learnflow_backend/cmd/api/router"
	"learnflow_backend/cmd/api/server"
	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/infrastructure/env"
	"net"
	"net/url"
	"strings"
	"time"
)

func main() {
	environment := env.GetStringEnv("ENVIRONMENT", "production")
	jsonLogger := bootstrap.NewLogger(environment)

	appConfig, err := getAppConfig(environment)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	dbInstance, redisClient, cleanup := bootstrap.MustInitInfra(appConfig.Database, jsonLogger)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := &app.App{
		Config: appConfig,
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

	srv := server.NewServer(r, application)

	if err := srv.Serve(); err != nil {
		jsonLogger.Fatal(err, nil)
	}
}

func getServerTimeout(appConfig *app.Config) error {
	durations := []struct {
		name     string
		val      *time.Duration
		envKey   string
		fallback time.Duration
	}{
		{"READ_HEADER_TIMEOUT", &appConfig.Timeouts.ReadHeaderTimeout, "READ_HEADER_TIMEOUT", 5 * time.Second},
		{"READ_TIMEOUT", &appConfig.Timeouts.ReadTimeout, "READ_TIMEOUT", 10 * time.Second},
		{"WRITE_TIMEOUT", &appConfig.Timeouts.WriteTimeout, "WRITE_TIMEOUT", 30 * time.Second},
		{"IDLE_TIMEOUT", &appConfig.Timeouts.IdleTimeout, "IDLE_TIMEOUT", 60 * time.Second},
		{"REQUEST_TIMEOUT", &appConfig.Timeouts.RequestTimeout, "REQUEST_TIMEOUT", 30 * time.Second},
	}
	for _, d := range durations {
		*d.val = env.GetDurationEnv(d.envKey, d.fallback)
		if *d.val <= 0 {
			return fmt.Errorf("%s must be positive, got %v", d.name, *d.val)
		}
	}
	return nil
}

// getAppConfig loads the API server Config. Blocks below mirror app.Config's field order
// so the loading logic can be scanned side-by-side with the struct.
func getAppConfig(environment string) (app.Config, error) {
	cfg := app.Config{}

	cfg.Port = env.GetIntEnv("PORT", 8080)
	cfg.Env = environment

	var err error
	cfg.Database, err = bootstrap.LoadDatabaseConfig()
	if err != nil {
		return cfg, err
	}

	cfg.Cors.TrustedOrigins, err = getCorsTrustedOrigins()
	if err != nil {
		return cfg, fmt.Errorf("CORS config: %w", err)
	}

	cfg.TrustedProxies, err = parseTrustedProxies(env.GetStringEnv("TRUSTED_PROXIES", ""))
	if err != nil {
		return cfg, fmt.Errorf("TRUSTED_PROXIES config: %w", err)
	}

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

	if jwtErr := getJWTConfig(&cfg); jwtErr != nil {
		return cfg, jwtErr
	}

	if err := getServerTimeout(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getJWTConfig(cfg *app.Config) error {
	cfg.Secret.JWTSecret = env.GetStringEnv("JWT_SECRET", "")
	cfg.Secret.JWTSecretPrev = env.GetStringEnv("JWT_SECRET_PREV", "")
	if cfg.Secret.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET cannot be empty")
	}
	if len(cfg.Secret.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes, got %d", len(cfg.Secret.JWTSecret))
	}
	cfg.Secret.JWTIssuer = env.GetStringEnv("JWT_ISSUER", "")
	if cfg.Secret.JWTIssuer == "" {
		return fmt.Errorf("JWT_ISSUER cannot be empty")
	}
	cfg.Secret.JWTAudience = env.GetStringEnv("JWT_AUDIENCE", "")
	if cfg.Secret.JWTAudience == "" {
		return fmt.Errorf("JWT_AUDIENCE cannot be empty")
	}
	return nil
}

func parseTrustedProxies(raw string) ([]net.IPNet, error) {
	if raw == "" {
		return nil, nil
	}
	var cidrs []net.IPNet
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		_, cidr, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		cidrs = append(cidrs, *cidr)
	}
	return cidrs, nil
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
