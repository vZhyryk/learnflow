package adminhttp_test

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const auditPath = "/api/v1/admin/audit/actions"

type auditFixture struct {
	*httpFixture
	svcErr     error
	svcActions []*auditdomain.AdminAction
	svcTotal   int
	called     bool
	gotActor   string
	gotType    auditdomain.AdminTargetType
	gotID      string
	gotParams  pagination.Params
}

func newAuditFixture() *auditFixture {
	af := &auditFixture{}
	svc := &mockService{
		getInstanceAdminActions: func(_ context.Context, actorID string, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
			af.called = true
			af.gotActor = actorID
			af.gotType, af.gotID, af.gotParams = targetType, targetID, params
			return af.svcActions, af.svcTotal, af.svcErr
		},
	}
	af.httpFixture = newHTTPFixture(svc, http.MethodGet, auditPath)

	return af
}

func auditQuery(targetType, targetID string) map[string]string {
	return map[string]string{"target_type": targetType, "target_id": targetID}
}

func TestGetInstanceAdminActionsRoute(t *testing.T) {
	Convey("GET "+auditPath, t, func() {
		af := newAuditFixture()

		Convey("A valid query returns the page and the total", func() {
			created := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
			af.svcActions = []*auditdomain.AdminAction{{
				ID: "a-1", AdminUserID: "admin-1", ActionType: auditdomain.ActionBlockUser,
				TargetType: auditdomain.TargetUser, TargetID: validUserID, CreatedAt: created,
			}}
			af.svcTotal = 7
			params := auditQuery("user", validUserID)
			params["page"], params["page_size"] = "2", "5"

			w := testutil.ServeHTTP(af.mux, withUser(af.newReq("", params)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(af.gotType, ShouldEqual, auditdomain.TargetUser)
			So(af.gotID, ShouldEqual, validUserID)
			So(af.gotParams, ShouldResemble, pagination.NewParams(2, 5))

			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 7)
			actions, ok := body["actions"].([]any)
			So(ok, ShouldBeTrue)
			So(actions, ShouldHaveLength, 1)
			first, ok := actions[0].(map[string]any)
			So(ok, ShouldBeTrue)
			So(first["admin_user_id"], ShouldEqual, "admin-1")
			So(first["action_type"], ShouldEqual, "block_user")
			So(first["created_at"], ShouldEqual, "2026-10-06T12:00:00Z")
		})

		Convey("Invalid input is a 400 and never reaches the service", func() {
			cases := map[string]map[string]string{
				"missing target_id":    {"target_type": "user"},
				"target_id not a UUID": auditQuery("user", "nope"),
				"missing target_type":  {"target_id": validUserID},
				"unknown target_type":  auditQuery("video", validUserID),
			}
			for name, query := range cases {
				w := testutil.ServeHTTP(af.mux, withUser(af.newReq("", query)))
				So(name+": "+http.StatusText(w.Code), ShouldEqual, name+": "+http.StatusText(http.StatusBadRequest))
			}
			So(af.called, ShouldBeFalse)
		})

		Convey("A service error maps through the error handler", func() {
			af.svcErr = admindomain.ErrUserNotFound
			w := testutil.ServeHTTP(af.mux, withUser(af.newReq("", auditQuery("user", validUserID))))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("An unexpected service error is a 500", func() {
			af.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(af.mux, withUser(af.newReq("", auditQuery("user", validUserID))))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestGetInstanceAdminActionsRouteEdgeCases(t *testing.T) {
	Convey("GET "+auditPath+" — edge cases", t, func() {
		af := newAuditFixture()

		Convey("Without a user in the context it panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(af.mux, af.newReq("", auditQuery("user", validUserID))) }, ShouldPanic)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				af.mux.ServeHTTP(&errWriter{}, withUser(af.newReq("", auditQuery("user", validUserID))))
			}, ShouldNotPanic)
		})

		Convey("Invalid input and the 400 response write fails → does not panic", func() {
			So(func() {
				af.mux.ServeHTTP(&errWriter{}, withUser(af.newReq("", auditQuery("video", validUserID))))
			}, ShouldNotPanic)
			So(af.called, ShouldBeFalse)
		})
	})
}

const failedJobsPath = "/api/v1/admin/dlq"

func TestGetFailedJobsRoute(t *testing.T) {
	Convey("GET "+failedJobsPath, t, func() {
		var gotParams pagination.Params
		var svcErr error
		jobs := []*auditdomain.FailedJob{}
		svc := &mockService{getFailedJobs: func(_ context.Context, params pagination.Params) ([]*auditdomain.FailedJob, int, error) {
			gotParams = params
			return jobs, 4, svcErr
		}}
		f := newHTTPFixture(svc, http.MethodGet, failedJobsPath)

		Convey("A valid request returns the page and the total", func() {
			errMsg := "smtp: connection refused"
			jobs = []*auditdomain.FailedJob{{
				ID: "job-1", EventType: "user.blocked", QueueName: "email", AttemptCount: 3,
				ErrorMessage: &errMsg, FailedAt: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
			}}

			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", map[string]string{"page": "2", "page_size": "5"})))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotParams, ShouldResemble, pagination.NewParams(2, 5))
			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 4)
			list, ok := body["failed_jobs"].([]any)
			So(ok, ShouldBeTrue)
			So(list, ShouldHaveLength, 1)
			first, ok := list[0].(map[string]any)
			So(ok, ShouldBeTrue)
			So(first["queue_name"], ShouldEqual, "email")
			So(first["attempt_count"], ShouldEqual, 3)
			So(first["failed_at"], ShouldEqual, "2026-10-06T12:00:00Z")
		})

		Convey("Missing page params fall back to the defaults", func() {
			testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(gotParams, ShouldResemble, pagination.NewParams(1, 20))
		})

		Convey("An unexpected service error is a 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Without a user in the context it panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(f.mux, f.newReq("", nil)) }, ShouldPanic)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				f.mux.ServeHTTP(&errWriter{}, withUser(f.newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

func TestGetInstanceAdminActionsRouteForbidden(t *testing.T) {
	Convey("GET "+auditPath+" — authorization", t, func() {
		af := newAuditFixture()
		af.svcErr = admindomain.ErrForbiddenUserAction

		Convey("A subadmin asking for the trail of an admin → 403, with the caller passed to the service", func() {
			w := testutil.ServeHTTP(af.mux, withUser(af.newReq("", auditQuery("user", validUserID))))

			So(w.Code, ShouldEqual, http.StatusForbidden)
			So(af.gotActor, ShouldEqual, "user-123")
		})
	})
}

const journalPath = "/api/v1/admin/audit"

type journalFixture struct {
	*httpFixture
	svcErr      error
	called      bool
	gotActor    string
	gotFilter   auditdomain.AdminActionFilter
	gotParams   pagination.Params
	svcActions  []*auditdomain.AdminAction
	svcTotalVal int
}

func newJournalFixture() *journalFixture {
	jf := &journalFixture{}
	svc := &mockService{
		getAdminActions: func(_ context.Context, actorID string, filter auditdomain.AdminActionFilter, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
			jf.called, jf.gotActor, jf.gotFilter, jf.gotParams = true, actorID, filter, params
			return jf.svcActions, jf.svcTotalVal, jf.svcErr
		},
	}
	jf.httpFixture = newHTTPFixture(svc, http.MethodGet, journalPath)

	return jf
}

func TestGetAdminActionsRoute(t *testing.T) {
	Convey("GET "+journalPath, t, func() {
		jf := newJournalFixture()

		Convey("Without a user in the context it panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(jf.mux, jf.newReq("", nil)) }, ShouldPanic)
		})

		Convey("Without a query the filter is empty and the journal is returned with its total", func() {
			jf.svcActions = []*auditdomain.AdminAction{{ID: "a-1", ActionType: auditdomain.ActionBlockUser}}
			jf.svcTotalVal = 7

			w := testutil.ServeHTTP(jf.mux, withUser(jf.newReq("", nil)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(jf.gotActor, ShouldEqual, "user-123")
			So(jf.gotFilter, ShouldResemble, auditdomain.AdminActionFilter{})
			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 7)
			actions, ok := body["actions"].([]any)
			So(ok, ShouldBeTrue)
			So(actions, ShouldHaveLength, 1)
		})

		Convey("Every query parameter reaches the service: admin, action type, RFC 3339 range and page", func() {
			query := map[string]string{
				"admin_user_id": validUserID,
				"action_type":   "block_user",
				"from":          "2026-10-01T00:00:00Z",
				"to":            "2026-10-02T03:04:05+02:00",
				"page":          "2",
				"page_size":     "5",
			}

			w := testutil.ServeHTTP(jf.mux, withUser(jf.newReq("", query)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(jf.gotFilter.AdminUserID, ShouldEqual, validUserID)
			So(jf.gotFilter.ActionType, ShouldEqual, auditdomain.ActionBlockUser)
			So(jf.gotFilter.From.Equal(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)), ShouldBeTrue)
			So(jf.gotFilter.To.Equal(time.Date(2026, 10, 2, 1, 4, 5, 0, time.UTC)), ShouldBeTrue)
			So(jf.gotParams, ShouldResemble, pagination.NewParams(2, 5))
		})

		Convey("A service error maps through the error handler", func() {
			jf.svcErr = admindomain.ErrForbiddenUserAction
			So(testutil.ServeHTTP(jf.mux, withUser(jf.newReq("", nil))).Code, ShouldEqual, http.StatusForbidden)

			jf.svcErr = testutil.ErrDBUnexpected
			So(testutil.ServeHTTP(jf.mux, withUser(jf.newReq("", nil))).Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("The success response write fails → does not panic", func() {
			So(func() { jf.mux.ServeHTTP(&errWriter{}, withUser(jf.newReq("", nil))) }, ShouldNotPanic)
		})
	})
}

func TestGetAdminActionsRouteInvalidQuery(t *testing.T) {
	cases := map[string]map[string]string{
		"admin id that is not a UUID":    {"admin_user_id": "nope"},
		"unknown action type":            {"action_type": "launch_rocket"},
		"from that is not RFC 3339":      {"from": "yesterday"},
		"to that is a date without time": {"to": "2026-10-01"},
		"from after to":                  {"from": "2026-10-02T00:00:00Z", "to": "2026-10-01T00:00:00Z"},
		"from equal to":                  {"from": "2026-10-01T00:00:00Z", "to": "2026-10-01T00:00:00Z"},
	}

	for name, query := range cases {
		Convey("GET "+journalPath+" with an "+name+" → 400 and never reaches the service", t, func() {
			jf := newJournalFixture()

			w := testutil.ServeHTTP(jf.mux, withUser(jf.newReq("", query)))

			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(jf.called, ShouldBeFalse)
		})
	}

	Convey("GET "+journalPath+" with an invalid query and a failing response write → does not panic", t, func() {
		jf := newJournalFixture()

		So(func() { jf.mux.ServeHTTP(&errWriter{}, withUser(jf.newReq("", map[string]string{"action_type": "x"}))) }, ShouldNotPanic)
	})
}
