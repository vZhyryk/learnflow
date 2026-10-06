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
	gotType    auditdomain.AdminTargetType
	gotID      string
	gotParams  pagination.Params
}

func newAuditFixture() *auditFixture {
	af := &auditFixture{}
	svc := &mockService{
		getInstanceAdminActions: func(_ context.Context, targetType auditdomain.AdminTargetType, targetID string, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
			af.called = true
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

		Convey("Without a user in the context it panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(af.mux, af.newReq("", auditQuery("user", validUserID))) }, ShouldPanic)
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
	})
}
