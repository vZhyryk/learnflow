package authhttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogin(t *testing.T) {
	Convey("POST /api/v1/auth/login", t, func() {
		var svcResult *authdomain.AuthTokens
		var svcErr error

		svc := &mockService{
			login: func(_ context.Context, _ authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
				return svcResult, svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
			r.Header.Set("User-Agent", "test-agent/1.0")
			return r
		}

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq("{invalid"))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Wrong type for Email → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":123,"Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"notanemail","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too short → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too long → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No User-Agent header → 400", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(`{"Email":"user@example.com","Password":"password123"}`))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("User-Agent too long → 400", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(`{"Email":"user@example.com","Password":"password123"}`))
			r.Header.Set("User-Agent", strings.Repeat("x", 2001))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrInvalidCredentials → 401", func() {
			svcErr = authdomain.ErrInvalidCredentials
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrUserNotFound → 401", func() {
			svcErr = authdomain.ErrUserNotFound
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrAccountLocked → 429 with Retry-After header", func() {
			svcErr = &authdomain.ErrAccountLockedError{LockedUntil: time.Now().Add(5 * time.Minute)}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusTooManyRequests)
			So(w.Header().Get("Retry-After"), ShouldNotBeEmpty)
		})

		Convey("Service ErrAccountBlocked → 403", func() {
			svcErr = authdomain.ErrAccountBlocked
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Service ErrEmailNotVerified → 403", func() {
			svcErr = authdomain.ErrEmailNotVerified
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid credentials → 200 with auth envelope", func() {
			svcResult = &authdomain.AuthTokens{AccessToken: "acc", RefreshToken: "ref", UserID: "user-123"}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["auth"], ShouldNotBeNil)
		})
	})
}
