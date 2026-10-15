package usersrepository

import (
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{repository.BaseRepository{DB: runner}}
}

// fakeProfile/fakeScanProfile mirror internal/auth/repository/mock_test.go's fixture
// of the same name. Not shared via testutil on purpose: usersdomain.UserProfile and
// authdomain.UserProfile are separate types by design (auth and users are independent
// bounded contexts per Clean Architecture layering), so a generic helper here would
// either need generics (forbidden in domain-adjacent code) or reflection.
func fakeProfile(now time.Time) *usersdomain.UserProfile {
	firstName, lastName, phoneNumber := "John", "Doe", "+380991234567"
	country, city, gender := "UA", "Kyiv", "male"
	timezone, bio := "Europe/Kiev", "bio text"
	avatarURL := ""
	return &usersdomain.UserProfile{
		UserID:      testUserID,
		FirstName:   &firstName,
		LastName:    &lastName,
		PhoneNumber: &phoneNumber,
		Country:     &country,
		City:        &city,
		DateOfBirth: nil,
		Gender:      &gender,
		UILanguage:  "uk",
		AvatarURL:   &avatarURL,
		Timezone:    &timezone,
		Bio:         &bio,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func fakeScanProfile(now time.Time) func(dest ...any) error {
	p := fakeProfile(now)
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = p.UserID
		*testutil.CastPtrStr(dest[1], 1) = p.FirstName
		*testutil.CastPtrStr(dest[2], 2) = p.LastName
		*testutil.CastPtrStr(dest[3], 3) = p.PhoneNumber
		*testutil.CastPtrStr(dest[4], 4) = p.Country
		*testutil.CastPtrStr(dest[5], 5) = p.City
		*testutil.CastPgtypeDate(dest[6], 6) = pgtype.Date{}
		*testutil.CastPtrStr(dest[7], 7) = p.Gender
		*testutil.CastStr(dest[8], 8) = p.UILanguage
		*testutil.CastPtrStr(dest[9], 9) = p.AvatarURL
		*testutil.CastPtrStr(dest[10], 10) = p.Timezone
		*testutil.CastPtrStr(dest[11], 11) = p.Bio
		*testutil.CastTime(dest[12], 12) = p.CreatedAt
		*testutil.CastTime(dest[13], 13) = p.UpdatedAt
		return nil
	}
}
