package main

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Env      string
	Database struct {
		DSN          string
		MaxIdleTime  string
		MaxOpenConns int
		MinOpenConns int
		MaxLifetime  string
	}

	SMTP SMTP
}

type SMTP struct {
	port     int
	host     string
	username string
	password string
	sender   string
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

// noCopy prevents App from being copied after first use.
// go vet detects accidental copies of types that embed noCopy.
type noCopy struct{}

func (*noCopy) Lock()   {} // intentionally empty — required by sync.Locker for go vet noCopy detection
func (*noCopy) Unlock() {} // intentionally empty — required by sync.Locker for go vet noCopy detection
