package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitiatePasswordResetUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiatePasswordReset(context.Background(), authdomain.RequestPasswordResetRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})

		Convey("When the user does not exist (silent no-op, prevents enumeration)", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiatePasswordReset(context.Background(), authdomain.RequestPasswordResetRequest{Email: "user@example.com"})

			So(err, ShouldBeNil)
		})
	})
}

func validInitiateResetGetUserByEmail(_ context.Context, _ string) (*authdomain.User, error) {
	return &authdomain.User{ID: "user-123", Email: "user@example.com"}, nil
}

func TestInitiatePasswordResetProfileLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When fetching the user profile fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: validInitiateResetGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiatePasswordReset(context.Background(), authdomain.RequestPasswordResetRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user profile")
		})
	})
}

func TestInitiatePasswordResetSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user exists and the token is issued", func() {
			var captured []any
			uRepo := &mockUserRepo{
				getUserByEmail: validInitiateResetGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123", FirstName: "Alice"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createPasswordResetToken: func(_ context.Context, t *authdomain.PasswordResetToken) (*authdomain.PasswordResetToken, error) {
					return t, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newCapturingOutbox(&captured), nil)

			err := srv.InitiatePasswordReset(context.Background(), authdomain.RequestPasswordResetRequest{Email: "user@example.com"})

			So(err, ShouldBeNil)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "password")
			So(captured[1], ShouldEqual, "user-123")
		})

		Convey("When creating the reset token fails", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: validInitiateResetGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createPasswordResetToken: func(_ context.Context, _ *authdomain.PasswordResetToken) (*authdomain.PasswordResetToken, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newNoopOutbox(), nil)

			err := srv.InitiatePasswordReset(context.Background(), authdomain.RequestPasswordResetRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create token")
		})
	})
}

func TestResetPasswordTokenLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token lookup fails", func() {
			tRepo := &mockTokenRepo{
				getPasswordResetToken: func(_ context.Context, _ string) (*authdomain.PasswordResetToken, error) {
					return nil, authdomain.ErrInvalidToken
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the token has expired", func() {
			tRepo := &mockTokenRepo{
				getPasswordResetToken: func(_ context.Context, _ string) (*authdomain.PasswordResetToken, error) {
					return &authdomain.PasswordResetToken{
						TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(-time.Hour)},
					}, nil
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(errors.Is(err, authdomain.ErrTokenExpired), ShouldBeTrue)
		})
	})
}

func validResetPasswordToken(_ context.Context, _ string) (*authdomain.PasswordResetToken, error) {
	return &authdomain.PasswordResetToken{
		TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, nil
}

func validResetPasswordGetUserByID(_ context.Context, _ string) (*authdomain.User, error) {
	return &authdomain.User{ID: "user-123"}, nil
}

func TestResetPasswordUserLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails", func() {
			tRepo := &mockTokenRepo{getPasswordResetToken: validResetPasswordToken}
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})

		Convey("When persisting the new password hash fails", func() {
			tRepo := &mockTokenRepo{getPasswordResetToken: validResetPasswordToken}
			uRepo := &mockUserRepo{
				getUserByID:        validResetPasswordGetUserByID,
				updatePasswordHash: testutil.AlwaysFailsDB2,
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update hash")
		})
	})
}

func TestResetPasswordMarkTokenUsedFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When marking the token as used fails", func() {
			tRepo := &mockTokenRepo{
				getPasswordResetToken: validResetPasswordToken,
				markPasswordResetTokenUsed: func(_ context.Context, _ string) error {
					return testutil.ErrDBUnexpected
				},
			}
			uRepo := &mockUserRepo{
				getUserByID:        validResetPasswordGetUserByID,
				updatePasswordHash: testutil.AlwaysNil2,
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mark token used")
		})
	})
}

func TestResetPasswordRevokeSessionsFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When revoking existing sessions fails", func() {
			tRepo := &mockTokenRepo{
				getPasswordResetToken:      validResetPasswordToken,
				markPasswordResetTokenUsed: testutil.AlwaysNil,
			}
			uRepo := &mockUserRepo{
				getUserByID:        validResetPasswordGetUserByID,
				updatePasswordHash: testutil.AlwaysNil2,
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke sessions")
		})
	})
}

func TestResetPasswordSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token is valid and the password resets", func() {
			var gotRevokeUserID string
			var gotReason authdomain.RevokeReason
			tRepo := &mockTokenRepo{
				getPasswordResetToken:      validResetPasswordToken,
				markPasswordResetTokenUsed: testutil.AlwaysNil,
			}
			uRepo := &mockUserRepo{
				getUserByID:        validResetPasswordGetUserByID,
				updatePasswordHash: testutil.AlwaysNil2,
			}
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, userID string, _ *string, reason authdomain.RevokeReason) error {
					gotRevokeUserID, gotReason = userID, reason
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, nil)

			err := srv.ResetPassword(context.Background(), authdomain.ResetPasswordRequest{Token: "tok", NewPassword: "new-password"})

			So(err, ShouldBeNil)
			So(gotRevokeUserID, ShouldEqual, "user-123")
			So(gotReason, ShouldEqual, authdomain.RevokeReasonPasswordReset)
		})
	})
}
