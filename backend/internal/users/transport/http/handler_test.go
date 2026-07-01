package usershttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usersdomain "learnflow_backend/internal/users/domain"
	usershttp "learnflow_backend/internal/users/transport/http"

	. "github.com/smartystreets/goconvey/convey"
)

// --- GET /api/v1/users/profile ---

func TestGetProfile(t *testing.T) {
	Convey("GET /api/v1/users/profile", t, func() {
		var svcResult *usersdomain.UserProfile
		var svcErr error

		svc := &mockService{
			getUserProfile: func(_ context.Context, _ string) (*usersdomain.UserProfile, error) {
				return svcResult, svcErr
			},
		}
		handler := usershttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, noopChain())

		Convey("When no user in context", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("When the profile exists", func() {
			svcResult = &usersdomain.UserProfile{UserID: "user-123"}
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["user"], ShouldNotBeNil)
		})

		Convey("When the profile is not found", func() {
			svcErr = usersdomain.ErrUserNotFound
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("When the service returns an unexpected error", func() {
			svcErr = errors.New("db down")
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

// --- PATCH /api/v1/users/profile ---

type changeProfileFixture struct {
	svcErr error
	mux    *http.ServeMux
	patch  func(body string) *http.Request
}

func newChangeProfileFixture() *changeProfileFixture {
	f := &changeProfileFixture{}
	svc := &mockService{
		changeUserProfile: func(_ context.Context, _ usersdomain.ChangeUserProfileRequest) error {
			return f.svcErr
		},
	}
	handler := usershttp.NewHTTPHandler(svc, newTestLogger())
	f.mux = http.NewServeMux()
	handler.RegisterRoutes(f.mux, noopChain())
	f.patch = func(body string) *http.Request {
		return httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/api/v1/users/profile", strings.NewReader(body))
	}
	return f
}

func TestChangeProfileRequestValidation(t *testing.T) {
	Convey("PATCH /api/v1/users/profile — request validation", t, func() {
		f := newChangeProfileFixture()

		Convey("When the body is invalid JSON", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{bad json`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When a field fails validation", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{"gender":"invalid_value"}`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When no user in context", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, f.patch(`{"first_name":"Jane"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func TestChangeProfileServiceOutcomes(t *testing.T) {
	Convey("PATCH /api/v1/users/profile — service outcomes", t, func() {
		f := newChangeProfileFixture()

		Convey("When the update succeeds", func() {
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(decodeBody(t, w.Body.Bytes())["message"], ShouldNotBeNil)
		})

		Convey("When the profile is not found", func() {
			f.svcErr = usersdomain.ErrUserNotFound
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("When the service returns a domain validation error", func() {
			f.svcErr = usersdomain.ErrFirstNameInvalid
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("When the service returns an unexpected error", func() {
			f.svcErr = errors.New("db down")
			w := httptest.NewRecorder()
			f.mux.ServeHTTP(w, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
