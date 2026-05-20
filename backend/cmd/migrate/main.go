// Package main is the entrypoint for the database migration runner.
package main

import (
	"errors"
	"flag"
	"fmt"
	"learnflow_backend/internal/infrastructure/config"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	traceLevel := logger.LevelFatal
	if env.GetStringEnv("ENVIRONMENT", "production") == "dev" {
		traceLevel = logger.LevelError
	}

	jsonLogger := logger.New(os.Stdout, nil, traceLevel)

	cfg := parseFlags()
	cfg, err := config.ResolveConfig(cfg, "migration")
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	if cfg.DSN == "" {
		jsonLogger.Fatal(errors.New("dsn is required"), nil)
	}

	m, err := migrate.New("file://migrations", cfg.DSN)
	if err != nil {
		jsonLogger.Fatal(fmt.Errorf("migrate init failed: %w", err), nil)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			jsonLogger.Error(fmt.Errorf("migrate close failed: %w", sourceErr), nil)
		}
		if dbErr != nil {
			jsonLogger.Error(fmt.Errorf("migrate close failed: %w", dbErr), nil)
		}
	}()

	err = run(m, cfg.Direction, cfg.Steps)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		jsonLogger.Fatal(fmt.Errorf("migrate failed: %w", err), nil)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		jsonLogger.Info("migrate: no change", nil)
	} else {
		jsonLogger.Info("migrate: done", nil)
	}
}

func parseFlags() config.Config {
	cfg := config.Config{}
	flag.StringVar(&cfg.DSN, "dsn", "", "PostgreSQL DSN (required)")
	flag.StringVar(&cfg.Direction, "direction", "", "Migration direction: up or down")
	flag.IntVar(&cfg.Steps, "steps", 0, "Number of migrations to apply (0 = all)")
	flag.Parse()
	return cfg
}

func run(m *migrate.Migrate, direction string, steps int) error {
	switch direction {
	case "down":
		if steps > 0 {
			if err := m.Steps(-steps); err != nil {
				return fmt.Errorf("migrate steps down: %w", err)
			}
			return nil
		}
		if err := m.Down(); err != nil {
			return fmt.Errorf("migrate down: %w", err)
		}
		return nil
	case "up":
		if steps > 0 {
			if err := m.Steps(steps); err != nil {
				return fmt.Errorf("migrate steps up: %w", err)
			}
			return nil
		}
		if err := m.Up(); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}
}
