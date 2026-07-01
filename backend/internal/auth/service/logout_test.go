package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func newLogoutTestContext(user *authdomain.User) context.Context {
	ctx := appcontext.WithUser(context.Background(), user)
	ctx = appcontext.WithJTI(ctx, "jti-123")
	ctx = appcontext.WithAccessTokenExpiresAt(ctx, time.Now().UTC().Add(15*time.Minute))
	return ctx
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
					return nil, errors.New("db connection lost")
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
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return &authdomain.UserSession{ID: "session-123", UserID: "user-123"}, nil
				},
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
				getUserSessionByRefreshToken: func(_ context.Context, _ string) (*authdomain.UserSession, error) {
					return &authdomain.UserSession{ID: "session-123", UserID: "user-123"}, nil
				},
				revokeUserSession: func(_ context.Context, _, _ string, _ authdomain.RevokeReason) error {
					return errors.New("db connection lost")
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
