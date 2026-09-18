package worker

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func ptrEntityType(et admindomain.EntityType) *admindomain.EntityType { return &et }
func ptrStr(s string) *string                                         { return &s }

func TestAnnouncementFanOutWorkerParseMessage(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When the message is invalid JSON", func() {
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, testutil.UnreachableRedis(), &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, _, key, err := w.parseMessage(context.Background(), "not-json")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unmarshal")
			So(key, ShouldBeEmpty)
		})

		Convey("When validatePayload fails (missing AnnouncementID)", func() {
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, testutil.UnreachableRedis(), &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, _, key, err := w.parseMessage(context.Background(), `{"announcement_id":""}`)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "missing fields")
			So(key, ShouldBeEmpty)
		})

		Convey("When the idempotency check fails (redis unreachable)", func() {
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return &admindomain.Announcement{ID: "ann-1"}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, testutil.UnreachableRedis(), annRep, defaultTestEntityConfigs())

			_, _, key, err := w.parseMessage(context.Background(), `{"announcement_id":"ann-1"}`)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "idempotency check")
			So(key, ShouldBeEmpty)
		})
	})
}

func TestAnnouncementFanOutWorkerValidatePayload(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When AnnouncementID is empty", func() {
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, err := w.validatePayload(context.Background(), events.AnnouncementPayload{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "missing fields")
		})

		Convey("When fetchAnnouncement fails", func() {
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			_, err := w.validatePayload(context.Background(), events.AnnouncementPayload{AnnouncementID: "ann-1"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fetchAnnouncement")
		})

		Convey("When checkEntityExists fails", func() {
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return &admindomain.Announcement{ID: "ann-1", EntityType: ptrEntityType("bogus")}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			_, err := w.validatePayload(context.Background(), events.AnnouncementPayload{AnnouncementID: "ann-1"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid announcement entity type")
		})

		Convey("When everything succeeds (platform-wide announcement)", func() {
			want := &admindomain.Announcement{ID: "ann-1"}
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return want, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			got, err := w.validatePayload(context.Background(), events.AnnouncementPayload{AnnouncementID: "ann-1"})

			So(err, ShouldBeNil)
			So(got, ShouldEqual, want)
		})
	})
}

func TestAnnouncementFanOutWorkerFetchAnnouncement(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When the repository returns an error", func() {
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			_, err := w.fetchAnnouncement(context.Background(), "ann-1")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fetchAnnouncement")
		})

		Convey("When the announcement doesn't exist (nil, nil)", func() {
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return nil, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			_, err := w.fetchAnnouncement(context.Background(), "ann-1")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "doesn't exists")
		})

		Convey("When the announcement is found", func() {
			want := &admindomain.Announcement{ID: "ann-1"}
			annRep := &mockAnnouncementRepo{
				getAnnouncementByID: func(_ context.Context, _ string) (*admindomain.Announcement, error) {
					return want, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, annRep, defaultTestEntityConfigs())

			got, err := w.fetchAnnouncement(context.Background(), "ann-1")

			So(err, ShouldBeNil)
			So(got, ShouldEqual, want)
		})
	})
}

