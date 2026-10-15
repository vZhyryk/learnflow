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

// insertOutboxRow inserts a 'pending' row for the poller's own publish-path tests. For
// arbitrary status/published_at combinations (e.g. cleanup-worker retention tests), use
// insertOutboxRowWithStatus (outbox_cleanup_integration_test.go) instead.
func insertOutboxRow(t *testing.T, pool db.QueryRunner, eventType events.EventType) string {
	t.Helper()
	return insertOutboxRowWithStatus(t, pool, eventType, "pending", nil)
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

func getOutboxRetryState(t *testing.T, pool db.QueryRunner, id string) (status string, attempts int, delayed bool) {
	t.Helper()

	err := pool.QueryRow(context.Background(),
		`SELECT status, attempt_count, available_at > NOW() FROM event_outbox WHERE id = $1`, id,
	).Scan(&status, &attempts, &delayed)
	if err != nil {
		t.Fatalf("getOutboxRetryState: %v", err)
	}

	return status, attempts, delayed
}

func TestOutboxPollerPublishFailureRetries(t *testing.T) {
	pool := testutil.NewTestPool(t)
	transactor := db.NewTransactor(pool)

	Convey("Given a pending event_outbox row and a publisher that always fails", t, func() {
		id := insertOutboxRow(t, pool, events.EventUserRegistered)
		t.Cleanup(func() {
			if _, err := pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id); err != nil {
				t.Logf("cleanup insertOutboxRow: %v", err)
			}
		})

		poller := NewOutboxPoller(pool, unAvailablePublisher(), testutil.NewTestLogger(), transactor)

		Convey("When the first publish fails, the row stays pending with a backoff delay", func() {
			poller.poll(context.Background())

			status, attempts, delayed := getOutboxRetryState(t, pool, id)
			So(status, ShouldEqual, "pending")
			So(attempts, ShouldEqual, 1)
			So(delayed, ShouldBeTrue)
		})

		Convey("When the delay has not elapsed, the next poll skips the row", func() {
			poller.poll(context.Background())
			poller.poll(context.Background())

			_, attempts, _ := getOutboxRetryState(t, pool, id)
			So(attempts, ShouldEqual, 1)
		})

		Convey("When the last allowed attempt fails, the row becomes failed", func() {
			_, err := pool.Exec(context.Background(),
				`UPDATE event_outbox SET attempt_count = $2 WHERE id = $1`, id, poller.maxAttemptsCount-1)
			So(err, ShouldBeNil)

			poller.poll(context.Background())

			status, attempts, _ := getOutboxRetryState(t, pool, id)
			So(status, ShouldEqual, "failed")
			So(attempts, ShouldEqual, poller.maxAttemptsCount)
		})
	})
}

func TestOutboxPollerPublishedRowDropsRawToken(t *testing.T) {
	pool := testutil.NewTestPool(t)
	transactor := db.NewTransactor(pool)

	Convey("Given a pending outbox row carrying a one-time token", t, func() {
		var id string
		err := pool.QueryRow(context.Background(),
			`INSERT INTO event_outbox (aggregate_type, aggregate_id, event_type, payload_json, status)
			 VALUES ('outbox_integration_test', gen_random_uuid(), $1, '{"user_id":"u-1","raw_token":"secret"}', 'pending')
			 RETURNING id`, string(events.EventUserRegistered)).Scan(&id)
		So(err, ShouldBeNil)
		t.Cleanup(func() {
			if _, delErr := pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id); delErr != nil {
				t.Logf("cleanup: %v", delErr)
			}
		})

		Convey("When it is published, the stored payload no longer holds the token", func() {
			poller := NewOutboxPoller(pool, availablePublisher(), testutil.NewTestLogger(), transactor)

			poller.poll(context.Background())

			var hasToken bool
			var hasUser bool
			err := pool.QueryRow(context.Background(),
				`SELECT payload_json ? 'raw_token', payload_json ? 'user_id' FROM event_outbox WHERE id = $1`, id).Scan(&hasToken, &hasUser)
			So(err, ShouldBeNil)
			So(getOutboxStatus(t, pool, id), ShouldEqual, "published")
			So(hasToken, ShouldBeFalse)
			So(hasUser, ShouldBeTrue)
		})
	})
}
