package authrepository

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetUserSessionByRefreshToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the profile exists", func() {
			row = &fakeRow{scanFn: fakeScanUserSession(now)}
			got, err := repo.GetUserSessionByRefreshToken(context.Background(), "refresh-token-hash")
			userSession := fakeUserSession(now)
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, userSession.ID)
			So(got.UserID, ShouldEqual, userSession.UserID)
			So(got.RefreshHash, ShouldEqual, userSession.RefreshHash)
			So(got.UserAgent, ShouldEqual, userSession.UserAgent)
			So(got.IPAddress, ShouldEqual, userSession.IPAddress)
			So(got.ExpiresAt, ShouldEqual, userSession.ExpiresAt)
			So(got.RevokedAt, ShouldEqual, userSession.RevokedAt)
			So(got.RevokeReason, ShouldEqual, userSession.RevokeReason)
			So(got.RevokedByUserID, ShouldEqual, userSession.RevokedByUserID)
			So(got.CreatedAt, ShouldEqual, userSession.CreatedAt)
			So(got.FailedAttemptCount, ShouldEqual, userSession.FailedAttemptCount)
			So(got.LastAttemptAt, ShouldEqual, userSession.LastAttemptAt)
			So(got.LockedUntil, ShouldEqual, userSession.LockedUntil)
			So(got.TokenVersion, ShouldEqual, userSession.TokenVersion)
			So(got.PreviousRefreshHash, ShouldEqual, userSession.PreviousRefreshHash)
			So(got.LastSeenAt, ShouldEqual, userSession.LastSeenAt)
			So(got.LastSeenIP, ShouldEqual, userSession.LastSeenIP)
		})

		Convey("When the profile does not exist", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserSessionByRefreshToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db connection lost") }}
			_, err := repo.GetUserSessionByRefreshToken(context.Background(), "refresh-token-hash")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}

func TestGetSessionByPrevHash(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the profile exists", func() {
			row = &fakeRow{scanFn: fakeScanUserSession(now)}
			got, err := repo.GetSessionByPrevHash(context.Background(), "refresh-token-hash")
			userSession := fakeUserSession(now)
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, userSession.ID)
			So(got.UserID, ShouldEqual, userSession.UserID)
			So(got.RefreshHash, ShouldEqual, userSession.RefreshHash)
			So(got.UserAgent, ShouldEqual, userSession.UserAgent)
			So(got.IPAddress, ShouldEqual, userSession.IPAddress)
			So(got.ExpiresAt, ShouldEqual, userSession.ExpiresAt)
			So(got.RevokedAt, ShouldEqual, userSession.RevokedAt)
			So(got.RevokeReason, ShouldEqual, userSession.RevokeReason)
			So(got.RevokedByUserID, ShouldEqual, userSession.RevokedByUserID)
			So(got.CreatedAt, ShouldEqual, userSession.CreatedAt)
			So(got.FailedAttemptCount, ShouldEqual, userSession.FailedAttemptCount)
			So(got.LastAttemptAt, ShouldEqual, userSession.LastAttemptAt)
			So(got.LockedUntil, ShouldEqual, userSession.LockedUntil)
			So(got.TokenVersion, ShouldEqual, userSession.TokenVersion)
			So(got.PreviousRefreshHash, ShouldEqual, userSession.PreviousRefreshHash)
			So(got.LastSeenAt, ShouldEqual, userSession.LastSeenAt)
			So(got.LastSeenIP, ShouldEqual, userSession.LastSeenIP)
		})

		Convey("When the profile does not exist", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetSessionByPrevHash(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db connection lost") }}
			_, err := repo.GetSessionByPrevHash(context.Background(), "refresh-token-hash")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}

func TestGetActiveSessionsByUserID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *fakeRow
		var retErr error
		repo := newTestRepo(&mockQueryRunner{
			queryRowsFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return pgx.Rows(&mockRows{rows: []*fakeRow{row}}), retErr
			},
		})

		Convey("When the profile exists", func() {
			row = &fakeRow{scanFn: fakeScanUserSession(now)}
			got, err := repo.GetActiveSessionsByUserID(context.Background(), "user-123")
			userSession := fakeUserSession(now)
			for _, session := range got {
				So(err, ShouldBeNil)
				So(session.ID, ShouldEqual, userSession.ID)
				So(session.UserID, ShouldEqual, userSession.UserID)
				So(session.RefreshHash, ShouldEqual, userSession.RefreshHash)
				So(session.UserAgent, ShouldEqual, userSession.UserAgent)
				So(session.IPAddress, ShouldEqual, userSession.IPAddress)
				So(session.ExpiresAt, ShouldEqual, userSession.ExpiresAt)
				So(session.RevokedAt, ShouldEqual, userSession.RevokedAt)
				So(session.RevokeReason, ShouldEqual, userSession.RevokeReason)
				So(session.RevokedByUserID, ShouldEqual, userSession.RevokedByUserID)
				So(session.CreatedAt, ShouldEqual, userSession.CreatedAt)
				So(session.FailedAttemptCount, ShouldEqual, userSession.FailedAttemptCount)
				So(session.LastAttemptAt, ShouldEqual, userSession.LastAttemptAt)
				So(session.LockedUntil, ShouldEqual, userSession.LockedUntil)
				So(session.TokenVersion, ShouldEqual, userSession.TokenVersion)
				So(session.PreviousRefreshHash, ShouldEqual, userSession.PreviousRefreshHash)
				So(session.LastSeenAt, ShouldEqual, userSession.LastSeenAt)
				So(session.LastSeenIP, ShouldEqual, userSession.LastSeenIP)
			}
		})

		Convey("When the profile does not exist", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			retErr = authdomain.ErrSessionNotFound
			_, err := repo.GetActiveSessionsByUserID(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db connection lost") }}
			_, err := repo.GetActiveSessionsByUserID(context.Background(), "refresh-token-hash")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}

func TestCreateUserSession(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the profile exists", func() {
			row = &fakeRow{scanFn: fakeScanUserSession(userSession.CreatedAt)}
			got, err := repo.CreateUserSession(context.Background(), userSession)

			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, userSession.ID)
			So(got.UserID, ShouldEqual, userSession.UserID)
			So(got.RefreshHash, ShouldEqual, userSession.RefreshHash)
			So(got.UserAgent, ShouldEqual, userSession.UserAgent)
			So(got.IPAddress, ShouldEqual, userSession.IPAddress)
			So(got.ExpiresAt, ShouldEqual, userSession.ExpiresAt)
			So(got.RevokedAt, ShouldEqual, userSession.RevokedAt)
			So(got.RevokeReason, ShouldEqual, userSession.RevokeReason)
			So(got.RevokedByUserID, ShouldEqual, userSession.RevokedByUserID)
			So(got.CreatedAt, ShouldEqual, userSession.CreatedAt)
			So(got.FailedAttemptCount, ShouldEqual, userSession.FailedAttemptCount)
			So(got.LastAttemptAt, ShouldEqual, userSession.LastAttemptAt)
			So(got.LockedUntil, ShouldEqual, userSession.LockedUntil)
			So(got.TokenVersion, ShouldEqual, userSession.TokenVersion)
			So(got.PreviousRefreshHash, ShouldEqual, userSession.PreviousRefreshHash)
			So(got.LastSeenAt, ShouldEqual, userSession.LastSeenAt)
			So(got.LastSeenIP, ShouldEqual, userSession.LastSeenIP)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db connection lost") }}
			_, err := repo.CreateUserSession(context.Background(), userSession)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}

