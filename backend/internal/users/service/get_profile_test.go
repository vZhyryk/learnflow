package usersservice

import (
	"context"
	"errors"
	"testing"
	"time"

	usersdomain "learnflow_backend/internal/users/domain"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetUserProfile(t *testing.T) {
	now := time.Now()

	Convey("Given a users service", t, func() {
		var repoResult *usersdomain.UserProfile
		var repoErr error
		svc := New(&mockUserProfileRepo{
			getUserProfileByID: func(_ context.Context, _ string) (*usersdomain.UserProfile, error) {
				return repoResult, repoErr
			},
		})

		Convey("When the profile exists", func() {
			repoResult = &usersdomain.UserProfile{
				UserID:    "user-123",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: now,
				UpdatedAt: now,
			}
			got, err := svc.GetUserProfile(context.Background(), "user-123")

			Convey("Then it returns the profile without error", func() {
				So(err, ShouldBeNil)
				So(got.UserID, ShouldEqual, "user-123")
				So(got.FirstName, ShouldEqual, "John")
				So(got.LastName, ShouldEqual, "Doe")
			})
		})

		Convey("When the profile is not found", func() {
			repoErr = usersdomain.ErrUserNotFound
			_, err := svc.GetUserProfile(context.Background(), "unknown-id")
			Convey("Then it returns ErrUserNotFound", func() {
				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
			})
		})

		Convey("When the repository returns an unexpected error", func() {
			repoErr = errors.New("db connection lost")
			_, err := svc.GetUserProfile(context.Background(), "user-123")
			Convey("Then it wraps and propagates the error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "db connection lost")
			})
		})
	})
}
