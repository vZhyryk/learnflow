package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/retry"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	attemptsCount = 3
)

// Config holds the event-type-specific callbacks and metadata for an EmailWorker.
type Config[T any] struct {
	EventType       string
	AggregationType string
	IdempotencyKey  func(T) string
	Validate        func(T) error
	Process         func(payload T, baseURL string, m Mailer) error
}

// EmailWorker is a generic Redis BLPop consumer that validates, deduplicates, and processes email events.
type EmailWorker[T any] struct {
	redisClient *redis.Client
	logger      *logger.Logger
	mailer      Mailer
	dlq         *DLQWriter
	baseURL     string
	cfg         Config[T]
}

// NewEmailWorker returns an EmailWorker wired with the provided dependencies and config.
func NewEmailWorker[T any](
	queryRunner db.QueryRunner,
	redisClient *redis.Client,
	jsonLogger *logger.Logger,
	m Mailer,
	baseURL string,
	cfg Config[T],
) *EmailWorker[T] {
	return &EmailWorker[T]{
		redisClient: redisClient,
		logger:      jsonLogger,
		mailer:      m,
		dlq:         NewDLQ(queryRunner, jsonLogger),
		baseURL:     baseURL,
		cfg:         cfg,
	}
}

// Run starts the BLPop event loop, processing messages until ctx is cancelled.
func (w *EmailWorker[T]) Run(ctx context.Context) {
	for {
		result, err := w.redisClient.BLPop(ctx, 5*time.Second, w.cfg.EventType).Result()
		isCont, isRet := w.handleRunBLPopErrors(err)
		if isCont {
			continue
		}

		if isRet {
			return
		}

		payload, key, msgErr := w.handleMessage(ctx, result[1])
		if errors.Is(msgErr, errAlreadyProcessed) {
			continue
		}
		if msgErr != nil {
			w.logger.Error(msgErr, nil)
			continue
		}

		w.processAndHandleFailure(ctx, payload, key)
	}
}

// processAndHandleFailure runs cfg.Process with retries; on exhaustion it DLQs the event
// and clears the idempotency key so a replay isn't skipped as already-processed.
func (w *EmailWorker[T]) processAndHandleFailure(ctx context.Context, payload *T, key string) {
	if err := retry.Do(ctx, attemptsCount, func() error {
		return w.cfg.Process(*payload, w.baseURL, w.mailer)
	}); err != nil {
		w.logger.Error(err, nil)
		w.dlq.Write(ctx, w.cfg.EventType, w.cfg.AggregationType, payload, err, attemptsCount)
		if delErr := w.redisClient.Del(ctx, key).Err(); delErr != nil {
			w.logger.Error(fmt.Errorf("%s: idempotency key cleanup: %w", w.cfg.EventType, delErr), nil)
		}
	}
}

func (w *EmailWorker[T]) handleMessage(ctx context.Context, message string) (result *T, idempotencyKey string, _ error) {
	var payload T
	if err := json.Unmarshal([]byte(message), &payload); err != nil {
		return nil, "", fmt.Errorf("%s: unmarshal: %w", w.cfg.EventType, err)
	}
	if err := w.cfg.Validate(payload); err != nil {
		return nil, "", err
	}
	key := w.cfg.IdempotencyKey(payload)
	ok, err := w.redisClient.SetNX(ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return nil, "", fmt.Errorf("%s: idempotency check: %w", w.cfg.EventType, err)
	}
	if !ok {
		return nil, "", errAlreadyProcessed
	}
	return &payload, key, nil
}

func (w *EmailWorker[T]) handleRunBLPopErrors(err error) (shouldContinue, shouldReturn bool) {
	if errors.Is(err, redis.Nil) {
		return true, false
	}
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, true
		}
		w.logger.Error(fmt.Errorf("%s: BLPop: %w", w.cfg.EventType, err), nil)
		return true, false
	}

	return false, false
}
