package usersdomain

import (
	"context"
)

// Service defines the business operations for the users module.
type Service interface {
	ChangeUserProfile(ctx context.Context, req ChangeUserProfileRequest) error
	GetUserProfile(ctx context.Context, userID string) (*UserProfile, error)
}

// UserProfileRepository defines persistence operations for UserProfile.
type UserProfileRepository interface {
	GetUserProfileByID(ctx context.Context, userID string) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, userProfile *UserProfile) error
}
