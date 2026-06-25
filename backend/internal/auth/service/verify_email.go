package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"
)

// VerifyEmail confirms a user's email address using the provided token.
func (s *Service) VerifyEmail(ctx context.Context, req authdomain.VerifyEmailRequest) (string, error) {
	sum := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(sum[:])

	var userID string
	err := s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		token, err := s.tokenRepo.GetEmailVerificationToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("verify_email.GetEmailVerificationToken: %w", err)
		}

		if token.ExpiresAt.Before(time.Now().UTC()) {
			return authdomain.ErrTokenExpired
		}

		err = s.userRepo.UpdateEmailVerifiedAt(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("verify_email.UpdateEmailVerifiedAt: %w", err)
		}

		err = s.userRepo.UpdateStatus(ctx, token.UserID, authdomain.StatusActive)
		if err != nil {
			return fmt.Errorf("verify_email.UpdateStatus: %w", err)
		}

		err = s.tokenRepo.MarkEmailVerificationTokenUsed(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("verify_email.MarkEmailVerificationTokenUsed: %w", err)
		}

		userID = token.UserID
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("verify_email: %w", err)
	}

	return userID, nil
}
