package authservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"time"
)

// Logout revokes the user's current session.
func (s *Service) Logout(ctx context.Context, req authdomain.LogoutRequest) (string, error) {
	user, ok := appcontext.UserFromContext(ctx)
	if !ok {
		return "", authdomain.ErrInvalidCredentials
	}

	refreshHash := sha256.Sum256([]byte(req.RefreshToken))
	refreshHashHex := hex.EncodeToString(refreshHash[:])

	session, err := s.sessionRepo.GetUserSessionByRefreshToken(ctx, refreshHashHex)
	if err != nil && !errors.Is(err, authdomain.ErrSessionNotFound) {
		return "", fmt.Errorf("logout: get session: %w", err)
	}

	if session == nil {
		return "", nil
	}

	if session.UserID != user.ID {
		return "", authdomain.ErrInvalidCredentials
	}

	if session.RevokedAt == nil {
		err = s.sessionRepo.RevokeUserSession(ctx, session.ID, user.ID, authdomain.RevokeReasonLogout)
		if err != nil {
			return "", fmt.Errorf("logout: revoke session: %w", err)
		}

		remaining := time.Until(req.AccessTokenExpiresAt)
		if remaining > 0 && req.JTI != "" {
			_, err := s.redis.SetNX(ctx, "blocklist:"+req.JTI, "1", remaining).Result()
			if err != nil {
				return "", fmt.Errorf("logout: session blocklist: %w", err)
			}
		}
	}

	return user.ID, nil
}
