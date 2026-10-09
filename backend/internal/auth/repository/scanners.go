package authrepository

import (
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/convert"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// lockNotAvailableCode is the SQLSTATE PostgreSQL returns when FOR UPDATE NOWAIT hits a row locked by another transaction.
const lockNotAvailableCode = "55P03"

// mapLockNotAvailable turns a lock_not_available error into ErrRequestInProgress and passes any other error through.
func mapLockNotAvailable(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == lockNotAvailableCode {
		return authdomain.ErrRequestInProgress
	}

	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

// dobLayout matches authdomain's DateOfBirth string format and the
// validator.IsValidDateOfBirth parse layout.
const dobLayout = "2006-01-02"

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
	var dob pgtype.Date
	err := row.Scan(
		&userProfile.UserID,
		&userProfile.FirstName,
		&userProfile.LastName,
		&userProfile.PhoneNumber,
		&userProfile.Country,
		&userProfile.City,
		&dob,
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
	userProfile.DateOfBirth = convert.FormatNullableDate(dob, dobLayout)
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
		return nil, mapLockNotAvailable(err)
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
		return nil, mapLockNotAvailable(err)
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
		return nil, mapLockNotAvailable(err)
	}
	return token, nil
}
