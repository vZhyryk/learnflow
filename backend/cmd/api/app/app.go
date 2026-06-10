// Package app defines the shared application container passed across HTTP layers.
package app

import (
	"context"
	"learnflow_backend/internal/infrastructure/logger"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Port     int
	Env      string
	Database struct {
		DSN          string
		MaxIdleTime  string
		MaxOpenConns int
		MaxLifetime  string
	}

	Cors struct {
		TrustedOrigins map[string]struct{}
	}

	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}

	JWTSecret string
}

// App is the shared application container injected into every handler and worker.
// App must not be copied after first use — always pass as *App.
type App struct {
	_      noCopy
	Config Config
	Logger *logger.Logger
	Wg     sync.WaitGroup
	DB     *pgxpool.Pool
	Ctx    context.Context
	Cancel context.CancelFunc
}

// noCopy prevents App from being copied after first use.
// go vet detects accidental copies of types that embed noCopy.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
