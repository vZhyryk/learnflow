package authservice

import (
	"context"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"

	"golang.org/x/crypto/bcrypt"
)

// ChangePassword updates the user's password after verifying the current one.
func (s *Service) ChangePassword(ctx context.Context, req authdomain.ChangePasswordRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.GetUserByID(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("change_password: get user: %w", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
		if err != nil {
			return authdomain.ErrWrongPassword
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.cost)
		if err != nil {
			return fmt.Errorf("change_password: hash password: %w", err)
		}

		err = s.userRepo.UpdatePasswordHash(ctx, user.ID, string(passwordHash))
		if err != nil {
			return fmt.Errorf("change_password: update hash: %w", err)
		}

		if req.IsAllSessionsLogout {
			return s.revokeUserSessions(ctx, "change_password", req.JTI, req.AccessTokenExpiresAt, func(ctx context.Context) error {
				return s.sessionRepo.RevokeAllUserSessions(ctx, req.UserID, nil, authdomain.RevokeReasonPasswordChanged)
			})
		}

		return nil
	})
}
