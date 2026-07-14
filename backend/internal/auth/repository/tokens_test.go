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

type rowHolder struct{ row *testutil.MockRow }

func newRowTestRepo() (*Repository, *rowHolder) {
	rh := &rowHolder{}
	repo := newTestRepo(&testutil.MockQueryRunner{
		QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return rh.row
		},
	})
	return repo, rh
}

func TestCreateEmailVerificationToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When creation succeeds", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.CreateEmailVerificationToken(context.Background(), &authdomain.EmailVerificationToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.UserID, ShouldEqual, "user_123")
			So(got.TokenHash, ShouldEqual, "some_password_hash")
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateEmailVerificationToken(context.Background(), &authdomain.EmailVerificationToken{})
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetEmailVerificationToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When token exists", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.GetEmailVerificationToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetEmailVerificationToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetEmailVerificationToken(context.Background(), "hash-abc")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

// runTokenMarkTests exercises the three standard outcomes shared by every
// Mark*TokenUsed repository method: unexpected DB error, 0 rows affected
// (already used / never existed), and success.
func runTokenMarkTests(t *testing.T, name string, mark func(repo *Repository, ctx context.Context, hash string) error) {
	Convey("Given a "+name+" repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = testutil.ErrDB
			err := mark(repo, context.Background(), "hash-abc")
			assertUnexpectedDBError(err, "db error")
		})

		Convey("When 0 rows are affected (token already used or never existed)", func() {
			err := mark(repo, context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
		})

		Convey("When mark succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := mark(repo, context.Background(), "hash-abc")
			So(err, ShouldBeNil)
		})
	})
}

func TestMarkEmailVerificationTokenUsed(t *testing.T) {
	runTokenMarkTests(t, "email verification tokens", (*Repository).MarkEmailVerificationTokenUsed)
}

func TestCreatePasswordResetToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When creation succeeds", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.CreatePasswordResetToken(context.Background(), &authdomain.PasswordResetToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreatePasswordResetToken(context.Background(), &authdomain.PasswordResetToken{})
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetPasswordResetToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When token exists", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.GetPasswordResetToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetPasswordResetToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetPasswordResetToken(context.Background(), "hash-abc")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestMarkPasswordResetTokenUsed(t *testing.T) {
	runTokenMarkTests(t, "password reset tokens", (*Repository).MarkPasswordResetTokenUsed)
}

func TestCreateEmailChangeToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When creation succeeds", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanEmailChangeToken(now)}
			got, err := repo.CreateEmailChangeToken(context.Background(), &authdomain.EmailChangeToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
				NewEmail:  "new@example.com",
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.NewEmail, ShouldEqual, "john@gmail.com")
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateEmailChangeToken(context.Background(), &authdomain.EmailChangeToken{})
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetEmailChangeToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When token exists", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanEmailChangeToken(now)}
			got, err := repo.GetEmailChangeToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.NewEmail, ShouldEqual, "john@gmail.com")
		})

		Convey("When token not found", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetEmailChangeToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetEmailChangeToken(context.Background(), "hash-abc")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestMarkEmailChangeTokenUsed(t *testing.T) {
	runTokenMarkTests(t, "email change tokens", (*Repository).MarkEmailChangeTokenUsed)
}

func TestCreateAccountRecoveryToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When creation succeeds", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.CreateAccountRecoveryToken(context.Background(), &authdomain.AccountRecoveryToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateAccountRecoveryToken(context.Background(), &authdomain.AccountRecoveryToken{})
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetAccountRecoveryToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		repo, rh := newRowTestRepo()

		Convey("When token exists", func() {
			rh.row = &testutil.MockRow{ScanFn: fakeScanToken(now)}
			got, err := repo.GetAccountRecoveryToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetAccountRecoveryToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			rh.row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.GetAccountRecoveryToken(context.Background(), "hash-abc")
			assertUnexpectedDBError(err, "db error")
		})
	})
}

func TestMarkAccountRecoveryTokenUsed(t *testing.T) {
	runTokenMarkTests(t, "account recovery tokens", (*Repository).MarkAccountRecoveryTokenUsed)
}

func TestDeleteExpiredTokens(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		var callCount int
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				callCount++
				return pgconn.NewCommandTag("DELETE 2"), fakeErr
			},
		})

		Convey("When all queries succeed", func() {
			total, err := repo.DeleteExpiredTokens(context.Background())
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 8) // 4 queries × 2 rows each
			So(callCount, ShouldEqual, 4)
		})

		Convey("When a query fails mid-way", func() {
			callCount = 0
			fakeErr = testutil.ErrDB
			_, err := repo.DeleteExpiredTokens(context.Background())
			assertUnexpectedDBError(err, "db error")
			So(callCount, ShouldEqual, 1)
		})
	})
}

