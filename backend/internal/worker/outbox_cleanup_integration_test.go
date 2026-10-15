//go:build integration

package worker

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func insertOutboxRowWithStatus(t *testing.T, pool db.QueryRunner, eventType events.EventType, status string, publishedAt *time.Time) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO event_outbox (aggregate_type, aggregate_id, event_type, payload_json, status, published_at)
		 VALUES ($1, gen_random_uuid(), $2, $3, $4, $5)
		 RETURNING id`,
		"outbox_integration_test", string(eventType), `{}`, status, publishedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertOutboxRowWithStatus: %v", err)
	}

	return id
}

func outboxRowExists(t *testing.T, pool db.QueryRunner, id string) bool {
	t.Helper()

	var exists bool
	if err := pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM event_outbox WHERE id = $1)`, id).Scan(&exists); err != nil {
		t.Fatalf("outboxRowExists: %v", err)
	}

	return exists
}

func TestOutboxCleanupWorkerPoll_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	w := NewOutboxCleanupWorker(pool, testutil.NewTestLogger(), 24*time.Hour)

	Convey("Given event_outbox rows backed by real Postgres", t, func() {
		Convey("When a published row is older than the 7-day retention window", func() {
			old := time.Now().Add(-8 * 24 * time.Hour)
			id := insertOutboxRowWithStatus(t, pool, events.EventUserRegistered, "published", &old)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(outboxRowExists(t, pool, id), ShouldBeFalse)
		})

		Convey("When a published row is within the 7-day retention window", func() {
			recent := time.Now().Add(-1 * 24 * time.Hour)
			id := insertOutboxRowWithStatus(t, pool, events.EventUserRegistered, "published", &recent)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(outboxRowExists(t, pool, id), ShouldBeTrue)
		})

		Convey("When a pending row has no published_at, poll leaves it untouched", func() {
			id := insertOutboxRowWithStatus(t, pool, events.EventUserRegistered, "pending", nil)
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE id = $1`, id) //nolint:errcheck // best-effort cleanup
			})

			w.poll(context.Background())

			So(outboxRowExists(t, pool, id), ShouldBeTrue)
		})
	})
}
