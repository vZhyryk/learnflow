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
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *authdomain.User) (string, error) {
	if m.createUser == nil {
		panic("mockUserRepo.createUser not set")
	}
	return m.createUser(ctx, user)
}

func (m *mockUserRepo) CreateUserProfile(ctx context.Context, profile *authdomain.UserProfile) error {
	if m.createUserProfile == nil {
		panic("mockUserRepo.createUserProfile not set")
	}
	return m.createUserProfile(ctx, profile)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
	if m.getUserByID == nil {
		panic("mockUserRepo.getUserByID not set")
	}
	return m.getUserByID(ctx, userID)
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*authdomain.User, error) {
	if m.getUserByEmail == nil {
		panic("mockUserRepo.getUserByEmail not set")
	}
	return m.getUserByEmail(ctx, email)
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, userID string, status authdomain.UserStatus) error {
	if m.updateStatus == nil {
		panic("mockUserRepo.updateStatus not set")
	}
	return m.updateStatus(ctx, userID, status)
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, userID string, role authdomain.UserRole) error {
	if m.updateRole == nil {
		panic("mockUserRepo.updateRole not set")
	}
	return m.updateRole(ctx, userID, role)
}

func (m *mockUserRepo) UpdateLastLoginAt(ctx context.Context, userID string) error {
	if m.updateLastLoginAt == nil {
		panic("mockUserRepo.updateLastLoginAt not set")
	}
	return m.updateLastLoginAt(ctx, userID)
}

func (m *mockUserRepo) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	if m.updatePasswordHash == nil {
		panic("mockUserRepo.updatePasswordHash not set")
	}
	return m.updatePasswordHash(ctx, userID, passwordHash)
}

func (m *mockUserRepo) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	if m.updateEmail == nil {
		panic("mockUserRepo.updateEmail not set")
	}
	return m.updateEmail(ctx, userID, newEmail)
}

func (m *mockUserRepo) UpdateEmailVerifiedAt(ctx context.Context, userID string) error {
	if m.updateEmailVerifiedAt == nil {
		panic("mockUserRepo.updateEmailVerifiedAt not set")
	}
	return m.updateEmailVerifiedAt(ctx, userID)
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, userID string) error {
	if m.deleteUser == nil {
		panic("mockUserRepo.deleteUser not set")
	}
	return m.deleteUser(ctx, userID)
}

func (m *mockUserRepo) IncrementFailedLogin(ctx context.Context, userID, lockInterval string, loginCountLimit int) error {
	if m.incrementFailedLogin == nil {
		panic("mockUserRepo.incrementFailedLogin not set")
	}
	return m.incrementFailedLogin(ctx, userID, lockInterval, loginCountLimit)
}

func (m *mockUserRepo) ResetFailedLogin(ctx context.Context, userID string) error {
	if m.resetFailedLogin == nil {
		panic("mockUserRepo.resetFailedLogin not set")
	}
	return m.resetFailedLogin(ctx, userID)
}

func (m *mockUserRepo) GetUserProfileByUserID(ctx context.Context, userID string) (*authdomain.UserProfile, error) {
	if m.getUserProfileByUserID == nil {
		panic("mockUserRepo.getUserProfileByUserID not set")
	}
	return m.getUserProfileByUserID(ctx, userID)
}

func (m *mockUserRepo) GetDeletedUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
	if m.getDeletedUserByID == nil {
		panic("mockUserRepo.getDeletedUserByID not set")
	}
	return m.getDeletedUserByID(ctx, userID)
}

func (m *mockUserRepo) RestoreUser(ctx context.Context, userID string) error {
	if m.restoreUser == nil {
		panic("mockUserRepo.restoreUser not set")
	}
	return m.restoreUser(ctx, userID)
}

func (m *mockUserRepo) GetDeletedUserByEmail(ctx context.Context, email string) (*authdomain.User, error) {
	if m.getDeletedUserByEmail == nil {
		panic("mockUserRepo.getDeletedUserByEmail not set")
	}
	return m.getDeletedUserByEmail(ctx, email)
}

