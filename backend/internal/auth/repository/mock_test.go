package authrepository

import (
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	. "github.com/smartystreets/goconvey/convey"
)

func assertUnexpectedDBError(err error, substr string) {
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, substr)
}

func newTestRepo(runner *testutil.MockQueryRunner) *Repository {
	return &Repository{db: runner}
}

func castPtrRevokeReason(v any, idx int) **authdomain.RevokeReason {
	s, ok := v.(**authdomain.RevokeReason)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected **authdomain.RevokeReason, got %T", idx, v))
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

func fakeProfile(now time.Time) *authdomain.UserProfile {
	firstName, lastName, phoneNumber := "John", "Doe", "+380991234567"
	country, city, gender := "UA", "Kyiv", "male"
	timezone, bio := "Europe/Kiev", "bio text"
	avatarURL := ""
	return &authdomain.UserProfile{
		UserID:      "user-123",
		FirstName:   &firstName,
		LastName:    &lastName,
		PhoneNumber: &phoneNumber,
		Country:     &country,
		City:        &city,
		DateOfBirth: nil,
		Gender:      &gender,
		UILanguage:  "uk",
		AvatarURL:   &avatarURL,
		Timezone:    &timezone,
		Bio:         &bio,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func fakeScanProfile(now time.Time) func(dest ...any) error {
	p := fakeProfile(now)
	return func(dest ...any) error {
		*testutil.CastStr(dest[0], 0) = p.UserID
		*testutil.CastPtrStr(dest[1], 1) = p.FirstName
		*testutil.CastPtrStr(dest[2], 2) = p.LastName
		*testutil.CastPtrStr(dest[3], 3) = p.PhoneNumber
		*testutil.CastPtrStr(dest[4], 4) = p.Country
		*testutil.CastPtrStr(dest[5], 5) = p.City
		*testutil.CastPgtypeDate(dest[6], 6) = pgtype.Date{}
		*testutil.CastPtrStr(dest[7], 7) = p.Gender
		*testutil.CastStr(dest[8], 8) = p.UILanguage
		*testutil.CastPtrStr(dest[9], 9) = p.AvatarURL
		*testutil.CastPtrStr(dest[10], 10) = p.Timezone
		*testutil.CastPtrStr(dest[11], 11) = p.Bio
		*testutil.CastTime(dest[12], 12) = p.CreatedAt
		*testutil.CastTime(dest[13], 13) = p.UpdatedAt
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
		*testutil.CastStr(dest[0], 0) = u.ID
		*testutil.CastStr(dest[1], 1) = u.Email
		*testutil.CastStr(dest[2], 2) = u.PasswordHash
		*castUserRole(dest[3], 3) = u.Role
		*castUserStatus(dest[4], 4) = u.Status
		*castPtrTime(dest[5], 5) = u.EmailVerifiedAt
		*castPtrTime(dest[6], 6) = u.LastLoginAt
		*castPtrTime(dest[7], 7) = u.DeletedAt
		*testutil.CastTime(dest[8], 8) = u.CreatedAt
		*testutil.CastTime(dest[9], 9) = u.UpdatedAt
		*castPtrTime(dest[10], 10) = u.PasswordChangedAt
		*castPtrTime(dest[11], 11) = u.EmailChangedAt
		*testutil.CastInt(dest[12], 12) = u.FailedLoginCount
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
		*testutil.CastStr(dest[0], 0) = userSession.ID
		*testutil.CastStr(dest[1], 1) = userSession.UserID
		*testutil.CastStr(dest[2], 2) = userSession.RefreshHash
		*testutil.CastPtrStr(dest[3], 3) = userSession.UserAgent
		*testutil.CastPtrStr(dest[4], 4) = userSession.IPAddress
		*testutil.CastTime(dest[5], 5) = userSession.ExpiresAt
		*castPtrTime(dest[6], 6) = userSession.RevokedAt
		*castPtrRevokeReason(dest[7], 7) = userSession.RevokeReason
		*testutil.CastPtrStr(dest[8], 8) = userSession.RevokedByUserID
		*testutil.CastTime(dest[9], 9) = userSession.CreatedAt
		*testutil.CastInt(dest[10], 10) = userSession.TokenVersion
		*castPtrTime(dest[11], 11) = userSession.LastAttemptAt
		*castPtrTime(dest[12], 12) = userSession.LockedUntil
		*testutil.CastInt(dest[13], 13) = userSession.FailedAttemptCount
		*testutil.CastPtrStr(dest[14], 14) = userSession.PreviousRefreshHash
		*castPtrTime(dest[15], 15) = userSession.LastSeenAt
		*testutil.CastPtrStr(dest[16], 16) = userSession.LastSeenIP
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
		*testutil.CastStr(dest[0], 0) = tb.ID
		*testutil.CastStr(dest[1], 1) = tb.UserID
		*testutil.CastStr(dest[2], 2) = tb.TokenHash
		*testutil.CastTime(dest[3], 3) = tb.ExpiresAt
		*testutil.CastTime(dest[4], 4) = tb.CreatedAt
		*castPtrTime(dest[5], 5) = tb.UsedAt
		*castPtrTime(dest[6], 6) = tb.InvalidatedAt
		*testutil.CastPtrStr(dest[7], 7) = tb.InvalidatedByUserID
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
		*testutil.CastStr(dest[0], 0) = t.ID
		*testutil.CastStr(dest[1], 1) = t.UserID
		*testutil.CastStr(dest[2], 2) = t.NewEmail
		*testutil.CastStr(dest[3], 3) = t.TokenHash
		*testutil.CastTime(dest[4], 4) = t.ExpiresAt
		*testutil.CastTime(dest[5], 5) = t.CreatedAt
		*castPtrTime(dest[6], 6) = t.UsedAt
		*castPtrTime(dest[7], 7) = t.InvalidatedAt
		*testutil.CastPtrStr(dest[8], 8) = t.InvalidatedByUserID
		return nil
	}
}
