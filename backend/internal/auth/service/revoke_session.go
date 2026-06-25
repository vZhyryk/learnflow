package authservice

import (
	"context"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"
)

func (s *Service) revokeAllUserSessions(ctx context.Context, userID, jti string, accessTokenExpiresAt time.Time) error {
	revokeErr := s.sessionRepo.RevokeAllUserSessions(ctx, userID, nil, authdomain.RevokeReasonEmailChanged)
	if revokeErr != nil {
		return fmt.Errorf("change_email: revoke sessions: %w", revokeErr)
	}

	// Redis SetNX is intentionally inside the transaction closure.
	// If Redis fails, the error propagates → InTransaction rolls back the DB
	// changes → email remains unchanged. The user gets a 500 but no state
	// divergence occurs (jti is never blocked while the email stays old).
	// Trade-off: Redis unavailability also prevents a successful email change.
	remaining := time.Until(accessTokenExpiresAt)
	if remaining > 0 && jti != "" {
		_, err := s.redis.SetNX(ctx, "blocklist:"+jti, "1", remaining).Result()
		if err != nil {
			return fmt.Errorf("change_email: session blocklist: %w", err)
		}
	}

	return nil
}