// mockSessionRepo implements authdomain.SessionRepository via function fields.
// Only set the fields needed for each test case — unset fields panic with a clear message.
type mockSessionRepo struct {
	createUserSession            func(ctx context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error)
	getUserSessionByRefreshToken func(ctx context.Context, token string) (*authdomain.UserSession, error)
	revokeUserSession            func(ctx context.Context, sessionID, revokedBy string, reason authdomain.RevokeReason) error
	revokeAllUserSessions        func(ctx context.Context, userID string, revokedBy *string, reason authdomain.RevokeReason) error
	getActiveSessionsByUserID    func(ctx context.Context, userID string) ([]*authdomain.UserSession, error)
	getSessionByPrevHash         func(ctx context.Context, prevToken string) (*authdomain.UserSession, error)
	updateSessionToken           func(ctx context.Context, sessionID, hash, ua, ip string) error
	updateFailedLoginAttempts    func(ctx context.Context, sessionID, lockInterval string, loginCountLimit int) error
}

func (m *mockSessionRepo) CreateUserSession(ctx context.Context, s *authdomain.UserSession) (*authdomain.UserSession, error) {
	if m.createUserSession == nil {
		panic("mockSessionRepo.createUserSession not set")
	}
	return m.createUserSession(ctx, s)
}

func (m *mockSessionRepo) GetUserSessionByRefreshToken(ctx context.Context, token string) (*authdomain.UserSession, error) {
	if m.getUserSessionByRefreshToken == nil {
		panic("mockSessionRepo.getUserSessionByRefreshToken not set")
	}
	return m.getUserSessionByRefreshToken(ctx, token)
}

func (m *mockSessionRepo) RevokeUserSession(ctx context.Context, sessionID, revokedByUserID string, revokeReason authdomain.RevokeReason) error {
	if m.revokeUserSession == nil {
		panic("mockSessionRepo.revokeUserSession not set")
	}
	return m.revokeUserSession(ctx, sessionID, revokedByUserID, revokeReason)
}

func (m *mockSessionRepo) RevokeAllUserSessions(ctx context.Context, userID string, revokedByUserID *string, revokeReason authdomain.RevokeReason) error {
	if m.revokeAllUserSessions == nil {
		panic("mockSessionRepo.revokeAllUserSessions not set")
	}
	return m.revokeAllUserSessions(ctx, userID, revokedByUserID, revokeReason)
}

func (m *mockSessionRepo) GetActiveSessionsByUserID(ctx context.Context, userID string) ([]*authdomain.UserSession, error) {
	if m.getActiveSessionsByUserID == nil {
		panic("mockSessionRepo.getActiveSessionsByUserID not set")
	}
	return m.getActiveSessionsByUserID(ctx, userID)
}

func (m *mockSessionRepo) UpdateSessionToken(ctx context.Context, sessionID, tokenHash, userAgent, ipAddress string) error {
	if m.updateSessionToken == nil {
		panic("mockSessionRepo.updateSessionToken not set")
	}
	return m.updateSessionToken(ctx, sessionID, tokenHash, userAgent, ipAddress)
}

func (m *mockSessionRepo) UpdateFailedLoginAttempts(ctx context.Context, sessionID, lockInterval string, loginCountLimit int) error {
	if m.updateFailedLoginAttempts == nil {
		panic("mockSessionRepo.updateFailedLoginAttempts not set")
	}
	return m.updateFailedLoginAttempts(ctx, sessionID, lockInterval, loginCountLimit)
}

func (m *mockSessionRepo) GetSessionByPrevHash(ctx context.Context, prevRefreshToken string) (*authdomain.UserSession, error) {
	if m.getSessionByPrevHash == nil {
		panic("mockSessionRepo.getSessionByPrevHash not set")
	}
	return m.getSessionByPrevHash(ctx, prevRefreshToken)
}

// mockTokenRepo implements authdomain.TokenRepository via function fields.
// Only set the fields needed for each test case — unset fields panic with a clear message.
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

func (m *mockTokenRepo) DeleteExpiredTokens(ctx context.Context) (int, error) {
	if m.deleteExpiredTokens == nil {
		panic("mockTokenRepo.deleteExpiredTokens not set")
	}
	return m.deleteExpiredTokens(ctx)
}

func (m *mockTokenRepo) CreateEmailVerificationToken(ctx context.Context, token *authdomain.EmailVerificationToken) (*authdomain.EmailVerificationToken, error) {
	if m.createEmailVerificationToken == nil {
		panic("mockTokenRepo.createEmailVerificationToken not set")
	}
	return m.createEmailVerificationToken(ctx, token)
}

