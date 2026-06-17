package main

import (
	"context"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/mailer"
	"sync"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	Env      string
	Database struct {
		DSN          string
		MaxIdleTime  string
		MaxOpenConns int
		MaxLifetime  string
	}
}

// App is the shared application container injected into every handler and worker.
// App must not be copied after first use — always pass as *App.
type App struct {
	_         noCopy
	Config    Config
	Logger    *logger.Logger
	Wg        sync.WaitGroup
	Outbox    *events.OutboxWriter
	Ctx       context.Context
	Cancel    context.CancelFunc
	Publisher *events.RedisPublisher
	Mailer    *mailer.Mailer
}

// noCopy prevents App from being copied after first use.
// go vet detects accidental copies of types that embed noCopy.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
