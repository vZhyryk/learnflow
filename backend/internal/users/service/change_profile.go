package usersservice

import (
	"context"
	"fmt"
	usersdomain "learnflow_backend/internal/users/domain"
)

// ChangeUserProfile applies a partial update to the user's profile.
func (s *Service) ChangeUserProfile(ctx context.Context, req usersdomain.ChangeUserProfileRequest) error {
	userProfile, err := s.usersRepo.GetUserProfileByID(ctx, *req.UserID)
	if err != nil {
		return fmt.Errorf("service.ChangeUserProfile: %w", err)
	}

	req.Apply(userProfile)

	return s.usersRepo.UpdateUserProfile(ctx, userProfile)
}
