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

// EmailWorker is a generic Redis BLPop consumer that validates, deduplicates, and
// processes email events. Every email-sending worker is just this type instantiated
// with a different payload T + Config[T] (EventType/Validate/Process) — see
// NewEmailVerificationWorker, NewEmailChangeWorker, NewPasswordResetWorker,
// NewRegistrationAttemptsWorker, NewAccountRecoveryWorker.
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
		isCont, isRet := handleRunBLPopErrors(err, w.cfg.EventType, w.logger)
		if isCont {
			continue
		}

		if isRet {
			return
		}

		// Defensive: a successful (err == nil) BLPOP is guaranteed by the Redis
		// protocol to return exactly [key, value] — this only trips if that
		// contract is ever violated (e.g. a client/library change).
		if len(result) < 2 {
			w.logger.Error(fmt.Errorf("%s: BLPop: unexpected result length %d, want 2", w.cfg.EventType, len(result)), nil)
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

		w.processAndHandleFailure(ctx, payload, key)
	}
}

// processAndHandleFailure: retry.Do (3x, backoff) -> still failing -> DLQ write + Del(key).
func (w *EmailWorker[T]) processAndHandleFailure(ctx context.Context, payload *T, key string) {
	if err := retry.Do(ctx, attemptsCount, func() error {
		return w.cfg.Process(*payload, w.baseURL, w.mailer)
	}); err != nil {
		w.logger.Error(err, nil)
		w.dlq.Write(ctx, w.cfg.EventType, w.cfg.AggregationType, payload, err, attemptsCount)
		// Del: clears the idempotency key so a later DLQ-requeue of this payload isn't
		// skipped as "already processed".
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

	// SetNX: dedupe against the same payload reappearing later (e.g. DLQ requeue) — not
	// concurrent workers, BLPop already hands each element to exactly one caller.
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

func handleRunBLPopErrors(err error, eventType string, jsonLogger *logger.Logger) (shouldContinue, shouldReturn bool) {
	if errors.Is(err, redis.Nil) {
		return true, false
	}
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, true
		}
		jsonLogger.Error(fmt.Errorf("%s: BLPop: %w", eventType, err), nil)
		return true, false
	}

	return false, false
}
