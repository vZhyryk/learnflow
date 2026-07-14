package usersrepository

import (
	"learnflow_backend/internal/infrastructure/convert"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

type rowScanner interface {
	Scan(dest ...any) error
}

// dobLayout matches usersdomain's DateOfBirth string format and the
// validator.IsValidDateOfBirth parse layout.
const dobLayout = "2006-01-02"

func scanUserProfile(row rowScanner) (*usersdomain.UserProfile, error) {
	user := &usersdomain.UserProfile{}
	var dob pgtype.Date
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.Country,
		&user.City,
		&dob,
		&user.Gender,
		&user.UILanguage,
		&user.AvatarURL,
		&user.Timezone,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.DateOfBirth = convert.FormatNullableDate(dob, dobLayout)
	return user, nil
}
