//go:build integration

package usersrepository

import (
	"context"
	"errors"
	"testing"

	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

const insertTestProfileSQL = `
	INSERT INTO user_profiles (
		user_id, first_name, last_name, phone_number, country, city,
		date_of_birth, gender, ui_language, avatar_url, timezone, bio
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

func insertTestUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "users-repo-integration"))
}

func softDeleteTestUser(t *testing.T, ctx context.Context, tx pgx.Tx, userID string) {
	t.Helper()

	_, err := tx.Exec(ctx, `UPDATE users SET deleted_at = now() WHERE id = $1`, userID)
	if err != nil {
		t.Fatalf("softDeleteTestUser: %v", err)
	}
}

func insertTestProfile(t *testing.T, ctx context.Context, tx pgx.Tx, p *usersdomain.UserProfile) {
	t.Helper()

	_, err := tx.Exec(ctx, insertTestProfileSQL,
		p.UserID, p.FirstName, p.LastName, p.PhoneNumber, p.Country, p.City,
		p.DateOfBirth, p.Gender, p.UILanguage, p.AvatarURL, p.Timezone, p.Bio,
	)
	if err != nil {
		t.Fatalf("insertTestProfile: %v", err)
	}
}

func fullTestProfile(userID string) *usersdomain.UserProfile {
	firstName, lastName, phoneNumber := "Anna", "Kowalska", "+48501234567"
	country, city, gender := "PL", "Warsaw", "female"
	avatarURL, timezone, bio := "https://cdn.example.com/avatar.png", "Europe/Warsaw", "Integration test fixture profile"
	dob := "1990-05-15"
	return &usersdomain.UserProfile{
		UserID:      userID,
		FirstName:   &firstName,
		LastName:    &lastName,
		PhoneNumber: &phoneNumber,
		Country:     &country,
		City:        &city,
		DateOfBirth: &dob,
		Gender:      &gender,
		UILanguage:  "pl",
		AvatarURL:   &avatarURL,
		Timezone:    &timezone,
		Bio:         &bio,
	}
}

func TestGetUserProfileByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a users repository backed by real Postgres", t, func() {
		Convey("When the profile has all fields populated", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				seed := fullTestProfile(userID)
				insertTestProfile(t, ctx, tx, seed)

				got, err := repo.GetUserProfileByID(ctx, userID)

				So(err, ShouldBeNil)
				So(got.UserID, ShouldEqual, userID)
				So(*got.FirstName, ShouldEqual, *seed.FirstName)
				So(*got.LastName, ShouldEqual, *seed.LastName)
				So(*got.PhoneNumber, ShouldEqual, *seed.PhoneNumber)
				So(*got.Country, ShouldEqual, *seed.Country)
				So(*got.City, ShouldEqual, *seed.City)
				So(got.DateOfBirth, ShouldNotBeNil)
				So(*got.DateOfBirth, ShouldEqual, *seed.DateOfBirth)
				So(*got.Gender, ShouldEqual, *seed.Gender)
				So(got.UILanguage, ShouldEqual, seed.UILanguage)
				So(*got.AvatarURL, ShouldEqual, *seed.AvatarURL)
				So(*got.Timezone, ShouldEqual, *seed.Timezone)
				So(*got.Bio, ShouldEqual, *seed.Bio)
				So(got.CreatedAt.IsZero(), ShouldBeFalse)
				So(got.UpdatedAt.IsZero(), ShouldBeFalse)
			})
		})

		Convey("When date_of_birth and other optional fields are NULL", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				seed := &usersdomain.UserProfile{
					UserID:     userID,
					UILanguage: "uk",
				}
				insertTestProfile(t, ctx, tx, seed)

				got, err := repo.GetUserProfileByID(ctx, userID)

				So(err, ShouldBeNil)
				So(got.DateOfBirth, ShouldBeNil)
				So(got.FirstName, ShouldBeNil)
				So(got.Country, ShouldBeNil)
				So(got.Gender, ShouldBeNil)
			})
		})

		Convey("When no profile exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				_, err := repo.GetUserProfileByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
			})
		})

		Convey("When the owning user is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				seed := fullTestProfile(userID)
				insertTestProfile(t, ctx, tx, seed)
				softDeleteTestUser(t, ctx, tx, userID)

				got, err := repo.GetUserProfileByID(ctx, userID)

				So(err, ShouldBeNil)
				So(got.UserID, ShouldEqual, userID)
			})
		})
	})
}

func TestUpdateUserProfile_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a users repository backed by real Postgres", t, func() {
		Convey("When updating an existing profile's fields", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

				update := fullTestProfile(userID)
				err := repo.UpdateUserProfile(ctx, update)
				So(err, ShouldBeNil)

				got, err := repo.GetUserProfileByID(ctx, userID)
				So(err, ShouldBeNil)
				So(*got.FirstName, ShouldEqual, *update.FirstName)
				So(*got.LastName, ShouldEqual, *update.LastName)
				So(*got.PhoneNumber, ShouldEqual, *update.PhoneNumber)
				So(*got.Country, ShouldEqual, *update.Country)
				So(*got.DateOfBirth, ShouldEqual, *update.DateOfBirth)
				So(*got.Gender, ShouldEqual, *update.Gender)
				So(*got.AvatarURL, ShouldEqual, *update.AvatarURL)
			})
		})

		Convey("When nulling out a previously-set date_of_birth", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				insertTestProfile(t, ctx, tx, fullTestProfile(userID))

				update := fullTestProfile(userID)
				update.DateOfBirth = nil
				So(repo.UpdateUserProfile(ctx, update), ShouldBeNil)

				got, err := repo.GetUserProfileByID(ctx, userID)
				So(err, ShouldBeNil)
				So(got.DateOfBirth, ShouldBeNil)
			})
		})

		Convey("When no profile row matches the user ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				update := fullTestProfile("00000000-0000-0000-0000-000000000000")
				err := repo.UpdateUserProfile(ctx, update)

				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeTrue)
			})
		})

		Convey("When the update violates a CHECK constraint at the DB level", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, tx)
				insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

				badGender := "not_a_valid_gender"
				update := fullTestProfile(userID)
				update.Gender = &badGender

				err := repo.UpdateUserProfile(ctx, update)

				So(err, ShouldNotBeNil)
				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeFalse)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514")
			})

			Convey("When the update violates the country CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					badCountry := "USA" // must be exactly 2 chars (ISO 3166-1 alpha-2)
					update := fullTestProfile(userID)
					update.Country = &badCountry

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514")
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_country_check")
				})
			})

			Convey("When the update violates the date_of_birth not-future CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					futureDOB := "2999-01-01"
					update := fullTestProfile(userID)
					update.DateOfBirth = &futureDOB

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514")
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_dob_not_future")
				})
			})

			Convey("When the update violates the date_of_birth min CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					tooOldDOB := "1899-12-31"
					update := fullTestProfile(userID)
					update.DateOfBirth = &tooOldDOB

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514")
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_dob_min")
				})
			})
		})
	})
}

func TestUserProfileForeignKeyRestrict_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a user with an existing profile", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			userID := insertTestUser(t, tx)
			insertTestProfile(t, ctx, tx, fullTestProfile(userID))

			Convey("When hard-deleting the owning user row", func() {
				_, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503")
			})
		})
	})
}
