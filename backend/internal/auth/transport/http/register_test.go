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

func TestRegister(t *testing.T) {
	Convey("POST /api/v1/auth/register", t, func() {
		var svcResult string
		var svcErr error

		svc := &mockService{
			register: func(_ context.Context, _ authdomain.RegisterRequest) (string, error) {
				return svcResult, svcErr
			},
		}
		h := authhttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		h.RegisterRoutes(mux, authhttp.AuthRouteChains{})

		newReq := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
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

		Convey("Invalid email format → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"notanemail","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Email too short → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"a@","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too short → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"short"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Password too long → 400", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"`+strings.Repeat("a", 73)+`"}`))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Service ErrUserAlreadyExists → 202 (email enumeration guard)", func() {
			svcErr = authdomain.ErrUserAlreadyExists
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusAccepted)
		})

		Convey("Unexpected service error → 500", func() {
			svcErr = errors.New("database failure")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Valid request → 201 with message", func() {
			svcResult = "user-123"
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, newReq(`{"Email":"user@example.com","Password":"password123"}`))
			So(w.Code, ShouldEqual, http.StatusCreated)
			body := decodeBody(t, w.Body.Bytes())
			So(body["message"], ShouldNotBeNil)
		})
	})
}
