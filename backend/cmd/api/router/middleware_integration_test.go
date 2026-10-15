//go:build integration

package router

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/rediskeys"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthenticateUserBlocklist_Integration(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { client.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup

	Convey("AuthenticateUser against real Redis", t, func() {
		route := newTestRouteHandler()
		route.App.Redis = redisinfra.NewInstance(client)

		userID := "auth-blocklist-" + time.Now().Format("150405.000000000")
		signed, err := route.token.GenerateAccessToken(&authdomain.User{ID: userID, Role: authdomain.UserRole("student")}, time.Minute)
		So(err, ShouldBeNil)
		claims, err := route.token.ValidateToken(signed)
		So(err, ShouldBeNil)

		ctx := context.Background()
		t.Cleanup(func() {
			client.Del(ctx, rediskeys.UserBlocked(userID), rediskeys.JTIBlocked(claims.ID), rediskeys.TokensRevokedBefore(userID)) //nolint:errcheck // best-effort cleanup
		})

		var gotUser *authdomain.User
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			gotUser, _ = appcontext.UserFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})
		serve := func() int {
			w, r := newTestRequest("/")
			r.Header.Set("Authorization", "Bearer "+signed)
			route.AuthenticateUser(next).ServeHTTP(w, r)
			return w.Code
		}

		Convey("When neither the user nor the token is blocked, next runs with the user in the context", func() {
			So(serve(), ShouldEqual, http.StatusOK)
			So(called, ShouldBeTrue)
			So(gotUser.ID, ShouldEqual, userID)
		})

		Convey("When the user is blocked in Redis, the still-valid token is rejected with 401", func() {
			So(route.App.Redis.BlockUser(ctx, userID, time.Minute), ShouldBeNil)

			So(serve(), ShouldEqual, http.StatusUnauthorized)
			So(called, ShouldBeFalse)
		})

		Convey("When the user is unblocked again, the same token is accepted", func() {
			So(route.App.Redis.BlockUser(ctx, userID, time.Minute), ShouldBeNil)
			So(route.App.Redis.UnBlockUser(ctx, userID), ShouldBeNil)

			So(serve(), ShouldEqual, http.StatusOK)
			So(called, ShouldBeTrue)
		})

		Convey("When all of the user's tokens were revoked after this one was issued, it is rejected with 401", func() {
			So(route.App.Redis.RevokeUserTokens(ctx, userID, time.Minute), ShouldBeNil)

			So(serve(), ShouldEqual, http.StatusUnauthorized)
			So(called, ShouldBeFalse)
		})

		Convey("When the revocation happened before this token was issued, it is accepted", func() {
			olderCutoff := time.Now().Add(-time.Hour).Unix()
			So(client.Set(ctx, rediskeys.TokensRevokedBefore(userID), olderCutoff, time.Minute).Err(), ShouldBeNil)

			So(serve(), ShouldEqual, http.StatusOK)
			So(called, ShouldBeTrue)
		})

		Convey("When another user's tokens were revoked, this token is accepted", func() {
			So(route.App.Redis.RevokeUserTokens(ctx, userID+"-other", time.Minute), ShouldBeNil)
			t.Cleanup(func() { client.Del(ctx, rediskeys.TokensRevokedBefore(userID+"-other")) }) //nolint:errcheck // best-effort cleanup

			So(serve(), ShouldEqual, http.StatusOK)
		})

		Convey("When the token's jti is blocklisted, it is rejected with 401", func() {
			So(client.Set(ctx, rediskeys.JTIBlocked(claims.ID), "1", time.Minute).Err(), ShouldBeNil)

			So(serve(), ShouldEqual, http.StatusUnauthorized)
			So(called, ShouldBeFalse)
		})
	})
}
