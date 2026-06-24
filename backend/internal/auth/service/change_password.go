package authservice

import (
	"context"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

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

		if req.NewPassword == req.OldPassword {
			return authdomain.ErrSamePassword
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), hashDefaultCost)
		if err != nil {
			return fmt.Errorf("change_password: hash password: %w", err)
		}

		err = s.userRepo.UpdatePasswordHash(ctx, user.ID, string(hash))
		if err != nil {
			return fmt.Errorf("change_password: update hash: %w", err)
		}

		if req.IsAllSessionsLogout {
			revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, req.UserID, nil, authdomain.RevokeReasonPasswordChanged)
			if revokeErr != nil {
				return fmt.Errorf("change_password: revoke sessions: %w", revokeErr)
			}

			remaining := time.Until(req.AccessTokenExpiresAt)
			if remaining > 0 && req.JTI != "" {
				_, err := s.redis.SetNX(ctx, "blocklist:"+req.JTI, "1", remaining).Result()
				if err != nil {
					return fmt.Errorf("change_password: session blocklist: %w", err)
				}
			}

		}

		return nil
	})
}
