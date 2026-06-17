package authrepository

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"

	"github.com/jackc/pgx/v5"
)

// CreateUserSession persists a new user session and returns it with DB-generated fields.
func (rep *Repository) CreateUserSession(ctx context.Context, session *authdomain.UserSession) (*authdomain.UserSession, error) {
	session, err := scanUserSession(rep.queryRunner(ctx).QueryRow(ctx, createUserSessionSQL, session.UserID, session.RefreshHash, session.UserAgent, session.IPAddress, session.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("repository.CreateUserSession: %w", err)
	}

	return session, nil
}

// GetUserSessionByRefreshToken retrieves an active session by its refresh token hash.
func (rep *Repository) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*authdomain.UserSession, error) {
	session, err := scanUserSession(rep.queryRunner(ctx).QueryRow(ctx, getSessionByTokenSQL, refreshToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetUserSessionByRefreshToken: %w", err)
	}

	return session, nil
}

// GetSessionByPrevHash retrieves a session by its previous refresh token hash (rotation detection).
func (rep *Repository) GetSessionByPrevHash(ctx context.Context, prevRefreshToken string) (*authdomain.UserSession, error) {
	session, err := scanUserSession(rep.queryRunner(ctx).QueryRow(ctx, getSessionByPrevHashSQL, prevRefreshToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, authdomain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetSessionByPrevHash: %w", err)
	}

	return session, nil
}

// RevokeUserSession marks a single session as revoked with the given reason.
func (rep *Repository) RevokeUserSession(ctx context.Context, sessionID, revokedByUserID string, revokeReason authdomain.RevokeReason) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, revokeUserSessionSQL, revokeReason, revokedByUserID, sessionID)
	if err != nil {
		return fmt.Errorf("repository.RevokeUserSession: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrSessionNotFound
	}
	return nil
}

// RevokeAllUserSessions revokes all active sessions for a user with the given reason.
func (rep *Repository) RevokeAllUserSessions(ctx context.Context, userID, revokedByUserID string, revokeReason authdomain.RevokeReason) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, revokeAllUserSessionsSQL, revokeReason, revokedByUserID, userID)
	if err != nil {
		return fmt.Errorf("repository.RevokeAllUserSessions: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrSessionNotFound
	}
	return nil
}

// GetActiveSessionsByUserID returns all non-revoked sessions for the given user.
func (rep *Repository) GetActiveSessionsByUserID(ctx context.Context, userID string) ([]*authdomain.UserSession, error) {
	rows, err := rep.queryRunner(ctx).Query(ctx, getActiveUserSessionSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetActiveSessionsByUserID: %w", err)
	}

	defer rows.Close()

	var sessions []*authdomain.UserSession
	for rows.Next() {
		session, err := scanUserSession(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.GetActiveSessionsByUserID scan: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.GetActiveSessionsByUserID rows: %w", err)
	}

	return sessions, nil
}

// UpdateSessionToken replaces the refresh token hash for a session (token rotation).
func (rep *Repository) UpdateSessionToken(ctx context.Context, sessionID, tokenHash, ipAddress string) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateSessionTokenSQL, sessionID, tokenHash, ipAddress)
	if err != nil {
		return fmt.Errorf("repository.UpdateSessionToken: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrSessionNotFound
	}
	return nil
}

// UpdateFailedLoginAttempts increments the failed attempt counter and locks the session when the limit is reached.
func (rep *Repository) UpdateFailedLoginAttempts(ctx context.Context, sessionID, lockInterval string, loginCountLimit int) error {
	tag, err := rep.queryRunner(ctx).Exec(ctx, updateFailedLoginAttemptsSQL, sessionID, loginCountLimit, lockInterval)
	if err != nil {
		return fmt.Errorf("repository.UpdateFailedLoginAttempts: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return authdomain.ErrSessionNotFound
	}
	return nil
}
