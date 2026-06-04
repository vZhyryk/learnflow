package authservice

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
)

// Service implements the auth domain Service interface.
type Service struct{}

func (s *Service) Login(ctx context.Context, req authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
	return nil, nil
}
func (s *Service) Logout(ctx context.Context, req authdomain.LogoutRequest) error {
	return nil
}
func (s *Service) Register(ctx context.Context, req authdomain.RegisterRequest) error {
	return nil
}
func (s *Service) Refresh(ctx context.Context, req authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
	return nil, nil
}
func (s *Service) VerifyEmail(ctx context.Context, req authdomain.VerifyEmailRequest) error {
	return nil
}
func (s *Service) ChangePassword(ctx context.Context, req authdomain.ChangePasswordRequest) error {
	return nil
}
func (s *Service) InitiatePasswordReset(ctx context.Context, req authdomain.RequestPasswordResetRequest) error {
	return nil
}
func (s *Service) ResetPassword(ctx context.Context, req authdomain.ResetPasswordRequest) error {
	return nil
}
func (s *Service) InitiateEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) error {
	return nil
}
func (s *Service) ChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest) error {
	return nil
}
func (s *Service) RecoverAccount(ctx context.Context, req authdomain.RecoverAccountRequest) error {
	return nil
}
