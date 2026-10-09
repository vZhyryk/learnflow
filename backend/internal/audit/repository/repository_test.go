package auditrepository

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateAdminAction(t *testing.T) {
	Convey("Given an admin repository", t, func() {
		var execErr error
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
				gotArgs = args
				return pgconn.NewCommandTag("INSERT 0 1"), execErr
			},
		})
		action := &auditdomain.AdminAction{
			AdminUserID: "admin-1",
			ActionType:  auditdomain.ActionBlockUser,
			TargetType:  auditdomain.TargetUser,
			TargetID:    "user-1",
		}

		Convey("When the insert fails", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.CreateAdminAction(context.Background(), action)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "audit.CreateAdminAction")
		})

		Convey("When it succeeds", func() {
			So(repo.CreateAdminAction(context.Background(), action), ShouldBeNil)
			So(gotArgs, ShouldResemble, []any{"admin-1", auditdomain.ActionBlockUser, auditdomain.TargetUser, "user-1", nil})
		})

		Convey("When the action carries details", func() {
			action.Details = map[string]any{"course_id": "course-1"}
			So(repo.CreateAdminAction(context.Background(), action), ShouldBeNil)
			So(gotArgs[4], ShouldResemble, map[string]any{"course_id": "course-1"})
		})
	})
}

func TestWasDeletedByAdmin(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		var scanErr error
		var deleted bool
		var gotArgs []any
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, args ...any) pgx.Row {
				gotArgs = args
				return &testutil.MockRow{ScanFn: func(dest ...any) error {
					*testutil.CastBool(dest[0], 0) = deleted
					return scanErr
				}}
			},
		})

		Convey("When the query fails, the error is wrapped", func() {
			scanErr = testutil.ErrDBUnexpected
			_, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.WasDeletedByAdmin")
		})

		Convey("When no row comes back, it reports false", func() {
			scanErr = pgx.ErrNoRows
			got, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldBeFalse)
		})

		Convey("When an admin deleted the account, it reports true for that target", func() {
			deleted = true
			got, err := repo.WasDeletedByAdmin(context.Background(), "user-1")
			So(err, ShouldBeNil)
			So(got, ShouldBeTrue)
			So(gotArgs, ShouldResemble, []any{"user-1"})
		})
	})
}

func TestNew(t *testing.T) {
	Convey("New returns an Audit backed by the given pool", t, func() {
		So(New(nil), ShouldNotBeNil)
	})
}

func fakeAdminAction(n int) *auditdomain.AdminAction {
	id := string(rune('a' + n))
	return &auditdomain.AdminAction{
		ID:          "action-" + id,
		AdminUserID: "admin-" + id,
		AdminName:   "Admin " + id,
		ActionType:  auditdomain.ActionBlockUser,
		TargetType:  auditdomain.TargetUser,
		TargetID:    "user-1",
		Details:     map[string]any{"course_id": "course-" + id},
		CreatedAt:   time.Date(2026, 10, n, 12, 0, 0, 0, time.UTC),
	}
}

func fakeAdminActionScan(a *auditdomain.AdminAction) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = a.ID
		*testutil.CastStr(dest[1], 1) = a.AdminUserID
		*testutil.CastStr(dest[2], 2) = a.AdminName
		*testutil.CastEnum[auditdomain.AdminActionType](dest[3], 3) = a.ActionType
		*testutil.CastEnum[auditdomain.AdminTargetType](dest[4], 4) = a.TargetType
		*testutil.CastStr(dest[5], 5) = a.TargetID
		*testutil.CastEnum[map[string]any](dest[6], 6) = a.Details
		*testutil.CastTime(dest[7], 7) = a.CreatedAt
		return nil
	}
}

type getActionsFixture struct {
	repo       *Audit
	count      int
	countErr   error
	queryErr   error
	rows       *testutil.MockRows
	countQuery string
	countArgs  []any
	listQuery  string
	listArgs   []any
}

