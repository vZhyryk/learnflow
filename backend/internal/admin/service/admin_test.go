package adminservice

import (
	"context"
	"errors"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetInstanceAdminActions(t *testing.T) {
	Convey("Given an admin service", t, func() {
		var gotType auditdomain.AdminTargetType
		var gotID string
		var gotParams pagination.Params
		actions := &mockAdminActionRepo{}
		srv := newTestUserService(&mockUserRepo{}, actions)
		params := pagination.NewParams(2, 10)

		Convey("When the repository fails the error is returned, not swallowed", func() {
			actions.getInstanceAdminActions = func(_ context.Context, _ auditdomain.AdminTargetType, _ string, _ pagination.Params) ([]*auditdomain.AdminAction, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}

			got, total, err := srv.GetInstanceAdminActions(context.Background(), auditdomain.TargetUser, validUserID, params)

			So(got, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetInstanceAdminActions")
		})

		Convey("When it succeeds it passes the target and page through and returns the page and total", func() {
			want := []*auditdomain.AdminAction{{ID: "a-1"}, {ID: "a-2"}}
			actions.getInstanceAdminActions = func(_ context.Context, targetType auditdomain.AdminTargetType, targetID string, p pagination.Params) ([]*auditdomain.AdminAction, int, error) {
				gotType, gotID, gotParams = targetType, targetID, p
				return want, 42, nil
			}

			got, total, err := srv.GetInstanceAdminActions(context.Background(), auditdomain.TargetCourse, validUserID, params)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 42)
			So(gotType, ShouldEqual, auditdomain.TargetCourse)
			So(gotID, ShouldEqual, validUserID)
			So(gotParams, ShouldResemble, params)
		})
	})
}

func TestGetFailedJobs(t *testing.T) {
	Convey("Given an admin service", t, func() {
		var gotParams pagination.Params
		actions := &mockAdminActionRepo{}
		srv := newTestUserService(&mockUserRepo{}, actions)
		params := pagination.NewParams(3, 15)

		Convey("When the repository fails the error is returned, not swallowed", func() {
			actions.getFailedJobs = func(_ context.Context, _ pagination.Params) ([]*auditdomain.FailedJob, int, error) {
				return nil, 0, testutil.ErrDBUnexpected
			}

			got, total, err := srv.GetFailedJobs(context.Background(), params)

			So(got, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetFailedJobs")
		})

		Convey("When it succeeds it passes the page through and returns the jobs and total", func() {
			want := []*auditdomain.FailedJob{{ID: "j-1"}, {ID: "j-2"}}
			actions.getFailedJobs = func(_ context.Context, p pagination.Params) ([]*auditdomain.FailedJob, int, error) {
				gotParams = p
				return want, 9, nil
			}

			got, total, err := srv.GetFailedJobs(context.Background(), params)

			So(err, ShouldBeNil)
			So(got, ShouldResemble, want)
			So(total, ShouldEqual, 9)
			So(gotParams, ShouldResemble, params)
		})
	})
}
