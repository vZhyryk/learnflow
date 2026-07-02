package authhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	authhttp "learnflow_backend/internal/auth/transport/http"
	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

type loginFixture struct {
	svcResult *authdomain.AuthTokens
	svcErr    error
	mux       *http.ServeMux
	newReq    func(body string) *http.Request
}

func newLoginFixture() *loginFixture {
	f := &loginFixture{}
	svc := &mockService{
		login: func(_ context.Context, _ authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
			return f.svcResult, f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		r.Header.Set("User-Agent", "test-agent/1.0")
		return r
	}
	return f
}

// doRequest fires body through f.mux and returns the recorded response.
func (f *loginFixture) doRequest(body string) *httptest.ResponseRecorder {
	return testutil.ServeHTTP(f.mux, f.newReq(body))
}

// doRequestWithWriter fires body through f.mux against an arbitrary
// http.ResponseWriter (e.g. errWriter, to exercise response-write-failure branches).
func (f *loginFixture) doRequestWithWriter(w http.ResponseWriter, body string) {
	f.mux.ServeHTTP(w, f.newReq(body))
}

const validLoginBody = `{"Email":"user@example.com","Password":"password123"}`

func TestLoginRequestValidation(t *testing.T) {
	Convey("POST /api/v1/auth/login — request validation", t, func() {
		f := newLoginFixture()

		Convey("Empty body → 400", func() {
			w := f.doRequest("")
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := f.doRequest("{invalid")
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Wrong type for Email → 400", func() {
			w := f.doRequest(`{"Email":123,"Password":"password123"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := f.doRequest(`{"Email":"notanemail","Password":"password123"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too short → 400", func() {
			w := f.doRequest(`{"Email":"user@example.com","Password":"short"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too long → 400", func() {
			w := f.doRequest(`{"Email":"user@example.com","Password":"` + strings.Repeat("a", 73) + `"}`)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No User-Agent header → 400", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(`{"Email":"user@example.com","Password":"password123"}`))
			w := testutil.ServeHTTP(f.mux, r)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("User-Agent too long → 400", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(`{"Email":"user@example.com","Password":"password123"}`))
			r.Header.Set("User-Agent", strings.Repeat("x", 2001))
			w := testutil.ServeHTTP(f.mux, r)
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON and the error response write also fails → does not panic", func() {
			So(func() { f.doRequestWithWriter(&errWriter{}, "{invalid") }, ShouldNotPanic)
		})

		Convey("Invalid email format and the error response write also fails → does not panic", func() {
			So(func() {
				f.doRequestWithWriter(&errWriter{}, `{"Email":"notanemail","Password":"password123"}`)
			}, ShouldNotPanic)
		})
	})
}

func TestLoginServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/auth/login — service outcomes", t, func() {
		f := newLoginFixture()

		Convey("Service ErrInvalidCredentials → 401", func() {
			f.svcErr = authdomain.ErrInvalidCredentials
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidCredentials and the error response write also fails → does not panic", func() {
			f.svcErr = authdomain.ErrInvalidCredentials
			So(func() { f.doRequestWithWriter(&errWriter{}, validLoginBody) }, ShouldNotPanic)
		})

		Convey("Service ErrUserNotFound → 401", func() {
			f.svcErr = authdomain.ErrUserNotFound
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrAccountLocked → 429 with Retry-After header", func() {
			f.svcErr = &authdomain.ErrAccountLockedError{LockedUntil: time.Now().Add(5 * time.Minute)}
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusTooManyRequests)
			So(w.Header().Get("Retry-After"), ShouldNotBeEmpty)
		})

		Convey("Service ErrAccountBlocked → 403", func() {
			f.svcErr = authdomain.ErrAccountBlocked
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Service ErrEmailNotVerified → 403", func() {
			f.svcErr = authdomain.ErrEmailNotVerified
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid credentials → 200 with auth envelope", func() {
			f.svcResult = &authdomain.AuthTokens{AccessToken: "acc", RefreshToken: "ref", UserID: "user-123"}
			w := f.doRequest(validLoginBody)
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["auth"], ShouldNotBeNil)
		})

		Convey("Valid credentials and the success response write fails → does not panic", func() {
			f.svcResult = &authdomain.AuthTokens{AccessToken: "acc", RefreshToken: "ref", UserID: "user-123"}
			So(func() { f.doRequestWithWriter(&errWriter{}, validLoginBody) }, ShouldNotPanic)
		})
	})
}
