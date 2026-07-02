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

// newTestOutboxPoller assembles an OutboxPoller from a mock runner and publisher,
// with a discard logger and a no-op transactor — the parts every test needs but none
// varies on. A nil publisher is replaced with an empty mockPublisher.
func newTestOutboxPoller(runner *testutil.MockQueryRunner, publisher *mockPublisher) *OutboxPoller {
	if publisher == nil {
		publisher = &mockPublisher{}
	}
	return NewOutboxPoller(runner, publisher, testutil.NewTestLogger(), testutil.NoopTransactor{})
}

func fakeOutboxEntryScan(entry OutboxEntry) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = entry.ID
		*castEventType(dest[1], 1) = entry.EventType
		*testutil.CastStr(dest[2], 2) = entry.PayloadJSON
		return nil
	}
}

func TestGetOutboxList(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When the query fails", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			p := newTestOutboxPoller(runner, nil)

			_, err := p.getOutboxList(context.Background())

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
			p := newTestOutboxPoller(runner, nil)

			_, err := p.getOutboxList(context.Background())

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
			p := newTestOutboxPoller(runner, nil)

			_, err := p.getOutboxList(context.Background())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return valid entries", func() {
			entry1 := OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered, PayloadJSON: `{"a":"b"}`}
			entry2 := OutboxEntry{ID: "entry-2", EventType: events.EventPasswordReset, PayloadJSON: `{"c":"d"}`}

			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: []*testutil.MockRow{
						{ScanFn: fakeOutboxEntryScan(entry1)},
						{ScanFn: fakeOutboxEntryScan(entry2)},
					}}, nil
				},
			}
			p := newTestOutboxPoller(runner, nil)

			entries, err := p.getOutboxList(context.Background())

			So(err, ShouldBeNil)
			So(entries, ShouldHaveLength, 2)
			So(entries[0], ShouldResemble, entry1)
			So(entries[1], ShouldResemble, entry2)
		})
	})
}

func TestHandleOutboxEntryUnknownEventType(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When the event_type is unknown", func() {
			var gotQuery string
			var gotArgs []any
			runner := &testutil.MockQueryRunner{
				ExecFn: capturingExecFn(&gotQuery, &gotArgs),
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					t.Fatal("Publish should not be called for an unknown event type")
					return nil
				},
			}
			p := newTestOutboxPoller(runner, publisher)

			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: "bogus.type"})

			So(gotQuery, ShouldEqual, queryMarkFailed)
			So(gotArgs[0], ShouldEqual, "entry-1")
		})
	})
}

func TestHandleOutboxEntryPublishFails(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When Publish fails", func() {
			var gotQuery string
			var gotArgs []any
			runner := &testutil.MockQueryRunner{
				ExecFn: capturingExecFn(&gotQuery, &gotArgs),
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					return testutil.ErrRedisUnavailable
				},
			}
			p := newTestOutboxPoller(runner, publisher)

			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered})

			So(gotQuery, ShouldEqual, queryMarkFailed)
			So(gotArgs[0], ShouldEqual, "entry-1")
		})
	})
}

func TestHandleOutboxEntryMarkFailedItselfFails(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When the mark-failed update itself fails", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, testutil.ErrDBUnexpected
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					return testutil.ErrRedisUnavailable
				},
			}
			p := newTestOutboxPoller(runner, publisher)

			// Nothing to assert beyond "this does not panic" — the Exec error is only logged.
			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered})
		})
	})
}

func TestHandleOutboxEntryMarkFailedNoRowsAffected(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When the mark-failed update affects no rows", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("UPDATE 0"), nil
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error {
					return testutil.ErrRedisUnavailable
				},
			}
			p := newTestOutboxPoller(runner, publisher)

			// Nothing to assert beyond "this does not panic" — a zero-rows update is only logged.
			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered})
		})
	})
}

func TestHandleOutboxEntrySuccess(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When publish and mark-published both succeed", func() {
			var gotQuery string
			var gotArgs []any
			var publishedType events.EventType
			var publishedPayload any
			runner := &testutil.MockQueryRunner{
				ExecFn: capturingExecFn(&gotQuery, &gotArgs),
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, et events.EventType, payload any) error {
					publishedType = et
					publishedPayload = payload
					return nil
				},
			}
			p := newTestOutboxPoller(runner, publisher)

			p.handleOutboxEntry(context.Background(), OutboxEntry{
				ID: "entry-1", EventType: events.EventUserRegistered, PayloadJSON: `{"a":"b"}`,
			})

			So(publishedType, ShouldEqual, events.EventUserRegistered)
			So(publishedPayload, ShouldResemble, json.RawMessage(`{"a":"b"}`))
			So(gotQuery, ShouldEqual, queryMarkPublished)
			So(gotArgs[0], ShouldEqual, "entry-1")
		})
	})
}

func TestHandleOutboxEntrySuccessNoRowsAffected(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When mark-published succeeds but affects no rows", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("UPDATE 0"), nil
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error { return nil },
			}
			p := newTestOutboxPoller(runner, publisher)

			// Nothing to assert beyond "this does not panic" — a zero-rows update is only logged.
			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered})
		})
	})
}

func TestHandleOutboxEntryMarkPublishedExecFails(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When mark-published Exec itself fails", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, testutil.ErrDBUnexpected
				},
			}
			publisher := &mockPublisher{
				publish: func(_ context.Context, _ events.EventType, _ any) error { return nil },
			}
			p := newTestOutboxPoller(runner, publisher)

			// Nothing to assert beyond "this does not panic" — the Exec error is only logged.
			p.handleOutboxEntry(context.Background(), OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered})
		})
	})
}

func TestOutboxPollerPoll(t *testing.T) {
	Convey("Given an OutboxPoller", t, func() {
		Convey("When getOutboxList fails inside the transaction", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			p := newTestOutboxPoller(runner, nil)

			// Nothing to assert beyond "this does not panic" — poll only logs transaction errors.
			p.poll(context.Background())
		})

		Convey("When there are pending entries", func() {
			entry := OutboxEntry{ID: "entry-1", EventType: events.EventUserRegistered, PayloadJSON: `{"a":"b"}`}
			queryCalls, publishCalls := 0, 0
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					queryCalls++
					return &testutil.MockRows{
						Rows: []*testutil.MockRow{{ScanFn: fakeOutboxEntryScan(entry)}},
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
			p := newTestOutboxPoller(runner, publisher)

			p.poll(context.Background())

			So(queryCalls, ShouldEqual, 1)
			So(publishCalls, ShouldEqual, 1)
		})
	})
}
