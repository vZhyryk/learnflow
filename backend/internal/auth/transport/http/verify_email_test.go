package authhttp_test

import (
	"context"
	"net/http"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

type verifyEmailFixture struct {
	*httpFixture
	svcResult string
	svcErr    error
}

func newVerifyEmailFixture() *verifyEmailFixture {
	f := &verifyEmailFixture{}
	svc := &mockService{
		verifyEmail: func(_ context.Context, _ authdomain.VerifyEmailRequest) (string, error) {
			return f.svcResult, f.svcErr
		},
	}
	f.httpFixture = newHTTPFixture(svc, http.MethodPost, "/api/v1/users/auth/email/verify")
	return f
}

func TestVerifyEmailRequestValidation(t *testing.T) {
	Convey("POST /api/v1/users/auth/email/verify — request validation", t, func() {
		f := newVerifyEmailFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq("{invalid"))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestVerifyEmailServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/users/auth/email/verify — service outcomes", t, func() {
		f := newVerifyEmailFixture()

		Convey("Service ErrTokenExpired → 400", func() {
			f.svcErr = authdomain.ErrTokenExpired
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			f.svcErr = authdomain.ErrTokenUsed
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			f.svcErr = authdomain.ErrInvalidToken
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with message", func() {
			f.svcResult = "user-123"
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid token and the success response write fails → does not panic", func() {
			f.svcResult = "user-123"
			So(func() { f.mux.ServeHTTP(&errWriter{}, f.newReq(`{"Token":"tok"}`)) }, ShouldNotPanic)
		})
	})
}
