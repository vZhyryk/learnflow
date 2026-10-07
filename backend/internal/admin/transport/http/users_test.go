package adminhttp_test

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
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

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				f.mux.ServeHTTP(&errWriter{}, withUser(f.newReq("", nil)))
			}, ShouldNotPanic)
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

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				f.mux.ServeHTTP(&errWriter{}, withUser(f.newReq("", nil)))
			}, ShouldNotPanic)
		})
	})
}

type userRoute struct {
	operation admindomain.UserAdminOperation
	method    string
	suffix    string
}

var userRoutes = []userRoute{
	{admindomain.DeleteUser, http.MethodDelete, ""},
	{admindomain.RestoreUser, http.MethodPut, "/restore"},
	{admindomain.BlockUser, http.MethodPut, "/block"},
	{admindomain.UnBlockUser, http.MethodPut, "/unblock"},
	{admindomain.AssignUserRole, http.MethodPut, "/subadmin"},
	{admindomain.RevokeUserRole, http.MethodDelete, "/subadmin"},
}

type changeUserFixture struct {
	*httpFixture
	svcErr                error
	gotUserID, gotAdminID string
	gotOp                 admindomain.UserAdminOperation
}

func newChangeUserFixture(route userRoute) *changeUserFixture {
	cf := &changeUserFixture{}
	svc := &mockService{
		changeUserField: func(_ context.Context, userID, adminID string, op admindomain.UserAdminOperation) error {
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

			Convey("Blocklist unavailable → 503", func() {
				cf.svcErr = fmt.Errorf("set user_blocked: %w: %w", admindomain.ErrBlocklistUnavailable, testutil.ErrRedisUnavailable)
				w := testutil.ServeHTTP(cf.mux, withUser(cf.newReq("", nil)))
				So(w.Code, ShouldEqual, http.StatusServiceUnavailable)
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

			Convey("Valid request and the success response write fails → does not panic", func() {
				So(func() {
					cf.mux.ServeHTTP(&errWriter{}, withUser(cf.newReq("", nil)))
				}, ShouldNotPanic)
			})
		})
	}
}

const grantItemID = "33333333-3333-3333-3333-333333333333"

type grantFixture struct {
	*httpFixture
	svcErr               error
	gotMethod            string
	gotUserID, gotItemID string
	gotAdminID           string
	gotAction            auditdomain.AdminActionType
}

func newGrantFixture() *grantFixture {
	gf := &grantFixture{}
	record := func(method string) func(_ context.Context, userID, itemID, adminID string, action auditdomain.AdminActionType) error {
		return func(_ context.Context, userID, itemID, adminID string, action auditdomain.AdminActionType) error {
			gf.gotMethod, gf.gotUserID, gf.gotItemID, gf.gotAdminID, gf.gotAction = method, userID, itemID, adminID, action
			return gf.svcErr
		}
	}
	svc := &mockService{
		grantUserCourseAccess:  record("course"),
		grantUserContentAccess: record("content"),
	}
	gf.httpFixture = newHTTPFixture(svc, http.MethodPost, "/api/v1/admin/users/"+validUserID+"/course-access")

	return gf
}

func grantBody(itemType string) string {
	return fmt.Sprintf(`{"item_id":%q,"item_type":%q}`, grantItemID, itemType)
}

func TestGrantUserAccessSuccess(t *testing.T) {
	for itemType, wantMethod := range map[string]string{"course": "course", "content": "content"} {
		Convey("POST course-access with item_type "+itemType, t, func() {
			gf := newGrantFixture()

			w := testutil.ServeHTTP(gf.mux, withUser(gf.newReq(grantBody(itemType), nil)))

			So(w.Code, ShouldEqual, http.StatusOK)
			So(gf.gotMethod, ShouldEqual, wantMethod)
			So(gf.gotUserID, ShouldEqual, validUserID)
			So(gf.gotItemID, ShouldEqual, grantItemID)
			So(gf.gotAdminID, ShouldEqual, "user-123")
			So(gf.gotAction, ShouldEqual, auditdomain.ActionGrantItemAccess)

			Convey("The success response write fails → does not panic", func() {
				So(func() {
					gf.mux.ServeHTTP(&errWriter{}, withUser(gf.newReq(grantBody(itemType), nil)))
				}, ShouldNotPanic)
			})
		})
	}
}

func TestGrantUserAccessInvalidRequest(t *testing.T) {
	Convey("POST course-access with an invalid body", t, func() {
		gf := newGrantFixture()

		Convey("Unknown item_type → 400", func() {
			w := testutil.ServeHTTP(gf.mux, withUser(gf.newReq(grantBody("video"), nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(gf.gotMethod, ShouldBeEmpty)
		})

		Convey("Item id is not a UUID → 400", func() {
			w := testutil.ServeHTTP(gf.mux, withUser(gf.newReq(`{"item_id":"nope","item_type":"course"}`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(gf.gotMethod, ShouldBeEmpty)
		})

		Convey("Malformed JSON → 400", func() {
			w := testutil.ServeHTTP(gf.mux, withUser(gf.newReq(`{`, nil)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestGrantUserAccessServiceErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
	}{
		{"Item not found → 404", admindomain.ErrItemNotFound, http.StatusNotFound},
		{"User not found → 404", admindomain.ErrUserNotFound, http.StatusNotFound},
		{"Access already granted → 409", admindomain.ErrAccessAlreadyGranted, http.StatusConflict},
		{"User in the wrong state → 409", admindomain.ErrInvalidUserState, http.StatusConflict},
		{"Invalid user id → 422", admindomain.ErrInvalidID, http.StatusUnprocessableEntity},
		{"Unexpected error → 500", testutil.ErrDBUnexpected, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		Convey("POST course-access: "+tc.name, t, func() {
			gf := newGrantFixture()
			gf.svcErr = tc.err

			w := testutil.ServeHTTP(gf.mux, withUser(gf.newReq(grantBody("course"), nil)))

			So(w.Code, ShouldEqual, tc.code)
		})
	}
}

func TestGrantUserAccessWithoutUserPanics(t *testing.T) {
	Convey("POST course-access without a user in context panics (middleware invariant violated)", t, func() {
		gf := newGrantFixture()
		So(func() { testutil.ServeHTTP(gf.mux, gf.newReq(grantBody("course"), nil)) }, ShouldPanic)
	})
}
