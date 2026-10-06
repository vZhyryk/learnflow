package adminhttp_test

import (
	"context"
	"fmt"
	admindomain "learnflow_backend/internal/admin/domain"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/testutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
