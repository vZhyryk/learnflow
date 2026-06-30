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

func TestChangePassword(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/change", t, func() {
		var svcErr error

		svc := &mockService{
			changePassword: func(_ context.Context, _ authdomain.ChangePasswordRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/password/change", strings.NewReader(body))
		}
		// validBody includes UserID matching the user injected by withUser.
		validBody := `{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"newpass12"}`

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty OldPassword → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"UserID":"user-123","OldPassword":"","NewPassword":"newpass12"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too short → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too long → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OldPassword == NewPassword → 400 (validation layer)", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"UserID":"user-123","OldPassword":"samepass1","NewPassword":"samepass1"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No user in context → 401", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(validBody))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("UserID mismatch → 401", func() {
			r := withUser(newReq(`{"UserID":"other-user","OldPassword":"oldpass12","NewPassword":"newpass12"}`))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrWrongPassword → 422", func() {
			svcErr = authdomain.ErrWrongPassword
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(newReq(validBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Service ErrSamePassword → 422", func() {
			svcErr = authdomain.ErrSamePassword
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(newReq(validBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(newReq(validBody)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(newReq(validBody)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
