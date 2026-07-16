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

func TestInitiatePasswordReset(t *testing.T) {
	Convey("POST /api/v1/users/auth/password/reset", t, func() {
		var svcErr error

		svc := &mockService{
			initiatePasswordReset: func(_ context.Context, _ authdomain.RequestPasswordResetRequest) error {
				return svcErr
			},
		}
		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/users/auth/password/reset")
		mux, newReq := f.mux, f.newReq

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"email":"notanemail"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrInvalidAccountState → 200 (account state guard)", func() {
			svcErr = authdomain.ErrInvalidAccountState
			w := testutil.ServeHTTP(mux, newReq(`{"email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq(`{"email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid email → 200 with message", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid email and the success response write fails → does not panic", func() {
			So(func() { mux.ServeHTTP(&errWriter{}, newReq(`{"email":"user@example.com"}`)) }, ShouldNotPanic)
		})
	})
}

type resetPasswordFixture struct {
	*httpFixture
	svcErr error
}

func newResetPasswordFixture() *resetPasswordFixture {
	f := &resetPasswordFixture{}
	svc := &mockService{
		resetPassword: func(_ context.Context, _ authdomain.ResetPasswordRequest) error {
			return f.svcErr
		},
	}
	f.httpFixture = newHTTPFixture(svc, http.MethodPut, "/api/v1/users/auth/password/reset")
	return f
}

func TestResetPasswordRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/reset — request validation", t, func() {
		f := newResetPasswordFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too short → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too long → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestResetPasswordServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/reset — service outcomes", t, func() {
		f := newResetPasswordFixture()

		Convey("Service ErrTokenExpired → 400", func() {
			f.svcErr = authdomain.ErrTokenExpired
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			f.svcErr = authdomain.ErrTokenUsed
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			f.svcErr = authdomain.ErrInvalidToken
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok","new_password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				f.mux.ServeHTTP(&errWriter{}, f.newReq(`{"token":"tok","new_password":"password123"}`))
			}, ShouldNotPanic)
		})
	})
}
