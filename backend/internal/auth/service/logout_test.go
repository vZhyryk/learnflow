package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

// newLogoutTestContext only carries the authenticated user. Service.Logout reads
// JTI/AccessTokenExpiresAt from the request (populated by the HTTP handler from
// context), never from ctx directly — so ctx has no reason to carry them here.
func newLogoutTestContext(user *authdomain.User) context.Context {
	return appcontext.WithUser(context.Background(), user)
}

func validLogoutGetUserSessionByRefreshToken(_ context.Context, _ string) (*authdomain.UserSession, error) {
	return &authdomain.UserSession{ID: "session-123", UserID: "user-123"}, nil
}

func TestLogoutNoUserInContext(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When there is no authenticated user in context", func() {
			srv := newTestService(nil, nil, nil, nil, nil)

			_, err := srv.Logout(context.Background(), authdomain.LogoutRequest{RefreshToken: "ref"})

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func TestLogoutSessionLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When looking up the session fails unexpectedly", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			_, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get session")
		})
	})
}

func TestLogoutSessionAlreadyGone(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session does not exist", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return nil, authdomain.ErrSessionNotFound
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			userID, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(err, ShouldBeNil)
			So(userID, ShouldBeEmpty)
		})
	})
}

func TestLogoutSessionBelongsToAnotherUser(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session belongs to a different user", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return &authdomain.UserSession{ID: "session-123", UserID: "someone-else"}, nil
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			_, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func TestLogoutAlreadyRevoked(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session was already revoked", func() {
			revokedAt := time.Now().UTC()
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return &authdomain.UserSession{ID: "session-123", UserID: "user-123", RevokedAt: &revokedAt}, nil
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			userID, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "user-123")
		})
	})
}

func TestLogoutRevokesActiveSession(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session is active", func() {
			var gotSessionID, gotRevokedBy string
			var gotReason authdomain.RevokeReason
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: validLogoutGetUserSessionByRefreshToken,
				revokeUserSession: func(_ context.Context, sessionID, revokedBy string, reason authdomain.RevokeReason) error {
					gotSessionID, gotRevokedBy, gotReason = sessionID, revokedBy, reason
					return nil
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			userID, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "user-123")
			So(gotSessionID, ShouldEqual, "session-123")
			So(gotRevokedBy, ShouldEqual, "user-123")
			So(gotReason, ShouldEqual, authdomain.RevokeReasonLogout)
		})

		Convey("When revoking the session fails", func() {
			sRepo := &mockSessionRepo{
				getUserSessionByRefreshToken: validLogoutGetUserSessionByRefreshToken,
				revokeUserSession: func(_ context.Context, _, _ string, _ authdomain.RevokeReason) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(nil, sRepo, nil, nil, nil)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			_, err := srv.Logout(ctx, authdomain.LogoutRequest{RefreshToken: "ref"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke sessions")
		})
	})
}

// logoutBlocklistRequest builds a LogoutRequest with a non-empty JTI and a future
// access token expiry so it exercises the SetNX blocklist branch in revokeUserSessions.
// Logout reads JTI/AccessTokenExpiresAt from the request (populated by the HTTP handler
// from context), not from ctx directly.
func logoutBlocklistRequest() authdomain.LogoutRequest {
	return authdomain.LogoutRequest{
		RefreshToken:         "ref",
		JTI:                  "jti-123",
		AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
	}
}

func activeLogoutSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		getUserSessionByRefreshToken: validLogoutGetUserSessionByRefreshToken,
		revokeUserSession: func(_ context.Context, _, _ string, _ authdomain.RevokeReason) error {
			return nil
		},
	}
}

func TestLogoutBlocklistsAccessToken(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session is active and redis blocklist succeeds", func() {
			var redisCalled bool
			var gotKey string
			redisClient := &mockRedis{
				setNX: func(_ context.Context, key string, _ any, _ time.Duration) *redis.BoolCmd {
					redisCalled = true
					gotKey = key
					return redis.NewBoolResult(true, nil)
				},
			}
			sRepo := activeLogoutSessionRepo()
			srv := newTestService(nil, sRepo, nil, nil, redisClient)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			userID, err := srv.Logout(ctx, logoutBlocklistRequest())

			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "user-123")
			So(redisCalled, ShouldBeTrue)
			So(gotKey, ShouldEqual, "blocklist:jti-123")
		})
	})
}

func TestLogoutBlocklistFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the session is active but redis blocklist fails", func() {
			redisClient := mockRedisSetNXError(testutil.ErrRedisUnavailable)
			sRepo := activeLogoutSessionRepo()
			srv := newTestService(nil, sRepo, nil, nil, redisClient)
			ctx := newLogoutTestContext(&authdomain.User{ID: "user-123"})

			_, err := srv.Logout(ctx, logoutBlocklistRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "session blocklist")
		})
	})
}
