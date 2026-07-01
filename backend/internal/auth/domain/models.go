package authdomain

import (
	"learnflow_backend/internal/shared/validator"
	"time"
)

// UserRole represents the permission level of a user account.
type UserRole string

// UserStatus represents the current state of a user account.
type UserStatus string

// RevokeReason describes why a session was invalidated.
type RevokeReason string

// Role constants.
const (
	RoleAdmin    UserRole = "admin"
	RoleSubAdmin UserRole = "subadmin"
	RoleUser     UserRole = "user"
)

// Status constants.
const (
	StatusActive              UserStatus = "active"
	StatusBlocked             UserStatus = "blocked"
	StatusDeleted             UserStatus = "deleted"
	StatusPendingVerification UserStatus = "pending_verification"
)

// RevokeReason constants.
const (
	RevokeReasonLogout             RevokeReason = "logout"
	RevokeReasonPasswordChanged    RevokeReason = "password_changed"
	RevokeReasonPasswordReset      RevokeReason = "password_reset"
	RevokeReasonEmailChanged       RevokeReason = "email_changed"
	RevokeReasonAdmin              RevokeReason = "admin"
	RevokeReasonSuspiciousActivity RevokeReason = "suspicious_activity"
	RevokeReasonTokenExpired       RevokeReason = "token_expired"
)

// Valid reports whether r is a valid RevokeReason.
func (r RevokeReason) Valid() bool {
	switch r {
	case
		RevokeReasonLogout,
		RevokeReasonPasswordChanged,
		RevokeReasonPasswordReset,
		RevokeReasonEmailChanged,
		RevokeReasonAdmin,
		RevokeReasonSuspiciousActivity,
		RevokeReasonTokenExpired:
		return true
	}
	return false
}

// User represents an authenticated account.
type User struct {
	ID                string
	Email             string
	PasswordHash      string
	Role              UserRole
	Status            UserStatus
	EmailVerifiedAt   *time.Time
	LastLoginAt       *time.Time
	DeletedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PasswordChangedAt *time.Time
	EmailChangedAt    *time.Time
	FailedLoginCount  int
	LastFailedLoginAt *time.Time
	LoginLockedUntil  *time.Time
}

// UserSession represents an active refresh-token session.
type UserSession struct {
	ID                  string
	UserID              string
	RefreshHash         string
	UserAgent           *string
	IPAddress           *string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	RevokeReason        *RevokeReason
	RevokedByUserID     *string
	CreatedAt           time.Time
	FailedAttemptCount  int
	LastAttemptAt       *time.Time
	LockedUntil         *time.Time
	TokenVersion        int
	PreviousRefreshHash *string
	LastSeenAt          *time.Time
	LastSeenIP          *string
}

// TokenBase holds fields common to all single-use auth tokens.
type TokenBase struct {
	ID                  string
	UserID              string
	TokenHash           string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UsedAt              *time.Time
	InvalidatedAt       *time.Time
	InvalidatedByUserID *string
}

// EmailVerificationToken is a single-use token for confirming a user's email address.
type EmailVerificationToken struct {
	TokenBase
}

// PasswordResetToken is a single-use token for resetting a forgotten password.
type PasswordResetToken struct {
	TokenBase
}

// EmailChangeToken is a single-use token for confirming an email address change.
type EmailChangeToken struct {
	TokenBase
	NewEmail string
}

// AccountRecoveryToken is a single-use token for recovering a deleted account.
type AccountRecoveryToken struct {
	TokenBase
}

// AuthTokens holds the access and refresh tokens issued after successful authentication.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
}