func TestRevokeUserSession(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When reason is invalid", func() {
			reason := authdomain.RevokeReason("invalid_reason")
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, reason)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid revoke reason")
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db connection lost")
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, *userSession.RevokeReason)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When the session is not found", func() {
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, *userSession.RevokeReason)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When revocation succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, *userSession.RevokeReason)
			So(err, ShouldBeNil)
		})
	})
}

func TestRevokeAllUserSessions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When reason is invalid", func() {
			reason := authdomain.RevokeReason("invalid_reason")
			err := repo.RevokeAllUserSessions(context.Background(), userSession.UserID, userSession.RevokedByUserID, reason)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "invalid revoke reason")
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db connection lost")
			err := repo.RevokeAllUserSessions(context.Background(), userSession.UserID, userSession.RevokedByUserID, *userSession.RevokeReason)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When revocation succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 3")
			err := repo.RevokeAllUserSessions(context.Background(), userSession.UserID, userSession.RevokedByUserID, *userSession.RevokeReason)
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateSessionToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db connection lost")
			err := repo.UpdateSessionToken(context.Background(), userSession.ID, "new-refresh-hash", "Mozilla/5.0", "127.0.0.1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When the session is not found", func() {
			err := repo.UpdateSessionToken(context.Background(), userSession.ID, "new-refresh-hash", "Mozilla/5.0", "127.0.0.1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateSessionToken(context.Background(), userSession.ID, "new-refresh-hash", "Mozilla/5.0", "127.0.0.1")
			So(err, ShouldBeNil)
		})
	})
}

func TestUpdateFailedLoginAttempts(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db connection lost")
			err := repo.UpdateFailedLoginAttempts(context.Background(), userSession.ID, "15 minutes", 5)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("When the session is not found", func() {
			err := repo.UpdateFailedLoginAttempts(context.Background(), userSession.ID, "15 minutes", 5)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.UpdateFailedLoginAttempts(context.Background(), userSession.ID, "15 minutes", 5)
			So(err, ShouldBeNil)
		})
	})
}
