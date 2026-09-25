package adminhttp_test

import (
	"context"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const validUserID = "11111111-1111-1111-1111-111111111111"

func TestGetUsersData(t *testing.T) {
	Convey("GET /api/v1/admin/users", t, func() {
		var svcErr error
		svc := &mockService{
			getUsersData: func(_ context.Context, _ pagination.Params) ([]*admindomain.UserData, int, error) {
				return []*admindomain.UserData{{UserID: "user-1"}}, 7, svcErr
			},
		}
		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/users")

		Convey("No user in context → panics (middleware invariant violated)", func() {
			So(func() { testutil.ServeHTTP(f.mux, f.newReq("", nil)) }, ShouldPanic)
		})

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with users and total", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["total"], ShouldEqual, 7)
			So(body["users"], ShouldHaveLength, 1)
		})
	})
}

func TestGetUserDataByID(t *testing.T) {
	Convey("GET /api/v1/admin/users/{id}", t, func() {
		var svcErr error
		var gotID string
		svc := &mockService{
			getUserDataByID: func(_ context.Context, id string) (*admindomain.UserData, error) {
				gotID = id
				return &admindomain.UserData{UserID: id}, svcErr
			},
		}
		f := newHTTPFixture(svc, http.MethodGet, "/api/v1/admin/users/"+validUserID)

		Convey("Invalid id → 422", func() {
			svcErr = admindomain.ErrInvalidID
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("User not found → 404", func() {
			svcErr = admindomain.ErrUserNotFound
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("Service returns an error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 and the path id (not the caller's) is fetched", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq("", nil)))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(gotID, ShouldEqual, validUserID)
			So(decodeBody(t, w.Body.Bytes())["user"], ShouldNotBeNil)
		})
	})
}

type userRoute struct {
	operation string
	method    string
	suffix    string
}

var userRoutes = []userRoute{
	{"DeleteUser", http.MethodDelete, ""},
	{"RestoreUser", http.MethodPut, "/restore"},
	{"BlockUser", http.MethodPut, "/block"},
	{"UnBlockUser", http.MethodPut, "/unblock"},
	{"AssignUserRole", http.MethodPut, "/subadmin"},
	{"RevokeUserRole", http.MethodDelete, "/subadmin"},
}

type changeUserFixture struct {
	*httpFixture
	svcErr                       error
	gotOp, gotUserID, gotAdminID string
}

func newChangeUserFixture(route userRoute) *changeUserFixture {
	cf := &changeUserFixture{}
	svc := &mockService{
		changeUserField: func(_ context.Context, op, userID, adminID string) error {
			cf.gotOp, cf.gotUserID, cf.gotAdminID = op, userID, adminID
			return cf.svcErr
		},
	}
	cf.httpFixture = newHTTPFixture(svc, route.method, "/api/v1/admin/users/"+validUserID+route.suffix)
	return cf
}

func TestChangeUserRoutesErrors(t *testing.T) {
	for _, route := range userRoutes {
		Convey(route.method+" /api/v1/admin/users/{id}"+route.suffix+" errors", t, func() {
			cf := newChangeUserFixture(route)

			Convey("User not found → 404", func() {
				cf.svcErr = admindomain.ErrUserNotFound
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})

			Convey("User in the wrong state → 409", func() {
				cf.svcErr = admindomain.ErrInvalidUserState
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusConflict)
			})

			Convey("Forbidden action on this user → 403", func() {
				cf.svcErr = admindomain.ErrForbiddenUserAction
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Invalid id → 422", func() {
				cf.svcErr = admindomain.ErrInvalidID
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
			})

			Convey("Service returns an error → 500", func() {
				cf.svcErr = testutil.ErrDBUnexpected
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	}
}

func TestChangeUserRoutesSuccess(t *testing.T) {
	for _, route := range userRoutes {
		Convey(route.method+" /api/v1/admin/users/{id}"+route.suffix, t, func() {
			cf := newChangeUserFixture(route)

			Convey("No user in context → panics (middleware invariant violated)", func() {
				So(func() { testutil.ServeHTTP(cf.mux, cf.newReq("", nil)) }, ShouldPanic)
			})

			Convey("Valid request → 200 with the operation, target id and admin id", func() {
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusOK)
				So(cf.gotOp, ShouldEqual, route.operation)
				So(cf.gotUserID, ShouldEqual, validUserID)
				So(cf.gotAdminID, ShouldEqual, "user-123")
			})
		})
	}
}
