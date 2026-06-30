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

func TestInitRecoverAccount(t *testing.T) {
	Convey("POST /api/v1/users/auth/account/recover", t, func() {
		var svcErr error

		svc := &mockService{
			initRecoverAccount: func(_ context.Context, _ authdomain.RequestRecoverAccountRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/users/auth/account/recover", strings.NewReader(body))
		}

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid email format → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"notanemail"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrInvalidAccountState → 200 (account state guard)", func() {
			svcErr = authdomain.ErrInvalidAccountState
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid email → 200 with message", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}

func TestRecoverAccount(t *testing.T) {
	Convey("PUT /api/v1/users/auth/account/recover", t, func() {
		var svcErr error

		svc := &mockService{
			recoverAccount: func(_ context.Context, _ authdomain.RecoverAccountRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/account/recover", strings.NewReader(body))
		}

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenExpired → 400", func() {
			svcErr = authdomain.ErrTokenExpired
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			svcErr = authdomain.ErrTokenUsed
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			svcErr = authdomain.ErrInvalidToken
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidAccountState → 200 (deleted account guard)", func() {
			svcErr = authdomain.ErrInvalidAccountState
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with message", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
