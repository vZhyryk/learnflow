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

func scanUserProfile(row rowScanner) (*authdomain.UserProfile, error) {
	userProfile := &authdomain.UserProfile{}
	err := row.Scan(
		&userProfile.UserID,
		&userProfile.FirstName,
		&userProfile.LastName,
		&userProfile.PhoneNumber,
		&userProfile.Country,
		&userProfile.City,
		&userProfile.DateOfBirth,
		&userProfile.Gender,
		&userProfile.UILanguage,
		&userProfile.AvatarURL,
		&userProfile.Timezone,
		&userProfile.Bio,
		&userProfile.CreatedAt,
		&userProfile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return userProfile, nil
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

func scanToken(row rowScanner) (*authdomain.TokenBase, error) {
	token := &authdomain.TokenBase{}
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UsedAt,
		&token.InvalidatedAt,
		&token.InvalidatedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func scanEmailChangeToken(row rowScanner) (*authdomain.EmailChangeToken, error) {
	token := &authdomain.EmailChangeToken{}
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.NewEmail,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UsedAt,
		&token.InvalidatedAt,
		&token.InvalidatedByUserID,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}
