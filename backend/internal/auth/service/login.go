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
	// loginFailLimit: max consecutive failed password attempts (tracked per user via
	// IncrementFailedLogin) before the account is temporarily locked. The repository also
	// exposes a session-level counter (UpdateFailedLoginAttempts) but no service code
	// currently calls it — see .planning/STATE.md Deferred Items (sarch-14).
	loginFailLimit = 5
)

// bcryptCompareHashAndPassword is a package-level indirection over bcrypt.CompareHashAndPassword.
// Tests substitute this var with a call-counting spy to assert the dummy-hash comparison in
// loginGetUser actually executes (the constant-time, user-enumeration-prevention guarantee),
// without relying on flaky wall-clock timing assertions.
var bcryptCompareHashAndPassword = bcrypt.CompareHashAndPassword

// Login authenticates a user and returns access/refresh tokens.
//
// Check order is intentional and security-sensitive — do not reorder:
//
//  1. loginGetUser: fetch user by email. If not found, run dummy bcrypt to prevent
//     user-enumeration via timing (found vs not-found responses take the same time).
//
//  2. LoginLockedUntil (brute-force lock) before bcrypt: skipping bcrypt for locked
//     accounts avoids ~100ms CPU cost per request during an active attack. Revealing
//     that an account is locked is acceptable — the attacker most likely triggered the
//     lock themselves, and the HTTP 429 response already communicates this explicitly.
//
//  3. bcrypt: always runs for existing, non-locked accounts regardless of outcome.
//     bcrypt takes the same ~100ms whether the password is correct or wrong, so the
//     result does not leak via timing.
//
//  4. Status checks (StatusBlocked, StatusPendingVerification) after bcrypt: intentional.
//     Checking status before bcrypt would create a timing oracle — a blocked account
//     with the correct password would return faster (no bcrypt) than with a wrong
//     password, letting an attacker confirm a valid password by measuring response time.
//     Placing status checks after bcrypt eliminates this difference.
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

// loginGetUser fetches the user by email. If the email does not exist, it runs a dummy
// bcrypt comparison against a precomputed hash to match the response time of a real
// password check, preventing user-enumeration via timing differences.
func (s *Service) loginGetUser(ctx context.Context, req authdomain.LoginRequest) (*authdomain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, authdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("login: get user: %w", err)
	}

	// Dummy bcrypt: keeps response time constant whether the email exists or not.
	// dummyPasswordHash is precomputed at startup — never generate it per-request
	// (GenerateFromPassword is slow by design and would invert the timing signature).
	bcryptCompareHashAndPassword(s.dummyPasswordHash, []byte(req.Password)) //nolint:errcheck // error is intentionally discarded — call exists only to consume constant time and prevent user enumeration via timing
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
