package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	. "github.com/smartystreets/goconvey/convey"
)

func newLoginTestUser(rawPassword string, status authdomain.UserStatus) *authdomain.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 4)
	if err != nil {
		panic(err)
	}
	return &authdomain.User{ID: "user-123", Role: authdomain.RoleUser, PasswordHash: string(hash), Status: status}
}

var validLoginReq = authdomain.LoginRequest{Email: "user@example.com", Password: "correct-password", UserAgent: "test-agent", IPAddress: "127.0.0.1"}

func TestLoginUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user does not exist", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})

		Convey("When the user lookup fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})
	})
}

func TestLoginAccountLocked(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the account is temporarily locked", func() {
			lockedUntil := time.Now().UTC().Add(10 * time.Minute)
			user := newLoginTestUser(validLoginReq.Password, authdomain.StatusActive)
			user.LoginLockedUntil = &lockedUntil
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			var lockErr *authdomain.ErrAccountLockedError
			So(errors.As(err, &lockErr), ShouldBeTrue)
			So(lockErr.LockedUntil, ShouldEqual, lockedUntil)
		})

		Convey("When the lock has already expired", func() {
			lockedUntil := time.Now().UTC().Add(-10 * time.Minute)
			user := newLoginTestUser(validLoginReq.Password, authdomain.StatusActive)
			user.LoginLockedUntil = &lockedUntil
			uRepo := &mockUserRepo{
				getUserByEmail:    func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
				resetFailedLogin:  func(_ context.Context, _ string) error { return nil },
				updateLastLoginAt: func(_ context.Context, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) { return s, nil },
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldBeNil)
		})
	})
}

func TestLoginWrongPassword(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the password is wrong and the failed-attempt counter increments", func() {
			var gotUserID string
			user := newLoginTestUser("actual-password", authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
				incrementFailedLogin: func(_ context.Context, userID, _ string, _ int) error {
					gotUserID = userID
					return nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
			So(gotUserID, ShouldEqual, "user-123")
		})

		Convey("When incrementing the failed-attempt counter fails unexpectedly", func() {
			user := newLoginTestUser("actual-password", authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
				incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
					return errors.New("db connection lost")
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "increment failed login")
		})

		Convey("When incrementing the failed-attempt counter races with user deletion", func() {
			user := newLoginTestUser("actual-password", authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil },
				incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
					return authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func TestLoginAccountStatus(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the account is blocked", func() {
			user := newLoginTestUser(validLoginReq.Password, authdomain.StatusBlocked)
			uRepo := &mockUserRepo{getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil }}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrAccountBlocked), ShouldBeTrue)
		})

		Convey("When the email is not yet verified", func() {
			user := newLoginTestUser(validLoginReq.Password, authdomain.StatusPendingVerification)
			uRepo := &mockUserRepo{getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil }}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrEmailNotVerified), ShouldBeTrue)
		})

		Convey("When the account is deleted", func() {
			user := newLoginTestUser(validLoginReq.Password, authdomain.StatusDeleted)
			uRepo := &mockUserRepo{getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil }}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func newActiveLoginUserRepo() *mockUserRepo {
	return &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
			return newLoginTestUser(validLoginReq.Password, authdomain.StatusActive), nil
		},
	}
}

func TestLoginCreateSessionFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When creating the session fails", func() {
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, _ *authdomain.UserSession) (*authdomain.UserSession, error) {
					return nil, errors.New("db connection lost")
				},
			}
			srv := newTestService(newActiveLoginUserRepo(), sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create session")
		})
	})
}

func TestLoginPostSessionUpdateFailures(t *testing.T) {
	activeUserRepo := newActiveLoginUserRepo

	Convey("Given an auth service", t, func() {
		Convey("When resetting the failed-login counter fails", func() {
			uRepo := activeUserRepo()
			uRepo.resetFailedLogin = func(_ context.Context, _ string) error { return errors.New("db connection lost") }
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) { return s, nil },
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "reset failed login")
		})

		Convey("When updating last-login-at fails", func() {
			uRepo := activeUserRepo()
			uRepo.resetFailedLogin = func(_ context.Context, _ string) error { return nil }
			uRepo.updateLastLoginAt = func(_ context.Context, _ string) error { return errors.New("db connection lost") }
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) { return s, nil },
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update last login")
		})
	})
}

func TestLoginSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the credentials are valid", func() {
			expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
			var gotSessionInput *authdomain.UserSession
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return newLoginTestUser(validLoginReq.Password, authdomain.StatusActive), nil
				},
				resetFailedLogin:  func(_ context.Context, _ string) error { return nil },
				updateLastLoginAt: func(_ context.Context, _ string) error { return nil },
			}
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) {
					gotSessionInput = s
					s.ExpiresAt = expiresAt
					return s, nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			got, err := srv.Login(context.Background(), validLoginReq)

			So(err, ShouldBeNil)
			So(got.AccessToken, ShouldNotBeEmpty)
			So(got.RefreshToken, ShouldNotBeEmpty)
			So(got.UserID, ShouldEqual, "user-123")
			So(got.ExpiresAt, ShouldEqual, expiresAt)
			So(gotSessionInput.UserID, ShouldEqual, "user-123")
			So(*gotSessionInput.UserAgent, ShouldEqual, validLoginReq.UserAgent)
			So(*gotSessionInput.IPAddress, ShouldEqual, validLoginReq.IPAddress)
		})
	})
}
