package authrepository

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
)

func (rep *Repository) CreateUserSession(ctx context.Context, session *authdomain.UserSession) error {
	return nil
}
func (rep *Repository) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*authdomain.UserSession, error) {
	return nil, nil
}
func (rep *Repository) RevokeUserSession(ctx context.Context, sessionID string, revokeReason authdomain.RevokeReason) error {
	return nil
}
func (rep *Repository) RevokeAllUserSessions(ctx context.Context, userID string, revokeReason authdomain.RevokeReason) error {
	return nil
}
func (rep *Repository) GetActiveSessionsByUserID(ctx context.Context, userID string) ([]*authdomain.UserSession, error) {
	return nil, nil
}
func (rep *Repository) UpdateSessionToken(ctx context.Context, sessionID, tokenHash string) error {
	return nil
}
func (rep *Repository) UpdateFailedLoginAttempts(ctx context.Context, sessionID string) error {
	return nil
}
func (rep *Repository) LockUserSession(ctx context.Context, sessionID string) error {
	return nil
}
