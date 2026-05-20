// Package app defines the shared application container passed across HTTP layers.
package app

import (
	"database/sql"
	"learnflow_backend/internal/infrastructure/config"
	"learnflow_backend/internal/infrastructure/logger"
	"sync"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Port   int
	Env    string
	DB     *sql.DB
	Config config.Config

	Cors struct {
		TrustedOrigins []string
	}

	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
}

// App is the shared application container injected into every handler and worker.
type App struct {
	Config Config
	Logger *logger.Logger
	Wg     sync.WaitGroup
}
