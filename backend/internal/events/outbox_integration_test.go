//go:build integration

package events

import (
	"context"
	"errors"
	"testing"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

const outboxIntegrationAggregateID = "00000000-0000-0000-0000-000000000001"

// countOutboxRows queries pool directly (never through the transaction under
// test) so it observes only what actually got committed.
func countOutboxRows(t *testing.T, pool db.QueryRunner, eventType EventType) int {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM event_outbox WHERE event_type = $1`, string(eventType)).Scan(&count)
	if err != nil {
		t.Fatalf("countOutboxRows: %v", err)
	}
	return count
}

func TestOutboxWriterEmit_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	transactor := db.NewTransactor(pool)
	writer := NewOutboxWriter(pool)

	Convey("Given an OutboxWriter backed by a real pool, called from inside a transaction", t, func() {
		Convey("When the outer transaction rolls back, Emit's insert is rolled back too — proving Emit wrote through the ctx tx, not a separate pool connection", func() {
			eventType := uniqueEventType("outbox.integration.rollback")
			wantErr := errors.New("forced rollback")

			err := transactor.InTransaction(context.Background(), func(ctx context.Context) error {
				if emitErr := writer.Emit(ctx, AggregationTypeUser, outboxIntegrationAggregateID, eventType, map[string]string{"a": "b"}); emitErr != nil {
					return emitErr
				}
				return wantErr
			})

			So(errors.Is(err, wantErr), ShouldBeTrue)
			So(countOutboxRows(t, pool, eventType), ShouldEqual, 0)
		})

		Convey("When the outer transaction commits, Emit's insert is persisted", func() {
			eventType := uniqueEventType("outbox.integration.commit")
			t.Cleanup(func() {
				pool.Exec(context.Background(), `DELETE FROM event_outbox WHERE event_type = $1`, string(eventType)) //nolint:errcheck // best-effort cleanup
			})

			err := transactor.InTransaction(context.Background(), func(ctx context.Context) error {
				return writer.Emit(ctx, AggregationTypeUser, outboxIntegrationAggregateID, eventType, map[string]string{"a": "b"})
			})

			So(err, ShouldBeNil)
			So(countOutboxRows(t, pool, eventType), ShouldEqual, 1)
		})
	})
}
