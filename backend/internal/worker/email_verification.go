package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/retry"
	"learnflow_backend/internal/shared/mailer"
	"time"

	"github.com/redis/go-redis/v9"
)

// EmailVerificationWorker sends a verification email when a user registers.
type EmailVerificationWorker struct {
	redisClient *redis.Client
	logger      *logger.Logger
	db          db.QueryRunner
	mailer      *mailer.Mailer
	dlq         *DLQWriter
}

// NewEmailVerificationWorker returns an EmailVerificationWorker wired with the given dependencies.
func NewEmailVerificationWorker(queryRunner db.QueryRunner, redisClient *redis.Client, jsonLogger *logger.Logger, m *mailer.Mailer) *EmailVerificationWorker {
	return &EmailVerificationWorker{db: queryRunner, redisClient: redisClient, logger: jsonLogger, mailer: m, dlq: NewDLQ(queryRunner, jsonLogger)}
}

var errAlreadyProcessed = errors.New("already processed")

// Run consumes user_registered events from Redis and sends verification emails until ctx is cancelled.
func (w *EmailVerificationWorker) Run(ctx context.Context) {
	for {
		result, err := w.redisClient.BLPop(ctx, 5*time.Second, string(events.EventUserRegistered)).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			w.logger.Error(fmt.Errorf("emailWorker: BLPop: %w", err), nil)
			continue
		}

		payload, key, msgErr := w.handleMessage(ctx, result[1])
		if errors.Is(msgErr, errAlreadyProcessed) {
			continue
		}

		if msgErr != nil {
			w.logger.Error(msgErr, nil)
			continue
		}

		if err := retry.Do(ctx, 3, func() error {
			return w.process(*payload)
		}); err != nil {
			w.redisClient.Del(ctx, key)
			w.logger.Error(err, map[string]any{"event": payload.UserID})
			w.dlq.Write(ctx, string(events.EventUserRegistered), string(events.AggregationTypeEmail), payload, err, 3)
		}
	}
}

func (w *EmailVerificationWorker) process(p events.UserRegisteredPayload) error {
	data := map[string]string{"URL": p.URL}
	recipient := mailer.CCuser{Mail: p.Email}

	if err := w.mailer.Send("email_verification.html", data, recipient, nil); err != nil {
		return fmt.Errorf("emailWorker.process: %w", err)
	}
	return nil
}

func (w *EmailVerificationWorker) handleMessage(ctx context.Context, message string) (*events.UserRegisteredPayload, string, error) {
	var payload events.UserRegisteredPayload
	if unmarshalErr := json.Unmarshal([]byte(message), &payload); unmarshalErr != nil {
		return nil, "", fmt.Errorf("emailWorker: unmarshal: %w", unmarshalErr)
	}

	if payload.UserID == "" || payload.Email == "" || payload.URL == "" {
		return nil, "", fmt.Errorf("emailWorker: invalid payload: missing fields")
	}

	key := "processed:email_verification:" + payload.UserID
	ok, err := w.redisClient.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return nil, "", fmt.Errorf("emailWorker: idempotency check: %w", err)
	}
	if !ok {
		return nil, "", errAlreadyProcessed
	}

	return &payload, key, nil
}
