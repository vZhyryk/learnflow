package authrepository

import authdomain "learnflow_backend/internal/auth/domain"

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*authdomain.User, error) {
	user := &authdomain.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.PasswordChangedAt,
		&user.EmailChangedAt,
		&user.FailedLoginCount,
		&user.LastFailedLoginAt,
		&user.LoginLockedUntil,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func scanUserSession(row rowScanner) (*authdomain.UserSession, error) {
	session := &authdomain.UserSession{}
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.RevokeReason,
		&session.RevokedByUserID,
		&session.CreatedAt,
		&session.FailedAttemptCount,
		&session.LastAttemptAt,
		&session.LockedUntil,
		&session.TokenVersion,
		&session.PreviousRefreshHash,
		&session.LastSeenAt,
		&session.LastSeenIP,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}
