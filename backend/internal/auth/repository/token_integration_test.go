//go:build integration

package authrepository

import (
	"context"
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func uniqueTokenHash(prefix string) string {
	return fmt.Sprintf("%s-token-hash-%s", prefix, uniqueSuffix())
}

func assertFreshToken(base authdomain.TokenBase, userId, hash string) {
	So(base.ID, ShouldNotBeEmpty)
	So(base.UserID, ShouldEqual, userId)
	So(base.TokenHash, ShouldEqual, hash)
	So(base.CreatedAt.IsZero(), ShouldBeFalse)
	So(base.UsedAt, ShouldBeNil)
	So(base.InvalidatedAt, ShouldBeNil)
	So(base.InvalidatedByUserID, ShouldBeNil)
}

func TestEmailVerificationToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("EmailVerificationToken", t, func() {
		Convey("Create and Get", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ev")
				expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

				created, err := repo.CreateEmailVerificationToken(ctx, &authdomain.EmailVerificationToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: expiresAt},
				})
				So(err, ShouldBeNil)
				assertFreshToken(created.TokenBase, userId, hash)

				got, err := repo.GetEmailVerificationToken(ctx, hash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				So(got.ID, ShouldEqual, created.ID)
				assertFreshToken(got.TokenBase, userId, hash)
			})
		})

		Convey("MarkEmailVerificationTokenUsed", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ev")
				_, err := repo.CreateEmailVerificationToken(ctx, &authdomain.EmailVerificationToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)},
				})
				So(err, ShouldBeNil)

				err = repo.MarkEmailVerificationTokenUsed(ctx, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetEmailVerificationToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkEmailVerificationTokenUsed(ctx, hash)
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("token does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, err := repo.GetEmailVerificationToken(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkEmailVerificationTokenUsed(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("expired token is not returned", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ev")

				_, err := repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO email_verification_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`,
					userId, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetEmailVerificationToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			})
		})

		Convey("invalidated token is not returned", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ev")
				_, err := repo.CreateEmailVerificationToken(ctx, &authdomain.EmailVerificationToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)},
				})
				So(err, ShouldBeNil)

				_, err = repo.queryRunner(ctx).Exec(ctx,
					`UPDATE email_verification_tokens SET invalidated_at = now(), invalidated_by_user_id = $2 WHERE token_hash = $1`,
					hash, userId)
				So(err, ShouldBeNil)

				_, err = repo.GetEmailVerificationToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			})
		})
	})
}

func TestPasswordResetToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("PasswordResetToken", t, func() {
		Convey("Create and Get", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("pr")
				expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

				created, err := repo.CreatePasswordResetToken(ctx, &authdomain.PasswordResetToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: expiresAt},
				})
				So(err, ShouldBeNil)
				assertFreshToken(created.TokenBase, userId, hash)

				got, err := repo.GetPasswordResetToken(ctx, hash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				So(got.ID, ShouldEqual, created.ID)
				assertFreshToken(got.TokenBase, userId, hash)
			})
		})

		Convey("MarkPasswordResetTokenUsed", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("pr")
				_, err := repo.CreatePasswordResetToken(ctx, &authdomain.PasswordResetToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)},
				})
				So(err, ShouldBeNil)

				err = repo.MarkPasswordResetTokenUsed(ctx, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetPasswordResetToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkPasswordResetTokenUsed(ctx, hash)
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("token does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, err := repo.GetPasswordResetToken(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkPasswordResetTokenUsed(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("expired token is not returned", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("pr")

				_, err := repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO password_reset_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`,
					userId, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetPasswordResetToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			})
		})
	})
}

func TestEmailChangeToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("EmailChangeToken", t, func() {
		Convey("Create and Get", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ec")
				newEmail := fmt.Sprintf("new-%s@example.com", hash)
				expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

				created, err := repo.CreateEmailChangeToken(ctx, &authdomain.EmailChangeToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: expiresAt},
					NewEmail:  newEmail,
				})
				So(err, ShouldBeNil)
				assertFreshToken(created.TokenBase, userId, hash)
				So(created.NewEmail, ShouldEqual, newEmail)

				got, err := repo.GetEmailChangeToken(ctx, hash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				So(got.ID, ShouldEqual, created.ID)
				So(got.NewEmail, ShouldEqual, newEmail)
				assertFreshToken(got.TokenBase, userId, hash)
			})
		})

		Convey("MarkEmailChangeTokenUsed", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ec")
				_, err := repo.CreateEmailChangeToken(ctx, &authdomain.EmailChangeToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)},
					NewEmail:  fmt.Sprintf("new-%s@example.com", hash),
				})
				So(err, ShouldBeNil)

				err = repo.MarkEmailChangeTokenUsed(ctx, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetEmailChangeToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkEmailChangeTokenUsed(ctx, hash)
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("token does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, err := repo.GetEmailChangeToken(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkEmailChangeTokenUsed(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("expired token is not returned", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ec")

				_, err := repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO email_change_tokens (user_id, token_hash, new_email, created_at, expires_at)
					 VALUES ($1, $2, $3, now() - interval '2 hours', now() - interval '1 hour')`,
					userId, hash, fmt.Sprintf("new-%s@example.com", hash))
				So(err, ShouldBeNil)

				_, err = repo.GetEmailChangeToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			})
		})
	})
}

func TestAccountRecoveryToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("AccountRecoveryToken", t, func() {
		Convey("Create and Get", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ar")
				expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

				created, err := repo.CreateAccountRecoveryToken(ctx, &authdomain.AccountRecoveryToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: expiresAt},
				})
				So(err, ShouldBeNil)
				assertFreshToken(created.TokenBase, userId, hash)

				got, err := repo.GetAccountRecoveryToken(ctx, hash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				So(got.ID, ShouldEqual, created.ID)
				assertFreshToken(got.TokenBase, userId, hash)
			})
		})

		Convey("MarkAccountRecoveryTokenUsed", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ar")
				_, err := repo.CreateAccountRecoveryToken(ctx, &authdomain.AccountRecoveryToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: hash, ExpiresAt: time.Now().Add(time.Hour)},
				})
				So(err, ShouldBeNil)

				err = repo.MarkAccountRecoveryTokenUsed(ctx, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetAccountRecoveryToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkAccountRecoveryTokenUsed(ctx, hash)
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("token does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, err := repo.GetAccountRecoveryToken(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)

				err = repo.MarkAccountRecoveryTokenUsed(ctx, "non-existent-hash")
				So(errors.Is(err, authdomain.ErrTokenUsed), ShouldBeTrue)
			})
		})

		Convey("expired token is not returned", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				hash := uniqueTokenHash("ar")

				_, err := repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO account_recovery_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`,
					userId, hash)
				So(err, ShouldBeNil)

				_, err = repo.GetAccountRecoveryToken(ctx, hash)
				So(errors.Is(err, authdomain.ErrInvalidToken), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteExpiredTokens_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("DeleteExpiredTokens", t, func() {
		Convey("removes only expired rows across all four token tables", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)

				validExpiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

				validEV, err := repo.CreateEmailVerificationToken(ctx, &authdomain.EmailVerificationToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: uniqueTokenHash("ev-valid"), ExpiresAt: validExpiresAt},
				})
				So(err, ShouldBeNil)
				validPR, err := repo.CreatePasswordResetToken(ctx, &authdomain.PasswordResetToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: uniqueTokenHash("pr-valid"), ExpiresAt: validExpiresAt},
				})
				So(err, ShouldBeNil)
				validEC, err := repo.CreateEmailChangeToken(ctx, &authdomain.EmailChangeToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: uniqueTokenHash("ec-valid"), ExpiresAt: validExpiresAt},
					NewEmail:  "valid@example.com",
				})
				So(err, ShouldBeNil)
				validAR, err := repo.CreateAccountRecoveryToken(ctx, &authdomain.AccountRecoveryToken{
					TokenBase: authdomain.TokenBase{UserID: userId, TokenHash: uniqueTokenHash("ar-valid"), ExpiresAt: validExpiresAt},
				})
				So(err, ShouldBeNil)

				expiredEVHash := uniqueTokenHash("ev-expired")
				_, err = repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO email_verification_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`, userId, expiredEVHash)
				So(err, ShouldBeNil)

				expiredPRHash := uniqueTokenHash("pr-expired")
				_, err = repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO password_reset_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`, userId, expiredPRHash)
				So(err, ShouldBeNil)

				expiredECHash := uniqueTokenHash("ec-expired")
				_, err = repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO email_change_tokens (user_id, token_hash, new_email, created_at, expires_at)
					 VALUES ($1, $2, $3, now() - interval '2 hours', now() - interval '1 hour')`,
					userId, expiredECHash, "expired@example.com")
				So(err, ShouldBeNil)

				expiredARHash := uniqueTokenHash("ar-expired")
				_, err = repo.queryRunner(ctx).Exec(ctx,
					`INSERT INTO account_recovery_tokens (user_id, token_hash, created_at, expires_at)
					 VALUES ($1, $2, now() - interval '2 hours', now() - interval '1 hour')`, userId, expiredARHash)
				So(err, ShouldBeNil)

				total, err := repo.DeleteExpiredTokens(ctx)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 4)

				_, err = repo.GetEmailVerificationToken(ctx, validEV.TokenHash)
				So(err, ShouldBeNil)
				_, err = repo.GetPasswordResetToken(ctx, validPR.TokenHash)
				So(err, ShouldBeNil)
				_, err = repo.GetEmailChangeToken(ctx, validEC.TokenHash)
				So(err, ShouldBeNil)
				_, err = repo.GetAccountRecoveryToken(ctx, validAR.TokenHash)
				So(err, ShouldBeNil)
			})
		})

		Convey("nothing to delete", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				total, err := repo.DeleteExpiredTokens(ctx)
				So(err, ShouldBeNil)
				So(total, ShouldEqual, 0)
			})
		})
	})
}
