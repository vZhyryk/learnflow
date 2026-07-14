package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
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

func validLoginReq() authdomain.LoginRequest {
	return authdomain.LoginRequest{Email: "user@example.com", Password: "correct-password", UserAgent: "test-agent", IPAddress: "127.0.0.1"}
}

func newActiveLoginUserRepo() *mockUserRepo {
	return &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
			return newLoginTestUser(validLoginReq().Password, authdomain.StatusActive), nil
		},
	}
}

func validLoginSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) { return s, nil },
	}
}

// loginGetUserByEmail returns a getUserByEmail closure that always resolves to user,
// regardless of the requested email.
func loginGetUserByEmail(user *authdomain.User) func(context.Context, string) (*authdomain.User, error) {
	return func(_ context.Context, _ string) (*authdomain.User, error) { return user, nil }
}

func TestLoginUserLookup(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the user does not exist", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})

		Convey("When the user lookup fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user")
		})
	})
}

// TestLoginConstantTimeUserEnumeration is a regression test for the timing-attack
// mitigation documented on Login/loginGetUser: when the email does not exist, a dummy
// bcrypt comparison must still run so the response time matches the "wrong password"
// path. It spies on the package-level bcryptCompareHashAndPassword hook to assert the
// dummy comparison actually executes, rather than asserting on wall-clock timing (flaky).
func TestLoginConstantTimeUserEnumeration(t *testing.T) {
	Convey("Given an auth service", t, func() {
		original := bcryptCompareHashAndPassword
		Reset(func() { bcryptCompareHashAndPassword = original })

		var calls int
		var lastHash []byte
		bcryptCompareHashAndPassword = func(hashedPassword, password []byte) error {
			calls++
			lastHash = hashedPassword
			return original(hashedPassword, password)
		}

		Convey("When the email does not exist, the dummy bcrypt comparison still runs", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
			So(calls, ShouldEqual, 1)
			So(lastHash, ShouldResemble, srv.dummyPasswordHash)
		})

		Convey("When the email exists, the real password hash is compared (not the dummy one)", func() {
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail:       loginGetUserByEmail(user),
				resetFailedLogin:     func(context.Context, string) error { return nil },
				updateLastLoginAt:    func(context.Context, string) error { return nil },
				incrementFailedLogin: func(context.Context, string, string, int) error { return nil },
			}
			srv := newTestService(uRepo, validLoginSessionRepo(), nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(err, ShouldBeNil)
			So(calls, ShouldEqual, 1)
			So(lastHash, ShouldResemble, []byte(user.PasswordHash))
			So(lastHash, ShouldNotResemble, srv.dummyPasswordHash)
		})
	})
}

func TestLoginAccountLocked(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the account is temporarily locked", func() {
			lockedUntil := time.Now().UTC().Add(10 * time.Minute)
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusActive)
			user.LoginLockedUntil = &lockedUntil
			uRepo := &mockUserRepo{
				getUserByEmail: loginGetUserByEmail(user),
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			var lockErr *authdomain.ErrAccountLockedError
			So(errors.As(err, &lockErr), ShouldBeTrue)
			So(lockErr.LockedUntil, ShouldEqual, lockedUntil)
		})

		Convey("When the lock has already expired", func() {
			lockedUntil := time.Now().UTC().Add(-10 * time.Minute)
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusActive)
			user.LoginLockedUntil = &lockedUntil
			uRepo := &mockUserRepo{
				getUserByEmail:    loginGetUserByEmail(user),
				resetFailedLogin:  testutil.AlwaysNil,
				updateLastLoginAt: testutil.AlwaysNil,
			}
			sRepo := validLoginSessionRepo()
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

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
				getUserByEmail: loginGetUserByEmail(user),
				incrementFailedLogin: func(_ context.Context, userID, _ string, _ int) error {
					gotUserID = userID
					return nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
			So(gotUserID, ShouldEqual, "user-123")
		})

		Convey("When incrementing the failed-attempt counter fails unexpectedly", func() {
			user := newLoginTestUser("actual-password", authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail: loginGetUserByEmail(user),
				incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
					return testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "increment failed login")
		})

		Convey("When incrementing the failed-attempt counter races with user deletion", func() {
			user := newLoginTestUser("actual-password", authdomain.StatusActive)
			uRepo := &mockUserRepo{
				getUserByEmail: loginGetUserByEmail(user),
				incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
					return authdomain.ErrUserNotFound
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})
	})
}

