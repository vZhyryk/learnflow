package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func validGetUserByID(_ context.Context, _ string) (*authdomain.User, error) {
	return newChangePasswordTestUser(), nil
}

func validChangePasswordUserRepo() *mockUserRepo {
	return &mockUserRepo{
		getUserByID:        validGetUserByID,
		updatePasswordHash: testutil.AlwaysNil2,
		resetFailedLogin:   testutil.AlwaysNil,
	}
}

func validChangePasswordSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
			return nil
		},
	}
}

func validChangePasswordRequest() authdomain.ChangePasswordRequest {
	return authdomain.ChangePasswordRequest{
		UserID:      TestUserID,
		OldPassword: "correct-old-password",
		NewPassword: "new-password",
	}
}

// changePasswordLogoutRequest builds a ChangePasswordRequest with IsAllSessionsLogout
// enabled, varying only the access token expiry so tests can exercise the blocklist
// skip/apply branches in revokeUserSessions.
func changePasswordLogoutRequest(accessTokenExpiresAt time.Time) authdomain.ChangePasswordRequest {
	return authdomain.ChangePasswordRequest{
		UserID:               TestUserID,
		OldPassword:          "correct-old-password",
		NewPassword:          "new-password",
		IsAllSessionsLogout:  true,
		JTI:                  "jti-123",
		AccessTokenExpiresAt: accessTokenExpiresAt,
	}
}

func TestChangePasswordUserLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{UserID: TestUserID})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})
	})
}

func TestChangePasswordWrongOldPassword(t *testing.T) {
	wrongPasswordRequest := authdomain.ChangePasswordRequest{UserID: TestUserID, OldPassword: "wrong-old-password", NewPassword: "new-password"}

	Convey("Given an auth service", t, func() {
		var incremented bool
		var gotUserID, gotInterval string
		var gotLimit int
		var incErr error
		uRepo := &mockUserRepo{
			getUserByID: validGetUserByID,
			incrementFailedLogin: func(_ context.Context, userID, lockInterval string, limit int) error {
				incremented, gotUserID, gotInterval, gotLimit = true, userID, lockInterval, limit
				return incErr
			},
		}
		srv := newTestService(uRepo, nil, nil, nil, nil)

		Convey("When the old password does not match, the failed attempt is counted with the login lock settings", func() {
			err := srv.ChangePassword(context.Background(), wrongPasswordRequest)

			So(errors.Is(err, authdomain.ErrWrongPassword), ShouldBeTrue)
			So(incremented, ShouldBeTrue)
			So(gotUserID, ShouldEqual, TestUserID)
			So(gotInterval, ShouldEqual, loginLockInterval)
			So(gotLimit, ShouldEqual, loginFailLimit)
		})

		Convey("When the user vanished before the counter update, it is still a wrong password", func() {
			incErr = authdomain.ErrUserNotFound

			err := srv.ChangePassword(context.Background(), wrongPasswordRequest)

			So(errors.Is(err, authdomain.ErrWrongPassword), ShouldBeTrue)
		})

		Convey("When counting the failed attempt fails, that error is returned instead", func() {
			incErr = testutil.ErrDBUnexpected

			err := srv.ChangePassword(context.Background(), wrongPasswordRequest)

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "increment failed login")
		})
	})
}

func TestChangePasswordLockedAccount(t *testing.T) {
	Convey("Given a user whose login is locked", t, func() {
		var incremented, updated bool
		lockedUntil := time.Now().UTC().Add(10 * time.Minute)
		user := newChangePasswordTestUser()
		user.LoginLockedUntil = &lockedUntil
		uRepo := &mockUserRepo{
			getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
			incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
				incremented = true
				return nil
			},
			updatePasswordHash: func(_ context.Context, _, _ string) error {
				updated = true
				return nil
			},
		}
		srv := newTestService(uRepo, nil, nil, nil, nil)

		Convey("A request with the correct old password is rejected with the lock and nothing is changed", func() {
			err := srv.ChangePassword(context.Background(), validChangePasswordRequest())

			var lockedErr *authdomain.ErrAccountLockedError
			So(errors.As(err, &lockedErr), ShouldBeTrue)
			So(lockedErr.LockedUntil.Equal(lockedUntil), ShouldBeTrue)
			So(updated, ShouldBeFalse)
		})

		Convey("A request with a wrong old password is rejected with the lock and is not counted again", func() {
			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{UserID: TestUserID, OldPassword: "wrong", NewPassword: "new-password"})

			var lockedErr *authdomain.ErrAccountLockedError
			So(errors.As(err, &lockedErr), ShouldBeTrue)
			So(incremented, ShouldBeFalse)
		})

		Convey("Once the lock has expired, the change goes through", func() {
			expired := time.Now().UTC().Add(-time.Minute)
			user.LoginLockedUntil = &expired
			uRepo.resetFailedLogin = testutil.AlwaysNil

			So(srv.ChangePassword(context.Background(), validChangePasswordRequest()), ShouldBeNil)
			So(updated, ShouldBeTrue)
		})
	})
}

