// Package main is the entrypoint for the database migration runner.
package main

import (
	"errors"
	"flag"
	"fmt"
	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Registers the pgx/v5 database driver with golang-migrate's driver registry.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	"learnflow_backend/migrations"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	environment := env.GetStringEnv("ENVIRONMENT", "production")
	jsonLogger := bootstrap.NewLogger(environment)

	cfg := parseFlags()
	cfg, err := resolveConfig(cfg)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	if cfg.DSN == "" {
		jsonLogger.Fatal(errors.New("dsn is required"), nil)
	}

	dsn := strings.Replace(cfg.DSN, "postgres://", "pgx5://", 1)
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		jsonLogger.Fatal(fmt.Errorf("migrate source init: %w", err), nil)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		jsonLogger.Fatal(fmt.Errorf("migrate init failed (dsn=%s): %w", db.MaskDSN(dsn), err), nil)
	}

	defer handleDefer(m, jsonLogger)

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

func parseFlags() migrationConfig {
	cfg := migrationConfig{}
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

type migrationConfig struct {
	DSN       string
	Direction string
	Steps     int
}

func resolveConfig(cfg migrationConfig) (migrationConfig, error) {
	out := cfg

	var err error
	if out.DSN == "" {
		out.DSN, err = db.BuildDSNFromEnv()
		if err != nil {
			return out, err
		}
	}

	return handleMigrationServiceConfig(out)
}

func handleMigrationServiceConfig(cfg migrationConfig) (migrationConfig, error) {
	if cfg.Direction == "" {
		cfg.Direction = env.GetStringEnv("DIRECTION", "up")
	}
	if cfg.Steps <= 0 {
		cfg.Steps = env.GetIntEnv("STEPS", 0)
	}
	if cfg.Direction != "up" && cfg.Direction != "down" {
		return cfg, fmt.Errorf("invalid direction: %s", cfg.Direction)
	}
	return cfg, nil
}

func handleDefer(m *migrate.Migrate, jsonLogger *logger.Logger) {
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		jsonLogger.Error(fmt.Errorf("migrate close failed: %w", sourceErr), nil)
	}
	if dbErr != nil {
		jsonLogger.Error(fmt.Errorf("migrate close failed: %w", dbErr), nil)
	}
}
