package usersservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"

	. "github.com/smartystreets/goconvey/convey"
)

func TestChangeUserProfile(t *testing.T) {
	now := time.Now()

	Convey("Given a users service", t, func() {
		var getResult *usersdomain.UserProfile
		var getErr error
		var updateErr error
		var capturedProfile *usersdomain.UserProfile

		svc := New(&mockUserProfileRepo{
			getUserProfileByID: func(_ context.Context, _ string) (*usersdomain.UserProfile, error) {
				return getResult, getErr
			},
			updateUserProfile: func(_ context.Context, p *usersdomain.UserProfile) error {
				capturedProfile = p
				return updateErr
			},
		})

		Convey("When valid partial fields are provided", func() {
			getResult = &usersdomain.UserProfile{
				UserID:    "user-123",
				FirstName: "John",
				LastName:  "Doe",
				CreatedAt: now,
				UpdatedAt: now,
			}
			userID, firstName := "user-123", "Jane"
			err := svc.ChangeUserProfile(context.Background(), usersdomain.ChangeUserProfileRequest{
				UserID:    &userID,
				FirstName: &firstName,
			})

			Convey("Then it applies only non-nil fields and saves", func() {
				So(err, ShouldBeNil)
				So(capturedProfile.FirstName, ShouldEqual, "Jane")
				So(capturedProfile.LastName, ShouldEqual, "Doe")
				So(capturedProfile.UserID, ShouldEqual, "user-123")
			})
		})

		Convey("When the profile is not found", func() {
			getErr = usersdomain.ErrUserNotFound
			userID := "unknown"
			err := svc.ChangeUserProfile(context.Background(), usersdomain.ChangeUserProfileRequest{
				UserID: &userID,
			})

			Convey("Then it propagates ErrUserNotFound", func() {
				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
			})
		})

		Convey("When the repository update fails", func() {
			getResult = &usersdomain.UserProfile{UserID: "user-123"}
			updateErr = testutil.ErrDBTimeout
			userID := "user-123"
			err := svc.ChangeUserProfile(context.Background(), usersdomain.ChangeUserProfileRequest{
				UserID: &userID,
			})

			Convey("Then it propagates the update error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "db timeout")
			})
		})
	})
}
