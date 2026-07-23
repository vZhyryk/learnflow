package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"learnflow_backend/internal/infrastructure/bootstrap"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/env"
	"learnflow_backend/internal/worker"
)

func main() {
	environment := env.GetStringEnv("ENVIRONMENT", "production")
	jsonLogger := bootstrap.NewLogger(environment)

	appConfig, err := getAppConfig(environment)
	if err != nil {
		jsonLogger.Fatal(err, nil)
	}

	dbInstance, redisClient, cleanup := bootstrap.MustInitInfra(appConfig.Database, jsonLogger)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApp(appConfig, jsonLogger, dbInstance, redisClient)

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
		worker.NewOutboxCleanupWorker(dbInstance, app.Logger, 24*time.Hour),
		// worker.NewBriefSubmittedWorker(dbInstance, app.Publisher, app.Logger, transactor),
		// worker.NewNotificationWorker(dbInstance, app.Publisher, app.Logger, transactor),
	}

	for _, w := range workers {
		app.Wg.Add(1)
		go func(w worker.Worker) {
			defer app.Wg.Done()
			worker.RunWithRecovery(ctx, app.Logger, w)
		}(w)
	}

	app.Logger.Info("starting worker", nil)

	// graceful shutdown при SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	cancel()
	app.Wg.Wait()
}

func getAppConfig(environment string) (Config, error) {
	cfg := Config{}
	cfg.Env = environment

	var err error
	cfg.Database, err = bootstrap.LoadDatabaseConfig()
	if err != nil {
		return cfg, err
	}

	if err := getMailerConfig(&cfg, environment); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getMailerConfig(cfg *Config, environment string) error {
	cfg.SMTP = SMTP{
		Host:     env.GetStringEnv("SMTP_HOST", "stub"),
		Username: env.GetStringEnv("SMTP_USERNAME", ""),
		Password: env.GetStringEnv("SMTP_PASSWORD", ""),
		Sender:   env.GetStringEnv("SMTP_SENDER", ""),
		Port:     env.GetIntEnv("SMTP_PORT", 587),
	}

	if cfg.SMTP.Host == "" || (cfg.SMTP.Host == "stub" && environment == "production") {
		return fmt.Errorf("smtp host is not valid")
	}

	if cfg.SMTP.Username == "" {
		return fmt.Errorf("smtp username is not valid")
	}

	if cfg.SMTP.Password == "" {
		return fmt.Errorf("smtp password is not valid")
	}

	if cfg.SMTP.Sender == "" {
		return fmt.Errorf("smtp sender is not valid")
	}

	return nil
}
