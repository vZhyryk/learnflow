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

func validChangeEmailToken(_ context.Context, _ string) (*authdomain.EmailChangeToken, error) {
	return &authdomain.EmailChangeToken{
		TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(time.Hour)},
		NewEmail:  "new@example.com",
	}, nil
}

func validChangeEmailUserRepo() *mockUserRepo {
	return &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return nil, authdomain.ErrUserNotFound },
		updateEmail:    testutil.AlwaysNil2,
	}
}

func validTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		getEmailChangeToken:      validChangeEmailToken,
		markEmailChangeTokenUsed: testutil.AlwaysNil,
	}
}

func validGetUserByEmail(_ context.Context, _ string) (*authdomain.User, error) {
	return nil, authdomain.ErrUserNotFound
}

// initiateEmailChangeGetUserByID returns the "current user" fixture shared by
// InitiateEmailChange tests that need a resolvable owner of the email-change request.
func initiateEmailChangeGetUserByID(_ context.Context, _ string) (*authdomain.User, error) {
	return &authdomain.User{ID: "user-123", Email: "old@example.com"}, nil
}

func validEmailChangeRequest() authdomain.EmailChangeRequest {
	return authdomain.EmailChangeRequest{Token: "tok", UserID: "user-123"}
}

func validRequestEmailChangeRequest() authdomain.RequestEmailChangeRequest {
	return authdomain.RequestEmailChangeRequest{UserID: "user-123", NewEmail: "new@example.com"}
}

func TestInitiateEmailChangeUserLookupFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})
	})
}

func TestInitiateEmailChangeSameEmail(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the new email equals the current email", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Email: "new@example.com"}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(errors.Is(err, authdomain.ErrEmailAlreadyInUse), ShouldBeTrue)
		})
	})
}

func TestInitiateEmailChangeAvailabilityCheck(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When checking whether the new email is taken fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByID: initiateEmailChangeGetUserByID,
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "check new email exists")
		})

		Convey("When the new email is already taken by another user", func() {
			uRepo := &mockUserRepo{
				getUserByID: initiateEmailChangeGetUserByID,
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "someone-else"}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(errors.Is(err, authdomain.ErrEmailAlreadyInUse), ShouldBeTrue)
		})
	})
}

func TestInitiateEmailChangeProfileLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When fetching the user profile fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByID:    initiateEmailChangeGetUserByID,
				getUserByEmail: validGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user profile")
		})

		Convey("When the user profile does not exist", func() {
			uRepo := &mockUserRepo{
				getUserByID:    initiateEmailChangeGetUserByID,
				getUserByEmail: validGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(errors.Is(err, authdomain.ErrUserNotFound), ShouldBeTrue)
		})
	})
}

func TestInitiateEmailChangeTokenIssued(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token is issued successfully", func() {
			var captured []any
			uRepo := &mockUserRepo{
				getUserByID:    initiateEmailChangeGetUserByID,
				getUserByEmail: validGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123", FirstName: "Alice"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createEmailChangeToken: func(_ context.Context, tok *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
					return tok, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newCapturingOutbox(&captured), nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldBeNil)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "email")
			So(captured[1], ShouldEqual, "user-123")
		})
	})
}

func TestInitiateEmailChangeTokenCreationFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When creating the token fails", func() {
			uRepo := &mockUserRepo{
				getUserByID:    initiateEmailChangeGetUserByID,
				getUserByEmail: validGetUserByEmail,
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createEmailChangeToken: func(_ context.Context, _ *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newNoopOutbox(), nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create token")
		})
	})
}

func TestChangeEmailTokenLookup(t *testing.T) {
	req := validEmailChangeRequest()

	Convey("Given an auth service", t, func() {
		Convey("When the token lookup fails", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: func(_ context.Context, _ string) (*authdomain.EmailChangeToken, error) {
					return nil, authdomain.ErrInvalidToken
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the token has expired", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: func(_ context.Context, _ string) (*authdomain.EmailChangeToken, error) {
					return &authdomain.EmailChangeToken{
						TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(-time.Hour)},
					}, nil
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(errors.Is(err, authdomain.ErrTokenExpired), ShouldBeTrue)
		})

		Convey("When the token belongs to a different user", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: func(_ context.Context, _ string) (*authdomain.EmailChangeToken, error) {
					return &authdomain.EmailChangeToken{
						TokenBase: authdomain.TokenBase{UserID: "someone-else", ExpiresAt: time.Now().UTC().Add(time.Hour)},
					}, nil
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})
	})
}

func TestChangeEmailNewEmailAvailability(t *testing.T) {
	req := validEmailChangeRequest()

	Convey("Given an auth service", t, func() {
		Convey("When the new email became taken meanwhile", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: validChangeEmailToken,
			}
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "someone-else"}, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(errors.Is(err, authdomain.ErrEmailAlreadyInUse), ShouldBeTrue)
		})

		Convey("When checking new email availability fails unexpectedly", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: validChangeEmailToken,
			}
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "check email taken")
		})
	})
}

func TestChangeEmailApplyFailures(t *testing.T) {
	req := validEmailChangeRequest()

	Convey("Given an auth service", t, func() {
		Convey("When updating the email fails", func() {
			tRepo := &mockTokenRepo{getEmailChangeToken: validChangeEmailToken}
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return nil, authdomain.ErrUserNotFound },
				updateEmail:    testutil.AlwaysFailsDB2,
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update email")
		})

		Convey("When marking the token as used fails", func() {
			tRepo := &mockTokenRepo{
				getEmailChangeToken: validChangeEmailToken,
				markEmailChangeTokenUsed: func(_ context.Context, _ string) error {
					return testutil.ErrDBUnexpected
				},
			}
			uRepo := validChangeEmailUserRepo()
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mark token used")
		})
	})
}

func TestChangeEmailWithoutSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is false", func() {
			var revokeCalled bool
			tRepo := validTokenRepo()
			uRepo := validChangeEmailUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					revokeCalled = true
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), validEmailChangeRequest())

			So(err, ShouldBeNil)
			So(revokeCalled, ShouldBeFalse)
		})
	})
}

func TestChangeEmailWithSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is true and revocation succeeds", func() {
			var gotUserID string
			var gotReason authdomain.RevokeReason
			tRepo := validTokenRepo()
			uRepo := validChangeEmailUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, userID string, _ *string, reason authdomain.RevokeReason) error {
					gotUserID, gotReason = userID, reason
					return nil
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, newSuccessfulMockRedis())

			err := srv.ChangeEmail(context.Background(), authdomain.EmailChangeRequest{
				Token:                "tok",
				UserID:               "user-123",
				IsAllSessionsLogout:  true,
				JTI:                  "jti-123",
				AccessTokenExpiresAt: time.Now().UTC().Add(15 * time.Minute),
			})

			So(err, ShouldBeNil)
			So(gotUserID, ShouldEqual, "user-123")
			So(gotReason, ShouldEqual, authdomain.RevokeReasonEmailChanged)
		})
	})
}

func TestChangeEmailSessionLogoutFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When IsAllSessionsLogout is true and revocation fails", func() {
			tRepo := validTokenRepo()
			uRepo := validChangeEmailUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, newSuccessfulMockRedis())

			err := srv.ChangeEmail(context.Background(), authdomain.EmailChangeRequest{
				Token: "tok", UserID: "user-123", IsAllSessionsLogout: true,
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke sessions")
		})
	})
}
