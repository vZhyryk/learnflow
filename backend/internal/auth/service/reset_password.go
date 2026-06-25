package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// InitiatePasswordReset sends a password reset token to the user's email.
func (s *Service) InitiatePasswordReset(ctx context.Context, req authdomain.RequestPasswordResetRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("init_password_reset: get user: %w", err)
		}

		if errors.Is(err, authdomain.ErrUserNotFound) {
			return nil
		}

		userProfile, err := s.userRepo.GetUserProfileByUserID(ctx, user.ID)
		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("init_password_reset: get user profile: %w", err)
		}

		rawToken, hashToken, err := tokens.GenerateSecureToken()
		if err != nil {
			return fmt.Errorf("init_password_reset: generate token: %w", err)
		}

		expiresAt := time.Now().UTC().Add(passwordResetTokenTTL)

		token := &authdomain.PasswordResetToken{
			TokenBase: authdomain.TokenBase{
				UserID:    user.ID,
				TokenHash: hashToken,
				ExpiresAt: expiresAt,
			},
		}

		if _, err = s.tokenRepo.CreatePasswordResetToken(ctx, token); err != nil {
			return fmt.Errorf("init_password_reset: create token: %w", err)
		}
		payload := events.InitPasswordResetToken{
			UserID:    user.ID,
			Email:     user.Email,
			ExpiresAt: expiresAt,
			RawToken:  rawToken,
			UserName:  userProfile.FirstName,
		}

		err = s.outbox.Emit(ctx, events.AggregationTypePassword, user.ID, events.EventPasswordReset, payload)
		if err != nil {
			return fmt.Errorf("init_password_reset: emit event: %w", err)
		}

		return nil
	})
}

// ResetPassword sets a new password using the provided reset token.
func (s *Service) ResetPassword(ctx context.Context, req authdomain.ResetPasswordRequest) error {
	sum := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(sum[:])

	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		token, err := s.tokenRepo.GetPasswordResetToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("reset_password: get token: %w", err)
		}

		if token.ExpiresAt.Before(time.Now().UTC()) {
			return authdomain.ErrTokenExpired
		}

		user, err := s.userRepo.GetUserByID(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("reset_password: get user: %w", err)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), hashDefaultCost)
		if err != nil {
			return fmt.Errorf("reset_password: hash password: %w", err)
		}

		err = s.userRepo.UpdatePasswordHash(ctx, user.ID, string(hash))
		if err != nil {
			return fmt.Errorf("reset_password: update hash: %w", err)
		}

		err = s.tokenRepo.MarkPasswordResetTokenUsed(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("reset_password: mark token used: %w", err)
		}

		revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, user.ID, nil, authdomain.RevokeReasonPasswordReset)
		if revokeErr != nil {
			return fmt.Errorf("reset_password: revoke sessions: %w", revokeErr)
		}

		return nil
	})
}
