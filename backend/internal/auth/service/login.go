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
	// loginFailLimit: consecutive failed attempts before a temporary lock (see sarch-14).
	loginFailLimit = 5
)

// bcryptCompareHashAndPassword indirects bcrypt.CompareHashAndPassword so tests can spy on
// the dummy-hash comparison in loginGetUser instead of asserting on wall-clock timing.
var bcryptCompareHashAndPassword = bcrypt.CompareHashAndPassword

// Login authenticates a user and returns access/refresh tokens. Check order (lock →
// bcrypt → status) is timing-attack-sensitive — do not reorder, see TestLoginConstantTimeUserEnumeration.
func (s *Service) Login(ctx context.Context, req authdomain.LoginRequest) (*authdomain.AuthTokens, error) {
	user, err := s.loginGetUser(ctx, req)
	if err != nil {
		return nil, err
	}

	if user.LoginLockedUntil != nil && user.LoginLockedUntil.After(time.Now().UTC()) {
		return nil, &authdomain.ErrAccountLockedError{LockedUntil: *user.LoginLockedUntil}
	}

	err = bcryptCompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
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

	return s.loginHandleSession(ctx, req, user)
}

// loginGetUser fetches the user by email, running a dummy bcrypt comparison on a miss to
// keep response timing indistinguishable from a real check (prevents user enumeration).
func (s *Service) loginGetUser(ctx context.Context, req authdomain.LoginRequest) (*authdomain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("login: get user: %w", err)
	}

	// Dummy bcrypt against a precomputed hash — keeps timing constant; never generate per-request.
	bcryptCompareHashAndPassword(s.dummyPasswordHash, []byte(req.Password)) //nolint:errcheck,gosec // discarded intentionally, only used to consume constant time
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
		ExpiresAt:   time.Now().UTC().Add(refreshTokenTTL),
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
