package worker

import (
	"context"
	"fmt"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func newTestOutboxCleanupWorker(runner *testutil.MockQueryRunner) *OutboxCleanupWorker {
	return NewOutboxCleanupWorker(runner, testutil.NewTestLogger(), 24*time.Hour)
}

func TestOutboxCleanupWorkerPollExecFails(t *testing.T) {
	Convey("Given an OutboxCleanupWorker", t, func() {
		Convey("When Exec fails", func() {
			calls := 0
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					calls++
					return pgconn.CommandTag{}, testutil.ErrDBUnexpected
				},
			}
			w := newTestOutboxCleanupWorker(runner)

			// Nothing to assert beyond "this does not panic" — the Exec error is only logged.
			w.poll(context.Background())

			So(calls, ShouldEqual, 1)
		})
	})
}

func TestOutboxCleanupWorkerPollPartialBatch(t *testing.T) {
	Convey("Given an OutboxCleanupWorker", t, func() {
		Convey("When a single batch deletes fewer rows than the batch size", func() {
			var gotQuery string
			calls := 0
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
					calls++
					gotQuery = sql
					return pgconn.NewCommandTag("DELETE 3"), nil
				},
			}
			w := newTestOutboxCleanupWorker(runner)

			w.poll(context.Background())

			So(calls, ShouldEqual, 1)
			So(gotQuery, ShouldEqual, deleteOutboxPublishedBatchSQL)
		})
	})
}

func fullThenPartialBatchExecFn(calls *int) func(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
		*calls++
		if *calls < 3 {
			return pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", outboxCleanupBatchSize)), nil
		}
		return pgconn.NewCommandTag("DELETE 0"), nil
	}
}

func TestOutboxCleanupWorkerPollLoopsUntilPartialBatch(t *testing.T) {
	Convey("Given an OutboxCleanupWorker", t, func() {
		Convey("When a full batch is deleted, poll loops until a partial batch is returned", func() {
			calls := 0
			runner := &testutil.MockQueryRunner{ExecFn: fullThenPartialBatchExecFn(&calls)}
			w := newTestOutboxCleanupWorker(runner)

			w.poll(context.Background())

			So(calls, ShouldEqual, 3)
		})
	})
}

func TestOutboxCleanupWorkerPollCtxAlreadyCancelled(t *testing.T) {
	Convey("Given an OutboxCleanupWorker", t, func() {
		Convey("When ctx is already cancelled", func() {
			calls := 0
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					calls++
					return pgconn.NewCommandTag("DELETE 0"), nil
				},
			}
			w := newTestOutboxCleanupWorker(runner)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w.poll(ctx)

			So(calls, ShouldEqual, 0)
		})
	})
}
