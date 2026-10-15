//go:build integration

package worker

import (
	"context"
	"testing"
	"time"

	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthCleanupWorkers_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given expired and live token and session rows", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			userID := testutil.InsertRandomTestUser(t, tx)
			past := time.Now().Add(-48 * time.Hour)
			expired := time.Now().Add(-24 * time.Hour)
			live := time.Now().Add(24 * time.Hour)

			insertToken := func(hash string, expiresAt time.Time) string {
				var id string
				err := tx.QueryRow(ctx,
					`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at, created_at)
					 VALUES ($1, $2, $3, $4) RETURNING id`, userID, hash, expiresAt, past).Scan(&id)
				So(err, ShouldBeNil)
				return id
			}
			exists := func(table, id string) bool {
				var n int
				So(tx.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE id = $1", id).Scan(&n), ShouldBeNil)
				return n == 1
			}

			expiredTokenID := insertToken("expired-token-hash", expired)
			liveTokenID := insertToken("live-token-hash", live)

			Convey("the token cleanup removes only expired tokens", func() {
				for _, w := range NewTokenCleanupWorkers(tx, testutil.NewTestLogger(), time.Hour) {
					w.(*CleanupWorker[TokenCleanupWorker]).poll(ctx)
				}

				So(exists("email_verification_tokens", expiredTokenID), ShouldBeFalse)
				So(exists("email_verification_tokens", liveTokenID), ShouldBeTrue)
			})

			Convey("the session cleanup keeps recently expired sessions and removes old ones", func() {
				insertSession := func(hash string, createdAt, expiresAt time.Time) string {
					var id string
					err := tx.QueryRow(ctx,
						`INSERT INTO user_sessions (user_id, refresh_hash, expires_at, created_at)
						 VALUES ($1, $2, $3, $4) RETURNING id`, userID, hash, expiresAt, createdAt).Scan(&id)
					So(err, ShouldBeNil)
					return id
				}
				longAgo := time.Now().Add(-90 * 24 * time.Hour)
				oldID := insertSession("old-session-hash", longAgo, longAgo.Add(time.Hour))
				recentID := insertSession("recent-session-hash", past, expired)

				NewSessionCleanupWorker(tx, testutil.NewTestLogger(), time.Hour).poll(ctx)

				So(exists("user_sessions", oldID), ShouldBeFalse)
				So(exists("user_sessions", recentID), ShouldBeTrue)
			})
		})
	})
}
