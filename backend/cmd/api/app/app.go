// Package app defines the shared application container passed across HTTP layers.
package app

import (
	"context"
	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/infrastructure/logger"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config holds all runtime configuration for the API server. Each field's doc comment
// names its source env var/flag — see cmd/api/main.go's getAppConfig/getJWTConfig.
type Config struct {
	Port     int    // PORT
	Env      string // ENVIRONMENT
	Database bootstrap.DatabaseConfig

	Cors struct {
		TrustedOrigins map[string]struct{} // CORS_TRUSTED_ORIGINS
	}

	TrustedProxies []net.IPNet // TRUSTED_PROXIES

	Limiter struct {
		Rps     float64 // -limiter-rps flag, falls back to LIMITER_RPS
		Burst   int     // -limiter-burst flag, falls back to LIMITER_BURST
		Enabled bool    // -limiter-enabled flag
	}

	Secret struct {
		JWTSecret     string // JWT_SECRET
		JWTSecretPrev string // JWT_SECRET_PREV
		JWTIssuer     string // JWT_ISSUER
		JWTAudience   string // JWT_AUDIENCE
	}

	Timeouts struct {
		ReadHeaderTimeout time.Duration // READ_HEADER_TIMEOUT
		ReadTimeout       time.Duration // READ_TIMEOUT
		WriteTimeout      time.Duration // WRITE_TIMEOUT
		IdleTimeout       time.Duration // IDLE_TIMEOUT
		RequestTimeout    time.Duration // REQUEST_TIMEOUT
	}
}

// App is the shared application container — must not be copied after first use, always
// pass as *App. Ctx/Cancel are the process-lifetime shutdown signal, not per-request.
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
