//go:build integration

package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func insertOutboxRow(t *testing.T, pool db.QueryRunner, eventType events.EventType) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO event_outbox (aggregate_type, aggregate_id, event_type, payload_json, status)
		 VALUES ($1, gen_random_uuid(), $2, $3, 'pending')
		 RETURNING id`,
		"outbox_poller_concurrency_test", string(eventType), `{}`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertOutboxRow: %v", err)
	}

	return id
}

func getOutboxStatus(t *testing.T, pool db.QueryRunner, id string) string {
	t.Helper()

	var status string
	if err := pool.QueryRow(context.Background(),
		`SELECT status FROM event_outbox WHERE id = $1`, id).Scan(&status); err != nil {
		t.Fatalf("getOutboxStatus: %v", err)
	}

	return status
}

func TestOutboxPollerPollConcurrentSkipLocked(t *testing.T) {
	pool := testutil.NewTestPool(t)
	transactor := db.NewTransactor(pool)

	Convey("Given a pending event_outbox row and two OutboxPollers backed by real Postgres", t, func() {
		id := insertOutboxRow(t, pool, events.EventUserRegistered)
		t.Cleanup(func() {
			_, err := pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id)
			if err != nil {
				t.Logf("cleanup insertOutboxRow: %v", err)
			}
		})

		var publishCount atomic.Int64
		publisher := &mockPublisher{
			publish: func(_ context.Context, _ events.EventType, _ any) error {
				publishCount.Add(1)
				return nil
			},
		}

		pollerA := NewOutboxPoller(pool, publisher, testutil.NewTestLogger(), transactor)
		pollerB := NewOutboxPoller(pool, publisher, testutil.NewTestLogger(), transactor)

		Convey("When both pollers poll concurrently, only one processes the row", func() {
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				pollerA.poll(context.Background())
			}()
			go func() {
				defer wg.Done()
				pollerB.poll(context.Background())
			}()

			wg.Wait()

			So(publishCount.Load(), ShouldEqual, int64(1))
			So(getOutboxStatus(t, pool, id), ShouldEqual, "published")
		})
	})
}
