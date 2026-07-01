package authrepository

import (
	"context"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockQueryRunner implements db.QueryRunner via function fields.
type mockQueryRunner struct {
	queryRowFn  func(ctx context.Context, sql string, args ...any) pgx.Row
	queryRowsFn func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	execFn      func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockQueryRunner) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFn(ctx, sql, args...)
}

func (m *mockQueryRunner) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.execFn(ctx, sql, args...)
}

func (m *mockQueryRunner) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.queryRowsFn(ctx, sql, args...)
}

type fakeRow struct {
	scanFn func(dest ...any) error
}

func (r *fakeRow) Scan(dest ...any) error { return r.scanFn(dest...) }

type mockRows struct {
	rows []*fakeRow
}

func (r *mockRows) Next() bool {
	return len(r.rows) > 0
}

func (r *mockRows) Scan(dest ...any) error {
	err := r.rows[0].Scan(dest...)
	r.rows = r.rows[1:]
	return err
}

func (r *mockRows) Close()                                       {}
func (r *mockRows) Err() error                                   { return nil }
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

func newTestRepo(runner *mockQueryRunner) *Repository {
	return &Repository{db: runner}
}

// castStr safely type-asserts a scan destination to *string, panicking with context on failure.
func castStr(v any, idx int) *string {
	s, ok := v.(*string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *string, got %T", idx, v))
	}
	return s
}

// castPtrStr safely type-asserts a scan destination to **string.
func castPtrStr(v any, idx int) **string {
	s, ok := v.(**string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected **string, got %T", idx, v))
	}
	return s
}

// castTime safely type-asserts a scan destination to *time.Time.
func castTime(v any, idx int) *time.Time {
	s, ok := v.(*time.Time)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *time.Time, got %T", idx, v))
	}
	return s
}

// castPtrRevokeReason safely type-asserts a scan destination to **authdomain.RevokeReason.
func castPtrRevokeReason(v any, idx int) **authdomain.RevokeReason {
	s, ok := v.(**authdomain.RevokeReason)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *time.Time, got %T", idx, v))
	}
	return s
}

func castPtrTime(v any, idx int) **time.Time {
	s, ok := v.(**time.Time)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected **time.Time, got %T", idx, v))
	}
	return s
}

func castInt(v any, idx int) *int {
	s, ok := v.(*int)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *int, got %T", idx, v))
	}
	return s
}

func fakeProfile(now time.Time) *authdomain.UserProfile {
	return &authdomain.UserProfile{
		UserID:      "user-123",
		FirstName:   "John",
		LastName:    "Doe",
		PhoneNumber: "+380991234567",
		Country:     "UA",
		City:        "Kyiv",
		DateOfBirth: nil,
		Gender:      "male",
		UILanguage:  "uk",
		AvatarURL:   "",
		Timezone:    "Europe/Kiev",
		Bio:         "bio text",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func fakeScanProfile(now time.Time) func(dest ...any) error {
	p := fakeProfile(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = p.UserID
		*castStr(dest[1], 1) = p.FirstName
		*castStr(dest[2], 2) = p.LastName
		*castStr(dest[3], 3) = p.PhoneNumber
		*castStr(dest[4], 4) = p.Country
		*castStr(dest[5], 5) = p.City
		*castPtrStr(dest[6], 6) = p.DateOfBirth
		*castStr(dest[7], 7) = p.Gender
		*castStr(dest[8], 8) = p.UILanguage
		*castStr(dest[9], 9) = p.AvatarURL
		*castStr(dest[10], 10) = p.Timezone
		*castStr(dest[11], 11) = p.Bio
		*castTime(dest[12], 12) = p.CreatedAt
		*castTime(dest[13], 13) = p.UpdatedAt
		return nil
	}
}

func castUserRole(v any, idx int) *authdomain.UserRole {
	s, ok := v.(*authdomain.UserRole)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *UserRole, got %T", idx, v))
	}
	return s
}

func castUserStatus(v any, idx int) *authdomain.UserStatus {
	s, ok := v.(*authdomain.UserStatus)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *UserStatus, got %T", idx, v))
	}
	return s
}

func fakeUser(now time.Time) *authdomain.User {
	return &authdomain.User{
		ID:                "user-123",
		Email:             "john@gmail.com",
		PasswordHash:      "some_password_hash",
		Role:              authdomain.RoleAdmin,
		Status:            authdomain.StatusActive,
		EmailVerifiedAt:   nil,
		LastLoginAt:       nil,
		DeletedAt:         nil,
		CreatedAt:         now,
		UpdatedAt:         now,
		PasswordChangedAt: nil,
		EmailChangedAt:    nil,
		FailedLoginCount:  0,
		LastFailedLoginAt: nil,
		LoginLockedUntil:  nil,
	}
}

