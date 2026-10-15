package authhttp_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitRecoverAccount(t *testing.T) {
	Convey("POST /api/v1/auth/account/recover", t, func() {
		var svcErr error
		svc := &mockService{
			initRecoverAccount: func(_ context.Context, _ authdomain.RequestRecoverAccountRequest) error {
				return svcErr
			},
		}
		f := newHTTPFixture(svc, http.MethodPost, "/api/v1/auth/account/recover")
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

		Convey("Service ErrDeletedByAdmin → 200 (same answer, no account enumeration)", func() {
			svcErr = authdomain.ErrDeletedByAdmin
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

type recoverAccountFixture struct {
	*httpFixture
	svcErr error
}

func newRecoverAccountFixture() *recoverAccountFixture {
	f := &recoverAccountFixture{}
	svc := &mockService{
		recoverAccount: func(_ context.Context, _ authdomain.RecoverAccountRequest) error {
			return f.svcErr
		},
	}
	f.httpFixture = newHTTPFixture(svc, http.MethodPut, "/api/v1/auth/account/recover")
	return f
}

func TestRecoverAccountRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/auth/account/recover — request validation", t, func() {
		f := newRecoverAccountFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestRecoverAccountServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/auth/account/recover — service outcomes", t, func() {
		f := newRecoverAccountFixture()

		Convey("Service ErrTokenExpired → 400", func() {
			f.svcErr = authdomain.ErrTokenExpired
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			f.svcErr = authdomain.ErrTokenUsed
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			f.svcErr = authdomain.ErrInvalidToken
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidAccountState → 403 (the token holder is told the account was not recovered)", func() {
			f.svcErr = authdomain.ErrInvalidAccountState
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
			So(testutil.DecodeBody(t, w.Body.Bytes())["code"], ShouldEqual, "account_not_recoverable")
		})

		Convey("Service ErrBlocklistUnavailable → 503", func() {
			f.svcErr = authdomain.ErrBlocklistUnavailable
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusServiceUnavailable)
		})

		Convey("Service ErrDeletedByAdmin → 403, also when wrapped by the service", func() {
			f.svcErr = fmt.Errorf("recover_account: %w", authdomain.ErrDeletedByAdmin)
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
			So(w.Body.String(), ShouldNotContainSubstring, "recover_account")
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with message", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid token and the success response write fails → does not panic", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, f.newReq(`{"token":"tok"}`)) }, ShouldNotPanic)
		})
	})
}
