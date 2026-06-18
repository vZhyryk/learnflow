package events

import (
	"context"
	"encoding/json"
	"fmt"
	"learnflow_backend/internal/infrastructure/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OutboxWriter persists domain events to the transactional outbox table.
type OutboxWriter struct {
	db db.QueryRunner
}

// NewOutboxWriter returns a new OutboxWriter backed by the given connection pool.
func NewOutboxWriter(pool *pgxpool.Pool) *OutboxWriter {
	return &OutboxWriter{db: pool}
}

// Emit serializes payload and inserts a pending event into the outbox table.
func (w *OutboxWriter) Emit(ctx context.Context, aggregateType AggregationType, aggregateID string, eventType EventType, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("outbox.Emit marshal: %w", err)
	}
	if _, err := w.db.Exec(ctx, queryInsertOutbox, string(aggregateType), aggregateID, string(eventType), string(data)); err != nil {
		return fmt.Errorf("outbox.Emit insert: %w", err)
	}

	return nil
}
