package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/tokens"
)

// Refresh rotates the refresh token and returns new access/refresh tokens.
func (s *Service) Refresh(ctx context.Context, req authdomain.RefreshRequest) (*authdomain.AuthTokens, error) {
	refreshHash := sha256.Sum256([]byte(req.RefreshToken))
	refreshHashHex := hex.EncodeToString(refreshHash[:])

	var user *authdomain.User
	var rawToken string
	session := &authdomain.UserSession{}

	err := s.transactor.InTransaction(ctx, func(ctx context.Context) error {
		uSession, err := s.sessionRepo.GetUserSessionByRefreshToken(ctx, refreshHashHex)
		if err != nil {
			// Token not found as current — check if it was already rotated (exists as previous_refresh_hash).
			// If yes: someone is replaying a rotated-out token, which indicates theft or a replay attack.
			// Revoke all sessions for that user as a precaution.
			prevSession, prevErr := s.sessionRepo.GetSessionByPrevHash(ctx, refreshHashHex)
			if prevErr != nil && !errors.Is(prevErr, authdomain.ErrSessionNotFound) {
				return fmt.Errorf("refresh: get prev session: %w", prevErr)
			}

			if prevSession != nil {
				revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, prevSession.UserID, nil, authdomain.RevokeReasonSuspiciousActivity)
				if revokeErr != nil {
					return fmt.Errorf("refresh: revoke all sessions (suspicious): %w", revokeErr)
				}
				return authdomain.ErrSessionRevoked
			}
			return err
		}

		user, err = s.userRepo.GetUserByID(ctx, uSession.UserID)
		if err != nil {
			return fmt.Errorf("refresh: get user: %w", err)
		}

		switch user.Status {
		case authdomain.StatusBlocked:
			return authdomain.ErrAccountBlocked
		case authdomain.StatusDeleted:
			return authdomain.ErrInvalidCredentials
		}

		// Second check — different scenario from the one above.
		// Token IS valid as current, but also appears as previous_refresh_hash in another session.
		// This means the same token was used to rotate a different session, which should not happen.
		// Indicates a race condition or token duplication attack — revoke all sessions.
		prevSession, err := s.sessionRepo.GetSessionByPrevHash(ctx, refreshHashHex)
		if err != nil && !errors.Is(err, authdomain.ErrSessionNotFound) {
			return fmt.Errorf("refresh: get prev session (reuse check): %w", err)
		}

		if prevSession != nil {
			revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, uSession.UserID, nil, authdomain.RevokeReasonSuspiciousActivity)
			if revokeErr != nil {
				return fmt.Errorf("refresh: revoke all sessions (reuse): %w", revokeErr)
			}
			return authdomain.ErrSessionRevoked
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

	accessToken, err := s.token.GenerateAccessToken(user, accessTokenTTL)
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
