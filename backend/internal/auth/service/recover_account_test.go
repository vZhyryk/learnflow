package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitRecoverAccountUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user lookup fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, errors.New("db connection lost")
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
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusActive}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(errors.Is(err, authdomain.ErrInvalidAccountState), ShouldBeTrue)
		})
	})
}

func TestInitRecoverAccountProfileLookup(t *testing.T) {
	deletedUser := &authdomain.User{ID: "user-123", Email: "user@example.com", Status: authdomain.StatusDeleted}

	Convey("Given an auth service", t, func() {
		Convey("When fetching the user profile fails", func() {
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return deletedUser, nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, errors.New("db connection lost")
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
					return &authdomain.UserProfile{UserID: "user-123"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createAccountRecoveryToken: func(_ context.Context, _ *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newNoopOutbox(), nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create token")
		})
	})
}

func TestInitRecoverAccountTokenIssued(t *testing.T) {
	deletedUser := &authdomain.User{ID: "user-123", Email: "user@example.com", Status: authdomain.StatusDeleted}

	Convey("Given an auth service", t, func() {
		Convey("When the token is issued successfully", func() {
			var captured []any
			uRepo := &mockUserRepo{
				getDeletedUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return deletedUser, nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123", FirstName: "Alice"}, nil
				},
			}
			tRepo := &mockTokenRepo{
				createAccountRecoveryToken: func(_ context.Context, t *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
					return t, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newCapturingOutbox(&captured), nil)

			err := srv.InitRecoverAccount(context.Background(), authdomain.RequestRecoverAccountRequest{Email: "user@example.com"})

			So(err, ShouldBeNil)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "account")
			So(captured[1], ShouldEqual, "user-123")
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
						TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(-time.Hour)},
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
		TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, nil
}

func TestRecoverAccountUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When fetching the deleted user fails", func() {
			tRepo := &mockTokenRepo{getAccountRecoveryToken: validRecoverAccountToken}
			uRepo := &mockUserRepo{
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, errors.New("db connection lost")
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
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusActive}, nil
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
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusDeleted}, nil
				},
				restoreUser: func(_ context.Context, _ string) error {
					return errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "restore user")
		})

		Convey("When marking the token as used fails", func() {
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken: validRecoverAccountToken,
				markAccountRecoveryTokenUsed: func(_ context.Context, _ string) error {
					return errors.New("db connection lost")
				},
			}
			uRepo := &mockUserRepo{
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusDeleted}, nil
				},
				restoreUser: func(_ context.Context, _ string) error { return nil },
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "mark token used")
		})
	})
}

func TestRecoverAccountSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the token is valid and the account is restored", func() {
			var gotRestoredUserID string
			tRepo := &mockTokenRepo{
				getAccountRecoveryToken: func(_ context.Context, _ string) (*authdomain.AccountRecoveryToken, error) {
					return &authdomain.AccountRecoveryToken{
						TokenBase: authdomain.TokenBase{UserID: "user-123", ExpiresAt: time.Now().UTC().Add(time.Hour)},
					}, nil
				},
				markAccountRecoveryTokenUsed: func(_ context.Context, _ string) error { return nil },
			}
			uRepo := &mockUserRepo{
				getDeletedUserByID: func(_ context.Context, _ string) (*authdomain.User, error) {
					return &authdomain.User{ID: "user-123", Status: authdomain.StatusDeleted}, nil
				},
				restoreUser: func(_ context.Context, userID string) error {
					gotRestoredUserID = userID
					return nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, nil, nil)

			err := srv.RecoverAccount(context.Background(), authdomain.RecoverAccountRequest{Token: "tok"})

			So(err, ShouldBeNil)
			So(gotRestoredUserID, ShouldEqual, "user-123")
		})
	})
}
