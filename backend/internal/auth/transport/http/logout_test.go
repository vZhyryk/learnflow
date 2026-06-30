package authhttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"

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
	handler := authhttp.NewHTTPHandler(svc, newTestLogger())
	f.mux = http.NewServeMux()
	handler.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.patch = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/logout", strings.NewReader(body))
	}
	return f
}

func TestLogout(t *testing.T) {
	Convey("POST /api/v1/auth/logout", t, func() {
		f := newLogoutFixture()

		Convey("Empty request body", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(``))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{bad json`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Wrong type for refresh_token", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token":1}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty refresh_token fails validation", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service returns ErrUserNotFound", func() {
			f.svcErr = authdomain.ErrUserNotFound
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service returns ErrInvalidCredentials", func() {
			f.svcErr = authdomain.ErrInvalidCredentials
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service returns ErrSessionNotFound", func() {
			f.svcErr = authdomain.ErrSessionNotFound
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Account blocked returns 403", func() {
			f.svcErr = authdomain.ErrAccountBlocked
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error returns 500", func() {
			f.svcErr = errors.New("database unavailable")
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request returns 200 with ok body", func() {
			f.svcResult = "user-123"
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"refresh_token": "ref"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["ok"], ShouldEqual, true)
		})
	})
}
