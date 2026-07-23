package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

const outboxCleanupBatchSize = 1000
const deleteOutboxPublishedBatchSQL = `
	DELETE FROM event_outbox
	WHERE id IN (
		SELECT id FROM event_outbox
		WHERE status = 'published' AND published_at < now() - interval '7 days'
		LIMIT 1000
	)
`

// OutboxCleanupWorker periodically deletes old published event_outbox rows so the table
// doesn't grow unbounded.
type OutboxCleanupWorker struct {
	pollInterval time.Duration
	db           db.QueryRunner
	logger       *logger.Logger
}

// NewOutboxCleanupWorker returns an OutboxCleanupWorker that runs on pollInterval.
func NewOutboxCleanupWorker(queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration) *OutboxCleanupWorker {
	return &OutboxCleanupWorker{db: queryRunner, logger: jsonLogger, pollInterval: pollInterval}
}

func (w *OutboxCleanupWorker) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, w.db)
}

// Run polls on pollInterval until ctx is cancelled.
func (w *OutboxCleanupWorker) Run(ctx context.Context) {
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

func (w *OutboxCleanupWorker) poll(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		tag, err := w.queryRunner(ctx).Exec(ctx, deleteOutboxPublishedBatchSQL)
		if err != nil {
			w.logger.Error(fmt.Errorf("outboxCleanupWorker.poll: %w", err), map[string]any{"worker": "outbox_cleanup"})
			return
		}

		if tag.RowsAffected() < outboxCleanupBatchSize {
			return
		}
	}
}
