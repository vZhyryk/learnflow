package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/tokens"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	loginLockInterval = "15 minutes"
	// loginFailLimit: max consecutive failed password attempts before the account is temporarily locked.
	// Distinct from loginCountLimit (session-level): this operates at the user level.
	loginFailLimit = 5
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
		return nil, fmt.Errorf("login: get user: %w", err)
	}

	dummyPasswordHash, err := bcrypt.GenerateFromPassword([]byte("dummy"), hashDefaultCost)
	if err != nil {
		return nil, fmt.Errorf("login: hash dummy: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.Password))
	if err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	return nil, authdomain.ErrInvalidCredentials
}

func (s *Service) loginHandleSession(ctx context.Context, req authdomain.LoginRequest, user *authdomain.User) (*authdomain.AuthTokens, error) {
	rawToken, tokenHash, err := tokens.GenerateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("login: generate token: %w", err)
	}

	accessToken, err := s.token.GenerateAccessToken(user, accessTokenTTL)
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
		UserID:       user.ID,
	}, nil
}
