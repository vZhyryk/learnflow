package authhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitRecoverAccount(t *testing.T) {
	Convey("POST /api/v1/users/auth/account/recover", t, func() {
		var svcErr error

		svc := &mockService{
			initRecoverAccount: func(_ context.Context, _ authdomain.RequestRecoverAccountRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/users/auth/account/recover", strings.NewReader(body))
		}

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"Email":"notanemail"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrInvalidAccountState → 200 (account state guard)", func() {
			svcErr = authdomain.ErrInvalidAccountState
			w := testutil.ServeHTTP(mux, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid email → 200 with message", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid email and the success response write fails → does not panic", func() {
			So(func() { mux.ServeHTTP(&errWriter{}, newReq(`{"Email":"user@example.com"}`)) }, ShouldNotPanic)
		})
	})
}

type recoverAccountFixture struct {
	svcErr error
	mux    *http.ServeMux
	newReq func(body string) *http.Request
}

func newRecoverAccountFixture() *recoverAccountFixture {
	f := &recoverAccountFixture{}
	svc := &mockService{
		recoverAccount: func(_ context.Context, _ authdomain.RecoverAccountRequest) error {
			return f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/account/recover", strings.NewReader(body))
	}
	return f
}

func TestRecoverAccountRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/users/auth/account/recover — request validation", t, func() {
		f := newRecoverAccountFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestRecoverAccountServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/users/auth/account/recover — service outcomes", t, func() {
		f := newRecoverAccountFixture()

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

		Convey("Service ErrInvalidAccountState → 200 (deleted account guard)", func() {
			f.svcErr = authdomain.ErrInvalidAccountState
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with message", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid token and the success response write fails → does not panic", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, f.newReq(`{"Token":"tok"}`)) }, ShouldNotPanic)
		})
	})
}
