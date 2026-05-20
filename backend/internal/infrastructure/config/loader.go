package config

import (
	"fmt"
	"learnflow_backend/internal/infrastructure/env"
)

// ResolveConfig fills in any zero-value fields from environment variables and
// applies service-specific defaults (e.g. migration direction for "migration").
func ResolveConfig(cfg Config, serviceName string) (Config, error) {
	out := cfg

	var err error
	if out.DSN == "" {
		out.DSN, err = BuildDSNFromEnv()
		if err != nil {
			return out, err
		}
	}

	switch serviceName {
	case "migration":
		return handleMigrationServiceConfig(out)
	default:
		return out, nil
	}
}

func handleMigrationServiceConfig(cfg Config) (Config, error) {
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
