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

// logoutEverywhere revokes all of the user's sessions through revokeSessions and then every access token issued so far.
// The revocation mark also covers the caller's own token, so no per-token (jti) blocklisting is needed on this path.
func (s *Service) logoutEverywhere(ctx context.Context, caller, userID string, revokeSessions func(ctx context.Context) error) error {
	if err := revokeSessions(ctx); err != nil {
		return fmt.Errorf("%s: revoke sessions: %w", caller, err)
	}

	return s.revokeAllUserTokens(ctx, caller, userID)
}

// revokeAllUserTokens invalidates every access token issued to userID so far (tokens issued later keep working).
// Redis cannot roll back, so callers run it as the last fallible step; a failure is reported as ErrBlocklistUnavailable.
func (s *Service) revokeAllUserTokens(ctx context.Context, caller, userID string) error {
	if err := s.blocklist.RevokeUserTokens(ctx, userID, tokens.BlockMarkTTL); err != nil {
		return fmt.Errorf("%s: revoke tokens: %w: %w", caller, authdomain.ErrBlocklistUnavailable, err)
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
		return fmt.Errorf("emit_token_event %s: generate token: %w", eventType, err)
	}

	expiresAt := time.Now().UTC().Add(ttl)

	payload, err := fn(ctx, rawToken, hashToken, expiresAt)
	if err != nil {
		return fmt.Errorf("emit_token_event %s: build payload: %w", eventType, err)
	}

	if err = s.outbox.Emit(ctx, aggregation, userID, eventType, payload); err != nil {
		return fmt.Errorf("emit_token_event %s: emit: %w", eventType, err)
	}

	return nil
}
