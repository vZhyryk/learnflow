package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/repository"
	"time"
)

// failImmediately makes markFailedSQL terminal on the first attempt (non-retryable errors).
const failImmediately = 1

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
	scanEntry      func(row repository.RowScanner) (PollerEntry[T], error)
}

// Poller is a generic transactional-poll worker: on each tick it selects a batch of rows
// under FOR UPDATE SKIP LOCKED, publishes each to Redis, and marks it success/failed.
// One generic engine, different table/queries per instantiation via SQLList[T]:
// NewOutboxPoller (event_outbox -> Redis, every 5s), NewAnnouncementDeliveryPoller and
// NewDLQRetryWorker (failed_jobs -> requeue to Redis, every 5m).
// A failed publish is retried until maxAttemptsCount; markFailedSQL receives
// (id, error text, maxAttemptsCount) and decides retry vs terminal failure.
type Poller[T any] struct {
	db               db.QueryRunner
	publisher        events.Publisher
	logger           *logger.Logger
	transactor       Transactor
	actionName       string
	maxAttemptsCount int
	pollInterval     time.Duration
	SQLList          SQLList[T]
}

// NewPoller returns a Poller wired with the given dependencies, queries, and poll interval.
func NewPoller[T any](queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor, actionName string, pollInterval time.Duration, maxAttemptsCount int, sqlList SQLList[T]) *Poller[T] {
	return &Poller[T]{db: queryRunner, publisher: publisher, logger: jsonLogger, transactor: transactor, actionName: actionName, pollInterval: pollInterval, SQLList: sqlList, maxAttemptsCount: maxAttemptsCount}
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

		p.handleEntries(ctx, entries)

		return nil
	})
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll transaction: %w", p.actionName, err), map[string]any{"worker": p.actionName})
	}
}

func (p *Poller[T]) getList(ctx context.Context) ([]PollerEntry[T], error) {
	rows, err := p.queryRunner(ctx).Query(ctx, p.SQLList.selectSQL)
	if err != nil {
		return nil, fmt.Errorf("%s.poll query: %w", p.actionName, err)
	}

	defer rows.Close()

	return repository.CollectRows(rows, fmt.Sprintf("%s.poll", p.actionName), p.SQLList.scanEntry)
}

// handleEntries publishes each entry, then marks all successfully published ones in one UPDATE.
func (p *Poller[T]) handleEntries(ctx context.Context, entries []PollerEntry[T]) {
	published := make([]string, 0, len(entries))
	for _, entry := range entries {
		if p.handleEntry(ctx, entry) {
			published = append(published, entry.ID)
		}
	}

	p.markPublished(ctx, published)
}

// handleEntry publishes one entry and reports whether it must be marked as published.
// Failures are recorded via markEntryFailed.
func (p *Poller[T]) handleEntry(ctx context.Context, entry PollerEntry[T]) bool {
	if !events.IsKnownEventType(entry.EventType) {
		p.markEntryFailed(ctx, fmt.Errorf("%s.poll: unknown event_type: %s", p.actionName, entry.EventType), entry.ID, failImmediately)
		return false
	}

	if err := p.publisher.Publish(ctx, entry.EventType, json.RawMessage(entry.PayloadJSON)); err != nil {
		p.markEntryFailed(ctx, fmt.Errorf("%s.poll publish: %w", p.actionName, err), entry.ID, p.maxAttemptsCount)
		return false
	}

	return true
}

func (p *Poller[T]) markPublished(ctx context.Context, ids []string) {
	if len(ids) == 0 {
		return
	}

	tag, err := p.queryRunner(ctx).Exec(ctx, p.SQLList.markSuccessSQL, ids)
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll update status: %w", p.actionName, err), map[string]any{"entry_ids": ids})
		return
	}

	if int(tag.RowsAffected()) != len(ids) {
		p.logger.Error(fmt.Errorf("%s.poll update status: updated %d of %d", p.actionName, tag.RowsAffected(), len(ids)), map[string]any{"entry_ids": ids})
	}
}

func (p *Poller[T]) markEntryFailed(ctx context.Context, failErr error, entryID string, maxAttempts int) {
	p.logger.Error(failErr, map[string]any{"entry_id": entryID})
	tag, err := p.queryRunner(ctx).Exec(ctx, p.SQLList.markFailedSQL, entryID, failErr.Error(), maxAttempts)
	if err != nil {
		p.logger.Error(fmt.Errorf("%s.poll mark failed: %w", p.actionName, err), map[string]any{"entry_id": entryID})
		return
	}

	if tag.RowsAffected() == 0 {
		p.logger.Error(fmt.Errorf("%s.poll mark failed: no changes", p.actionName), map[string]any{"entry_id": entryID})
	}
}
