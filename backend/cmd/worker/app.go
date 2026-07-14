package main

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config holds all runtime configuration for the worker process.
//
// Each field's doc comment names the env var it is sourced from —
// see cmd/worker/main.go's getAppConfig/getMailerConfig for the loading logic.
type Config struct {
	Env      string // ENVIRONMENT
	Database bootstrap.DatabaseConfig

	SMTP SMTP
}

// SMTP holds outgoing-mail server configuration.
type SMTP struct {
	Port     int    // SMTP_PORT
	Host     string // SMTP_HOST
	Username string // SMTP_USERNAME
	Password string // SMTP_PASSWORD
	Sender   string // SMTP_SENDER
}

// App is the shared application container injected into every handler and worker.
// App must not be copied after first use — always pass as *App.
type App struct {
	_           noCopy
	Config      Config
	Logger      *logger.Logger
	Wg          sync.WaitGroup
	Outbox      *events.OutboxWriter
	Publisher   *events.RedisPublisher
	Mailer      *mailer.Mailer
	RedisClient *redis.Client
}

// NewApp wraps raw infrastructure dependencies (DB pool, Redis client) into the
// App container, mirroring how cmd/api/router.NewRouter wraps cmd/api/app.App's
// raw DB/Redis into repositories and services — both entrypoints gather raw
// dependencies in main() first, then hand them to a single factory step.
func NewApp(cfg Config, log *logger.Logger, dbInstance *pgxpool.Pool, redisClient *redis.Client) *App {
	return &App{
		Config:      cfg,
		Logger:      log,
		Outbox:      events.NewOutboxWriter(dbInstance),
		Publisher:   events.NewRedisPublisher(redisClient),
		RedisClient: redisClient,
		Mailer:      mailer.New(cfg.SMTP.Port, cfg.SMTP.Host, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Sender),
	}
}

// noCopy prevents App from being copied after first use.
// go vet detects accidental copies of types that embed noCopy.
type noCopy struct{}

func (*noCopy) Lock()   {} // intentionally empty — required by sync.Locker for go vet noCopy detection
func (*noCopy) Unlock() {} // intentionally empty — required by sync.Locker for go vet noCopy detection
