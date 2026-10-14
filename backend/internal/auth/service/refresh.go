package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/tokens"
	"time"
)

// refreshReuseGrace is how long after a rotation the previous refresh token is still accepted. A client
// that lost the rotation response (or a second tab) retries with the old token; inside the window that is
// treated as a retry and rotated again instead of being read as token theft.
const refreshReuseGrace = 20 * time.Second

// Refresh rotates the refresh token and returns new access/refresh tokens.
func (s *Service) Refresh(ctx context.Context, req authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
	refreshHashHex := tokens.MakeHash(req.RefreshToken)

	var user *authdomain.User
	var rawToken string
	session := &authdomain.UserSession{}

	graceSession, reuseErr := s.checkRefreshSessionPrevHash(ctx, refreshHashHex)
	if reuseErr != nil {
		return nil, reuseErr
	}

	err := s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		uSession, err := s.sessionRepo.GetUserSessionByRefreshToken(ctx, refreshHashHex)
		if err != nil {
			if graceSession == nil || !errors.Is(err, authdomain.ErrSessionNotFound) {
				return err
			}
			uSession = graceSession
		}

		user, err = s.getRefreshUserAndCheckStatusByID(ctx, uSession.UserID)
		if err != nil {
			return err
		}

		rToken, tokenHash, err := tokens.GenerateSecureToken()
		if err != nil {
			return fmt.Errorf("refresh: generate token: %w", err)
		}

		err = s.sessionRepo.UpdateSessionToken(ctx, uSession.ID, tokenHash, req.UserAgent, req.IPAddress)
		if err != nil {
			return fmt.Errorf("refresh: update session token: %w", err)
		}

		rawToken = rToken
		session = uSession

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("refresh: transaction: %w", err)
	}

	accessToken, err := s.token.GenerateAccessToken(user, tokens.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("refresh: generate access token: %w", err)
	}

	return &authdomain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresAt:    session.ExpiresAt,
		UserID:       session.UserID,
	}, nil
}

// checkRefreshSessionPrevHash detects reuse of a rotated refresh token. Inside refreshReuseGrace of the
// rotation it returns the session so the caller can rotate it again; outside, it revokes every session of
// the user and fails.
func (s *Service) checkRefreshSessionPrevHash(ctx context.Context, refreshHashHex string) (*authdomain.UserSession, error) {
	prevSession, err := s.sessionRepo.GetSessionByPrevHash(ctx, refreshHashHex)
	if err != nil && !errors.Is(err, authdomain.ErrSessionNotFound) {
		return nil, fmt.Errorf("refresh: get prev session (reuse check): %w", err)
	}

	if prevSession == nil {
		return nil, nil
	}

	if withinRefreshGrace(prevSession) {
		return prevSession, nil
	}

	// Tokens first: a Redis failure then fails the request before any session is touched.
	if tokensErr := s.revokeAllUserTokens(ctx, "refresh", prevSession.UserID); tokensErr != nil {
		return nil, tokensErr
	}

	revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, prevSession.UserID, nil, authdomain.RevokeReasonSuspiciousActivity)
	if revokeErr != nil {
		return nil, fmt.Errorf("refresh: revoke all sessions (reuse): %w", revokeErr)
	}

	return nil, authdomain.ErrSessionRevoked
}

func withinRefreshGrace(session *authdomain.UserSession) bool {
	return session.RevokedAt == nil && session.LastSeenAt != nil && time.Since(*session.LastSeenAt) <= refreshReuseGrace
}

func (s *Service) getRefreshUserAndCheckStatusByID(ctx context.Context, userID string) (*authdomain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("refresh: get user: %w", err)
	}

	switch user.Status {
	case authdomain.StatusBlocked:
		return nil, authdomain.ErrAccountBlocked
	case authdomain.StatusDeleted:
		return nil, authdomain.ErrInvalidCredentials
	}

	return user, nil
}
