//go:build integration

package router

import (
	"context"
	"learnflow_backend/cmd/api/app"
	authdomain "learnflow_backend/internal/auth/domain"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	smokeSecret   = "router-smoke-secret-that-is-long-enough-32b"
	smokeIssuer   = "learnflow"
	smokeAudience = "learnflow-users"
)

func newSmokeRouter(t *testing.T) (*RouteHandler, *redisinfra.Instance) {
	t.Helper()

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { client.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup
	redisInstance := &redisinfra.Instance{Client: client}

	a := &app.App{Logger: testutil.NewTestLogger(), DB: testutil.NewTestPool(t), Redis: redisInstance}
	a.Config.Secret.JWTSecret = smokeSecret
	a.Config.Secret.JWTIssuer = smokeIssuer
	a.Config.Secret.JWTAudience = smokeAudience
	a.Config.Timeouts.RequestTimeout = 5 * time.Second
	a.Config.Limiter.Enabled = false
	a.Config.Cors.TrustedOrigins = map[string]struct{}{"http://localhost:3000": {}}

	route, err := NewRouter(a)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}

	return route, redisInstance
}

func serve(route *RouteHandler, method, path, bearer string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), method, path, http.NoBody)
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	route.Router.ServeHTTP(w, r)

	return w
}

func TestNewRouterSmoke_Integration(t *testing.T) {
	Convey("Given the fully wired router on real Postgres and Redis", t, func() {
		route, redisInstance := newSmokeRouter(t)
		signer := tokens.NewTokens(smokeSecret, "", smokeIssuer, smokeAudience)
		tokenFor := func(userID string, role authdomain.UserRole) string {
			signed, err := signer.GenerateAccessToken(&authdomain.User{ID: userID, Role: role}, time.Minute)
			So(err, ShouldBeNil)
			return signed
		}
		userID := "11111111-1111-4111-8111-" + time.Now().Format("150405000000")
		t.Cleanup(func() { redisInstance.UnBlockUser(context.Background(), userID) }) //nolint:errcheck // best-effort cleanup

		Convey("The helper routes respond", func() {
			So(serve(route, http.MethodGet, "/health", "").Code, ShouldEqual, http.StatusOK)
			So(serve(route, http.MethodGet, "/readiness", "").Code, ShouldEqual, http.StatusOK)
			So(serve(route, http.MethodGet, "/metrics", "").Code, ShouldEqual, http.StatusOK)
		})

		Convey("An unknown path returns 404", func() {
			So(serve(route, http.MethodGet, "/no-such-route", "").Code, ShouldEqual, http.StatusNotFound)
		})

		Convey("The admin users route requires a token", func() {
			So(serve(route, http.MethodGet, "/api/v1/admin/users", "").Code, ShouldEqual, http.StatusUnauthorized)
		})

		Convey("The admin users route rejects a student token with 403", func() {
			w := serve(route, http.MethodGet, "/api/v1/admin/users", tokenFor(userID, authdomain.UserRole("student")))

			So(w.Code, ShouldEqual, http.StatusForbidden)
		})

		Convey("The admin users route accepts an admin token", func() {
			w := serve(route, http.MethodGet, "/api/v1/admin/users", tokenFor(userID, authdomain.RoleAdmin))

			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("A blocked admin's still-valid token is rejected with 401 before the role check", func() {
			token := tokenFor(userID, authdomain.RoleAdmin)
			So(redisInstance.BlockUser(context.Background(), userID, time.Minute), ShouldBeNil)

			So(serve(route, http.MethodGet, "/api/v1/admin/users", token).Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}
