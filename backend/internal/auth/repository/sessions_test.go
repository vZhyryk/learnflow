package authrepository

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetUserSessionByRefreshToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the session exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUserSession(now)}
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

		Convey("When the session does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetUserSessionByRefreshToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetUserSessionByRefreshToken(context.Background(), "refresh-token-hash")
			assertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetSessionByPrevHash(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the session exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUserSession(now)}
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

		Convey("When the session does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetSessionByPrevHash(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetSessionByPrevHash(context.Background(), "refresh-token-hash")
			assertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetActiveSessionsByUserID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		var retErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return pgx.Rows(&testutil.MockRows{Rows: []*testutil.MockRow{row}}), retErr
			},
		})

		Convey("When the session exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUserSession(now)}
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

		Convey("When the session does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			retErr = authdomain.ErrSessionNotFound
			_, err := repo.GetActiveSessionsByUserID(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetActiveSessionsByUserID(context.Background(), "refresh-token-hash")
			assertUnexpectedDBError(err, "db connection lost")
		})

		Convey("When rows.Err returns an error after iteration", func() {
			rowsErrRepo := newTestRepo(&testutil.MockQueryRunner{
				QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
					return pgx.Rows(&testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}), nil
				},
			})
			_, err := rowsErrRepo.GetActiveSessionsByUserID(context.Background(), "user-123")
			assertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestCreateUserSession(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the session exists", func() {
			row = &testutil.MockRow{ScanFn: fakeScanUserSession(userSession.CreatedAt)}
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
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.CreateUserSession(context.Background(), userSession)
			assertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestRevokeUserSession(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	userSession := fakeUserSession(now)

	Convey("Given a users repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When reason is invalid", func() {
			reason := authdomain.RevokeReason("invalid_reason")
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, reason)

			assertUnexpectedDBError(err, "invalid revoke reason")
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDBUnexpected
			err := repo.RevokeUserSession(context.Background(), userSession.ID, *userSession.RevokedByUserID, *userSession.RevokeReason)
			assertUnexpectedDBError(err, "db connection lost")
		})

		Convey("When the session is missing or already revoked (idempotent double-revoke)", func() {
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
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When reason is invalid", func() {
			reason := authdomain.RevokeReason("invalid_reason")
			err := repo.RevokeAllUserSessions(context.Background(), userSession.UserID, userSession.RevokedByUserID, reason)

			assertUnexpectedDBError(err, "invalid revoke reason")
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDBUnexpected
			err := repo.RevokeAllUserSessions(context.Background(), userSession.UserID, userSession.RevokedByUserID, *userSession.RevokeReason)
			assertUnexpectedDBError(err, "db connection lost")
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
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDBUnexpected
			err := repo.UpdateSessionToken(context.Background(), userSession.ID, "new-refresh-hash", "Mozilla/5.0", "127.0.0.1")
			assertUnexpectedDBError(err, "db connection lost")
		})

		Convey("When the session is missing, revoked, or expired (0 rows)", func() {
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
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDBUnexpected
			err := repo.UpdateFailedLoginAttempts(context.Background(), userSession.ID, "15 minutes", 5)
			assertUnexpectedDBError(err, "db connection lost")
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
