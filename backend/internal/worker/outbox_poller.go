package worker

import (
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/shared/repository"
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
    	WHERE status = 'pending' AND available_at <= NOW()
    	ORDER BY created_at
    	LIMIT 100
    	FOR UPDATE SKIP LOCKED
	`

	// queryMarkPublished also drops raw_token: a published row must not keep a usable one-time token
	// in the table (or its backups) for the whole retention period.
	queryMarkPublished = `
    	UPDATE event_outbox
    	SET status = 'published', attempt_count = attempt_count + 1, published_at = NOW(), updated_at = NOW(),
    		payload_json = payload_json - 'raw_token'
    	WHERE id = ANY($1::uuid[])
	`

	queryMarkFailed = `
    	UPDATE event_outbox
  		SET status    = CASE WHEN attempt_count + 1 >= $3 THEN 'failed' ELSE 'pending' END,
      	attempt_count = attempt_count + 1,
      	available_at  = NOW() + make_interval(secs => power(2, attempt_count)),
      	last_error    = $2,
      	updated_at    = NOW()
    	WHERE id = $1
	`
)

func scanOutboxEntry(row repository.RowScanner) (PollerEntry[OutboxPoller], error) {
	var entry PollerEntry[OutboxPoller]
	err := row.Scan(&entry.ID, &entry.EventType, &entry.PayloadJSON)
	return entry, err
}

// NewOutboxPoller returns a Poller that reads pending outbox entries and publishes them
// to Redis, polling on a ticker until ctx is cancelled.
func NewOutboxPoller(queryRunner db.QueryRunner, publisher events.Publisher, jsonLogger *logger.Logger, transactor Transactor) *Poller[OutboxPoller] {
	return NewPoller[OutboxPoller](queryRunner, publisher, jsonLogger, transactor, "outboxPoller", 5*time.Second, 3, SQLList[OutboxPoller]{
		selectSQL:      querySelectAndLockPending,
		markSuccessSQL: queryMarkPublished,
		markFailedSQL:  queryMarkFailed,
		scanEntry:      scanOutboxEntry,
	})
}
