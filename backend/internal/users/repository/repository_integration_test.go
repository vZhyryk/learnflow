//go:build integration

package usersrepository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

// dummyPasswordHash is a well-formed 60-char bcrypt hash — the users.password_hash
// column is varchar(60) NOT NULL, so fixtures must satisfy the column width even
// though this module never reads or verifies the hash itself.
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

const insertTestUserSQL = `
	INSERT INTO users (email, password_hash, role, status)
	VALUES ($1, $2, 'user', 'active')
	RETURNING id`

const insertTestProfileSQL = `
	INSERT INTO user_profiles (
		user_id, first_name, last_name, phone_number, country, city,
		date_of_birth, gender, ui_language, avatar_url, timezone, bio
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

// randomTestEmail avoids colliding with the unique-active-email index across
// concurrent test runs sharing the same database.
func randomTestEmail(t *testing.T) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("randomTestEmail: %v", err)
	}
	return fmt.Sprintf("users-repo-integration-%s@example.com", hex.EncodeToString(buf))
}

// insertTestUser creates a row in `users` so a `user_profiles` row can satisfy its
// FK (ON DELETE RESTRICT) — the users module owns no user-creation SQL itself
// (that lives in internal/auth/repository), so fixtures replicate the shape
// directly rather than importing across bounded contexts.
func insertTestUser(t *testing.T, ctx context.Context, tx pgx.Tx) string {
	t.Helper()

	var id string
	err := tx.QueryRow(ctx, insertTestUserSQL, randomTestEmail(t), dummyPasswordHash).Scan(&id)
	if err != nil {
		t.Fatalf("insertTestUser: %v", err)
	}
	return id
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
				userID := insertTestUser(t, ctx, tx)
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
				userID := insertTestUser(t, ctx, tx)
				seed := &usersdomain.UserProfile{
					UserID:     userID,
					UILanguage: "uk", // NOT NULL DEFAULT 'uk' at the column level, but repository always inserts explicitly
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

		// user_profiles has no deleted_at of its own and getProfileByUserIDSQL never
		// joins against users.deleted_at — this is intentional (a soft-deleted user's
		// profile row stays readable/writable since ON DELETE RESTRICT keeps it alive),
		// but it's exactly the kind of cross-table invariant a mocked QueryRunner can't
		// verify, so it's pinned down here.
		Convey("When the owning user is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, ctx, tx)
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
				userID := insertTestUser(t, ctx, tx)
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
				userID := insertTestUser(t, ctx, tx)
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

		// The service layer validates gender via ChangeUserProfileRequest.Validate()
		// before calling the repository, but the repository itself performs no
		// validation — the DB CHECK constraint (user_profiles_gender_check) is the
		// last line of defense against bad data reaching the table directly. A mock
		// QueryRunner would happily accept any string here.
		Convey("When the update violates a CHECK constraint at the DB level", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userID := insertTestUser(t, ctx, tx)
				insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

				badGender := "not_a_valid_gender"
				update := fullTestProfile(userID)
				update.Gender = &badGender

				err := repo.UpdateUserProfile(ctx, update)

				So(err, ShouldNotBeNil)
				So(errors.Is(err, usersdomain.ErrUserNotFound), ShouldBeFalse)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
			})

			Convey("When the update violates the country CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, ctx, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					badCountry := "USA" // must be exactly 2 chars (ISO 3166-1 alpha-2)
					update := fullTestProfile(userID)
					update.Country = &badCountry

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514") // check_violation
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_country_check")
				})
			})

			Convey("When the update violates the date_of_birth not-future CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, ctx, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					futureDOB := "2999-01-01"
					update := fullTestProfile(userID)
					update.DateOfBirth = &futureDOB

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514") // check_violation
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_dob_not_future")
				})
			})

			Convey("When the update violates the date_of_birth min CHECK constraint", func() {
				testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
					repo := &Repository{db: tx}
					userID := insertTestUser(t, ctx, tx)
					insertTestProfile(t, ctx, tx, &usersdomain.UserProfile{UserID: userID, UILanguage: "uk"})

					tooOldDOB := "1899-12-31"
					update := fullTestProfile(userID)
					update.DateOfBirth = &tooOldDOB

					err := repo.UpdateUserProfile(ctx, update)

					So(err, ShouldNotBeNil)
					var pgErr *pgconn.PgError
					So(errors.As(err, &pgErr), ShouldBeTrue)
					So(pgErr.Code, ShouldEqual, "23514") // check_violation
					So(pgErr.ConstraintName, ShouldEqual, "user_profiles_dob_min")
				})
			})
		})
	})
}

// TestUserProfileForeignKeyRestrict_Integration pins down that user_profiles'
// FK to users is ON DELETE RESTRICT — a hard delete of a user with an existing
// profile must fail at the DB level. This is infrastructure the users module
// relies on (soft delete is the only supported deletion path) but doesn't
// enforce itself in Go, so a mocked QueryRunner can't catch a regression here.
func TestUserProfileForeignKeyRestrict_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a user with an existing profile", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			userID := insertTestUser(t, ctx, tx)
			insertTestProfile(t, ctx, tx, fullTestProfile(userID))

			Convey("When hard-deleting the owning user row", func() {
				_, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})
	})
}