// Token tables (email_verification_tokens, password_reset_tokens,
// email_change_tokens, account_recovery_tokens) have no deleted_at column
// and are not in the soft-deletable table list
// (.claude/rules/db-conventions.md). Their own analog of a soft-delete
// filter is `used_at IS NULL AND expires_at > now() AND invalidated_at IS
// NULL` on lookup, and `used_at IS NULL` on mark-used. These tests are a
// regression guard on those filters so a used/expired/invalidated token can
// never be redeemed twice.

func TestGetEmailVerificationTokenFiltersUsedExpiredInvalidated(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, _ ...any) pgx.Row {
				gotQuery = sql
				return &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			},
		})

		Convey("When looking up a token by its hash", func() {
			_, err := repo.GetEmailVerificationToken(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			So(gotQuery, ShouldContainSubstring, "used_at IS NULL")
			So(gotQuery, ShouldContainSubstring, "expires_at > now()")
			So(gotQuery, ShouldContainSubstring, "invalidated_at IS NULL")
		})
	})
}

func TestGetPasswordResetTokenFiltersUsedExpiredInvalidated(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, _ ...any) pgx.Row {
				gotQuery = sql
				return &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			},
		})

		Convey("When looking up a token by its hash", func() {
			_, err := repo.GetPasswordResetToken(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			So(gotQuery, ShouldContainSubstring, "used_at IS NULL")
			So(gotQuery, ShouldContainSubstring, "expires_at > now()")
			So(gotQuery, ShouldContainSubstring, "invalidated_at IS NULL")
		})
	})
}

func TestGetEmailChangeTokenFiltersUsedExpiredInvalidated(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, _ ...any) pgx.Row {
				gotQuery = sql
				return &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			},
		})

		Convey("When looking up a token by its hash", func() {
			_, err := repo.GetEmailChangeToken(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			So(gotQuery, ShouldContainSubstring, "used_at IS NULL")
			So(gotQuery, ShouldContainSubstring, "expires_at > now()")
			So(gotQuery, ShouldContainSubstring, "invalidated_at IS NULL")
		})
	})
}

func TestGetAccountRecoveryTokenFiltersUsedExpiredInvalidated(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, sql string, _ ...any) pgx.Row {
				gotQuery = sql
				return &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			},
		})

		Convey("When looking up a token by its hash", func() {
			_, err := repo.GetAccountRecoveryToken(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			So(gotQuery, ShouldContainSubstring, "used_at IS NULL")
			So(gotQuery, ShouldContainSubstring, "expires_at > now()")
			So(gotQuery, ShouldContainSubstring, "invalidated_at IS NULL")
		})
	})
}

// runTokenMarkQueryFilterTest asserts the shared Mark*TokenUsed query shape:
// it must only affect rows where used_at IS NULL, so an already-used token
// can never be marked used (and thus never re-validated) twice.
func runTokenMarkQueryFilterTest(t *testing.T, name string, mark func(repo *Repository, ctx context.Context, hash string) error) {
	Convey("Given a "+name+" repository", t, func() {
		var gotQuery string
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
				gotQuery = sql
				return pgconn.NewCommandTag("UPDATE 1"), nil
			},
		})

		Convey("When marking a token used", func() {
			err := mark(repo, context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(gotQuery, ShouldContainSubstring, "used_at IS NULL")
		})
	})
}

func TestMarkEmailVerificationTokenUsedFiltersAlreadyUsedTokens(t *testing.T) {
	runTokenMarkQueryFilterTest(t, "email verification tokens", (*Repository).MarkEmailVerificationTokenUsed)
}

func TestMarkPasswordResetTokenUsedFiltersAlreadyUsedTokens(t *testing.T) {
	runTokenMarkQueryFilterTest(t, "password reset tokens", (*Repository).MarkPasswordResetTokenUsed)
}

func TestMarkEmailChangeTokenUsedFiltersAlreadyUsedTokens(t *testing.T) {
	runTokenMarkQueryFilterTest(t, "email change tokens", (*Repository).MarkEmailChangeTokenUsed)
}

func TestMarkAccountRecoveryTokenUsedFiltersAlreadyUsedTokens(t *testing.T) {
	runTokenMarkQueryFilterTest(t, "account recovery tokens", (*Repository).MarkAccountRecoveryTokenUsed)
}
