package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
)

// OutboxEntry holds a single row fetched from the event_outbox table.
type OutboxEntry struct {
	ID          string
	EventType   events.EventType
	PayloadJSON string
}

// OutboxPoller reads pending outbox entries and publishes them to Redis.
type OutboxPoller struct {
	db         db.QueryRunner
	publisher  events.Publisher
	logger     *logger.Logger
	transactor Transactor
}

// NewOutboxPoller returns an OutboxPoller wired with the given dependencies.
func NewOutboxPoller(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *OutboxPoller {
	return &OutboxPoller{db: queryRunner, publisher: publisher, logger: jsonLogger, transactor: transactor}
}

func (p *OutboxPoller) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, p.db)
}

const (
	querySelectAndLockPending = `
    	SELECT id, event_type, payload_json
    	FROM event_outbox
    	WHERE status = 'pending' AND (locked_until < now() OR locked_until IS NULL)
    	ORDER BY created_at
    	LIMIT 100
    	FOR UPDATE SKIP LOCKED
	`

	queryMarkPublished = `
    	UPDATE event_outbox
    	SET status = 'published', attempt_count = attempt_count + 1, published_at = NOW(), updated_at = NOW()
    	WHERE id = $1
	`

	queryMarkFailed = `
    	UPDATE event_outbox
    	SET status = 'failed', attempt_count = attempt_count + 1, updated_at = NOW()
    	WHERE id = $1
	`
)

// Run polls the outbox on a 5-second ticker until ctx is cancelled.
func (p *OutboxPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
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

func (p *OutboxPoller) poll(ctx context.Context) {
	err := p.transactor.InTransaction(ctx, func(ctx context.Context) error {
		entries, err := p.getOutboxList(ctx)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			p.handleOutboxEntry(ctx, entry)
		}

		return nil
	})
	if err != nil {
		p.logger.Error(fmt.Errorf("outboxPoller.poll transaction: %w", err), nil)
	}
}

func (p *OutboxPoller) markEntryFailed(ctx context.Context, failErr error, entryID string) {
	p.logger.Error(failErr, nil)
	tag, err := p.queryRunner(ctx).Exec(ctx, queryMarkFailed, entryID)
	if err != nil {
		p.logger.Error(fmt.Errorf("outboxPoller.poll mark failed: %w", err), nil)
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("outboxPoller.poll mark failed: no changes"), nil)
		return
	}
}

func (p *OutboxPoller) getOutboxList(ctx context.Context) ([]OutboxEntry, error) {
	rows, err := p.queryRunner(ctx).Query(ctx, querySelectAndLockPending)
	if err != nil {
		return nil, fmt.Errorf("outboxPoller.poll query: %w", err)
	}

	defer rows.Close()

	var entries []OutboxEntry

	for rows.Next() {
		var entry OutboxEntry
		if err := rows.Scan(&entry.ID, &entry.EventType, &entry.PayloadJSON); err != nil {
			return nil, fmt.Errorf("outboxPoller.poll scan: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("outboxPoller.poll rows: %w", err)
	}

	return entries, nil
}

func (p *OutboxPoller) handleOutboxEntry(ctx context.Context, entry OutboxEntry) {
	if !events.IsKnownEventType(entry.EventType) {
		p.markEntryFailed(ctx, fmt.Errorf("outboxPoller: unknown event_type: %s", entry.EventType), entry.ID)
		return
	}

	if err := p.publisher.Publish(ctx, entry.EventType, json.RawMessage(entry.PayloadJSON)); err != nil {
		p.markEntryFailed(ctx, fmt.Errorf("outboxPoller.poll publish: %w", err), entry.ID)
		return
	}

	tag, err := p.queryRunner(ctx).Exec(ctx, queryMarkPublished, entry.ID)
	if err != nil {
		p.logger.Error(fmt.Errorf("outboxPoller.poll update status: %w", err), nil)
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("outboxPoller.poll update status: not updated"), nil)
	}
}
