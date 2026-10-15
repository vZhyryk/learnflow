package usershttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learnflow_backend/internal/shared/testutil"
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
		handler := usershttp.NewHTTPHandler(svc, testutil.NewTestLogger())
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux, noopChain())

		Convey("When no user in context", func() {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody)
			So(func() { testutil.ServeHTTP(mux, r) }, ShouldPanic)
		})

		Convey("When the profile exists", func() {
			svcResult = &usersdomain.UserProfile{UserID: "user-123"}
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := testutil.ServeHTTP(mux, r)
			So(w.Code, ShouldEqual, http.StatusOK)
			body := decodeBody(t, w.Body.Bytes())
			So(body["user"], ShouldNotBeNil)
		})

		Convey("When the profile exists and the success response write fails", func() {
			svcResult = &usersdomain.UserProfile{UserID: "user-123"}
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			So(func() { mux.ServeHTTP(&errWriter{}, r) }, ShouldNotPanic)
		})

		Convey("When the profile is not found", func() {
			svcErr = usersdomain.ErrUserNotFound
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := testutil.ServeHTTP(mux, r)
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("When the profile is not found and the error response write also fails", func() {
			svcErr = usersdomain.ErrUserNotFound
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			So(func() { mux.ServeHTTP(&errWriter{}, r) }, ShouldNotPanic)
		})

		Convey("When the service returns an unexpected error", func() {
			svcErr = testutil.ErrDBUnexpected
			r := withUser(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/users/profile", http.NoBody))
			w := testutil.ServeHTTP(mux, r)
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

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
	handler := usershttp.NewHTTPHandler(svc, testutil.NewTestLogger())
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
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{bad json`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When the body is invalid JSON and the error response write also fails", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, withUser(f.patch(`{bad json`))) }, ShouldNotPanic)
		})

		Convey("When a field fails validation", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{"gender":"invalid_value"}`)))
			So(w.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("When a field fails validation and the error response write also fails", func() {
			So(func() {
				f.mux.ServeHTTP(&errWriter{}, withUser(f.patch(`{"gender":"invalid_value"}`)))
			}, ShouldNotPanic)
		})

		Convey("When no user in context", func() {
			So(func() { testutil.ServeHTTP(f.mux, f.patch(`{"first_name":"Jane"}`)) }, ShouldPanic)
		})
	})
}

func TestChangeProfileServiceOutcomes(t *testing.T) {
	Convey("PATCH /api/v1/users/profile — service outcomes", t, func() {
		f := newChangeProfileFixture()

		Convey("When the update succeeds", func() {
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(decodeBody(t, w.Body.Bytes())["message"], ShouldNotBeNil)
		})

		Convey("When the update succeeds and the success response write fails", func() {
			So(func() { f.mux.ServeHTTP(&errWriter{}, withUser(f.patch(`{"first_name":"Jane"}`))) }, ShouldNotPanic)
		})

		Convey("When the profile is not found", func() {
			f.svcErr = usersdomain.ErrUserNotFound
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("When the service returns a domain validation error", func() {
			f.svcErr = usersdomain.ErrFirstNameInvalid
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusUnprocessableEntity)
		})

		Convey("When the service returns an unexpected error", func() {
			f.svcErr = testutil.ErrDBUnexpected
			w := testutil.ServeHTTP(f.mux, withUser(f.patch(`{"first_name":"Jane"}`)))
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}
