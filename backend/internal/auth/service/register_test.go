package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func registerExistingUser() *authdomain.User {
	return &authdomain.User{ID: "user-123", Email: "user@example.com"}
}

func validRegisterRequest() authdomain.RegisterRequest {
	return authdomain.RegisterRequest{Email: "user@example.com", Password: "password123"}
}

func TestRegisterExistingEmailLookupFailures(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When checking for an existing user fails unexpectedly", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user by email")
		})

		Convey("When the email is already registered and the profile lookup fails", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return registerExistingUser(), nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get user profile")
		})
	})
}

func TestRegisterExistingEmailNotifyGuard(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the email is already registered and notifying the user fails", func() {
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return registerExistingUser(), nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					return &authdomain.UserProfile{UserID: "user-123"}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, newFailingOutbox(testutil.ErrDBUnexpected), nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "inform user")
		})

		Convey("When the email is already registered (email-enumeration guard)", func() {
			var captured []any
			uRepo := &mockUserRepo{
				getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
					return registerExistingUser(), nil
				},
				getUserProfileByUserID: func(_ context.Context, _ string) (*authdomain.UserProfile, error) {
					aliceName := "Alice"
					return &authdomain.UserProfile{UserID: "user-123", FirstName: &aliceName}, nil
				},
			}
			srv := newTestService(uRepo, nil, nil, newCapturingOutbox(&captured), nil)

			id, err := srv.Register(context.Background(), validRegisterRequest())

			So(errors.Is(err, authdomain.ErrUserAlreadyExists), ShouldBeTrue)
			So(id, ShouldBeEmpty)
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "user")
			So(captured[1], ShouldEqual, "user-123")
		})
	})
}

func newRegisterNewUserRepo() *mockUserRepo {
	return &mockUserRepo{
		getUserByEmail: func(_ context.Context, _ string) (*authdomain.User, error) {
			return nil, authdomain.ErrUserNotFound
		},
	}
}

func TestRegisterCreateUserFailures(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When creating the user fails", func() {
			uRepo := newRegisterNewUserRepo()
			uRepo.createUser = func(_ context.Context, _ *authdomain.User) (string, error) {
				return "", testutil.ErrDBUnexpected
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create user")
		})

		Convey("When creating the user profile fails", func() {
			uRepo := newRegisterNewUserRepo()
			uRepo.createUser = func(_ context.Context, _ *authdomain.User) (string, error) { return "user-123", nil }
			uRepo.createUserProfile = func(_ context.Context, _ *authdomain.UserProfile) error {
				return testutil.ErrDBUnexpected
			}
			srv := newTestService(uRepo, nil, nil, nil, nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create user profile")
		})
	})
}

func TestRegisterCreateVerificationTokenFails(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When creating the verification token fails", func() {
			uRepo := newRegisterNewUserRepo()
			uRepo.createUser = func(_ context.Context, _ *authdomain.User) (string, error) { return "user-123", nil }
			uRepo.createUserProfile = func(_ context.Context, _ *authdomain.UserProfile) error { return nil }
			tRepo := &mockTokenRepo{
				createEmailVerificationToken: func(_ context.Context, _ *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newNoopOutbox(), nil)

			_, err := srv.Register(context.Background(), validRegisterRequest())

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create verification token")
		})
	})
}

func TestRegisterSuccess(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When registration succeeds", func() {
			var capturedProfile *authdomain.UserProfile
			var captured []any
			uRepo := newRegisterNewUserRepo()
			uRepo.createUser = func(_ context.Context, _ *authdomain.User) (string, error) { return "user-123", nil }
			uRepo.createUserProfile = func(_ context.Context, p *authdomain.UserProfile) error {
				capturedProfile = p
				return nil
			}
			tRepo := &mockTokenRepo{
				createEmailVerificationToken: func(_ context.Context, t *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error) {
					return t, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newCapturingOutbox(&captured), nil)

			id, err := srv.Register(context.Background(), authdomain.RegisterRequest{
				Email: "user@example.com", Password: "password123", FirstName: "Alice",
			})

			So(err, ShouldBeNil)
			So(id, ShouldEqual, "user-123")
			So(capturedProfile.UserID, ShouldEqual, "user-123")
			So(*capturedProfile.FirstName, ShouldEqual, "Alice")
			So(captured, ShouldNotBeEmpty)
			So(captured[0], ShouldEqual, "user")
			So(captured[1], ShouldEqual, "user-123")
		})
	})
}