func newGetActionsFixture() *getActionsFixture {
	f := &getActionsFixture{rows: &testutil.MockRows{}}
	f.repo = newTestRepo(&testutil.MockQueryRunner{
		QueryRowFn: func(_ context.Context, sql string, args ...any) pgx.Row {
			f.countQuery, f.countArgs = sql, args
			return &testutil.MockRow{ScanFn: func(dest ...any) error {
				*testutil.CastInt(dest[0], 0) = f.count
				return f.countErr
			}}
		},
		QueryFn: func(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
			f.listQuery, f.listArgs = sql, args
			return f.rows, f.queryErr
		},
	})

	return f
}

func TestGetInstanceAdminActions(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		f := newGetActionsFixture()
		params := pagination.NewParams(2, 10)
		call := func() ([]*auditdomain.AdminAction, int, error) {
			return f.repo.GetInstanceAdminActions(context.Background(), auditdomain.TargetUser, "user-1", params)
		}

		Convey("When the count query fails", func() {
			f.countErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetInstanceAdminActionsCount")
		})

		Convey("When the list query fails", func() {
			f.queryErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetInstanceAdminActions")
		})

		Convey("When a row fails to scan", func() {
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			f.rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("When there are no entries", func() {
			actions, count, err := call()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
			So(actions, ShouldNotBeNil)
			So(actions, ShouldBeEmpty)
		})

		Convey("When two entries exist", func() {
			first, second := fakeAdminAction(1), fakeAdminAction(2)
			f.count = 25
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeAdminActionScan(first)},
				{ScanFn: fakeAdminActionScan(second)},
			}}

			actions, count, err := call()

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 25)
			So(actions, ShouldResemble, []*auditdomain.AdminAction{first, second})
		})

		Convey("It filters by target and pages the list", func() {
			_, _, err := call()

			So(err, ShouldBeNil)
			So(f.countQuery, ShouldEqual, getInstanceAdminActionsCountSQL)
			So(f.countArgs, ShouldResemble, []any{auditdomain.TargetUser, "user-1"})
			So(f.listQuery, ShouldEqual, getInstanceAdminActionsSQL)
			So(f.listArgs, ShouldResemble, []any{auditdomain.TargetUser, "user-1", params.Limit(), params.Offset()})
		})
	})
}

func fakeFailedJob(n int) *auditdomain.FailedJob {
	note, errMsg := "retried manually", "smtp: connection refused"
	resolved := time.Date(2026, 10, n, 13, 0, 0, 0, time.UTC)
	return &auditdomain.FailedJob{
		ID:             "job-" + string(rune('a'+n)),
		EventType:      "user.blocked",
		QueueName:      "email",
		AttemptCount:   3,
		ErrorMessage:   &errMsg,
		FailedAt:       time.Date(2026, 10, n, 12, 0, 0, 0, time.UTC),
		ResolvedAt:     &resolved,
		ResolutionNote: &note,
		CreatedAt:      time.Date(2026, 10, n, 11, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 10, n, 14, 0, 0, 0, time.UTC),
	}
}

func fakeFailedJobScan(j *auditdomain.FailedJob) func(dest ...any) error {
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = j.ID
		*testutil.CastStr(dest[1], 1) = j.EventType
		*testutil.CastStr(dest[2], 2) = j.QueueName
		*testutil.CastInt(dest[3], 3) = j.AttemptCount
		*testutil.CastPtrStr(dest[4], 4) = j.ErrorMessage
		*testutil.CastTime(dest[5], 5) = j.FailedAt
		*testutil.CastPtrTime(dest[6], 6) = j.ResolvedAt
		*testutil.CastPtrStr(dest[7], 7) = j.ResolutionNote
		*testutil.CastTime(dest[8], 8) = j.CreatedAt
		*testutil.CastTime(dest[9], 9) = j.UpdatedAt
		return nil
	}
}

