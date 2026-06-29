package authservice

import (
	"context"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

// mockUserRepo implements authdomain.UserRepository via function fields.
// Only set the fields needed for each test case — unset fields panic with a clear message.
type mockUserRepo struct {
	getUserByEmail         func(ctx context.Context, email string) (*authdomain.User, error)
	createUser             func(ctx context.Context, user *authdomain.User) (string, error)
	createUserProfile      func(ctx context.Context, profile *authdomain.UserProfile) error
	getUserByID            func(ctx context.Context, userID string) (*authdomain.User, error)
	updateStatus           func(ctx context.Context, userID string, status authdomain.UserStatus) error
	updatePasswordHash     func(ctx context.Context, userID, hash string) error
	updateEmail            func(ctx context.Context, userID, email string) error
	updateEmailVerifiedAt  func(ctx context.Context, userID string) error
	incrementFailedLogin   func(ctx context.Context, userID, lockInterval string, limit int) error
	resetFailedLogin       func(ctx context.Context, userID string) error
	updateLastLoginAt      func(ctx context.Context, userID string) error
	getUserProfileByUserID func(ctx context.Context, userID string) (*authdomain.UserProfile, error)
	getDeletedUserByEmail  func(ctx context.Context, email string) (*authdomain.User, error)
	getDeletedUserByID     func(ctx context.Context, userID string) (*authdomain.User, error)
	restoreUser            func(ctx context.Context, userID string) error
	deleteUser             func(ctx context.Context, userID string) error
	updateRole             func(ctx context.Context, userID string, role authdomain.UserRole) error
	updateLastLoginAtRepo  func(ctx context.Context, userID string) error
}

// mockSessionRepo implements authdomain.SessionRepository
type mockSessionRepo struct {
	createUserSession            func(ctx context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error)
	getUserSessionByRefreshToken func(ctx context.Context, token string) (*authdomain.UserSession, error)
	revokeUserSession            func(ctx context.Context, sessionID, revokedBy string, reason authdomain.RevokeReason) error
	revokeAllUserSessions        func(ctx context.Context, userID string, revokedBy *string, reason authdomain.RevokeReason) error
	getSessionByPrevHash         func(ctx context.Context, prevToken string) (*authdomain.UserSession, error)
	updateSessionToken           func(ctx context.Context, sessionID, hash, ua, ip string) error
}

// mockTokenRepo implements authdomain.TokenRepository
type mockTokenRepo struct {
	createEmailVerificationToken   func(ctx context.Context, t *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error)
	getEmailVerificationToken      func(ctx context.Context, token string) (*authdomain.EmailVerificationToken, error)
	markEmailVerificationTokenUsed func(ctx context.Context, hash string) error
	createPasswordResetToken       func(ctx context.Context, t *authdomain.PasswordResetToken) (*authdomain.PasswordResetToken, error)
	getPasswordResetToken          func(ctx context.Context, token string) (*authdomain.PasswordResetToken, error)
	markPasswordResetTokenUsed     func(ctx context.Context, hash string) error
	createEmailChangeToken         func(ctx context.Context, t *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error)
	getEmailChangeToken            func(ctx context.Context, token string) (*authdomain.EmailChangeToken, error)
	markEmailChangeTokenUsed       func(ctx context.Context, hash string) error
	createAccountRecoveryToken     func(ctx context.Context, t *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error)
	getAccountRecoveryToken        func(ctx context.Context, token string) (*authdomain.AccountRecoveryToken, error)
	markAccountRecoveryTokenUsed   func(ctx context.Context, hash string) error
	deleteExpiredTokens            func(ctx context.Context) (int, error)
}

// mockTransactor calls fn(ctx) immediately — no real transaction.
type mockTransactor struct{}

func (m *mockTransactor) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// mockRedis implements RedisOps for JTI blocklist tests.
type mockRedis struct {
	setNX func(ctx context.Context, key string, value any, exp time.Duration) *redis.BoolCmd
}