func TestAnnouncementFanOutWorkerCheckEntityExists(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

		Convey("When EntityType is nil (platform-wide)", func() {
			err := w.checkEntityExists(context.Background(), &admindomain.Announcement{})
			So(err, ShouldBeNil)
		})

		Convey("When EntityType is unknown", func() {
			err := w.checkEntityExists(context.Background(), &admindomain.Announcement{
				EntityType: ptrEntityType("bogus"), EntityID: ptrStr("x"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid announcement entity type")
		})

		Convey("When checkExists returns an error", func() {
			w.entityConfigs[admindomain.CourseEntityType] = entityConfig{
				checkExists: func(_ context.Context, _ string) (bool, error) { return false, testutil.ErrDBUnexpected },
			}
			err := w.checkEntityExists(context.Background(), &admindomain.Announcement{
				EntityType: ptrEntityType(admindomain.CourseEntityType), EntityID: ptrStr("course-1"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid payload")
		})

		Convey("When the entity doesn't exist", func() {
			w.entityConfigs[admindomain.CourseEntityType] = entityConfig{
				checkExists: func(_ context.Context, _ string) (bool, error) { return false, nil },
			}
			err := w.checkEntityExists(context.Background(), &admindomain.Announcement{
				EntityType: ptrEntityType(admindomain.CourseEntityType), EntityID: ptrStr("course-1"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "doesn't exists")
		})

		Convey("When the entity exists", func() {
			err := w.checkEntityExists(context.Background(), &admindomain.Announcement{
				EntityType: ptrEntityType(admindomain.CourseEntityType), EntityID: ptrStr("course-1"),
			})
			So(err, ShouldBeNil)
		})
	})
}

func TestAnnouncementFanOutWorkerGenerateIdempotencyKey(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

		Convey("When generating the idempotency key", func() {
			key := w.generateIdempotencyKey(events.AnnouncementPayload{AnnouncementID: "ann-1"})
			So(key, ShouldEqual, "announcement:approved:ann-1")
		})
	})
}

func TestAnnouncementFanOutWorkerProcessAndHandleFailureSuccess(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When fanOut succeeds on the first attempt", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: nil}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			// Nothing to assert beyond "this does not panic and doesn't call DLQ/Del" —
			// no redisClient is set, so a Del call here would panic on a nil client.
			w.processAndHandleFailure(context.Background(), &events.AnnouncementPayload{AnnouncementID: "ann-1"}, &admindomain.Announcement{ID: "ann-1"}, "key")
		})
	})
}

func TestRowPlaceholders(t *testing.T) {
	Convey("rowPlaceholders", t, func() {
		Convey("When building placeholders for a single row", func() {
			So(rowPlaceholders(1, 2), ShouldEqual, "($1, $2)")
		})

		Convey("When building placeholders starting past $1", func() {
			So(rowPlaceholders(3, 2), ShouldEqual, "($3, $4)")
		})
	})
}

func TestAnnouncementFanOutWorkerBuildInsertBatch(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

		Convey("When building an insert batch for 2 recipients", func() {
			query, args := w.buildInsertBatch([]string{"user-1", "user-2"}, &admindomain.Announcement{ID: "ann-1"})

			So(query, ShouldContainSubstring, "($1, $2),\n($3, $4)")
			So(args, ShouldResemble, []any{"user-1", "ann-1", "user-2", "ann-1"})
		})
	})
}

func TestAnnouncementFanOutWorkerInsertBatch(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When Exec fails", func() {
			runner := &testutil.MockQueryRunner{ExecFn: testutil.AlwaysFailsExec}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.insertBatch(context.Background(), "INSERT ...", nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "insertBatch")
		})

		Convey("When some rows are skipped by ON CONFLICT DO NOTHING (retry dedup), it's not an error", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("INSERT 0 1"), nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.insertBatch(context.Background(), "INSERT ...", nil)

			So(err, ShouldBeNil)
		})

		Convey("When all rows are inserted", func() {
			runner := &testutil.MockQueryRunner{
				ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("INSERT 0 2"), nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.insertBatch(context.Background(), "INSERT ...", nil)

			So(err, ShouldBeNil)
		})
	})
}

func TestAnnouncementFanOutWorkerBuildRecipientQuery(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

		Convey("When the announcement is platform-wide (EntityType nil)", func() {
			query, args, err := w.buildRecipientQuery(&admindomain.Announcement{})

			So(err, ShouldBeNil)
			So(query, ShouldEqual, getAnnouncementUserListSQL)
			So(args, ShouldBeNil)
		})

		Convey("When EntityType is unknown", func() {
			_, _, err := w.buildRecipientQuery(&admindomain.Announcement{EntityType: ptrEntityType("bogus")})

			So(err, ShouldNotBeNil)
		})

		Convey("When EntityType is article (no EntityID arg)", func() {
			query, args, err := w.buildRecipientQuery(&admindomain.Announcement{
				EntityType: ptrEntityType(admindomain.ArticleEntityType), EntityID: ptrStr("article-1"),
			})

			So(err, ShouldBeNil)
			So(query, ShouldEqual, getAnnouncementUserListSQL)
			So(args, ShouldBeNil)
		})

		Convey("When EntityType is course (takes EntityID arg)", func() {
			query, args, err := w.buildRecipientQuery(&admindomain.Announcement{
				EntityType: ptrEntityType(admindomain.CourseEntityType), EntityID: ptrStr("course-1"),
			})

			So(err, ShouldBeNil)
			So(query, ShouldEqual, getCourseAnnouncementUserListSQL)
			So(args, ShouldResemble, []any{ptrStr("course-1")})
		})
	})
}

