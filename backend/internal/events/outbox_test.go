package events_test

import (
	"context"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOutboxWriterEmit(t *testing.T) {
	Convey("Given an OutboxWriter", t, func() {
		Convey("When the payload cannot be marshaled to JSON", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					t.Fatal("Exec should not be called when marshal fails")
					return pgconn.CommandTag{}, nil
				},
			}
			w := events.NewOutboxWriterWithRunner(runner)
			err := w.Emit(context.Background(), events.AggregationTypeUser, "user-123", events.EventUserRegistered, make(chan int))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "marshal")
		})

		Convey("When Exec fails", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: testutil.AlwaysFailsExec,
			}
			w := events.NewOutboxWriterWithRunner(runner)

			err := w.Emit(context.Background(), events.AggregationTypeUser, "user-123", events.EventUserRegistered, map[string]string{"a": "b"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "insert")
		})

		Convey("When the insert succeeds", func() {
			var gotArgs []any
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
					gotArgs = args
					return pgconn.CommandTag{}, nil
				},
			}
			w := events.NewOutboxWriterWithRunner(runner)

			err := w.Emit(context.Background(), events.AggregationTypeUser, "user-123", events.EventUserRegistered, map[string]string{"a": "b"})

			So(err, ShouldBeNil)
			So(gotArgs[0], ShouldEqual, string(events.AggregationTypeUser))
			So(gotArgs[1], ShouldEqual, "user-123")
			So(gotArgs[2], ShouldEqual, string(events.EventUserRegistered))
			So(gotArgs[3], ShouldEqual, `{"a":"b"}`)
		})
	})
}
