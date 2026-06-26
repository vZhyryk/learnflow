package usersrepository

import (
	usersdomain "learnflow_backend/internal/users/domain"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUserProfile(row rowScanner) (*usersdomain.UserProfile, error) {
	user := &usersdomain.UserProfile{}
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.Country,
		&user.City,
		&user.DateOfBirth,
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
	return user, nil
}
