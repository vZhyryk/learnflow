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

type registerFixture struct {
	svcResult string
	svcErr    error
	mux       *http.ServeMux
	newReq    func(body string) *http.Request
}

func newRegisterFixture() *registerFixture {
	f := &registerFixture{}
	svc := &mockService{
		register: func(_ context.Context, _ authdomain.RegisterRequest) (string, error) {
			return f.svcResult, f.svcErr
		},
	}
	h := authhttp.NewHTTPHandler(svc, newTestLogger())
	f.mux = http.NewServeMux()
	h.RegisterRoutes(f.mux, authhttp.AuthRouteChains{})
	f.newReq = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	}
	return f
}

func TestRegisterRequestValidation(t *testing.T) {
	Convey("POST /api/v1/auth/register — request validation", t, func() {
		f := newRegisterFixture()

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

		Convey("Invalid email format → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"notanemail","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Email too short → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"a@","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too short → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"user@example.com","Password":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too long → 400", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"user@example.com","Password":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestRegisterServiceOutcomes(t *testing.T) {
	Convey("POST /api/v1/auth/register — service outcomes", t, func() {
		f := newRegisterFixture()

		Convey("Service ErrUserAlreadyExists → 202 (email enumeration guard)", func() {
			f.svcErr = authdomain.ErrUserAlreadyExists
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusAccepted)
		})

		Convey("Unexpected service error → 500", func() {
			f.svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			f.svcResult = "user-123"
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