func TestLoginAccountStatus(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the account is blocked", func() {
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusBlocked)
			uRepo := &mockUserRepo{getUserByEmail: loginGetUserByEmail(user)}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrAccountBlocked), ShouldBeTrue)
		})

		Convey("When the email is not yet verified", func() {
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusPendingVerification)
			uRepo := &mockUserRepo{getUserByEmail: loginGetUserByEmail(user)}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrEmailNotVerified), ShouldBeTrue)
		})

		Convey("When the account is deleted", func() {
			user := newLoginTestUser(validLoginReq().Password, authdomain.StatusDeleted)
			uRepo := &mockUserRepo{getUserByEmail: loginGetUserByEmail(user)}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
		})

		// Regression test for the security invariant documented on Login: status checks run
		// after bcrypt, so a blocked account with a WRONG password must still fail as
		// ErrInvalidCredentials, not ErrAccountBlocked — reordering the checks in login.go
		// would turn account status into a timing/response oracle.
		Convey("When the account is blocked and the password is also wrong, bcrypt failure wins over status", func() {
			user := newLoginTestUser("actual-password", authdomain.StatusBlocked)
			var incremented bool
			uRepo := &mockUserRepo{
				getUserByEmail: loginGetUserByEmail(user),
				incrementFailedLogin: func(_ context.Context, _, _ string, _ int) error {
					incremented = true
					return nil
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(errors.Is(err, authdomain.ErrInvalidCredentials), ShouldBeTrue)
			So(errors.Is(err, authdomain.ErrAccountBlocked), ShouldBeFalse)
			So(incremented, ShouldBeTrue)
		})
	})
}

func TestLoginCreateSessionFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When creating the session fails", func() {
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, _ *authdomain.UserSession) (*authdomain.UserSession, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(newActiveLoginUserRepo(), sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

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
			uRepo.resetFailedLogin = testutil.AlwaysFailsDB
			sRepo := validLoginSessionRepo()
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "reset failed login")
		})

		Convey("When updating last-login-at fails", func() {
			uRepo := activeUserRepo()
			uRepo.resetFailedLogin = testutil.AlwaysNil
			uRepo.updateLastLoginAt = testutil.AlwaysFailsDB
			sRepo := validLoginSessionRepo()
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			_, err := srv.Login(context.Background(), validLoginReq())

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
			uRepo := newActiveLoginUserRepo()
			uRepo.resetFailedLogin = testutil.AlwaysNil
			uRepo.updateLastLoginAt = testutil.AlwaysNil
			sRepo := &mockSessionRepo{
				createUserSession: func(_ context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) {
					gotSessionInput = s
					s.ExpiresAt = expiresAt
					return s, nil
				},
			}
			srv := newTestService(uRepo, sRepo, nil, nil, nil)

			got, err := srv.Login(context.Background(), validLoginReq())

			So(err, ShouldBeNil)
			So(got.AccessToken, ShouldNotBeEmpty)
			So(got.RefreshToken, ShouldNotBeEmpty)
			So(got.UserID, ShouldEqual, "user-123")
			So(got.ExpiresAt, ShouldEqual, expiresAt)
			So(gotSessionInput.UserID, ShouldEqual, "user-123")
			So(*gotSessionInput.UserAgent, ShouldEqual, validLoginReq().UserAgent)
			So(*gotSessionInput.IPAddress, ShouldEqual, validLoginReq().IPAddress)
		})
	})
}
