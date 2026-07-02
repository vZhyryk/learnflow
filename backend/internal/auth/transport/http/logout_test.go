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

type logoutFixture struct {
	svcResult string
	svcErr    error
	mux       *http.ServeMux
	patch     func(string) *http.Request
}

func newLogoutFixture() *logoutFixture {
	f := &logoutFixture{}
	svc := &mockService{
		logout: func(_ context.Context, _ authdomain.LogoutRequest) (string, error) {
			return f.svcResult, f.svcErr
		},
	}
	handler := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	f.mux = http.NewServeMux()
	handler.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.patch = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/logout", strings.NewReader(body))
	}
	return f
}

func TestLogoutRequestValidation(t *testing.T) {
	Convey("POST /api/v1/auth/logout — request validation", t, func() {
		f := newLogoutFixture()

		Convey("Empty request body", func() {
			w := testutil.ServeHTTP(f.mux, f.patch(``))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON", func() {
			w := testutil.ServeHTTP(f.mux, f.patch(`{bad json`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Wrong type for refresh_token", func() {
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token":1}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty refresh_token fails validation", func() {
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestLogoutServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/auth/logout — service outcomes", t, func() {
		f := newLogoutFixture()

		Convey("Service returns ErrUserNotFound", func() {
			f.svcErr = authdomain.ErrUserNotFound
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service returns ErrInvalidCredentials", func() {
			f.svcErr = authdomain.ErrInvalidCredentials
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service returns ErrSessionNotFound", func() {
			f.svcErr = authdomain.ErrSessionNotFound
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Account blocked returns 403", func() {
			f.svcErr = authdomain.ErrAccountBlocked
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error returns 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request returns 200 with ok body", func() {
			f.svcResult = "user-123"
			w := testutil.ServeHTTP(f.mux, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["ok"], ShouldEqual, true)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			f.svcResult = "user-123"
			So(func() { f.mux.ServeHTTP(&errWriter{}, f.patch(`{"refresh_token": "ref"}`)) }, ShouldNotPanic)
		})
	})
}
