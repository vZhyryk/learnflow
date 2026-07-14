package usersservice

import (
	"context"
	"fmt"
	usersdomain "learnflow_backend/internal/users/domain"
)

// GetUserProfile returns the profile for the given user ID.
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	userProfile, err := s.usersRepo.GetUserProfileByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.GetUserProfile: %w", err)
	}
	return userProfile, nil
}
