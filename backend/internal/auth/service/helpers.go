package authservice

import (
	"context"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/shared/tokens"
	"time"
)

func (s *Service) revokeUserSessions(ctx context.Context, caller, jti string, accessTokenExpiresAt time.Time, fn func(ctx context.Context) error) error {
	if err := fn(ctx); err != nil {
		return fmt.Errorf("%s: revoke sessions: %w", caller, err)
	}

	// BlockToken stays inside the tx closure: a Redis failure rolls the DB change back too.
	remaining := time.Until(accessTokenExpiresAt)
	if remaining > 0 && jti != "" {
		if err := s.blocklist.BlockToken(ctx, jti, remaining); err != nil {
			return fmt.Errorf("%s: session blocklist: %w: %w", caller, authdomain.ErrBlocklistUnavailable, err)
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
		return err
	}

	return s.outbox.Emit(ctx, aggregation, userID, eventType, payload)
}
