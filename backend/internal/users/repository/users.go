package usersrepository

import (
	"context"
	"errors"
	"fmt"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/jackc/pgx/v5"
)

// GetUserProfileByID fetches the profile for the given user ID.
func (rep *Repository) GetUserProfileByID(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	user, err := scanUserProfile(rep.queryRunner(ctx).QueryRow(ctx, getProfileByUserIDSQL, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usersdomain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserProfileByID: %w", err)
	}

	return user, nil
}

// UpdateUserProfile persists profile changes for an existing user.
func (rep *Repository) UpdateUserProfile(ctx context.Context, userProfile *usersdomain.UserProfile) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateProfileSQL, userProfile.UserID, userProfile.FirstName, userProfile.LastName, userProfile.PhoneNumber, userProfile.Country, userProfile.City, userProfile.DateOfBirth, userProfile.Gender, userProfile.UILanguage, userProfile.AvatarURL, userProfile.Timezone, userProfile.Bio)
	if err != nil {
		return fmt.Errorf("repository.UpdateUserProfile: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return usersdomain.ErrUserNotFound
	}

	return nil
}
