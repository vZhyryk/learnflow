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

func TestInitRecoverAccountUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})

		Convey("When no deleted user exists (silent no-op, prevents enumeration)", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldBeNil)
		})

		Convey("When the account is not actually deleted", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: TestUserID, Status: authdomain.StatusActive}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(errors.Is(err, authdomain.ErrInvalidAccountState), ShouldBeTrue)
		})
	})
}

func fakeRecoverAccountDeletedUser() *authdomain.User {
	return &authdomain.User{ID: TestUserID, Email: "user@example.com", Status: authdomain.StatusDeleted}
}

func TestInitRecoverAccountDeletedByAdmin(t *testing.T) {
	deletedUser := fakeRecoverAccountDeletedUser()

	Convey("Given a soft-deleted user", t, func() {
		// getUserProfileByUserID is left unset: reaching it would panic, proving the check short-circuits.
		uRepo := &mockUserRepo{
			getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
				return deletedUser, nil
			},
		}
		req := authdomain.RequestRecoverAccountRequest{Email: "user@example.com"}

		Convey("When an admin deleted the account, recovery is refused before any token is issued", func() {
			srv := newTestService(uRepo, nil, nil, nil, nil)
			srv.audit = mockAuditDeletedByAdmin(true, nil)

			err := srv.InitRecoverAccount(context.Background(), req)

			So(errors.Is(err, authdomain.ErrDeletedByAdmin), ShouldBeTrue)
		})

		Convey("When the audit lookup fails, the error is wrapped", func() {
			srv := newTestService(uRepo, nil, nil, nil, nil)
			srv.audit = mockAuditDeletedByAdmin(false, testutil.ErrDBUnexpected)

			err := srv.InitRecoverAccount(context.Background(), req)

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "check deleted by admin")
		})
	})
}

func TestRecoverAccountDeletedByAdmin(t *testing.T) {
	Convey("Given a valid token for a soft-deleted user", t, func() {
		tRepo := &mockTokenRepo{getAccountRecoveryToken: validRecoverAccountToken}
		// restoreUser is left unset: reaching it would panic, proving the account is not restored.
		uRepo := &mockUserRepo{getDeletedUserByID: recoverAccountGetDeletedUserByID}
		req := authdomain.RecoverAccountRequest{Token: "tok"}

		Convey("When an admin deleted the account, it is not restored", func() {
			srv := newTestService(uRepo, nil, tRepo, nil, nil)
			srv.audit = mockAuditDeletedByAdmin(true, nil)

			err := srv.RecoverAccount(context.Background(), req)

			So(errors.Is(err, authdomain.ErrDeletedByAdmin), ShouldBeTrue)
		})

		Convey("When the audit lookup fails, the error is wrapped", func() {
			srv := newTestService(uRepo, nil, tRepo, nil, nil)
			srv.audit = mockAuditDeletedByAdmin(false, testutil.ErrDBUnexpected)

			err := srv.RecoverAccount(context.Background(), req)

			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "check deleted by admin")
		})
	})
}

func TestInitRecoverAccountProfileLookup(t *testing.T) {
	deletedUser := fakeRecoverAccountDeletedUser()

	Convey("Given an auth service", t, func() {
		Convey("When fetching the user profile fails", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return deletedUser, nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user profile")
		})

		Convey("When creating the recovery token fails", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return deletedUser, nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: TestUserID}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createAccountRecoveryToken: func(_ context.Context, _ *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, testutil.NewNoopOutbox(), nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create token")
		})
	})
}

func TestInitRecoverAccountTokenIssued(t *testing.T) {
	deletedUser := fakeRecoverAccountDeletedUser()

	Convey("Given an auth service", t, func() {
		Convey("When the token is issued successfully", func() {
			var captured []any
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return deletedUser, nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					aliceName := "Alice"
					return &authdomain.UserProfile{UserID: TestUserID, FirstName: &aliceName}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createAccountRecoveryToken: func(_ context.Context, t *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
					return t, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, testutil.NewCapturingOutbox(&captured), nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldBeNil)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "account")
			So(captured[1], ShouldEqual, TestUserID)
		})
	})
}

func TestRecoverAccountTokenLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token lookup fails", func() {
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken: func(_ context.Context, _ string) (*authdomain.AccountRecoveryToken, error) {
					return nil, authdomain.ErrInvalidToken
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the token has expired", func() {
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken: func(_ context.Context, _ string) (*authdomain.AccountRecoveryToken, error) {
					return &authdomain.AccountRecoveryToken{
						TokenBase: authdomain.TokenBase{UserID: TestUserID, ExpiresAt: time.Now().UTC().Add(-time.Hour)},
					}, nil
				},
			}
			srv := newTestService(nil, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(errors.Is(err, authdomain.ErrTokenExpired), ShouldBeTrue)
		})
	})
}

func validRecoverAccountToken(_ context.Context, _ string) (*authdomain.AccountRecoveryToken, error) {
	return &authdomain.AccountRecoveryToken{
		TokenBase: authdomain.TokenBase{UserID: TestUserID, ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, nil
}

func recoverAccountGetDeletedUserByID(_ context.Context, _ string) (*authdomain.User, error) {
	return &authdomain.User{ID: TestUserID, Status: authdomain.StatusDeleted}, nil
}

func TestRecoverAccountUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When fetching the deleted user fails", func() {
			tRepo := &mockTokenRepo{getAccountRecoveryToken: validRecoverAccountToken}
			uRepo := &mockUserRepo{
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get deleted user")
		})

		Convey("When the account is not actually deleted", func() {
			tRepo := &mockTokenRepo{getAccountRecoveryToken: validRecoverAccountToken}
			uRepo := &mockUserRepo{
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: TestUserID, Status: authdomain.StatusActive}, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(errors.Is(err, authdomain.ErrInvalidAccountState), ShouldBeTrue)
		})
	})
}

func TestRecoverAccountRestoreFailures(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When restoring the user fails", func() {
			tRepo := &mockTokenRepo{getAccountRecoveryToken: validRecoverAccountToken}
			uRepo := &mockUserRepo{
				getDeletedUserByID: recoverAccountGetDeletedUserByID,
				restoreUser:        testutil.AlwaysFailsDB,
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "restore user")
		})

		Convey("When marking the token as used fails", func() {
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken:      validRecoverAccountToken,
				markAccountRecoveryTokenUsed: testutil.AlwaysFailsDB,
			}
			uRepo := &mockUserRepo{
				getDeletedUserByID: recoverAccountGetDeletedUserByID,
				restoreUser:        testutil.AlwaysNil,
			}
			var unblocked []string
			srv := newTestService(uRepo, nil, tRepo, nil, mockBlocklistUnBlockUser(&unblocked, nil))

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mark token used")
			So(unblocked, ShouldBeEmpty)
		})

		Convey("When clearing the Redis block mark fails", func() {
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken:      validRecoverAccountToken,
				markAccountRecoveryTokenUsed: testutil.AlwaysNil,
			}
			uRepo := &mockUserRepo{
				getDeletedUserByID: recoverAccountGetDeletedUserByID,
				restoreUser:        testutil.AlwaysNil,
			}
			srv := newTestService(uRepo, nil, tRepo, nil, mockBlocklistUnBlockUser(nil, testutil.ErrRedisUnavailable))

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(errors.Is(err, testutil.ErrRedisUnavailable), ShouldBeTrue)
			So(errors.Is(err, authdomain.ErrBlocklistUnavailable), ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, "clear user_blocked")
		})
	})
}

func TestRecoverAccountSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token is valid and the account is restored", func() {
			var gotRestoredUserID string
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken:      validRecoverAccountToken,
				markAccountRecoveryTokenUsed: testutil.AlwaysNil,
			}
			uRepo := &mockUserRepo{
				getDeletedUserByID: recoverAccountGetDeletedUserByID,
				restoreUser: func(_ context.Context, userID string) error {
					gotRestoredUserID = userID
					return nil
				},
			}
			var unblocked []string
			srv := newTestService(uRepo, nil, tRepo, nil, mockBlocklistUnBlockUser(&unblocked, nil))

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldBeNil)
			So(gotRestoredUserID, ShouldEqual, TestUserID)
			So(unblocked, ShouldResemble, []string{TestUserID})
		})
	})
}
