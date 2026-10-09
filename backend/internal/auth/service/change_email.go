package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"strings"
	"time"
)

// InitiateEmailChange sends an email change confirmation token to the user's new address. An address that belongs to
// another account is answered like a free one (no token, no email), so the endpoint cannot be used to probe emails.
// The password check runs outside the transaction so a failed attempt still counts towards the login lock.
func (s *Service) InitiateEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) error {
	user, err := s.getUserForEmailChange(ctx, req)
	if err != nil {
		return err
	}

	if verifyErr := s.verifyCurrentPassword(ctx, user, req.Password, "init_email_change"); verifyErr != nil {
		return verifyErr
	}

	userProfile, err := s.getProfileForNewEmail(ctx, req, user.ID)
	if errors.Is(err, authdomain.ErrEmailAlreadyInUse) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		return s.emitTokenEvent(ctx, req.UserID, emailChangeTokenTTL, events.AggregationTypeEmail, events.EventEmailChange,
			func(ctx context.Context, rawToken, hashToken string, expiresAt time.Time) (any, error) {
				token := &authdomain.EmailChangeToken{
					TokenBase: authdomain.TokenBase{
						UserID:    req.UserID,
						TokenHash: hashToken,
						ExpiresAt: expiresAt,
					},
					NewEmail: req.NewEmail,
				}

				if _, err := s.tokenRepo.CreateEmailChangeToken(ctx, token); err != nil {
					return nil, fmt.Errorf("init_email_change: create token: %w", err)
				}

				return events.TokenPayload{
					UserID:    req.UserID,
					Email:     req.NewEmail,
					ExpiresAt: expiresAt,
					RawToken:  rawToken,
					UserName:  userProfile.GetFirstName(),
				}, nil
			})
	})
}

func (s *Service) getUserForEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) (*authdomain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("init_email_change: get user: %w", err)
	}
	if strings.EqualFold(user.Email, req.NewEmail) {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	return user, nil
}

func (s *Service) getProfileForNewEmail(ctx context.Context, req authdomain.RequestEmailChangeRequest, userID string) (*authdomain.UserProfile, error) {
	existingUserAtNewEmail, err := s.userRepo.GetUserByEmail(ctx, req.NewEmail)
	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("init_email_change: check new email exists: %w", err)
	}

	if existingUserAtNewEmail != nil && existingUserAtNewEmail.ID != req.UserID {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	userProfile, err := s.userRepo.GetUserProfileByUserID(ctx, userID)
	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("init_email_change: get user profile: %w", err)
	}

	if errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, err
	}

	return userProfile, nil
}

// ChangeEmail applies the email change after token verification.
func (s *Service) ChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest) error {
	tokenHash := tokens.MakeHash(req.Token)
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		token, err := s.getTokenAndValidateUserChangeEmail(ctx, req, tokenHash)
		if err != nil {
			return err
		}

		err = s.userRepo.UpdateEmail(ctx, token.UserID, token.NewEmail)
		if err != nil {
			return fmt.Errorf("change_email: update email: %w", err)
		}

		err = s.tokenRepo.MarkEmailChangeTokenUsed(ctx, tokenHash)
		if err != nil {
			return fmt.Errorf("change_email: mark token used: %w", err)
		}

		return s.logoutEverywhere(ctx, "change_email", token.UserID, func(ctx context.Context) error {
			return s.sessionRepo.RevokeAllUserSessions(ctx, token.UserID, nil, authdomain.RevokeReasonEmailChanged)
		})
	})
}

func (s *Service) getTokenAndValidateUserChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest, tokenHash string) (*authdomain.EmailChangeToken, error) {
	token, err := s.tokenRepo.GetEmailChangeToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("change_email: get token: %w", err)
	}

	if token.IsExpired() {
		return nil, authdomain.ErrTokenExpired
	}

	if token.UserID != req.UserID {
		return nil, authdomain.ErrInvalidToken
	}

	existingUserAtNewEmail, err := s.userRepo.GetUserByEmail(ctx, token.NewEmail)
	if err == nil && existingUserAtNewEmail != nil {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("change_email: check email taken: %w", err)
	}

	return token, nil
}
