package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminrepository "learnflow_backend/internal/admin/repository"
	articlerepository "learnflow_backend/internal/article/repository"
	contentrepository "learnflow_backend/internal/content/repository"
	courserepository "learnflow_backend/internal/courses/repository"
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
		jsonLogger.Fatal(errors.New("BASE_URL env var is required and must not be empty"), nil)
	}

	adminRepo := adminrepository.NewRepository(dbInstance)
	contentRepo := contentrepository.NewRepository(dbInstance)
	articleRepo := articlerepository.NewRepository(dbInstance)
	courseRepo := courserepository.NewRepository(dbInstance)

	cleanUpPollInterval := 24 * time.Hour

	workers := []worker.Worker{
		worker.NewOutboxPoller(dbInstance, app.Publisher, app.Logger, transactor),
		worker.NewDLQRetryWorker(dbInstance, app.Publisher, app.Logger, transactor),
		worker.NewAnnouncementDeliveryPoller(dbInstance, app.Publisher, app.Logger, transactor),
		worker.NewEmailVerificationWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
		worker.NewEmailChangeWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
		worker.NewPasswordResetWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
		worker.NewRegistrationAttemptsWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
		worker.NewAccountRecoveryWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
		worker.NewOutboxCleanupWorker(dbInstance, app.Logger, cleanUpPollInterval),
		worker.NewAnnouncementCleanUpWorker(dbInstance, app.Logger, cleanUpPollInterval),
		worker.NewAnnouncementFanOutWorker(dbInstance, redisClient.Client, app.Logger, adminRepo, contentRepo, courseRepo, articleRepo),
		worker.NewAnnouncementDeliveryWorker(dbInstance, redisClient.Client, app.Logger, app.Mailer, baseURL),
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
		return errors.New("SMTP_HOST env var is required in production and must not be the stub value")
	}

	if cfg.SMTP.Username == "" {
		return errors.New("SMTP_USERNAME env var is required and must not be empty")
	}

	if cfg.SMTP.Password == "" {
		return errors.New("SMTP_PASSWORD env var is required and must not be empty")
	}

	if cfg.SMTP.Sender == "" {
		return errors.New("SMTP_SENDER env var is required and must not be empty")
	}

	return nil
}
