package usershttp_test

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	usersdomain "learnflow_backend/internal/users/domain"

	"github.com/justinas/alice"
)

func noopChain() alice.Chain { return alice.New() }

// errWriter, decodeBody, and withUser are shared test helpers defined once in testutil.
type errWriter = testutil.ErrWriter

var decodeBody = testutil.DecodeBody

var withUser = testutil.WithUser

// mockService implements usersdomain.Service via function fields.
type mockService struct {
	getUserProfile    func(ctx context.Context, userID string) (*usersdomain.UserProfile, error)
	changeUserProfile func(ctx context.Context, req usersdomain.ChangeUserProfileRequest) error
}

func (m *mockService) GetUserProfile(ctx context.Context, userID string) (*usersdomain.UserProfile, error) {
	if m.getUserProfile == nil {
		panic("mockService.getUserProfile not set")
	}
	return m.getUserProfile(ctx, userID)
}

func (m *mockService) ChangeUserProfile(ctx context.Context, req usersdomain.ChangeUserProfileRequest) error {
	if m.changeUserProfile == nil {
		panic("mockService.changeUserProfile not set")
	}
	return m.changeUserProfile(ctx, req)
}
