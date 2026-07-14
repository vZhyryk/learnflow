package usersservice

import (
	"context"
	usersdomain "learnflow_backend/internal/users/domain"
)

// mockUserProfileRepo implements usersdomain.UserProfileRepository via function fields.
// Only set the fields needed for each test case — unset fields panic with a clear message.
type mockUserProfileRepo struct {
	getUserProfileByID func(ctx context.Context, userID string) (*usersdomain.UserProfile, error)
	updateUserProfile  func(ctx context.Context, userProfile *usersdomain.UserProfile) error
}

func (m *mockUserProfileRepo) GetUserProfileByID(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	if m.getUserProfileByID == nil {
		panic("mockUserProfileRepo.getUserProfileByID not set")
	}
	return m.getUserProfileByID(ctx, userID)
}

func (m *mockUserProfileRepo) UpdateUserProfile(ctx context.Context, userProfile *usersdomain.UserProfile) error {
	if m.updateUserProfile == nil {
		panic("mockUserProfileRepo.updateUserProfile not set")
	}
	return m.updateUserProfile(ctx, userProfile)
}

func newTestService(repo *mockUserProfileRepo) *Service {
	return New(repo)
}
