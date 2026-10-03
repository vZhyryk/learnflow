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

		if err := s.ensureNotDeletedByAdmin(ctx, user.ID); err != nil {
			return err
		}

		return s.issueRecoveryToken(ctx, user)
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

		if token.IsExpired() {
			return authdomain.ErrTokenExpired
		}

		user, err := s.userRepo.GetDeletedUserByID(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("recover_account: get deleted user: %w", err)
		}

		if user.Status != authdomain.StatusDeleted {
			return authdomain.ErrInvalidAccountState
		}

		if err := s.ensureNotDeletedByAdmin(ctx, user.ID); err != nil {
			return err
		}

		return s.restoreAccount(ctx, token.UserID, tokenHash)
	})
}

// restoreAccount reactivates the user and consumes the token; the Redis mark is cleared last because it cannot be rolled back.
// If COMMIT then fails, the mark is gone while the DB still says deleted, until the user retries.
func (s *Service) restoreAccount(ctx context.Context, userID, tokenHash string) error {
	if err := s.userRepo.RestoreUser(ctx, userID); err != nil {
		return fmt.Errorf("recover_account: restore user: %w", err)
	}

	if err := s.tokenRepo.MarkAccountRecoveryTokenUsed(ctx, tokenHash); err != nil {
		return fmt.Errorf("recover_account: mark token used: %w", err)
	}

	if err := s.blocklist.UnBlockUser(ctx, userID); err != nil {
		return fmt.Errorf("recover_account: clear user_blocked: %w: %w", authdomain.ErrBlocklistUnavailable, err)
	}

	return nil
}

// issueRecoveryToken creates the recovery token for user and queues the recovery email.
func (s *Service) issueRecoveryToken(ctx context.Context, user *authdomain.User) error {
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
			return events.TokenPayload{
				UserID:    user.ID,
				Email:     user.Email,
				ExpiresAt: expiresAt,
				RawToken:  rawToken,
				UserName:  userProfile.GetFirstName(),
			}, nil
		},
	)
}

// ensureNotDeletedByAdmin returns ErrDeletedByAdmin so recovery cannot undo an admin's deletion.
func (s *Service) ensureNotDeletedByAdmin(ctx context.Context, userID string) error {
	deletedByAdmin, err := s.audit.WasDeletedByAdmin(ctx, userID)
	if err != nil {
		return fmt.Errorf("recover_account: check deleted by admin: %w", err)
	}

	if deletedByAdmin {
		return authdomain.ErrDeletedByAdmin
	}

	return nil
}
