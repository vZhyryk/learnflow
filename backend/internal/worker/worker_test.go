package worker

import (
	"context"
	"errors"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDLQWriterWrite(t *testing.T) {
	Convey("Given a DLQWriter", t, func() {
		Convey("When processErr is nil", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					t.Fatal("Exec should not be called when processErr is nil")
					return pgconn.CommandTag{}, nil
				},
			}
			d := NewDLQ(runner, testutil.NewTestLogger())

			d.Write(context.Background(), "email.verify", "email_queue", map[string]string{"a": "b"}, nil, 3)
		})

		Convey("When the payload cannot be marshaled", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					t.Fatal("Exec should not be called when marshal fails")
					return pgconn.CommandTag{}, nil
				},
			}
			d := NewDLQ(runner, testutil.NewTestLogger())

			d.Write(context.Background(), "email.verify", "email_queue", make(chan int), errors.New("process failed"), 3)
		})

		Convey("When the insert succeeds", func() {
			var gotArgs []any
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
					gotArgs = args
					return pgconn.CommandTag{}, nil
				},
			}
			d := NewDLQ(runner, testutil.NewTestLogger())

			d.Write(context.Background(), "email.verify", "email_queue", map[string]string{"a": "b"}, errors.New("process failed"), 3)

			So(gotArgs[0], ShouldEqual, "email.verify")
			So(gotArgs[1], ShouldEqual, "email_queue")
			So(gotArgs[2], ShouldEqual, `{"a":"b"}`)
			So(gotArgs[3], ShouldEqual, "process failed")
			So(gotArgs[4], ShouldEqual, 3)
		})
	})
}
