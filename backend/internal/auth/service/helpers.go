package authservice

import (
	"context"
	"fmt"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"time"
)

func (s *Service) revokeUserSessions(ctx context.Context, caller, jti string, accessTokenExpiresAt time.Time, fn func(ctx context.Context) error) error {
	if err := fn(ctx); err != nil {
		return fmt.Errorf("%s: revoke sessions: %w", caller, err)
	}

	// Redis SetNX stays inside the tx closure so a Redis failure rolls back the DB change too —
	// no state divergence, at the cost of Redis outages also blocking the DB change.
	remaining := time.Until(accessTokenExpiresAt)
	if remaining > 0 && jti != "" {
		_, err := s.redisClient.SetNX(ctx, "blocklist:"+jti, "1", remaining).Result()
		if err != nil {
			return fmt.Errorf("%s: session blocklist: %w", caller, err)
		}
	}

	return nil
}

func (s *Service) emitTokenEvent(
	ctx context.Context,
	userID string,
	ttl time.Duration,
	aggregation events.AggregationType,
	eventType events.EventType,
	fn func(ctx context.Context, rawToken, hashToken string, expiresAt time.Time) (any, error),
) error {
	rawToken, hashToken, err := tokens.GenerateSecureToken()
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(ttl)

	payload, err := fn(ctx, rawToken, hashToken, expiresAt)
	if err != nil {
		return fmt.Errorf("emitTokenEvent: %w", err)
	}

	return s.outbox.Emit(ctx, aggregation, userID, eventType, payload)
}
