package authdomain

import "errors"

var (
	// ErrUserNotFound is returned when no user matches the lookup criteria.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when registering with an email already in use.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrAccountBlocked is returned when the account has been suspended.
	ErrAccountBlocked = errors.New("account is blocked")

	// ErrAccountDeleted is returned when the account has been soft-deleted.
	ErrAccountDeleted = errors.New("account is deleted")

	// ErrAccountLocked is returned when the account has exceeded login attempt limits.
	ErrAccountLocked = errors.New("account is temporarily locked")
	// ErrInvalidCredentials is returned on wrong email or password.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrEmailNotVerified is returned when the account email has not been confirmed.
	ErrEmailNotVerified = errors.New("email not verified")

	// ErrInvalidToken is returned when a single-use token cannot be found or has an invalid signature.
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when a single-use token has passed its expiry time.
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenUsed is returned when a single-use token has already been consumed.
	ErrTokenUsed = errors.New("token already used")

	// ErrInvalidCredentialFormat is returned when email or password does not meet format requirements.
	ErrInvalidCredentialFormat = errors.New("invalid credential format")

	// ErrSessionNotFound is returned when no session matches the lookup criteria.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionRevoked is returned when the session was explicitly revoked.
	ErrSessionRevoked = errors.New("session has been revoked")
	// ErrSessionExpired is returned when the session has passed its expiry time.
	ErrSessionExpired = errors.New("session expired")

	// ErrWrongPassword is returned when the supplied current password does not match.
	ErrWrongPassword = errors.New("wrong current password")
	// ErrSamePassword is returned when the new password is identical to the current one.
	ErrSamePassword = errors.New("new password must differ from current")

	// ErrEmailAlreadyInUse is returned when the requested new email is taken by another account.
	ErrEmailAlreadyInUse = errors.New("email is already in use")
)
