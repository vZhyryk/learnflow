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

type refreshFixture struct {
	svcResult *authdomain.AuthTokens
	svcErr    error
	mux       *http.ServeMux
	newReq    func(body string) *http.Request
}

func newRefreshFixture() *refreshFixture {
	f := &refreshFixture{}
	svc := &mockService{
		refresh: func(_ context.Context, _ authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
			return f.svcResult, f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	}
	return f
}

func TestRefreshRequestValidation(t *testing.T) {
	Convey("POST /api/v1/auth/refresh — request validation", t, func() {
		f := newRefreshFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq("{invalid"))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty RefreshToken → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestRefreshServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/auth/refresh — service outcomes", t, func() {
		f := newRefreshFixture()

		Convey("Service ErrSessionNotFound → 401", func() {
			f.svcErr = authdomain.ErrSessionNotFound
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrSessionExpired → 401", func() {
			f.svcErr = authdomain.ErrSessionExpired
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrSessionRevoked → 401", func() {
			f.svcErr = authdomain.ErrSessionRevoked
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidCredentials → 401", func() {
			f.svcErr = authdomain.ErrInvalidCredentials
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrAccountBlocked → 403", func() {
			f.svcErr = authdomain.ErrAccountBlocked
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with auth envelope", func() {
			f.svcResult = &authdomain.AuthTokens{AccessToken: "new-acc", RefreshToken: "new-ref", UserID: "user-123"}
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["auth"], ShouldNotBeNil)
		})

		Convey("Valid token and the success response write fails → does not panic", func() {
			f.svcResult = &authdomain.AuthTokens{AccessToken: "new-acc", RefreshToken: "new-ref", UserID: "user-123"}
			So(func() { f.mux.ServeHTTP(&errWriter{}, f.newReq(`{"RefreshToken":"ref"}`)) }, ShouldNotPanic)
		})
	})
}
