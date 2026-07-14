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

func TestVerifyEmailTokenLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token lookup fails", func() {
			tRepo := &mockTokenRepo{
				getEmailVerificationToken: func(_ context.Context, _ string) (*authdomain.EmailVerificationToken, error) {
					return nil, authdomain.ErrInvalidToken
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			_, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "tok"})

			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the token has expired", func() {
			tRepo := &mockTokenRepo{
				getEmailVerificationToken: func(_ context.Context, _ string) (*authdomain.EmailVerificationToken, error) {
					return &authdomain.EmailVerificationToken{
						TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(-time.Hour)},
					}, nil
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			_, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "tok"})

			So(errors.Is(err, authdomain.ErrTokenExpired), ShouldBeTrue)
		})
	})
}

func validVerifyEmailToken(_ context.Context, _ string) (*authdomain.EmailVerificationToken, error) {
	return &authdomain.EmailVerificationToken{
		TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, nil
}

func TestVerifyEmailUserUpdateFailures(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When marking the email verified fails", func() {
			uRepo := &mockUserRepo{
				updateEmailVerifiedAt: func(_ context.Context, _ string) error {
					return testutil.ErrDBUnexpected
				},
			}
			tRepo := &mockTokenRepo{getEmailVerificationToken: validVerifyEmailToken}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			_, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "UpdateEmailVerifiedAt")
		})

		Convey("When activating the account status fails", func() {
			uRepo := &mockUserRepo{
				updateEmailVerifiedAt: testutil.AlwaysNil,
				updateStatus: func(_ context.Context, _ string, _ authdomain.UserStatus) error {
					return testutil.ErrDBUnexpected
				},
			}
			tRepo := &mockTokenRepo{getEmailVerificationToken: validVerifyEmailToken}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			_, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "UpdateStatus")
		})
	})
}

func TestVerifyEmailMarkTokenUsedFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When marking the token as used fails", func() {
			uRepo := &mockUserRepo{
				updateEmailVerifiedAt: testutil.AlwaysNil,
				updateStatus:          func(_ context.Context, _ string, _ authdomain.UserStatus) error { return nil },
			}
			tRepo := &mockTokenRepo{
				getEmailVerificationToken: validVerifyEmailToken,
				markEmailVerificationTokenUsed: func(_ context.Context, _ string) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			_, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "MarkEmailVerificationTokenUsed")
		})
	})
}

func TestVerifyEmailSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token is valid and unused", func() {
			var gotUpdatedStatusUserID string
			var gotMarkedUsedHash string
			uRepo := &mockUserRepo{
				updateEmailVerifiedAt: testutil.AlwaysNil,
				updateStatus: func(_ context.Context, userID string, status authdomain.UserStatus) error {
					gotUpdatedStatusUserID = userID
					So(status, ShouldEqual, authdomain.StatusActive)
					return nil
				},
			}
			tRepo := &mockTokenRepo{
				getEmailVerificationToken: validVerifyEmailToken,
				markEmailVerificationTokenUsed: func(_ context.Context, hash string) error {
					gotMarkedUsedHash = hash
					return nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			userID, err := srv.VerifyEmail(context.Background(), authdomain.VerifyEmailRequest{Token: "raw-token"})

			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "user-123")
			So(gotUpdatedStatusUserID, ShouldEqual, "user-123")
			So(gotMarkedUsedHash, ShouldNotBeEmpty)
		})
	})
}