func fakeScanUser(now time.Time) func(dest ...any) error {
	u := fakeUser(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = u.ID
		*castStr(dest[1], 1) = u.Email
		*castStr(dest[2], 2) = u.PasswordHash
		*castUserRole(dest[3], 3) = u.Role
		*castUserStatus(dest[4], 4) = u.Status
		*castPtrTime(dest[5], 5) = u.EmailVerifiedAt
		*castPtrTime(dest[6], 6) = u.LastLoginAt
		*castPtrTime(dest[7], 7) = u.DeletedAt
		*castTime(dest[8], 8) = u.CreatedAt
		*castTime(dest[9], 9) = u.UpdatedAt
		*castPtrTime(dest[10], 10) = u.PasswordChangedAt
		*castPtrTime(dest[11], 11) = u.EmailChangedAt
		*castInt(dest[12], 12) = u.FailedLoginCount
		*castPtrTime(dest[13], 13) = u.LastFailedLoginAt
		*castPtrTime(dest[14], 14) = u.LoginLockedUntil
		return nil
	}
}

func fakeUserSession(now time.Time) *authdomain.UserSession {
	return &authdomain.UserSession{
		ID:                  "session-123",
		UserID:              "user-123",
		RefreshHash:         "some_refresh_hash",
		UserAgent:           &[]string{"user-agent-string"}[0],
		IPAddress:           &[]string{"ip-address"}[0],
		ExpiresAt:           now,
		RevokedAt:           &now,
		RevokeReason:        &[]authdomain.RevokeReason{authdomain.RevokeReasonLogout}[0],
		RevokedByUserID:     &[]string{"user-456"}[0],
		CreatedAt:           now,
		FailedAttemptCount:  0,
		LastAttemptAt:       &now,
		LockedUntil:         &now,
		TokenVersion:        0,
		PreviousRefreshHash: &[]string{"previous_refresh_hash"}[0],
		LastSeenAt:          &now,
		LastSeenIP:          &[]string{"last_ip_address"}[0],
	}
}

func fakeScanUserSession(now time.Time) func(dest ...any) error {
	userSession := fakeUserSession(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = userSession.ID
		*castStr(dest[1], 1) = userSession.UserID
		*castStr(dest[2], 2) = userSession.RefreshHash
		*castPtrStr(dest[3], 3) = userSession.UserAgent
		*castPtrStr(dest[4], 4) = userSession.IPAddress
		*castTime(dest[5], 5) = userSession.ExpiresAt
		*castPtrTime(dest[6], 6) = userSession.RevokedAt
		*castPtrRevokeReason(dest[7], 7) = userSession.RevokeReason
		*castPtrStr(dest[8], 8) = userSession.RevokedByUserID
		*castTime(dest[9], 9) = userSession.CreatedAt
		*castInt(dest[10], 10) = userSession.TokenVersion
		*castPtrTime(dest[11], 11) = userSession.LastAttemptAt
		*castPtrTime(dest[12], 12) = userSession.LockedUntil
		*castInt(dest[13], 13) = userSession.FailedAttemptCount
		*castPtrStr(dest[14], 14) = userSession.PreviousRefreshHash
		*castPtrTime(dest[15], 15) = userSession.LastSeenAt
		*castPtrStr(dest[16], 16) = userSession.LastSeenIP
		return nil
	}
}

func fakeToken(now time.Time) *authdomain.TokenBase {
	return &authdomain.TokenBase{
		ID:                  "session_123",
		UserID:              "user_123",
		TokenHash:           "some_password_hash",
		ExpiresAt:           now,
		CreatedAt:           now,
		UsedAt:              nil,
		InvalidatedAt:       nil,
		InvalidatedByUserID: nil,
	}
}

func fakeScanToken(now time.Time) func(dest ...any) error {
	tb := fakeToken(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = tb.ID
		*castStr(dest[1], 1) = tb.UserID
		*castStr(dest[2], 2) = tb.TokenHash
		*castTime(dest[3], 3) = tb.ExpiresAt
		*castTime(dest[4], 4) = tb.CreatedAt
		*castPtrTime(dest[5], 5) = tb.UsedAt
		*castPtrTime(dest[6], 6) = tb.InvalidatedAt
		*castPtrStr(dest[7], 7) = tb.InvalidatedByUserID
		return nil
	}
}

func fakeEmailChangeToken(now time.Time) *authdomain.EmailChangeToken {
	return &authdomain.EmailChangeToken{
		TokenBase: *fakeToken(now),
		NewEmail:  "john@gmail.com",
	}
}

func fakeScanEmailChangeToken(now time.Time) func(dest ...any) error {
	t := fakeEmailChangeToken(now)
	return func(dest ...any) error {
		*castStr(dest[0], 0) = t.ID
		*castStr(dest[1], 1) = t.UserID
		*castStr(dest[2], 2) = t.NewEmail
		*castStr(dest[3], 3) = t.TokenHash
		*castTime(dest[4], 4) = t.ExpiresAt
		*castTime(dest[5], 5) = t.CreatedAt
		*castPtrTime(dest[6], 6) = t.UsedAt
		*castPtrTime(dest[7], 7) = t.InvalidatedAt
		*castPtrStr(dest[8], 8) = t.InvalidatedByUserID
		return nil
	}
}
