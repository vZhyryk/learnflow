// Package app defines the shared application container passed across HTTP layers.
package app

import (
	"context"
	"learnflow_backend/internal/infrastructure/logger"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Port     int
	Env      string
	Database struct {
		DSN          string
		MaxIdleTime  string
		MaxOpenConns int
		MinOpenConns int
		MaxLifetime  string
	}

	Cors struct {
		TrustedOrigins map[string]struct{}
	}

	TrustedProxies []net.IPNet

	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}

	Secret struct {
		JWTSecret     string
		JWTSecretPrev string
		JWTIssuer     string
		JWTAudience   string
	}

	Timeouts struct {
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		RequestTimeout    time.Duration
	}
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
	Redis  *redis.Client
}

// noCopy prevents App from being copied after first use.
// go vet detects accidental copies of types that embed noCopy.
type noCopy struct{}

func (*noCopy) Lock()   {} // intentionally empty — required by sync.Locker for go vet noCopy detection
func (*noCopy) Unlock() {} // intentionally empty — required by sync.Locker for go vet noCopy detection
