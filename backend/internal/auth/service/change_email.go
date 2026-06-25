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
	"strings"
	"time"
)

// InitiateEmailChange sends an email change confirmation token to the user's new address.
func (s *Service) InitiateEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		userProfile, err := s.userRepo.GetUserProfileByUserID(ctx, req.UserID)
		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("init_email_change: get user profile: %w", err)
		}

		if errors.Is(err, authdomain.ErrUserNotFound) {
			return err
		}

		user, err := s.userRepo.GetUserByID(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("init_email_change: get user: %w", err)
		}
		if strings.EqualFold(user.Email, req.NewEmail) {
			return authdomain.ErrEmailAlreadyInUse
		}

		nUser, err := s.userRepo.GetUserByEmail(ctx, req.NewEmail)
		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("init_email_change: check new email exists: %w", err)
		}

		if nUser != nil && nUser.ID != req.UserID {
			return authdomain.ErrEmailAlreadyInUse
		}

		rawToken, hashToken, err := tokens.GenerateSecureToken()
		if err != nil {
			return fmt.Errorf("init_email_change: generate token: %w", err)
		}

		expiresAt := time.Now().UTC().Add(emailChangeTokenTTL)

		token := &authdomain.EmailChangeToken{
			TokenBase: authdomain.TokenBase{
				UserID:    req.UserID,
				TokenHash: hashToken,
				ExpiresAt: expiresAt,
			},
			NewEmail: req.NewEmail,
		}

		if _, err = s.tokenRepo.CreateEmailChangeToken(ctx, token); err != nil {
			return fmt.Errorf("init_email_change: create token: %w", err)
		}

		payload := events.InitEmailChangeToken{
			UserID:    req.UserID,
			Email:     req.NewEmail,
			ExpiresAt: expiresAt,
			RawToken:  rawToken,
			UserName:  userProfile.FirstName,
		}

		err = s.outbox.Emit(ctx, events.AggregationTypeEmail, req.UserID, events.EventEmailChange, payload)
		if err != nil {
			return fmt.Errorf("init_email_change: emit event: %w", err)
		}

		return nil
	})
}

// ChangeEmail applies the email change after token verification.
func (s *Service) ChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest) error {
	sum := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(sum[:])

	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		token, err := s.tokenRepo.GetEmailChangeToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("change_email: get token: %w", err)
		}

		if token.ExpiresAt.Before(time.Now().UTC()) {
			return authdomain.ErrTokenExpired
		}

		if token.UserID != req.UserID {
			return authdomain.ErrInvalidToken
		}

		newMailUser, err := s.userRepo.GetUserByEmail(ctx, token.NewEmail)
		if err == nil && newMailUser != nil {
			return authdomain.ErrEmailAlreadyInUse
		}

		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("change_email: check email taken: %w", err)
		}

		err = s.userRepo.UpdateEmail(ctx, token.UserID, token.NewEmail)
		if err != nil {
			return fmt.Errorf("change_email: update email: %w", err)
		}

		err = s.tokenRepo.MarkEmailChangeTokenUsed(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("change_email: mark token used: %w", err)
		}

		if req.IsAllSessionsLogout {
			revokeErr := s.revokeAllUserSessions(ctx, token.UserID, req.JTI, req.AccessTokenExpiresAt)
			if revokeErr != nil {
				return revokeErr
			}
		}

		return nil
	})
}
