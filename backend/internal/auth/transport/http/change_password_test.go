package authhttp_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

const validChangePasswordBody = `{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"newpass12"}`

type changePasswordFixture struct {
	*httpFixture
	svcErr error
}

func newChangePasswordFixture() *changePasswordFixture {
	f := &changePasswordFixture{}
	svc := &mockService{
		changePassword: func(_ context.Context, _ authdomain.ChangePasswordRequest) error {
			return f.svcErr
		},
	}
	f.httpFixture = newHTTPFixture(svc, http.MethodPut, "/api/v1/users/auth/password/change")
	return f
}

func TestChangePasswordRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/change — request validation", t, func() {
		f := newChangePasswordFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty OldPassword → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"UserID":"user-123","OldPassword":"","NewPassword":"newpass12"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too short → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too long → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OldPassword == NewPassword → 400 (validation layer)", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"UserID":"user-123","OldPassword":"samepass1","NewPassword":"samepass1"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No user in context → 401", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(validChangePasswordBody))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("UserID mismatch → 401", func() {
			r := withUser(f.newReq(`{"UserID":"other-user","OldPassword":"oldpass12","NewPassword":"newpass12"}`))
			w := testutil.ServeHTTP(f.mux, r)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestChangePasswordServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/change — service outcomes", t, func() {
		f := newChangePasswordFixture()

		Convey("Service ErrWrongPassword → 422", func() {
			f.svcErr = authdomain.ErrWrongPassword
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Service ErrSamePassword → 422", func() {
			f.svcErr = authdomain.ErrSamePassword
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, withUser(f.newReq(validChangePasswordBody))) }, ShouldNotPanic)
		})
	})
}
