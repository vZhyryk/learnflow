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

// InitiateEmailChange sends an email change confirmation token to the user's new address.
func (s *Service) InitiateEmailChange(ctx context.Context, req authdomain.RequestEmailChangeRequest) error {
	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		userProfile, err := s.getExistsUserProfileChangeEmail(ctx, req)
		if err != nil {
			return err
		}
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
				_, err := s.tokenRepo.CreateEmailChangeToken(ctx, token)

				if err != nil {
					return nil, fmt.Errorf("init_email_change: create token: %w", err)
				}

				return events.InitEmailChangeToken{
					UserID:    req.UserID,
					Email:     req.NewEmail,
					ExpiresAt: expiresAt,
					RawToken:  rawToken,
					UserName:  userProfile.FirstName,
				}, nil
			},
		)
	})
}

func (s *Service) getExistsUserProfileChangeEmail(ctx context.Context, req authdomain.RequestEmailChangeRequest) (*authdomain.UserProfile, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("init_email_change: get user: %w", err)
	}
	if strings.EqualFold(user.Email, req.NewEmail) {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	nUser, err := s.userRepo.GetUserByEmail(ctx, req.NewEmail)
	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("init_email_change: check new email exists: %w", err)
	}

	if nUser != nil && nUser.ID != req.UserID {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	userProfile, err := s.userRepo.GetUserProfileByUserID(ctx, user.ID)
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

		if req.IsAllSessionsLogout {
			return s.revokeUserSessions(ctx, "change_email", req.JTI, req.AccessTokenExpiresAt, func(ctx context.Context) error {
				return s.sessionRepo.RevokeAllUserSessions(ctx, token.UserID, nil, authdomain.RevokeReasonEmailChanged)
			})
		}

		return nil
	})
}

func (s *Service) getTokenAndValidateUserChangeEmail(ctx context.Context, req authdomain.EmailChangeRequest, tokenHash string) (*authdomain.EmailChangeToken, error) {
	token, err := s.tokenRepo.GetEmailChangeToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("change_email: get token: %w", err)
	}

	if token.ExpiresAt.Before(time.Now().UTC()) {
		return nil, authdomain.ErrTokenExpired
	}

	if token.UserID != req.UserID {
		return nil, authdomain.ErrInvalidToken
	}

	newMailUser, err := s.userRepo.GetUserByEmail(ctx, token.NewEmail)
	if err == nil && newMailUser != nil {
		return nil, authdomain.ErrEmailAlreadyInUse
	}

	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("change_email: check email taken: %w", err)
	}

	return token, nil
}
