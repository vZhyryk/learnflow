package adminservice

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
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

			got, total, err := srv.GetInstanceAdminActions(context.Background(), testAdminID, auditdomain.TargetCourse, validUserID, params)

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

			got, total, err := srv.GetInstanceAdminActions(context.Background(), testAdminID, auditdomain.TargetCourse, validUserID, params)

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

func usersByID(users map[string]*admindomain.UserData) *mockUserRepo {
	return &mockUserRepo{getUserDataByID: func(_ context.Context, id string) (*admindomain.UserData, error) {
		user, ok := users[id]
		if !ok {
			return nil, admindomain.ErrUserNotFound
		}
		return user, nil
	}}
}

func TestGetInstanceAdminActionsAuthorization(t *testing.T) {
	const targetID = "44444444-4444-4444-4444-444444444444"

	Convey("Given the audit trail of a user", t, func() {
		var repoCalled bool
		actions := &mockAdminActionRepo{getInstanceAdminActions: func(_ context.Context, _ auditdomain.AdminTargetType, _ string, _ pagination.Params) ([]*auditdomain.AdminAction, int, error) {
			repoCalled = true
			return []*auditdomain.AdminAction{{ID: "a-1"}}, 1, nil
		}}
		users := map[string]*admindomain.UserData{
			testAdminID: activeUserWithRole(admindomain.RoleAdmin),
			targetID:    activeUserWithRole(admindomain.RoleUser),
		}
		read := func(targetType auditdomain.AdminTargetType) error {
			_, _, err := newTestUserService(usersByID(users), actions).GetInstanceAdminActions(context.Background(), testAdminID, targetType, targetID, pagination.NewParams(1, 10))
			return err
		}

		Convey("An admin reads the trail of a plain user, a subadmin and another admin", func() {
			for _, role := range []admindomain.UserRole{admindomain.RoleUser, admindomain.RoleSubAdmin, admindomain.RoleAdmin} {
				users[targetID] = activeUserWithRole(role)
				repoCalled = false

				So(read(auditdomain.TargetUser), ShouldBeNil)
				So(repoCalled, ShouldBeTrue)
			}
		})

		Convey("A subadmin reads the trail of a plain user", func() {
			users[testAdminID] = activeUserWithRole(admindomain.RoleSubAdmin)

			So(read(auditdomain.TargetUser), ShouldBeNil)
			So(repoCalled, ShouldBeTrue)
		})

		Convey("A subadmin cannot read the trail of an admin or another subadmin, and the repository is never asked", func() {
			users[testAdminID] = activeUserWithRole(admindomain.RoleSubAdmin)

			for _, role := range []admindomain.UserRole{admindomain.RoleAdmin, admindomain.RoleSubAdmin} {
				users[targetID] = activeUserWithRole(role)

				So(errors.Is(read(auditdomain.TargetUser), admindomain.ErrForbiddenUserAction), ShouldBeTrue)
			}
			So(repoCalled, ShouldBeFalse)
		})

		Convey("A subadmin gets 'user not found' for a missing target instead of an empty page", func() {
			users[testAdminID] = activeUserWithRole(admindomain.RoleSubAdmin)
			delete(users, targetID)

			So(errors.Is(read(auditdomain.TargetUser), admindomain.ErrUserNotFound), ShouldBeTrue)
		})

		Convey("A blocked actor and a plain-user actor are forbidden", func() {
			users[testAdminID] = &admindomain.UserData{Role: admindomain.RoleAdmin, Status: admindomain.StatusBlocked}
			So(errors.Is(read(auditdomain.TargetUser), admindomain.ErrForbiddenUserAction), ShouldBeTrue)

			users[testAdminID] = activeUserWithRole(admindomain.RoleUser)
			So(errors.Is(read(auditdomain.TargetUser), admindomain.ErrForbiddenUserAction), ShouldBeTrue)
			So(repoCalled, ShouldBeFalse)
		})

		Convey("The trail of a non-user target needs no role lookup", func() {
			users[testAdminID] = activeUserWithRole(admindomain.RoleSubAdmin)

			So(read(auditdomain.TargetCourse), ShouldBeNil)
			So(repoCalled, ShouldBeTrue)
		})
	})
}

func TestGetAdminActions(t *testing.T) {
	Convey("Given the general audit journal", t, func() {
		var gotFilter auditdomain.AdminActionFilter
		var gotParams pagination.Params
		var repoErr error
		actions := &mockAdminActionRepo{getAdminActions: func(_ context.Context, filter auditdomain.AdminActionFilter, params pagination.Params) ([]*auditdomain.AdminAction, int, error) {
			gotFilter, gotParams = filter, params
			if repoErr != nil {
				return nil, 0, repoErr
			}
			return []*auditdomain.AdminAction{{ID: "a-1"}, {ID: "a-2"}}, 9, nil
		}}
		users := map[string]*admindomain.UserData{testAdminID: activeUserWithRole(admindomain.RoleAdmin)}
		read := func() ([]*auditdomain.AdminAction, int, error) {
			return newTestUserService(usersByID(users), actions).GetAdminActions(context.Background(), testAdminID, auditdomain.AdminActionFilter{ActionType: auditdomain.ActionBlockUser}, pagination.NewParams(2, 5))
		}

		Convey("An admin gets the page and total, and the filter and page reach the repository", func() {
			got, total, err := read()

			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(total, ShouldEqual, 9)
			So(gotFilter.ActionType, ShouldEqual, auditdomain.ActionBlockUser)
			So(gotParams, ShouldResemble, pagination.NewParams(2, 5))
		})

		Convey("A subadmin, a plain user and a blocked admin are forbidden and the repository is never asked", func() {
			for _, actor := range []*admindomain.UserData{
				activeUserWithRole(admindomain.RoleSubAdmin),
				activeUserWithRole(admindomain.RoleUser),
				{Role: admindomain.RoleAdmin, Status: admindomain.StatusBlocked},
			} {
				users[testAdminID] = actor
				gotParams = pagination.Params{}

				_, _, err := read()

				So(errors.Is(err, admindomain.ErrForbiddenUserAction), ShouldBeTrue)
				So(gotParams, ShouldResemble, pagination.Params{})
			}
		})

		Convey("When the actor cannot be loaded, the error is wrapped", func() {
			delete(users, testAdminID)

			_, _, err := read()

			So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetAdminActions")
		})

		Convey("When the repository fails, the error is wrapped", func() {
			repoErr = testutil.ErrDBUnexpected

			got, total, err := read()

			So(got, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "service.GetAdminActions")
		})
	})
}
