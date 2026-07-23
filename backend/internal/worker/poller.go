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

// entryScanner is the minimal Scan capability poll needs from a pgx.Rows cursor —
// decoupled from pgx directly, mirrors courses/repository's rowScanner pattern.
type entryScanner interface {
	Scan(dest ...any) error
}

// PollerEntry is a row fetched by a Poller[T] — T only distinguishes poller instantiations
// at the type level, it does not affect these fields.
type PollerEntry[T any] struct {
	ID           string
	EventType    events.EventType
	PayloadJSON  string
	QueueName    string
	AttemptCount int
}

// SQLList holds the queries and scan logic for one poller instantiation. scanEntry is
// required because pollers over different tables (event_outbox vs failed_jobs) select
// different columns in different order — there is no single fixed Scan shape shared by
// both.
type SQLList[T any] struct {
	selectSQL      string
	markSuccessSQL string
	markFailedSQL  string
	scanEntry      func(row entryScanner) (PollerEntry[T], error)
}

// Poller is a generic transactional-poll worker: on each tick it selects a batch of rows
// under FOR UPDATE SKIP LOCKED, publishes each to Redis, and marks it success/failed.
type Poller[T any] struct {
	db           db.QueryRunner
	publisher    events.Publisher
	logger       *logger.Logger
	transactor   Transactor
	actionName   string
	pollInterval time.Duration
	SQLList      SQLList[T]
}

// NewPoller returns a Poller wired with the given dependencies, queries, and poll interval.
func NewPoller[T any](queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor, actionName string, pollInterval time.Duration, sqlList SQLList[T]) *Poller[T] {
	return &Poller[T]{db: queryRunner, publisher: publisher, logger: jsonLogger, transactor: transactor, actionName: actionName, pollInterval: pollInterval, SQLList: sqlList}
}

func (p *Poller[T]) queryRunner(ctx context.Context) db.QueryRunner {
	return db.FallbackQueryRunner(ctx, p.db)
}

// Run polls on pollInterval until ctx is cancelled.
func (p *Poller[T]) Run(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
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

func (p *Poller[T]) poll(ctx context.Context) {
	err := p.transactor.InTransaction(ctx, func(ctx context.Context) error {
		entries, err := p.getList(ctx)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			p.handleEntry(ctx, entry)
		}

		return nil
	})
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll transaction: %w", p.actionName, err), map[string]any{"worker": "dlq_retry"})
	}
}

func (p *Poller[T]) getList(ctx context.Context) ([]PollerEntry[T], error) {
	rows, err := p.queryRunner(ctx).Query(ctx, p.SQLList.selectSQL)
	if err != nil {
		return nil, fmt.Errorf("%s.poll query: %w", p.actionName, err)
	}

	defer rows.Close()

	var entries []PollerEntry[T]

	for rows.Next() {
		entry, err := p.SQLList.scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("%s.poll scan: %w", p.actionName, err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s.poll rows: %w", p.actionName, err)
	}

	return entries, nil
}

func (p *Poller[T]) handleEntry(ctx context.Context, entry PollerEntry[T]) {
	if !events.IsKnownEventType(entry.EventType) {
		p.markEntryFailed(ctx, fmt.Errorf("%s: unknown event_type: %s", p.actionName, entry.EventType), entry.ID)
		return
	}

	if err := p.publisher.Publish(ctx, entry.EventType, json.RawMessage(entry.PayloadJSON)); err != nil {
		p.markEntryFailed(ctx, fmt.Errorf("%s.poll publish: %w", p.actionName, err), entry.ID)
		return
	}

	tag, err := p.queryRunner(ctx).Exec(ctx, p.SQLList.markSuccessSQL, entry.ID)
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll update status: %w", p.actionName, err), map[string]any{"entry_id": entry.ID, "event_type": string(entry.EventType)})
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("%s.poll update status: not updated", p.actionName), map[string]any{"entry_id": entry.ID, "event_type": string(entry.EventType)})
	}
}

func (p *Poller[T]) markEntryFailed(ctx context.Context, failErr error, entryID string) {
	p.logger.Error(failErr, map[string]any{"entry_id": entryID})
	tag, err := p.queryRunner(ctx).Exec(ctx, p.SQLList.markFailedSQL, entryID, failErr.Error())
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll mark failed: %w", p.actionName, err), map[string]any{"entry_id": entryID})
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("%s.poll mark failed: no changes", p.actionName), map[string]any{"entry_id": entryID})
		return
	}
}