func TestChangePasswordResetsFailedLogin(t *testing.T) {
	Convey("Given an auth service", t, func() {
		var resetUserID string
		var resetErr error
		uRepo := validChangePasswordUserRepo()
		uRepo.resetFailedLogin = func(_ context.Context, userID string) error {
			resetUserID = userID
			return resetErr
		}
		srv := newTestService(uRepo, nil, nil, nil, nil)

		Convey("When the password is changed, the failed login counter and lock are cleared", func() {
			So(srv.ChangePassword(context.Background(), validChangePasswordRequest()), ShouldBeNil)
			So(resetUserID, ShouldEqual, TestUserID)
		})

		Convey("When clearing the counter fails, the error is returned", func() {
			resetErr = testutil.ErrDBUnexpected

			err := srv.ChangePassword(context.Background(), validChangePasswordRequest())

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "reset failed login")
		})
	})
}

func TestChangePasswordUpdateHashFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When persisting the new password hash fails", func() {
			uRepo := &mockUserRepo{
				getUserByID:        validGetUserByID,
				updatePasswordHash: testutil.AlwaysFailsDB2,
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), validChangePasswordRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update hash")
		})
	})
}

func TestChangePasswordWithoutSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is false", func() {
			var revokeCalled bool
			uRepo := validChangePasswordUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					revokeCalled = true
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), validChangePasswordRequest())

			So(err, ShouldBeNil)
			So(revokeCalled, ShouldBeFalse)
		})
	})
}

func TestChangePasswordWithSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is true and revocation succeeds", func() {
			var gotUserID string
			var gotReason authdomain.RevokeReason
			uRepo := validChangePasswordUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, userID string, _ *string, reason authdomain.RevokeReason) error {
					gotUserID, gotReason = userID, reason
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, newSuccessfulMockBlocklist())

			err := srv.ChangePassword(context.Background(), changePasswordLogoutRequest(time.Now().UTC().Add(15*time.Minute)))

			So(err, ShouldBeNil)
			So(gotUserID, ShouldEqual, TestUserID)
			So(gotReason, ShouldEqual, authdomain.RevokeReasonPasswordChanged)
		})

		Convey("When IsAllSessionsLogout is true and revocation fails", func() {
			uRepo := validChangePasswordUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, newSuccessfulMockBlocklist())

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID:              TestUserID,
				OldPassword:         "correct-old-password",
				NewPassword:         "new-password",
				IsAllSessionsLogout: true,
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke sessions")
		})
	})
}

func TestChangePasswordSessionBlocklistFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is true and blocklisting the JTI fails", func() {
			uRepo := validChangePasswordUserRepo()
			sRepo := validChangePasswordSessionRepo()
			blocklist := mockBlocklistBlockTokenError(testutil.ErrRedisUnavailable)
			srv := newTestService(uRepo, sRepo, nil, nil, blocklist)

			err := srv.ChangePassword(context.Background(), changePasswordLogoutRequest(time.Now().UTC().Add(15*time.Minute)))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "session blocklist")
			So(errors.Is(err, authdomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})
	})
}

func TestChangePasswordSkipsBlocklistWhenTokenExpired(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the access token is already expired, blocklisting is skipped", func() {
			var blocked bool
			uRepo := validChangePasswordUserRepo()
			sRepo := validChangePasswordSessionRepo()
			blocklist := newSuccessfulMockBlocklist()
			blocklist.blockToken = func(_ context.Context, _ string, _ time.Duration) error {
				blocked = true
				return nil
			}
			srv := newTestService(uRepo, sRepo, nil, nil, blocklist)

			err := srv.ChangePassword(context.Background(), changePasswordLogoutRequest(time.Now().UTC().Add(-time.Minute)))

			So(err, ShouldBeNil)
			So(blocked, ShouldBeFalse)
		})
	})
}

func TestChangePasswordRevokesAllTokens(t *testing.T) {
	Convey("Given an auth service", t, func() {
		var revokedUserID string
		var revokedTTL time.Duration
		var revokeErr error
		blocklist := newSuccessfulMockBlocklist()
		blocklist.revokeUserTokens = func(_ context.Context, userID string, ttl time.Duration) error {
			revokedUserID, revokedTTL = userID, ttl
			return revokeErr
		}
		srv := newTestService(validChangePasswordUserRepo(), validChangePasswordSessionRepo(), nil, nil, blocklist)

		Convey("When all sessions are logged out, every access token issued so far is revoked for one token lifetime", func() {
			err := srv.ChangePassword(context.Background(), changePasswordLogoutRequest(time.Now().UTC().Add(15*time.Minute)))

			So(err, ShouldBeNil)
			So(revokedUserID, ShouldEqual, TestUserID)
			So(revokedTTL, ShouldEqual, tokens.BlockMarkTTL)
		})

		Convey("When the sessions stay, the user's other access tokens are not touched", func() {
			err := srv.ChangePassword(context.Background(), validChangePasswordRequest())

			So(err, ShouldBeNil)
			So(revokedUserID, ShouldBeEmpty)
		})

		Convey("When the Redis revocation fails, the change fails with ErrBlocklistUnavailable", func() {
			revokeErr = testutil.ErrRedisUnavailable

			err := srv.ChangePassword(context.Background(), changePasswordLogoutRequest(time.Now().UTC().Add(15*time.Minute)))

			So(errors.Is(err, authdomain.ErrBlocklistUnavailable), ShouldBeTrue)
			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
		})
	})
}
