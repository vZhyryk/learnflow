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

// validChangePasswordBody includes UserID matching the user injected by withUser.
const validChangePasswordBody = `{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"newpass12"}`

type changePasswordFixture struct {
	svcErr error
	mux    *http.ServeMux
	newReq func(body string) *http.Request
}

func newChangePasswordFixture() *changePasswordFixture {
	f := &changePasswordFixture{}
	svc := &mockService{
		changePassword: func(_ context.Context, _ authdomain.ChangePasswordRequest) error {
			return f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, newTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/password/change", strings.NewReader(body))
	}
	return f
}

func TestChangePasswordRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/change — request validation", t, func() {
		f := newChangePasswordFixture()

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty OldPassword → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"UserID":"user-123","OldPassword":"","NewPassword":"newpass12"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too short → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("NewPassword too long → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"UserID":"user-123","OldPassword":"oldpass12","NewPassword":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OldPassword == NewPassword → 400 (validation layer)", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"UserID":"user-123","OldPassword":"samepass1","NewPassword":"samepass1"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No user in context → 401", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(validChangePasswordBody))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("UserID mismatch → 401", func() {
			r := withUser(f.newReq(`{"UserID":"other-user","OldPassword":"oldpass12","NewPassword":"newpass12"}`))
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestChangePasswordServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/users/auth/password/change — service outcomes", t, func() {
		f := newChangePasswordFixture()

		Convey("Service ErrWrongPassword → 422", func() {
			f.svcErr = authdomain.ErrWrongPassword
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Service ErrSamePassword → 422", func() {
			f.svcErr = authdomain.ErrSamePassword
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.newReq(validChangePasswordBody)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
