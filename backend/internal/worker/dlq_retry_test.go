package worker

import (
	"context"
	"encoding/json"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func newTestDLQRetryWorker(runner *testutil.MockQueryRunner, publisher *mockPublisher) *Poller[DLQRetryWorker] {
	if publisher == nil {
		publisher = &mockPublisher{}
	}
	return NewDLQRetryWorker(runner, publisher, testutil.NewTestLogger(), testutil.NoopTransactor{})
}

func fakeFailedJobScan(job PollerEntry[DLQRetryWorker]) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = job.ID
		*castEventType(dest[1], 1) = job.EventType
		*testutil.CastStr(dest[2], 2) = job.QueueName
		*testutil.CastStr(dest[3], 3) = job.PayloadJSON
		*testutil.CastInt(dest[4], 4) = job.AttemptCount
		return nil
	}
}

func TestGetFailedJobs(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When the query fails", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: testutil.AlwaysFailsQuery,
			}
			w := newTestDLQRetryWorker(runner, nil)

			_, err := w.getList(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "query")
		})

		Convey("When a row fails to scan", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{
						Rows: []*testutil.MockRow{
							{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}},
					}, nil
				},
			}
			w := newTestDLQRetryWorker(runner, nil)

			_, err := w.getList(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{
						Rows: nil, RowsErr: testutil.ErrDBUnexpected,
					}, nil
				},
			}
			w := newTestDLQRetryWorker(runner, nil)

			_, err := w.getList(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return valid entries", func() {
			payload1, payload2 := `{"a":"b"}`, `{"c":"d"}`
			job1 := PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, QueueName: "email_queue", PayloadJSON: payload1, AttemptCount: 1}
			job2 := PollerEntry[DLQRetryWorker]{ID: "job-2", EventType: events.EventPasswordReset, QueueName: "email_queue", PayloadJSON: payload2, AttemptCount: 2}

			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: []*testutil.MockRow{
						{ScanFn: fakeFailedJobScan(job1)},
						{ScanFn: fakeFailedJobScan(job2)},
					}}, nil
				},
			}
			w := newTestDLQRetryWorker(runner, nil)

			entries, err := w.getList(context.Background())

			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 2)
			So(entries[0], ShouldResemble, job1)
			So(entries[1], ShouldResemble, job2)
		})
	})
}

func TestHandleFailedJobEntryUnknownEventType(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When the event_type is unknown", func() {
			var gotQuery string
			var gotArgs []any
			runner := &testutil.MockQueryRunner{ExecFn: capturingExecFn(&gotQuery, &gotArgs)}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					t.Fatal("Publish should not be called for an unknown event type")
					return nil
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: "bogus.type"})

			So(gotQuery, ShouldEqual, markAsRetriedFailedSQL)
			So(gotArgs[0], ShouldEqual, "job-1")
			So(gotArgs[1], ShouldContainSubstring, "unknown event_type")
		})
	})
}

func TestHandleFailedJobEntryPublishFails(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When Publish fails, the entry is marked failed (no retry within this call)", func() {
			var gotQuery string
			var gotArgs []any
			runner := &testutil.MockQueryRunner{ExecFn: capturingExecFn(&gotQuery, &gotArgs)}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					return testutil.ErrRedisUnavailable
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			payload := `{"a":"b"}`
			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, PayloadJSON: payload})

			So(gotQuery, ShouldEqual, markAsRetriedFailedSQL)
			So(gotArgs[0], ShouldEqual, "job-1")
			So(gotArgs[1], ShouldContainSubstring, "publish")
		})
	})
}

func TestHandleFailedJobEntryMarkFailedItselfFails(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When the mark-failed update itself fails", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: testutil.AlwaysFailsExec,
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					return testutil.ErrRedisUnavailable
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			// Nothing to assert beyond "this does not panic" — the Exec error is only logged.
			payload := `{"a":"b"}`
			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, PayloadJSON: payload})
		})
	})
}

func TestHandleFailedJobEntrySuccess(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When retry succeeds", func() {
			var gotQuery string
			var gotArgs []any
			var publishedType events.EventType
			var publishedPayload any
			runner := &testutil.MockQueryRunner{ExecFn: capturingExecFn(&gotQuery, &gotArgs)}
			publisher := &mockPublisher{
				publish: func(_ context.Context, et events.EventType, payload any) error {
					publishedType = et
					publishedPayload = payload
					return nil
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			payload := `{"a":"b"}`
			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{
				ID: "job-1", EventType: events.EventUserRegistered, PayloadJSON: payload,
			})

			So(publishedType, ShouldEqual, events.EventUserRegistered)
			So(publishedPayload, ShouldResemble, json.RawMessage(`{"a":"b"}`))
			So(gotQuery, ShouldEqual, markAsRetriedSuccessSQL)
			So(gotArgs[0], ShouldEqual, "job-1")
		})
	})
}

func TestHandleFailedJobEntrySuccessNoRowsAffected(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When retry succeeds but the update affects no rows", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("UPDATE 0"), nil
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error { return nil },
			}
			w := newTestDLQRetryWorker(runner, publisher)

			// Nothing to assert beyond "this does not panic" — a zero-rows update is only logged.
			payload := `{"a":"b"}`
			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, PayloadJSON: payload})
		})
	})
}

func TestHandleFailedJobEntrySuccessButMarkRetriedExecFails(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When Publish succeeds but the mark-retried-success update fails", func() {
			var publishedType events.EventType
			var publishedPayload any
			runner := &testutil.MockQueryRunner{
				ExecFn: testutil.AlwaysFailsExec,
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, et events.EventType, payload any) error {
					publishedType = et
					publishedPayload = payload
					return nil
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			payload := `{"a":"b"}`
			// Nothing to assert beyond "this does not panic" — the Exec error is only logged, and the
			// transaction still commits at the poll() level despite this failure (see handleEntry's
			// void signature).
			w.handleEntry(context.Background(), PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, PayloadJSON: payload})

			So(publishedType, ShouldEqual, events.EventUserRegistered)
			So(publishedPayload, ShouldResemble, json.RawMessage(`{"a":"b"}`))
		})
	})
}

func TestDLQRetryWorkerPoll(t *testing.T) {
	Convey("Given a DLQRetryWorker", t, func() {
		Convey("When getList fails inside the transaction", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: testutil.AlwaysFailsQuery,
			}
			w := newTestDLQRetryWorker(runner, nil)

			// Nothing to assert beyond "this does not panic" — poll only logs transaction errors.
			w.poll(context.Background())
		})

		Convey("When there are pending entries", func() {
			payload := `{"a":"b"}`
			job := PollerEntry[DLQRetryWorker]{ID: "job-1", EventType: events.EventUserRegistered, QueueName: "email_queue", PayloadJSON: payload, AttemptCount: 1}
			queryCalls, publishCalls := 0, 0
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					queryCalls++
					return &testutil.MockRows{
						Rows: []*testutil.MockRow{{ScanFn: fakeFailedJobScan(job)}},
					}, nil
				},
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("UPDATE 1"), nil
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					publishCalls++
					return nil
				},
			}
			w := newTestDLQRetryWorker(runner, publisher)

			w.poll(context.Background())

			So(queryCalls, ShouldEqual, 1)
			So(publishCalls, ShouldEqual, 1)
		})
	})
}