func TestAnnouncementFanOutWorkerQueryRecipientsErrors(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When the query fails", func() {
			runner := &testutil.MockQueryRunner{QueryFn: testutil.AlwaysFailsQuery}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, err := w.queryRecipients(context.Background(), "SELECT ...", nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "queryRecipients")
		})

		Convey("When a row fails to scan", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{
						Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}},
					}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, err := w.queryRecipients(context.Background(), "SELECT ...", nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: nil, RowsErr: testutil.ErrDBUnexpected}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			_, err := w.queryRecipients(context.Background(), "SELECT ...", nil)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "queryRecipients")
		})
	})
}

func TestAnnouncementFanOutWorkerQueryRecipientsSuccess(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When rows return valid recipient IDs", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: []*testutil.MockRow{
						{ScanFn: func(dest ...any) error { *testutil.CastStr(dest[0], 0) = "user-1"; return nil }},
						{ScanFn: func(dest ...any) error { *testutil.CastStr(dest[0], 0) = "user-2"; return nil }},
					}}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			got, err := w.queryRecipients(context.Background(), "SELECT ...", nil)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, []string{"user-1", "user-2"})
		})
	})
}

func TestAnnouncementFanOutWorkerFanOutErrors(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When the entity type is invalid", func() {
			w := newTestAnnouncementFanOutWorker(&testutil.MockQueryRunner{}, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.fanOut(context.Background(), &admindomain.Announcement{EntityType: ptrEntityType("bogus")})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fanOut")
		})

		Convey("When queryRecipients fails", func() {
			runner := &testutil.MockQueryRunner{QueryFn: testutil.AlwaysFailsQuery}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.fanOut(context.Background(), &admindomain.Announcement{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fanOut")
		})

		Convey("When the batch insert fails", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: []*testutil.MockRow{
						{ScanFn: func(dest ...any) error { *testutil.CastStr(dest[0], 0) = "user-1"; return nil }},
					}}, nil
				},
				ExecFn: testutil.AlwaysFailsExec,
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.fanOut(context.Background(), &admindomain.Announcement{ID: "ann-1"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fanOut")
		})
	})
}

func TestAnnouncementFanOutWorkerFanOutSuccess(t *testing.T) {
	Convey("Given an AnnouncementFanOutWorker", t, func() {
		Convey("When there are no recipients", func() {
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: nil}, nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.fanOut(context.Background(), &admindomain.Announcement{ID: "ann-1"})

			So(err, ShouldBeNil)
		})

		Convey("When recipients are found and the batch insert succeeds", func() {
			var gotQuery string
			var gotArgs []any
			runner := &testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return &testutil.MockRows{Rows: []*testutil.MockRow{
						{ScanFn: func(dest ...any) error { *testutil.CastStr(dest[0], 0) = "user-1"; return nil }},
					}}, nil
				},
				ExecFn: func(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
					gotQuery, gotArgs = sql, args
					return pgconn.NewCommandTag("INSERT 0 1"), nil
				},
			}
			w := newTestAnnouncementFanOutWorker(runner, nil, &mockAnnouncementRepo{}, defaultTestEntityConfigs())

			err := w.fanOut(context.Background(), &admindomain.Announcement{ID: "ann-1"})

			So(err, ShouldBeNil)
			So(gotQuery, ShouldContainSubstring, "INSERT INTO announcement_email_deliveries")
			So(gotArgs, ShouldResemble, []any{"user-1", "ann-1"})
		})
	})
}
