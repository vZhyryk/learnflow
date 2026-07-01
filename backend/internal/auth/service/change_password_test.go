package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"

	. "github.com/smartystreets/goconvey/convey"
)

func newSuccessfulMockRedis() *mockRedis {
	return &mockRedis{
		setNX: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.BoolCmd {
			return redis.NewBoolResult(true, nil)
		},
	}
}

func newChangePasswordTestUser() *authdomain.User {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-old-password"), 4)
	if err != nil {
		panic(err)
	}
	return &authdomain.User{ID: "user-123", PasswordHash: string(hash)}
}

func TestChangePasswordUserLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{UserID: "user-123"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})
	})
}

func TestChangePasswordWrongOldPassword(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the old password does not match", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "wrong-old-password", NewPassword: "new-password",
			})

			So(errors.Is(err, authdomain.ErrWrongPassword), ShouldBeTrue)
		})
	})
}

func TestChangePasswordUpdateHashFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When persisting the new password hash fails", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error {
					return errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update hash")
		})
	})
}

func TestChangePasswordWithoutSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is false", func() {
			var revokeCalled bool
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					revokeCalled = true
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
			})

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
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, userID string, _ *string, reason authdomain.RevokeReason) error {
					gotUserID, gotReason = userID, reason
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, newSuccessfulMockRedis())

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
				IsAllSessionsLogout: true, JTI: "jti-123", AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			})

			So(err, ShouldBeNil)
			So(gotUserID, ShouldEqual, "user-123")
			So(gotReason, ShouldEqual, authdomain.RevokeReasonPasswordChanged)
		})

		Convey("When IsAllSessionsLogout is true and revocation fails", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, newSuccessfulMockRedis())

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
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
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return nil
				},
			}
			redisClient := &mockRedis{
				setNX: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.BoolCmd {
					return redis.NewBoolResult(false, errors.New("redis unavailable"))
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, redisClient)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
				IsAllSessionsLogout: true, JTI: "jti-123", AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "session blocklist")
		})
	})
}

func TestChangePasswordSkipsBlocklistWhenTokenExpired(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the access token is already expired, blocklisting is skipped", func() {
			var redisCalled bool
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newChangePasswordTestUser(), nil
				},
				updatePasswordHash: func(_ context.Context, _, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return nil
				},
			}
			redisClient := &mockRedis{
				setNX: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.BoolCmd {
					redisCalled = true
					return redis.NewBoolResult(true, nil)
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, redisClient)

			err := srv.ChangePassword(context.Background(), authdomain.ChangePasswordRequest{
				UserID: "user-123", OldPassword: "correct-old-password", NewPassword: "new-password",
				IsAllSessionsLogout: true, JTI: "jti-123", AccessTokenExpiresAt: time.Now().UTC().Add(-time.Minute),
			})

			So(err, ShouldBeNil)
			So(redisCalled, ShouldBeFalse)
		})
	})
}
