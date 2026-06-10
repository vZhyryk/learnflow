package authdomain

import (
	"context"
)

// Service defines all authentication use cases.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*AuthTokens, error)
	Logout(ctx context.Context, req LogoutRequest) error
	Register(ctx context.Context, req RegisterRequest) error
	Refresh(ctx context.Context, req RefreshRequest) (*AuthTokens, error)
	VerifyEmail(ctx context.Context, req VerifyEmailRequest) error
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	InitiatePasswordReset(ctx context.Context, req RequestPasswordResetRequest) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	InitiateEmailChange(ctx context.Context, req RequestEmailChangeRequest) error
	ChangeEmail(ctx context.Context, req EmailChangeRequest) error
	RecoverAccount(ctx context.Context, req RecoverAccountRequest) error
}

// SessionRepository defines persistence operations for user sessions.
type SessionRepository interface {
	CreateUserSession(ctx context.Context, session *UserSession) (*UserSession, error)
	GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error)
	RevokeUserSession(ctx context.Context, sessionID, revokedByUserID string, revokeReason RevokeReason) error
	RevokeAllUserSessions(ctx context.Context, userID, revokedByUserID string, revokeReason RevokeReason) error
	GetActiveSessionsByUserID(ctx context.Context, userID string) ([]*UserSession, error)
	UpdateSessionToken(ctx context.Context, sessionID, tokenHash, IPAddress string) error
	UpdateFailedLoginAttempts(ctx context.Context, sessionID, lockInterval string, loginCountLimit int) error
	GetSessionByPrevHash(ctx context.Context, prevRefreshToken string) (*UserSession, error)
}

// TokenRepository defines persistence operations for single-use auth tokens.
type TokenRepository interface {
	DeleteExpiredTokens(ctx context.Context) (int, error)
	CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) (*EmailVerificationToken, error)
	GetEmailVerificationToken(ctx context.Context, token string) (*EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) (*PasswordResetToken, error)
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error
	CreateEmailChangeToken(ctx context.Context, token *EmailChangeToken) (*EmailChangeToken, error)
	GetEmailChangeToken(ctx context.Context, token string) (*EmailChangeToken, error)
	MarkEmailChangeTokenUsed(ctx context.Context, tokenHash string) error
	CreateAccountRecoveryToken(ctx context.Context, token *AccountRecoveryToken) (*AccountRecoveryToken, error)
	GetAccountRecoveryToken(ctx context.Context, token string) (*AccountRecoveryToken, error)
	MarkAccountRecoveryTokenUsed(ctx context.Context, tokenHash string) error
}

// UserRepository defines persistence operations for User.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (string, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateStatus(ctx context.Context, userID string, status UserStatus) error
	UpdateRole(ctx context.Context, userID string, role UserRole) error
	UpdateLastLoginAt(ctx context.Context, userID string) error
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
	UpdateEmail(ctx context.Context, userID, newEmail string) error
	UpdateEmailVerifiedAt(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error
	IncrementFailedLogin(ctx context.Context, userID, lockInterval string, loginCountLimit int) error
	ResetFailedLogin(ctx context.Context, userID string) error
}
