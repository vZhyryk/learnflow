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

func TestInitiatePasswordReset(t *testing.T) {
	Convey("POST /api/v1/users/auth/password/reset", t, func() {
		var svcErr error

		svc := &mockService{
			initiatePasswordReset: func(_ context.Context, _ authdomain.RequestPasswordResetRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/users/auth/password/reset", strings.NewReader(body))
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

func TestResetPassword(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/reset", t, func() {
		var svcErr error

		svc := &mockService{
			resetPassword: func(_ context.Context, _ authdomain.ResetPasswordRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/password/reset", strings.NewReader(body))
		}

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too short → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too long → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenExpired → 400", func() {
			svcErr = authdomain.ErrTokenExpired
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			svcErr = authdomain.ErrTokenUsed
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			svcErr = authdomain.ErrInvalidToken
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Token":"tok","NewPassword":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
