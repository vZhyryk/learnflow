package usersservice

import (
	"context"
	usersdomain "learnflow_backend/internal/users/domain"
)

type mockUserProfileRepo struct {
	getUserProfileByID func(ctx context.Context, userID string) (*usersdomain.UserProfile, error)
	updateUserProfile  func(ctx context.Context, userProfile *usersdomain.UserProfile) error
}

func (m *mockUserProfileRepo) GetUserProfileByID(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	return m.getUserProfileByID(ctx, userID)
}

func (m *mockUserProfileRepo) UpdateUserProfile(ctx context.Context, userProfile *usersdomain.UserProfile) error {
	return m.updateUserProfile(ctx, userProfile)
}
