package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

type FailedJob struct {
	ID           string
	EventType    events.EventType
	QueueName    string
	PayloadJSON  string
	AttemptCount int
}

type DLQRetryWorker struct {
	db         db.QueryRunner
	publisher  events.Publisher
	logger     *logger.Logger
	transactor Transactor
}

// NewDLQRetryWorker returns a DLQRetryWorker that retries failed jobs on a 5-minute ticker.
func NewDLQRetryWorker(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *DLQRetryWorker {
	return &DLQRetryWorker{
		db:         queryRunner,
		publisher:  publisher,
		logger:     jsonLogger,
		transactor: transactor,
	}
}

func (rep *DLQRetryWorker) queryRunner(ctx context.Context) db.QueryRunner {
	if tx, ok := db.ExtractTx(ctx); ok {
		return tx
	}
	return rep.db
}

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

func (p *DLQRetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *DLQRetryWorker) poll(ctx context.Context) {
	err := p.transactor.InTransaction(ctx, func(ctx context.Context) error {
		entries, err := p.getFailedJobs(ctx)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			p.handleFailedJobEntry(ctx, entry)
		}

		return nil
	})
	if err != nil {
		p.logger.Error(fmt.Errorf("failedJobs.poll transaction: %w", err), map[string]any{"worker": "dlq_retry"})
	}
}

func (p *DLQRetryWorker) getFailedJobs(ctx context.Context) ([]FailedJob, error) {
	rows, err := p.queryRunner(ctx).Query(ctx, selectFailedJobSQL)
	if err != nil {
		return nil, fmt.Errorf("failedJobs.poll query: %w", err)
	}

	defer rows.Close()

	var entries []FailedJob

	for rows.Next() {
		var entry FailedJob
		if err := rows.Scan(&entry.ID, &entry.EventType, &entry.QueueName, &entry.PayloadJSON, &entry.AttemptCount); err != nil {
			return nil, fmt.Errorf("failedJobs.poll scan: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failedJobs.poll rows: %w", err)
	}

	return entries, nil
}

func (p *DLQRetryWorker) handleFailedJobEntry(ctx context.Context, entry FailedJob) {
	if !events.IsKnownEventType(entry.EventType) {
		p.markEntryFailed(ctx, fmt.Errorf("failedJobs: unknown event_type: %s", entry.EventType), entry.ID)
		return
	}

	if err := p.publisher.Publish(ctx, entry.EventType, json.RawMessage(entry.PayloadJSON)); err != nil {
		p.markEntryFailed(ctx, fmt.Errorf("failedJobs.poll publish: %w", err), entry.ID)
		return
	}

	tag, err := p.queryRunner(ctx).Exec(ctx, markAsRetriedSuccessSQL, entry.ID)
	if err != nil {
		p.logger.Error(fmt.Errorf("failedJobs.poll update status: %w", err), map[string]any{"entry_id": entry.ID, "event_type": string(entry.EventType)})
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("failedJobs.poll update status: not updated"), map[string]any{"entry_id": entry.ID, "event_type": string(entry.EventType)})
	}
}

func (p *DLQRetryWorker) markEntryFailed(ctx context.Context, failErr error, entryID string) {
	p.logger.Error(failErr, map[string]any{"entry_id": entryID})
	tag, err := p.queryRunner(ctx).Exec(ctx, markAsRetriedFailedSQL, entryID, failErr.Error())
	if err != nil {
		p.logger.Error(fmt.Errorf("failedJobs.poll mark failed: %w", err), map[string]any{"entry_id": entryID})
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("failedJobs.poll mark failed: no changes"), map[string]any{"entry_id": entryID})
		return
	}
}
