package authservice

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"
	"learnflow_backend/internal/shared/tokens"
)

// Logout revokes the user's current session.
func (s *Service) Logout(ctx context.Context, req authdomain.LogoutRequest) (string, error) {
	user, ok := appcontext.UserFromContext(ctx)
	if !ok {
		return "", authdomain.ErrInvalidCredentials
	}

	refreshHashHex := tokens.MakeHash(req.RefreshToken)
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
		err = s.revokeUserSessions(ctx, "logout", req.JTI, req.AccessTokenExpiresAt, func(ctx context.Context) error {
			return s.sessionRepo.RevokeUserSession(ctx, session.ID, user.ID, authdomain.RevokeReasonLogout)
		})
		if err != nil {
			return "", err
		}
	}

	return user.ID, nil
}