func TestGetFailedJobs(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		f := newGetActionsFixture()
		params := pagination.NewParams(2, 10)
		call := func() ([]*auditdomain.FailedJob, int, error) {
			return f.repo.GetFailedJobs(context.Background(), params)
		}

		Convey("When the count query fails", func() {
			f.countErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetFailedJobsCount")
		})

		Convey("When the list query fails", func() {
			f.queryErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetFailedJobs")
		})

		Convey("When a row fails to scan", func() {
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			f.rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("When there are no failed jobs", func() {
			jobs, count, err := call()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
			So(jobs, ShouldNotBeNil)
			So(jobs, ShouldBeEmpty)
		})

		Convey("When two failed jobs exist", func() {
			first, second := fakeFailedJob(1), fakeFailedJob(2)
			f.count = 12
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeFailedJobScan(first)},
				{ScanFn: fakeFailedJobScan(second)},
			}}

			jobs, count, err := call()

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 12)
			So(jobs, ShouldResemble, []*auditdomain.FailedJob{first, second})
		})

		Convey("It pages the list and counts the whole table", func() {
			_, _, err := call()

			So(err, ShouldBeNil)
			So(f.countQuery, ShouldEqual, getFailedJobsCountSQL)
			So(f.countArgs, ShouldBeEmpty)
			So(f.listQuery, ShouldEqual, getFailedJobsSQL)
			So(f.listArgs, ShouldResemble, []any{params.Limit(), params.Offset()})
		})
	})
}

func TestGetAdminActions(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		f := newGetActionsFixture()
		params := pagination.NewParams(2, 10)
		call := func() ([]*auditdomain.AdminAction, int, error) {
			return f.repo.GetAdminActions(context.Background(), auditdomain.AdminActionFilter{}, params)
		}

		Convey("When the count query fails", func() {
			f.countErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetAdminActionsCount")
		})

		Convey("When the list query fails", func() {
			f.queryErr = testutil.ErrDBUnexpected
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "audit.GetAdminActions")
		})

		Convey("When a row fails to scan", func() {
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}}}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			f.rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, _, err := call()
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When there are no entries the page is empty, not nil", func() {
			actions, count, err := call()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)
			So(actions, ShouldNotBeNil)
			So(actions, ShouldBeEmpty)
		})

		Convey("When two entries exist, both are returned with the total", func() {
			first, second := fakeAdminAction(1), fakeAdminAction(2)
			f.count = 25
			f.rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeAdminActionScan(first)},
				{ScanFn: fakeAdminActionScan(second)},
			}}

			actions, count, err := call()

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 25)
			So(actions, ShouldResemble, []*auditdomain.AdminAction{first, second})
		})
	})
}

func TestGetAdminActionsFilterArgs(t *testing.T) {
	Convey("Given an audit repository", t, func() {
		f := newGetActionsFixture()
		params := pagination.NewParams(2, 10)
		from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		to := from.Add(24 * time.Hour)
		filter := auditdomain.AdminActionFilter{}
		call := func() error {
			_, _, err := f.repo.GetAdminActions(context.Background(), filter, params)
			return err
		}

		Convey("An empty filter passes NULL for every condition, so nothing is filtered", func() {
			err := call()

			So(err, ShouldBeNil)
			So(f.countQuery, ShouldEqual, getAdminActionsCountSQL)
			So(f.countArgs, ShouldResemble, []any{nil, nil, (*time.Time)(nil), (*time.Time)(nil)})
			So(f.listQuery, ShouldEqual, getAdminActionsSQL)
			So(f.listArgs, ShouldResemble, []any{nil, nil, (*time.Time)(nil), (*time.Time)(nil), params.Limit(), params.Offset()})
		})

		Convey("A full filter passes every value to both the count and the list query", func() {
			filter = auditdomain.AdminActionFilter{AdminUserID: "admin-1", ActionType: auditdomain.ActionBlockUser, From: &from, To: &to}

			err := call()

			So(err, ShouldBeNil)
			So(f.countArgs, ShouldResemble, []any{"admin-1", string(auditdomain.ActionBlockUser), &from, &to})
			So(f.listArgs, ShouldResemble, []any{"admin-1", string(auditdomain.ActionBlockUser), &from, &to, params.Limit(), params.Offset()})
		})
	})
}
