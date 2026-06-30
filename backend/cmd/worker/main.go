package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/infrastructure/logger"
	lredis "learnflow_backend/internal/infrastructure/redis"
	"learnflow_backend/internal/shared/mailer"
	"learnflow_backend/internal/worker"

	"github.com/redis/go-redis/v9"
)

func main() {
	traceLevel := logger.LevelFatal

	environment := env.GetStringEnv("ENVIRONMENT", "production")
	if environment == "dev" {
		traceLevel = logger.LevelError
	}

	jsonLogger := logger.New(os.Stdout, nil, traceLevel)

	appCfg, err := getAppConfig(environment)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	dbInstance, err := db.InitDatabase(appCfg.Database.DSN, appCfg.Database.MaxIdleTime, appCfg.Database.MaxLifetime, int32(appCfg.Database.MaxOpenConns), int32(appCfg.Database.MinOpenConns)) //nolint:gosec // bounded by runtime config, cannot overflow int32
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	defer dbInstance.Close()

	redisClient, err := getRedis()
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			jsonLogger.Fatal(closeErr, nil)
		}
	}()

	smtp, err := getSMTP(environment)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := App{
		Config:      appCfg,
		Logger:      jsonLogger,
		Outbox:      events.NewOutboxWriter(dbInstance),
		Publisher:   events.NewRedisPublisher(redisClient),
		RedisClient: redisClient,
		Mailer:      mailer.New(smtp.port, smtp.host, smtp.username, smtp.password, smtp.sender),
	}

	transactor := db.NewTransactor(dbInstance)

	baseURL := env.GetStringEnv("BASE_URL", "")
	if baseURL == "" {
		jsonLogger.Fatal(fmt.Errorf("base url is not valid"), nil)
	}

	workers := []worker.Worker{
		worker.NewOutboxPoller(dbInstance, app.Publisher, app.Logger, transactor),
		worker.NewDLQRetryWorker(dbInstance, app.Publisher, app.Logger, transactor),
		worker.NewEmailVerificationWorker(dbInstance, redisClient, app.Logger, app.Mailer, baseURL),
		worker.NewEmailChangeWorker(dbInstance, redisClient, app.Logger, app.Mailer, baseURL),
		worker.NewPasswordResetWorker(dbInstance, redisClient, app.Logger, app.Mailer, baseURL),
		worker.NewRegistrationAttemptsWorker(dbInstance, redisClient, app.Logger, app.Mailer, baseURL),
		worker.NewAccountRecoveryWorker(dbInstance, redisClient, app.Logger, app.Mailer, baseURL),
		// worker.NewBriefSubmittedWorker(dbInstance, app.Publisher, app.Logger, transactor),
		// worker.NewNotificationWorker(dbInstance, app.Publisher, app.Logger, transactor),
	}

	for _, w := range workers {
		app.Wg.Add(1)
		go app.handleWorker(ctx, w)
	}

	app.Logger.Info("starting worker", nil)

	// graceful shutdown при SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()
	app.Wg.Wait()
}

func (app *App) handleWorker(ctx context.Context, w worker.Worker) {
	defer app.Wg.Done()
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					app.Logger.Error(fmt.Errorf("worker panic: %v\n%s", r, debug.Stack()), nil)
				}
			}()
			w.Run(ctx)
		}()
		if ctx.Err() != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func getRedis() (*redis.Client, error) {
	return lredis.InitRedis(env.GetStringEnv("REDIS_ADDR", "redis:6379"), env.GetStringEnv("REDIS_PASSWORD", ""))
}

func getAppConfig(environment string) (Config, error) {
	cfg := Config{}
	cfg.Env = environment

	dsn, err := db.BuildDSNFromEnv()
	if err != nil {
		return cfg, fmt.Errorf("failed to resolve database DSN: %w", err)
	}

	maxOpenConns, minOpenConns, maxIdleTime, maxLifetime := getDatabaseConfig()

	cfg.Database.DSN = dsn
	cfg.Database.MaxIdleTime = maxIdleTime
	cfg.Database.MaxOpenConns = maxOpenConns
	cfg.Database.MinOpenConns = minOpenConns
	cfg.Database.MaxLifetime = maxLifetime

	return cfg, nil
}

func getDatabaseConfig() (maxOpenConns, minOpenConns int, maxIdleTime, maxLifetime string) {
	maxOpenConns = max(env.GetIntEnv("DB_OPEN_CONNECTION_LIMIT", 25), runtime.NumCPU()*4)
	minOpenConns = env.GetIntEnv("DB_MIN_CONNECTION_LIMIT", 2)
	maxIdleTime = env.GetStringEnv("DB_MAX_IDLE_TIME", "30m")
	maxLifetime = env.GetStringEnv("DB_MAX_LIFETIME", "1h")
	return maxOpenConns, minOpenConns, maxIdleTime, maxLifetime
}

func getSMTP(environment string) (SMTP, error) {
	smtp := SMTP{
		host:     env.GetStringEnv("SMTP_HOST", "stub"),
		username: env.GetStringEnv("SMTP_USERNAME", ""),
		password: env.GetStringEnv("SMTP_PASSWORD", ""),
		sender:   env.GetStringEnv("SMTP_SENDER", ""),
		port:     env.GetIntEnv("SMTP_PORT", 587),
	}

	if smtp.host == "" || (smtp.host == "stub" && environment == "production") {
		return smtp, fmt.Errorf("smtp host is not valid")
	}

	if smtp.username == "" {
		return smtp, fmt.Errorf("smtp username is not valid")
	}

	if smtp.password == "" {
		return smtp, fmt.Errorf("smtp password is not valid")
	}

	if smtp.sender == "" {
		return smtp, fmt.Errorf("smtp sender is not valid")
	}

	return smtp, nil
}