// UserProfile holds public profile data for a user.
type UserProfile struct {
	UserID      string
	FirstName   string
	LastName    string
	PhoneNumber string
	Country     string
	City        string
	DateOfBirth *string
	Gender      string
	UILanguage  string
	AvatarURL   string
	Timezone    string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RegisterRequest carries credentials and profile data for new account creation.
type RegisterRequest struct {
	Email       string
	Password    string
	FirstName   string
	LastName    string
	PhoneNumber string
	Country     string
	City        string
	Gender      string
	DateOfBirth *string
	UILanguage  string
	AvatarURL   string
	Timezone    string
	Bio         string
}

// Validate checks that email and password meet format requirements.
func (r *RegisterRequest) Validate() error {
	if len(r.Email) < 3 || !validator.MatchesEmail(r.Email) {
		return ErrInvalidCredentialFormat
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return ErrInvalidCredentialFormat
	}

	return nil
}

// LoginRequest carries credentials and client metadata for session creation.
type LoginRequest struct {
	Email     string
	Password  string
	UserAgent string `json:"-"`
	IPAddress string `json:"-"`
}

// Validate checks that LoginRequest fields meet required criteria.
func (r *LoginRequest) Validate() error {
	if len(r.Email) < 3 || !validator.MatchesEmail(r.Email) {
		return ErrInvalidCredentialFormat
	}
	if len(r.Password) < 8 || len(r.Password) > 72 {
		return ErrInvalidCredentialFormat
	}
	if r.UserAgent == "" || len(r.UserAgent) > 2000 {
		return ErrInvalidCredentialFormat
	}

	return nil
}

// RefreshRequest carries a refresh token and client metadata for token rotation.
type RefreshRequest struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

// Validate checks that the refresh request fields are valid.
func (r *RefreshRequest) Validate() error {
	if r.RefreshToken == "" {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// LogoutRequest carries the refresh token to be revoked on logout.
type LogoutRequest struct {
	RefreshToken         string `json:"refresh_token"`
	JTI                  string
	AccessTokenExpiresAt time.Time
}

// Validate checks that the logout request fields are valid.
func (r *LogoutRequest) Validate() error {
	if r.RefreshToken == "" {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// VerifyEmailRequest carries the token submitted to confirm an email address.
type VerifyEmailRequest struct {
	Token string
}

// Validate checks that the token field is present.
func (r *VerifyEmailRequest) Validate() error {
	if r.Token == "" {
		return ErrInvalidCredentialFormat
	}

	return nil
}

// RequestPasswordResetRequest carries the email for which a reset link should be sent.
type RequestPasswordResetRequest struct {
	Email string
}

// Validate checks that the request password reset fields are valid.
func (r *RequestPasswordResetRequest) Validate() error {
	if len(r.Email) < 3 || !validator.MatchesEmail(r.Email) {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// ResetPasswordRequest carries the reset token and the new password.
type ResetPasswordRequest struct {
	Token       string
	NewPassword string
}

// Validate checks that the reset password fields are valid.
func (r *ResetPasswordRequest) Validate() error {
	if r.Token == "" {
		return ErrInvalidCredentialFormat
	}
	if len(r.NewPassword) < 8 || len(r.NewPassword) > 72 {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// ChangePasswordRequest carries the user ID, current password, and desired new password.
type ChangePasswordRequest struct {
	UserID               string
	OldPassword          string
	NewPassword          string
	IsAllSessionsLogout  bool
	JTI                  string
	AccessTokenExpiresAt time.Time
}

// Validate checks that the change password fields are valid.
func (r *ChangePasswordRequest) Validate() error {
	if r.OldPassword == "" {
		return ErrInvalidCredentialFormat
	}
	if len(r.NewPassword) < 8 || len(r.NewPassword) > 72 {
		return ErrInvalidCredentialFormat
	}

	if r.NewPassword == r.OldPassword {
		return ErrWrongPassword
	}
	return nil
}

// RequestEmailChangeRequest carries the user ID and the desired new email address.
type RequestEmailChangeRequest struct {
	UserID   string
	NewEmail string
}

// Validate checks that the request email change fields are valid.
func (r *RequestEmailChangeRequest) Validate() error {
	if len(r.NewEmail) < 3 || !validator.MatchesEmail(r.NewEmail) {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// EmailChangeRequest carries the token submitted to confirm an email address change.
type EmailChangeRequest struct {
	Token                string
	UserID               string
	IsAllSessionsLogout  bool
	JTI                  string
	AccessTokenExpiresAt time.Time
}

// Validate checks that the email change fields are valid.
func (r *EmailChangeRequest) Validate() error {
	if r.Token == "" {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// RecoverAccountRequest carries the recovery token submitted to restore a deleted account.
type RecoverAccountRequest struct {
	Token string
}

// Validate checks that the recover account fields are valid.
func (r *RecoverAccountRequest) Validate() error {
	if r.Token == "" {
		return ErrInvalidCredentialFormat
	}
	return nil
}

// RequestRecoverAccountRequest carries the email for which account recovery should be initiated.
type RequestRecoverAccountRequest struct {
	Email string
}

// Validate checks that the recover account request fields are valid.
func (r *RequestRecoverAccountRequest) Validate() error {
	if len(r.Email) < 3 || !validator.MatchesEmail(r.Email) {
		return ErrInvalidCredentialFormat
	}
	return nil
}
