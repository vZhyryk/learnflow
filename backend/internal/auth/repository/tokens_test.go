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

func TestCreateEmailVerificationToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.CreateEmailVerificationToken(context.Background(), &authdomain.EmailVerificationToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.UserID, ShouldEqual, "user_123")
			So(got.TokenHash, ShouldEqual, "some_password_hash")
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.CreateEmailVerificationToken(context.Background(), &authdomain.EmailVerificationToken{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestGetEmailVerificationToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When token exists", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.GetEmailVerificationToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetEmailVerificationToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.GetEmailVerificationToken(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestMarkEmailVerificationTokenUsed(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db error")
			err := repo.MarkEmailVerificationTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})

		Convey("When token is already used (0 rows affected)", func() {
			err := repo.MarkEmailVerificationTokenUsed(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
		})

		Convey("When mark succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.MarkEmailVerificationTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
		})
	})
}

func TestCreatePasswordResetToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.CreatePasswordResetToken(context.Background(), &authdomain.PasswordResetToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.CreatePasswordResetToken(context.Background(), &authdomain.PasswordResetToken{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestGetPasswordResetToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When token exists", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.GetPasswordResetToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetPasswordResetToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.GetPasswordResetToken(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestMarkPasswordResetTokenUsed(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db error")
			err := repo.MarkPasswordResetTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})

		Convey("When token is already used", func() {
			err := repo.MarkPasswordResetTokenUsed(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
		})

		Convey("When mark succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.MarkPasswordResetTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateEmailChangeToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			row = &fakeRow{scanFn: fakeScanEmailChangeToken(now)}
			got, err := repo.CreateEmailChangeToken(context.Background(), &authdomain.EmailChangeToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
				NewEmail:  "new@example.com",
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.NewEmail, ShouldEqual, "john@gmail.com")
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.CreateEmailChangeToken(context.Background(), &authdomain.EmailChangeToken{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestGetEmailChangeToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When token exists", func() {
			row = &fakeRow{scanFn: fakeScanEmailChangeToken(now)}
			got, err := repo.GetEmailChangeToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
			So(got.NewEmail, ShouldEqual, "john@gmail.com")
		})

		Convey("When token not found", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetEmailChangeToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.GetEmailChangeToken(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestMarkEmailChangeTokenUsed(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db error")
			err := repo.MarkEmailChangeTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})

		Convey("When token is already used", func() {
			err := repo.MarkEmailChangeTokenUsed(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
		})

		Convey("When mark succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.MarkEmailChangeTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
		})
	})
}

func TestCreateAccountRecoveryToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.CreateAccountRecoveryToken(context.Background(), &authdomain.AccountRecoveryToken{
				TokenBase: authdomain.TokenBase{UserID: "user-123", TokenHash: "hash-abc", ExpiresAt: now},
			})
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.CreateAccountRecoveryToken(context.Background(), &authdomain.AccountRecoveryToken{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestGetAccountRecoveryToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a tokens repository", t, func() {
		var row *fakeRow
		repo := newTestRepo(&mockQueryRunner{
			queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When token exists", func() {
			row = &fakeRow{scanFn: fakeScanToken(now)}
			got, err := repo.GetAccountRecoveryToken(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
			So(got.ID, ShouldEqual, "session_123")
		})

		Convey("When token not found", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetAccountRecoveryToken(context.Background(), "unknown")
			So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &fakeRow{scanFn: func(_ ...any) error { return errors.New("db error") }}
			_, err := repo.GetAccountRecoveryToken(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})
	})
}

func TestMarkAccountRecoveryTokenUsed(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		var execTag pgconn.CommandTag
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, fakeErr
			},
		})

		Convey("When the database returns an unexpected error", func() {
			fakeErr = errors.New("db error")
			err := repo.MarkAccountRecoveryTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
		})

		Convey("When token is already used", func() {
			err := repo.MarkAccountRecoveryTokenUsed(context.Background(), "hash-abc")
			So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
		})

		Convey("When mark succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			err := repo.MarkAccountRecoveryTokenUsed(context.Background(), "hash-abc")
			So(err, ShouldBeNil)
		})
	})
}

func TestDeleteExpiredTokens(t *testing.T) {
	Convey("Given a tokens repository", t, func() {
		var fakeErr error
		callCount := 0
		repo := newTestRepo(&mockQueryRunner{
			execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
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
			fakeErr = errors.New("db error")
			_, err := repo.DeleteExpiredTokens(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db error")
			So(callCount, ShouldEqual, 1)
		})
	})
}