func TestRegisterMaxLengthPasswordHashing(t *testing.T) {
	Convey("Given an auth service", t, func() {
		Convey("When the password is exactly 72 bytes of multi-byte runes (Validate's own limit)", func() {
			uRepo := newRegisterNewUserRepo()
			uRepo.createUser = func(_ context.Context, _ *authdomain.User) (string, error) { return "user-123", nil }
			uRepo.createUserProfile = func(_ context.Context, _ *authdomain.UserProfile) error { return nil }
			tRepo := &mockTokenRepo{
				createEmailVerificationToken: func(_ context.Context, t *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error) {
					return t, nil
				},
			}
			srv := newTestService(uRepo, nil, tRepo, newNoopOutbox(), nil)

			// 36 Cyrillic runes * 2 bytes = 72 bytes, at bcrypt's own byte limit,
			// so GenerateFromPassword must not return bcrypt.ErrPasswordTooLong.
			password := strings.Repeat("я", 36)
			id, err := srv.Register(context.Background(), authdomain.RegisterRequest{
				Email: "user@example.com", Password: password,
			})

			So(err, ShouldBeNil)
			So(id, ShouldEqual, "user-123")
		})
	})
}

func TestNewUserProfileFromRegisterRequest(t *testing.T) {
	Convey("newUserProfileFromRegisterRequest", t, func() {
		Convey("When optional fields are empty", func() {
			req := authdomain.RegisterRequest{Email: "user@example.com", Password: "password123"}

			p := newUserProfileFromRegisterRequest("user-123", req)

			So(p.UserID, ShouldEqual, "user-123")
			So(p.FirstName, ShouldBeNil)
			So(p.LastName, ShouldBeNil)
			So(p.PhoneNumber, ShouldBeNil)
			So(p.Country, ShouldBeNil)
			So(p.City, ShouldBeNil)
			So(p.Gender, ShouldBeNil)
			So(p.DateOfBirth, ShouldBeNil)
			So(p.AvatarURL, ShouldBeNil)
			So(p.Timezone, ShouldBeNil)
			So(p.Bio, ShouldBeNil)
			So(p.UILanguage, ShouldEqual, defaultUILanguage)
		})

		Convey("When optional fields are provided", func() {
			dob := "1990-01-01"
			req := authdomain.RegisterRequest{
				Email:       "user@example.com",
				Password:    "password123",
				FirstName:   "Alice",
				LastName:    "Doe",
				PhoneNumber: "+48123456789",
				Country:     "PL",
				City:        "Warsaw",
				Gender:      "female",
				DateOfBirth: &dob,
				UILanguage:  "pl",
				AvatarURL:   "https://example.com/avatar.png",
				Timezone:    "Europe/Warsaw",
				Bio:         "hello",
			}

			p := newUserProfileFromRegisterRequest("user-456", req)

			So(p.UserID, ShouldEqual, "user-456")
			So(*p.FirstName, ShouldEqual, "Alice")
			So(*p.LastName, ShouldEqual, "Doe")
			So(*p.PhoneNumber, ShouldEqual, "+48123456789")
			So(*p.Country, ShouldEqual, "PL")
			So(*p.City, ShouldEqual, "Warsaw")
			So(*p.Gender, ShouldEqual, "female")
			So(p.DateOfBirth, ShouldEqual, &dob)
			So(*p.AvatarURL, ShouldEqual, "https://example.com/avatar.png")
			So(*p.Timezone, ShouldEqual, "Europe/Warsaw")
			So(*p.Bio, ShouldEqual, "hello")
			So(p.UILanguage, ShouldEqual, "pl")
		})
	})
}
