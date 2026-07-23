package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"time"
)

// OutboxPoller is a phantom type parameterizing Poller[T] for event_outbox — it carries
// no fields of its own, it only keeps Poller[OutboxPoller] distinct from Poller[FailedJob]
// at the type level.
type OutboxPoller struct{}

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
    	SET status = 'failed', attempt_count = attempt_count + 1, updated_at = NOW(), last_error = $2
    	WHERE id = $1
	`
)

func scanOutboxEntry(row entryScanner) (PollerEntry[OutboxPoller], error) {
	var entry PollerEntry[OutboxPoller]
	err := row.Scan(&entry.ID, &entry.EventType, &entry.PayloadJSON)
	return entry, err
}

// NewOutboxPoller returns a Poller that reads pending outbox entries and publishes them
// to Redis, polling on a ticker until ctx is cancelled.
func NewOutboxPoller(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *Poller[OutboxPoller] {
	return NewPoller[OutboxPoller](queryRunner, publisher, jsonLogger, transactor, "outboxPoller", 5*time.Second, SQLList[OutboxPoller]{
		selectSQL:      querySelectAndLockPending,
		markSuccessSQL: queryMarkPublished,
		markFailedSQL:  queryMarkFailed,
		scanEntry:      scanOutboxEntry,
	})
}
