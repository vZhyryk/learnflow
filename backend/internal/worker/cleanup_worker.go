package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

const cleanupBatchSize = 1000

// CleanupWorker is a generic ticker-based retention worker: on each tick it deletes rows
// matching cleanUpQuery in batches until a batch affects fewer than cleanupBatchSize rows.
// One generic engine, different table/query per instantiation — see
// NewOutboxCleanupWorker (event_outbox) and NewAnnouncementCleanUpWorker
// (announcement_email_deliveries).
type CleanupWorker[T any] struct {
	pollInterval time.Duration
	cleanUpQuery string
	workerName   string
	db           db.QueryRunner
	logger       *logger.Logger
}

// NewCleanupWorker returns a CleanupWorker wired with the given query and poll interval.
func NewCleanupWorker[T any](queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration, cleanUpQuery, workerName string) *CleanupWorker[T] {
	return &CleanupWorker[T]{db: queryRunner, logger: jsonLogger, pollInterval: pollInterval, cleanUpQuery: cleanUpQuery, workerName: workerName}
}

func (w *CleanupWorker[T]) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, w.db)
}

// Run polls on pollInterval until ctx is cancelled.
func (w *CleanupWorker[T]) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *CleanupWorker[T]) poll(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		tag, err := w.queryRunner(ctx).Exec(ctx, w.cleanUpQuery)
		if err != nil {
			w.logger.Error(fmt.Errorf("%s.poll: %w", w.workerName, err), map[string]any{"worker": w.workerName})
			return
		}

		if tag.RowsAffected() < cleanupBatchSize {
			return
		}
	}
}
