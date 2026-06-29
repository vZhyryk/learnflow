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

func TestChangeProfile(t *testing.T) {
	Convey("PATCH /api/v1/users/profile", t, func() {
		var svcErr error

		svc := &mockService{
			changeUserProfile: func(_ context.Context, _ usersdomain.ChangeUserProfileRequest) error {
				return svcErr
			},
		}
		handler := usershttp.NewHTTPHandler(svc, newTestLogger())
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, noopChain())

		patch := func(body string) *http.Request {
			return httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/api/v1/users/profile", strings.NewReader(body))
		}

		Convey("When the body is invalid JSON", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{bad json`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When a field fails validation", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{"gender":"invalid_value"}`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When no user in context", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, patch(`{"first_name":"Jane"}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("When the update succeeds", func() {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(decodeBody(t, w.Body.Bytes())["message"], ShouldNotBeNil)
		})

		Convey("When the profile is not found", func() {
			svcErr = usersdomain.ErrUserNotFound
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("When the service returns a domain validation error", func() {
			svcErr = usersdomain.ErrFirstNameInvalid
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("When the service returns an unexpected error", func() {
			svcErr = errors.New("db down")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, withUser(patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
