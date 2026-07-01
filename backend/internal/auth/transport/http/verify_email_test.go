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

type verifyEmailFixture struct {
	svcResult string
	svcErr    error
	mux       *http.ServeMux
	newReq    func(body string) *http.Request
}

func newVerifyEmailFixture() *verifyEmailFixture {
	f := &verifyEmailFixture{}
	svc := &mockService{
		verifyEmail: func(_ context.Context, _ authdomain.VerifyEmailRequest) (string, error) {
			return f.svcResult, f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, newTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/users/auth/email/verify", strings.NewReader(body))
	}
	return f
}

func TestVerifyEmailRequestValidation(t *testing.T) {
	Convey("POST /api/v1/users/auth/email/verify — request validation", t, func() {
		f := newVerifyEmailFixture()

		Convey("Empty body → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid JSON → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq("{invalid"))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestVerifyEmailServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/users/auth/email/verify — service outcomes", t, func() {
		f := newVerifyEmailFixture()

		Convey("Service ErrTokenExpired → 400", func() {
			f.svcErr = authdomain.ErrTokenExpired
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			f.svcErr = authdomain.ErrTokenUsed
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			f.svcErr = authdomain.ErrInvalidToken
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid token → 200 with message", func() {
			f.svcResult = "user-123"
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