func (m *mockTokenRepo) GetEmailVerificationToken(ctx context.Context, token string) (*authdomain.EmailVerificationToken, error) {
	if m.getEmailVerificationToken == nil {
		panic("mockTokenRepo.getEmailVerificationToken not set")
	}
	return m.getEmailVerificationToken(ctx, token)
}

func (m *mockTokenRepo) MarkEmailVerificationTokenUsed(ctx context.Context, tokenHash string) error {
	if m.markEmailVerificationTokenUsed == nil {
		panic("mockTokenRepo.markEmailVerificationTokenUsed not set")
	}
	return m.markEmailVerificationTokenUsed(ctx, tokenHash)
}

func (m *mockTokenRepo) CreatePasswordResetToken(ctx context.Context, token *authdomain.PasswordResetToken) (*authdomain.PasswordResetToken, error) {
	if m.createPasswordResetToken == nil {
		panic("mockTokenRepo.createPasswordResetToken not set")
	}
	return m.createPasswordResetToken(ctx, token)
}

func (m *mockTokenRepo) GetPasswordResetToken(ctx context.Context, token string) (*authdomain.PasswordResetToken, error) {
	if m.getPasswordResetToken == nil {
		panic("mockTokenRepo.getPasswordResetToken not set")
	}
	return m.getPasswordResetToken(ctx, token)
}

func (m *mockTokenRepo) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	if m.markPasswordResetTokenUsed == nil {
		panic("mockTokenRepo.markPasswordResetTokenUsed not set")
	}
	return m.markPasswordResetTokenUsed(ctx, tokenHash)
}

func (m *mockTokenRepo) CreateEmailChangeToken(ctx context.Context, token *authdomain.EmailChangeToken) (*authdomain.EmailChangeToken, error) {
	if m.createEmailChangeToken == nil {
		panic("mockTokenRepo.createEmailChangeToken not set")
	}
	return m.createEmailChangeToken(ctx, token)
}

func (m *mockTokenRepo) GetEmailChangeToken(ctx context.Context, token string) (*authdomain.EmailChangeToken, error) {
	if m.getEmailChangeToken == nil {
		panic("mockTokenRepo.getEmailChangeToken not set")
	}
	return m.getEmailChangeToken(ctx, token)
}

func (m *mockTokenRepo) MarkEmailChangeTokenUsed(ctx context.Context, tokenHash string) error {
	if m.markEmailChangeTokenUsed == nil {
		panic("mockTokenRepo.markEmailChangeTokenUsed not set")
	}
	return m.markEmailChangeTokenUsed(ctx, tokenHash)
}

func (m *mockTokenRepo) CreateAccountRecoveryToken(ctx context.Context, token *authdomain.AccountRecoveryToken) (*authdomain.AccountRecoveryToken, error) {
	if m.createAccountRecoveryToken == nil {
		panic("mockTokenRepo.createAccountRecoveryToken not set")
	}
	return m.createAccountRecoveryToken(ctx, token)
}

func (m *mockTokenRepo) GetAccountRecoveryToken(ctx context.Context, token string) (*authdomain.AccountRecoveryToken, error) {
	if m.getAccountRecoveryToken == nil {
		panic("mockTokenRepo.getAccountRecoveryToken not set")
	}
	return m.getAccountRecoveryToken(ctx, token)
}

func (m *mockTokenRepo) MarkAccountRecoveryTokenUsed(ctx context.Context, tokenHash string) error {
	if m.markAccountRecoveryTokenUsed == nil {
		panic("mockTokenRepo.markAccountRecoveryTokenUsed not set")
	}
	return m.markAccountRecoveryTokenUsed(ctx, tokenHash)
}

// mockRedis implements RedisOps for JTI blocklist tests.
type mockRedis struct {
	setNX func(ctx context.Context, key string, value any, exp time.Duration) *redis.BoolCmd
}

func (m *mockRedis) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	if m.setNX == nil {
		panic("mockRedis.setNX not set")
	}
	return m.setNX(ctx, key, value, expiration)
}

// mockRedisSetNXError returns a mockRedis whose SetNX always fails with err.
func mockRedisSetNXError(err error) *mockRedis {
	return &mockRedis{
		setNX: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.BoolCmd {
			return redis.NewBoolResult(false, err)
		},
	}
}
