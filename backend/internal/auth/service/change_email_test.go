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
		TokenBase: authdomain.TokenBase{UserID: TestUserID, ExpiresAt: time.Now().UTC().Add(time.Hour)},
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

func initiateEmailChangeGetUserByID(_ context.Context, _ string) (*authdomain.User, error) {
	user := newChangePasswordTestUser()
	user.Email = "old@example.com"
	return user, nil
}

func validEmailChangeRequest() authdomain.EmailChangeRequest {
	return authdomain.EmailChangeRequest{Token: "tok", UserID: TestUserID}
}

func validRequestEmailChangeRequest() authdomain.RequestEmailChangeRequest {
	return authdomain.RequestEmailChangeRequest{UserID: TestUserID, NewEmail: "new@example.com", Password: "correct-old-password"}
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

func TestInitiateEmailChangePassword(t *testing.T) {
	Convey("Given an auth service", t, func() {
		var incremented bool
		var user *authdomain.User
		uRepo := &mockUserRepo{
			getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
			incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
				incremented = true
				return nil
			},
		}
		srv := newTestService(uRepo, nil, nil, nil, nil)
		user = newChangePasswordTestUser()
		user.Email = "old@example.com"

		Convey("When the password is wrong, the failed attempt is counted and nothing else runs", func() {
			req := validRequestEmailChangeRequest()
			req.Password = "wrong-password"

			err := srv.InitiateEmailChange(context.Background(), req)

			So(errors.Is(err, authdomain.ErrWrongPassword), ShouldBeTrue)
			So(incremented, ShouldBeTrue)
		})

		Convey("When the account is locked, it is rejected before the password is checked", func() {
			lockedUntil := time.Now().UTC().Add(time.Hour)
			user.LoginLockedUntil = &lockedUntil

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			var lockedErr *authdomain.ErrAccountLockedError
			So(errors.As(err, &lockedErr), ShouldBeTrue)
			So(incremented, ShouldBeFalse)
		})

		Convey("When the password is wrong and the new email is taken, the password error wins", func() {
			uRepo.getUserByEmail = func(_ context.Context, _ string) (*authdomain.User, error) {
				return &authdomain.User{ID: "someone-else"}, nil
			}
			req := validRequestEmailChangeRequest()
			req.Password = "wrong-password"

			err := srv.InitiateEmailChange(context.Background(), req)

			So(errors.Is(err, authdomain.ErrWrongPassword), ShouldBeTrue)
			So(errors.Is(err, authdomain.ErrEmailAlreadyInUse), ShouldBeFalse)
		})
	})
}

func TestInitiateEmailChangeSameEmail(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the new email equals the current email", func() {
			uRepo := &mockUserRepo{
				getUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: TestUserID, Email: "new@example.com"}, nil
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

			So(err, ShouldBeNil)
		})

		Convey("When the new email is taken by another user, no token is created and nothing is emitted", func() {
			var tokenCreated bool
			var captured []any
			uRepo := &mockUserRepo{
				getUserByID: initiateEmailChangeGetUserByID,
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "someone-else"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createEmailChangeToken: func(_ context.Context, tok *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
					tokenCreated = true
					return tok, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, testutil.NewCapturingOutbox(&captured), nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldBeNil)
			So(tokenCreated, ShouldBeFalse)
			So(captured, ShouldBeEmpty)
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
					aliceName := "Alice"
					return &authdomain.UserProfile{UserID: TestUserID, FirstName: &aliceName}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createEmailChangeToken: func(_ context.Context, tok *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
					return tok, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, testutil.NewCapturingOutbox(&captured), nil)

			err := srv.InitiateEmailChange(context.Background(), validRequestEmailChangeRequest())

			So(err, ShouldBeNil)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "email")
			So(captured[1], ShouldEqual, TestUserID)
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
					return &authdomain.UserProfile{UserID: TestUserID}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createEmailChangeToken: func(_ context.Context, _ *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, testutil.NewNoopOutbox(), nil)

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
						TokenBase: authdomain.TokenBase{UserID: TestUserID, ExpiresAt: time.Now().UTC().Add(-time.Hour)},
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
				getEmailChangeToken:      validChangeEmailToken,
				markEmailChangeTokenUsed: testutil.AlwaysFailsDB,
			}
			uRepo := validChangeEmailUserRepo()
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.ChangeEmail(context.Background(), req)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mark token used")
		})
	})
}

func TestChangeEmailWithSessionLogout(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the change is applied, all sessions are revoked", func() {
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
			srv := newTestService(uRepo, sRepo, tRepo, nil, newSuccessfulMockBlocklist())

			err := srv.ChangeEmail(context.Background(), validEmailChangeRequest())

			So(err, ShouldBeNil)
			So(gotUserID, ShouldEqual, TestUserID)
			So(gotReason, ShouldEqual, authdomain.RevokeReasonEmailChanged)
		})
	})
}

func TestChangeEmailSessionLogoutFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When session revocation fails", func() {
			tRepo := validTokenRepo()
			uRepo := validChangeEmailUserRepo()
			sRepo := &mockSessionRepo{
				revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, sRepo, tRepo, nil, newSuccessfulMockBlocklist())

			err := srv.ChangeEmail(context.Background(), validEmailChangeRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "revoke sessions")
		})
	})
}

func TestChangeEmailRevokesAllTokens(t *testing.T) {
	Convey("Given an auth service", t, func() {
		var revokedUserID string
		var revokeErr error
		blocklist := newSuccessfulMockBlocklist()
		blocklist.revokeUserTokens = func(_ context.Context, userID string, _ time.Duration) error {
			revokedUserID = userID
			return revokeErr
		}
		sRepo := &mockSessionRepo{revokeAllUserSessions: func(_ context.Context, _ string, _ *string, _ authdomain.RevokeReason) error { return nil }}
		srv := newTestService(validChangeEmailUserRepo(), sRepo, validTokenRepo(), nil, blocklist)
		logoutRequest := validEmailChangeRequest()

		Convey("Every access token issued so far is revoked", func() {
			So(srv.ChangeEmail(context.Background(), logoutRequest), ShouldBeNil)
			So(revokedUserID, ShouldEqual, TestUserID)
		})

		Convey("When the Redis revocation fails, the change fails with ErrBlocklistUnavailable", func() {
			revokeErr = testutil.ErrRedisUnavailable

			err := srv.ChangeEmail(context.Background(), logoutRequest)

			So(errors.Is(err, authdomain.ErrBlocklistUnavailable), ShouldBeTrue)
		})
	})
}
