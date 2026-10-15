package worker

import (
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

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
type OutboxCleanupWorker string

// NewOutboxCleanupWorker returns an OutboxCleanupWorker that runs on pollInterval.
func NewOutboxCleanupWorker(queryRunner db.QueryRunner, jsonLogger *logger.Logger, pollInterval time.Duration) *CleanupWorker[OutboxCleanupWorker] {
	return NewCleanupWorker[OutboxCleanupWorker](queryRunner, jsonLogger, pollInterval, deleteOutboxPublishedBatchSQL, "OutboxCleanupWorker")
}
