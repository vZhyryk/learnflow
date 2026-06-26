package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"time"
)

// InitRecoverAccount sends an account recovery email for a soft-deleted user.
func (s *Service) InitRecoverAccount(ctx context.Context, req authdomain.RequestRecoverAccountRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.GetDeletedUserByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
			return fmt.Errorf("init_recover_account: get user: %w", err)
		}

		if errors.Is(err, authdomain.ErrUserNotFound) {
			return nil
		}

		if user.Status != authdomain.StatusDeleted {
			return authdomain.ErrInvalidAccountState
		}

		userProfile, err := s.userRepo.GetUserProfileByUserID(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("init_recover_account: get user profile: %w", err)
		}

		return s.emitTokenEvent(ctx, user.ID, accountRecoverTokenTTL, events.AggregationTypeAccount, events.EventAccountRecovery,
			func(ctx context.Context, rawToken, hashToken string, expiresAt time.Time) (any, error) {
				token := &authdomain.AccountRecoveryToken{
					TokenBase: authdomain.TokenBase{
						UserID:    user.ID,
						TokenHash: hashToken,
						ExpiresAt: expiresAt,
					},
				}
				_, err := s.tokenRepo.CreateAccountRecoveryToken(ctx, token)
				if err != nil {
					return nil, fmt.Errorf("init_recover_account: create token: %w", err)
				}
				return events.InitAccountRecoveryToken{
					UserID:    user.ID,
					Email:     user.Email,
					ExpiresAt: expiresAt,
					RawToken:  rawToken,
					UserName:  userProfile.FirstName,
				}, nil
			},
		)
	})
}

// RecoverAccount restores a soft-deleted account using the provided recovery token.
func (s *Service) RecoverAccount(ctx context.Context, req authdomain.RecoverAccountRequest) error {
	tokenHash := tokens.MakeHash(req.Token)
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		token, err := s.tokenRepo.GetAccountRecoveryToken(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("recover_account: get token: %w", err)
		}

		if token.ExpiresAt.Before(time.Now().UTC()) {
			return authdomain.ErrTokenExpired
		}

		user, err := s.userRepo.GetDeletedUserByID(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("recover_account: get deleted user: %w", err)
		}

		if user.Status != authdomain.StatusDeleted {
			return authdomain.ErrInvalidAccountState
		}

		err = s.userRepo.RestoreUser(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("recover_account: restore user: %w", err)
		}

		err = s.tokenRepo.MarkAccountRecoveryTokenUsed(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("recover_account: mark token used: %w", err)
		}

		return nil
	})
}
