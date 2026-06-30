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

func TestRefresh(t *testing.T) {
	Convey("POST /api/v1/auth/refresh", t, func() {
		var svcResult *authdomain.AuthTokens
		var svcErr error

		svc := &mockService{
			refresh: func(_ context.Context, _ authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
				return svcResult, svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
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

		Convey("Empty RefreshToken → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrSessionNotFound → 401", func() {
			svcErr = authdomain.ErrSessionNotFound
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrSessionExpired → 401", func() {
			svcErr = authdomain.ErrSessionExpired
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrSessionRevoked → 401", func() {
			svcErr = authdomain.ErrSessionRevoked
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidCredentials → 401", func() {
			svcErr = authdomain.ErrInvalidCredentials
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrAccountBlocked → 403", func() {
			svcErr = authdomain.ErrAccountBlocked
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with auth envelope", func() {
			svcResult = &authdomain.AuthTokens{AccessToken: "new-acc", RefreshToken: "new-ref", UserID: "user-123"}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"RefreshToken":"ref"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["auth"], ShouldNotBeNil)
		})
	})
}
