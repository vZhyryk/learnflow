package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ChangePassword updates the user's password after verifying the current one. Failed checks count towards the login
// lock and run outside the transaction so the counter survives the error; a locked account is rejected before bcrypt.
func (s *Service) ChangePassword(ctx context.Context, req authdomain.ChangePasswordRequest) error {
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("change_password: get user: %w", err)
	}

	if err := s.verifyOldPassword(ctx, user, req.OldPassword); err != nil {
		return err
	}

	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.cost)
		if err != nil {
			return fmt.Errorf("change_password: hash password: %w", err)
		}

		err = s.userRepo.UpdatePasswordHash(ctx, user.ID, string(passwordHash))
		if err != nil {
			return fmt.Errorf("change_password: update hash: %w", err)
		}

		if err = s.userRepo.ResetFailedLogin(ctx, user.ID); err != nil {
			return fmt.Errorf("change_password: reset failed login: %w", err)
		}

		if req.IsAllSessionsLogout {
			err = s.revokeUserSessions(ctx, "change_password", req.JTI, req.AccessTokenExpiresAt, func(ctx context.Context) error {
				return s.sessionRepo.RevokeAllUserSessions(ctx, req.UserID, nil, authdomain.RevokeReasonPasswordChanged)
			})
			if err != nil {
				return err
			}

			return s.revokeAllUserTokens(ctx, "change_password", req.UserID)
		}

		return nil
	})
}

// verifyOldPassword rejects a locked account before bcrypt, and counts a wrong password towards the login lock.
func (s *Service) verifyOldPassword(ctx context.Context, user *authdomain.User, oldPassword string) error {
	if user.LoginLockedUntil != nil && user.LoginLockedUntil.After(time.Now().UTC()) {
		return &authdomain.ErrAccountLockedError{LockedUntil: *user.LoginLockedUntil}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		incErr := s.userRepo.IncrementFailedLogin(ctx, user.ID, loginLockInterval, loginFailLimit)
		if incErr != nil && !errors.Is(incErr, authdomain.ErrUserNotFound) {
			return fmt.Errorf("change_password: increment failed login: %w", incErr)
		}
		return authdomain.ErrWrongPassword
	}

	return nil
}
