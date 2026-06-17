package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	loginLockInterval = "15 minutes"
	loginFailLimit    = 5
)

// Login authenticates a user and returns access/refresh tokens.
func (s *Service) Login(ctx context.Context, req authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
	user, err := s.loginGetUser(ctx, req)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		incErr := s.userRepo.IncrementFailedLogin(ctx, user.ID, loginLockInterval, loginFailLimit)
		if incErr != nil && !errors.Is(incErr, authdomain.ErrUserNotFound) {
			return nil, fmt.Errorf("login: increment failed login: %w", incErr)
		}
		return nil, authdomain.ErrInvalidCredentials
	}

	switch user.Status {
	case authdomain.StatusBlocked:
		return nil, authdomain.ErrAccountBlocked
	case authdomain.StatusPendingVerification:
		return nil, authdomain.ErrEmailNotVerified
	case authdomain.StatusDeleted:
		return nil, authdomain.ErrInvalidCredentials
	}

	if user.LoginLockedUntil != nil && user.LoginLockedUntil.After(time.Now()) {
		return nil, authdomain.ErrAccountLocked
	}

	return s.loginHandleSession(ctx, req, user)
}

func (s *Service) loginGetUser(ctx context.Context, req authdomain.LoginRequest) (*authdomain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("service.loginGetUser: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(s.dummyPasswordHash, []byte(req.Password))
	if err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	return nil, authdomain.ErrInvalidCredentials
}

func (s *Service) loginHandleSession(ctx context.Context, req authdomain.LoginRequest, user *authdomain.User) (*authdomain.AuthTokens, error) {
	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("login: generate token: %w", err)
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("login: generate access token: %w", err)
	}

	sessionInput := &authdomain.UserSession{
		UserID:      user.ID,
		RefreshHash: tokenHash,
		UserAgent:   &req.UserAgent,
		IPAddress:   &req.IPAddress,
		ExpiresAt:   time.Now().Add(refreshTokenTTL),
	}

	var createdSession *authdomain.UserSession
	err = s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		createdSession, err = s.sessionRepo.CreateUserSession(ctx, sessionInput)
		if err != nil {
			return fmt.Errorf("login: create session: %w", err)
		}
		if err = s.userRepo.ResetFailedLogin(ctx, user.ID); err != nil {
			return fmt.Errorf("login: reset failed login: %w", err)
		}
		if err = s.userRepo.UpdateLastLoginAt(ctx, user.ID); err != nil {
			return fmt.Errorf("login: update last login: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("login: transaction: %w", err)
	}

	return &authdomain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresAt:    createdSession.ExpiresAt,
	}, nil
}

// Logout revokes the user's current session.
func (s *Service) Logout(_ context.Context, _ authdomain.LogoutRequest) error {
	return nil
}

// Register creates a new user account and sends an email verification token.
func (s *Service) Register(ctx context.Context, req authdomain.RegisterRequest) error {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, authdomain.ErrUserNotFound) {
		return err
	}

	if user != nil {
		return authdomain.ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), hashDefaultCost)
	if err != nil {
		return fmt.Errorf("register: hash password: %w", err)
	}

	user = &authdomain.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         authdomain.RoleUser,
		Status:       authdomain.StatusPendingVerification,
	}

	return s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		id, err := s.userRepo.CreateUser(ctx, user)
		if err != nil {
			return err
		}

		rawToken, hashToken, err := generateSecureToken()
		if err != nil {
			return fmt.Errorf("register: generate token: %w", err)
		}

		token := &authdomain.EmailVerificationToken{
			TokenBase: authdomain.TokenBase{
				UserID:    id,
				TokenHash: hashToken,
				ExpiresAt: time.Now().Add(emailVerificationTokenTTL),
			},
		}

		if _, err = s.tokenRepo.CreateEmailVerificationToken(ctx, token); err != nil {
			return fmt.Errorf("register: create verification token: %w", err)
		}

		payload := events.UserRegisteredPayload{
			UserID: id,
			Email:  user.Email,
			URL:    fmt.Sprintf("/api/v1/users/auth/email/verify/%s", rawToken),
		}
		err = s.outbox.Emit(ctx, events.AggregationTypeUser, id, events.EventUserRegistered, payload)
		if err != nil {
			return fmt.Errorf("register: emit event: %w", err)
		}

		return nil
	})
}

// Refresh rotates the refresh token and returns new access/refresh tokens.
func (s *Service) Refresh(_ context.Context, _ authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
	return nil, nil
}

// VerifyEmail confirms a user's email address using the provided token.
func (s *Service) VerifyEmail(_ context.Context, _ authdomain.VerifyEmailRequest) error {
	return nil
}

// ChangePassword updates the user's password after verifying the current one.
func (s *Service) ChangePassword(_ context.Context, _ authdomain.ChangePasswordRequest) error {
	return nil
}

// InitiatePasswordReset sends a password reset token to the user's email.
func (s *Service) InitiatePasswordReset(_ context.Context, _ authdomain.RequestPasswordResetRequest) error {
	return nil
}

// ResetPassword sets a new password using the provided reset token.
func (s *Service) ResetPassword(_ context.Context, _ authdomain.ResetPasswordRequest) error {
	return nil
}

// InitiateEmailChange sends an email change confirmation token to the user's new address.
func (s *Service) InitiateEmailChange(_ context.Context, _ authdomain.RequestEmailChangeRequest) error {
	return nil
}

// ChangeEmail applies the email change after token verification.
func (s *Service) ChangeEmail(_ context.Context, _ authdomain.EmailChangeRequest) error {
	return nil
}

// RecoverAccount restores a soft-deleted account using the provided recovery token.
func (s *Service) RecoverAccount(_ context.Context, _ authdomain.RecoverAccountRequest) error {
	return nil
}
