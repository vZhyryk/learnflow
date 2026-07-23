package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

// DLQRetryWorker is a phantom type parameterizing Poller[T] for failed_jobs, mirroring
// OutboxPoller for event_outbox.
type DLQRetryWorker struct{}

const (
	selectFailedJobSQL = `
			SELECT id, event_type, queue_name, payload_json, attempt_count
			FROM failed_jobs
			WHERE resolved_at IS NULL AND attempt_count < 6
			ORDER BY failed_at ASC
			LIMIT 50
			FOR UPDATE SKIP LOCKED
		`
	markAsRetriedSuccessSQL = `
	  		UPDATE failed_jobs
  				SET resolved_at = now(),
				resolution_note = 'retried_successfully',
				updated_at = now()
			WHERE id = $1
	`

	markAsRetriedFailedSQL = `
  			UPDATE failed_jobs
			SET attempt_count     = attempt_count + 1,
				error_message     = $2,
				resolution_note   = CASE WHEN attempt_count + 1 >= 6 THEN 'max_retries_exceeded' ELSE NULL END,
				resolved_at       = CASE WHEN attempt_count + 1 >= 6 THEN now() ELSE NULL END,
				updated_at        = now()
			WHERE id = $1
	`
)

// scanFailedJob reads the 5 columns selectFailedJobSQL selects — failed_jobs has
// queue_name/attempt_count, unlike event_outbox.
func scanFailedJob(row entryScanner) (PollerEntry[DLQRetryWorker], error) {
	var entry PollerEntry[DLQRetryWorker]
	err := row.Scan(&entry.ID, &entry.EventType, &entry.QueueName, &entry.PayloadJSON, &entry.AttemptCount)
	return entry, err
}

// NewDLQRetryWorker returns a Poller that retries failed_jobs entries on a 5-minute ticker.
func NewDLQRetryWorker(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *Poller[DLQRetryWorker] {
	return NewPoller[DLQRetryWorker](queryRunner, publisher, jsonLogger, transactor, "failedJobs", 5*time.Minute, SQLList[DLQRetryWorker]{
		selectSQL:      selectFailedJobSQL,
		markSuccessSQL: markAsRetriedSuccessSQL,
		markFailedSQL:  markAsRetriedFailedSQL,
		scanEntry:      scanFailedJob,
	})
}
