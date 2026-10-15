package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

// newTestPoller controls pollInterval directly — NewOutboxPoller hardcodes 5s, too slow for a test.
func newTestPoller(runner *testutil.MockQueryRunner, publisher *mockPublisher, pollInterval time.Duration) *Poller[OutboxPoller] {
	if publisher == nil {
		publisher = &mockPublisher{}
	}
	return &Poller[OutboxPoller]{
		db:           runner,
		publisher:    publisher,
		logger:       testutil.NewTestLogger(),
		transactor:   testutil.NoopTransactor{},
		actionName:   "testPoller",
		pollInterval: pollInterval,
		SQLList: SQLList[OutboxPoller]{
			selectSQL:      querySelectAndLockPending,
			markSuccessSQL: queryMarkPublished,
			markFailedSQL:  queryMarkFailed,
			scanEntry:      scanOutboxEntry,
		},
	}
}

func TestPollerRun(t *testing.T) {
	Convey("Given a Poller", t, func() {
		Convey("When ctx is cancelled, Run stops after ticking at least once", func() {
			var pollCount atomic.Int32
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					pollCount.Add(1)
					return &testutil.MockRows{Rows: nil}, nil
				},
			}
			p := newTestPoller(runner, nil, 5*time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()

			p.Run(ctx)

			So(pollCount.Load(), ShouldBeGreaterThan, 0)
		})

		Convey("When ctx is already cancelled, Run returns without polling", func() {
			var pollCount atomic.Int32
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					pollCount.Add(1)
					return &testutil.MockRows{Rows: nil}, nil
				},
			}
			p := newTestPoller(runner, nil, time.Hour)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			p.Run(ctx)

			So(pollCount.Load(), ShouldEqual, 0)
		})
	})
}
