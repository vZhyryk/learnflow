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

func TestInitiateEmailChange(t *testing.T) {
	Convey("POST /api/v1/users/auth/email/change", t, func() {
		var svcErr error

		svc := &mockService{
			initiateEmailChange: func(_ context.Context, _ authdomain.RequestEmailChangeRequest) error {
				return svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/users/auth/email/change", strings.NewReader(body))
		}

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Invalid NewEmail format → 400", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"NewEmail":"notanemail"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No user in context → 401", func() {
			w := testutil.ServeHTTP(mux, newReq(`{"NewEmail":"new@example.com"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrEmailAlreadyInUse → 401", func() {
			svcErr = authdomain.ErrEmailAlreadyInUse
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"NewEmail":"new@example.com"}`)))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"NewEmail":"new@example.com"}`)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(mux, withUser(newReq(`{"NewEmail":"new@example.com"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() {
				mux.ServeHTTP(&errWriter{}, withUser(newReq(`{"NewEmail":"new@example.com"}`)))
			}, ShouldNotPanic)
		})
	})
}

type changeEmailFixture struct {
	svcErr error
	mux    *http.ServeMux
	newReq func(body string) *http.Request
}

func newChangeEmailFixture() *changeEmailFixture {
	f := &changeEmailFixture{}
	svc := &mockService{
		changeEmail: func(_ context.Context, _ authdomain.EmailChangeRequest) error {
			return f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, testutil.NewTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/v1/users/auth/email/change", strings.NewReader(body))
	}
	return f
}

func TestChangeEmailRequestValidation(t *testing.T) {
	Convey("PUT /api/v1/users/auth/email/change — request validation", t, func() {
		f := newChangeEmailFixture()

		Convey("Empty body → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(""))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Empty Token → 400", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":""}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("No user in context → 401", func() {
			w := testutil.ServeHTTP(f.mux, f.newReq(`{"Token":"tok"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestChangeEmailServiceOutcomes(t *testing.T) {
	Convey("PUT /api/v1/users/auth/email/change — service outcomes", t, func() {
		f := newChangeEmailFixture()

		Convey("Service ErrTokenExpired → 400", func() {
			f.svcErr = authdomain.ErrTokenExpired
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(`{"Token":"tok"}`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrTokenUsed → 401", func() {
			f.svcErr = authdomain.ErrTokenUsed
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(`{"Token":"tok"}`)))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Service ErrInvalidToken → 401", func() {
			f.svcErr = authdomain.ErrInvalidToken
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(`{"Token":"tok"}`)))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(`{"Token":"tok"}`)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 200 with message", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.newReq(`{"Token":"tok"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})

		Convey("Valid request and the success response write fails → does not panic", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, withUser(f.newReq(`{"Token":"tok"}`))) }, ShouldNotPanic)
		})
	})
}
