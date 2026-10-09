//go:build integration

package authservice

import (
	"context"
	"errors"
	authdomain "learnflow_backend/internal/auth/domain"
	authrepository "learnflow_backend/internal/auth/repository"
	"learnflow_backend/internal/infrastructure/db"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	"learnflow_backend/internal/shared/rediskeys"
	"learnflow_backend/internal/shared/testutil"
	"learnflow_backend/internal/shared/tokens"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

// The service commits through a real transactor, so these tests write to the pool and clean up after themselves.
func newRefreshIntegrationService(t *testing.T, pool *pgxpool.Pool) (*Service, *authrepository.Repository) {
	t.Helper()

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	t.Cleanup(func() { redisClient.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup

	repo := authrepository.NewRepository(pool)
	srv, err := New(
		Repos{UserRepo: repo, SessionRepo: repo, TokenRepo: repo, Transactor: db.NewTransactor(pool)},
		Utils{
			Token:     tokens.NewTokens("refresh-integration-secret-0123456789", "", "learnflow-test", "learnflow-test"),
			Blocklist: redisinfra.NewInstance(redisClient),
		},
		Options{BcryptCost: 4},
	)
	if err != nil {
		t.Fatalf("authservice.New: %v", err)
	}

	return srv, repo
}

func seedRefreshSession(t *testing.T, repo *authrepository.Repository, userID string) (rawToken string) {
	t.Helper()

	raw, hash, err := tokens.GenerateSecureToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	agent, ip := "integration-agent", "127.0.0.1"
	if _, err = repo.CreateUserSession(context.Background(), &authdomain.UserSession{
		UserID: userID, RefreshHash: hash, UserAgent: &agent, IPAddress: &ip, ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	return raw
}

func activeSessionCount(t *testing.T, pool *pgxpool.Pool, userID string) int {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&count)
	if err != nil {
		t.Fatalf("count active sessions: %v", err)
	}

	return count
}

func TestRefreshReplayRevokesAllSessions_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an auth service backed by real Postgres", t, func() {
		ctx := context.Background()
		srv, repo := newRefreshIntegrationService(t, pool)
		userID := testutil.InsertRandomTestUser(t, pool)
		t.Cleanup(func() {
			pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID) //nolint:errcheck // best-effort cleanup of throwaway data
			pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)              //nolint:errcheck // best-effort cleanup of throwaway data
		})
		rotatedRaw := seedRefreshSession(t, repo, userID)
		seedRefreshSession(t, repo, userID)
		req := authdomain.RefreshRequest{RefreshToken: rotatedRaw, UserAgent: "integration-agent", IPAddress: "127.0.0.1"}

		Convey("When the token is used once it rotates and keeps every session active", func() {
			got, err := srv.Refresh(ctx, req)

			So(err, ShouldBeNil)
			So(got.RefreshToken, ShouldNotEqual, rotatedRaw)
			So(activeSessionCount(t, pool, userID), ShouldEqual, 2)
		})

		Convey("When the rotated-out token is replayed every session of the user is revoked and stays revoked", func() {
			_, err := srv.Refresh(ctx, req)
			So(err, ShouldBeNil)

			_, err = srv.Refresh(ctx, authdomain.RefreshRequest{RefreshToken: rotatedRaw, UserAgent: "attacker", IPAddress: "203.0.113.9"})

			So(errors.Is(err, authdomain.ErrSessionRevoked), ShouldBeTrue)
			So(activeSessionCount(t, pool, userID), ShouldEqual, 0)

			var reason string
			So(pool.QueryRow(ctx, `SELECT revoke_reason FROM user_sessions WHERE user_id = $1 AND revoked_at IS NOT NULL LIMIT 1`, userID).Scan(&reason), ShouldBeNil)
			So(reason, ShouldEqual, string(authdomain.RevokeReasonSuspiciousActivity))
		})

		Convey("When the rotated-out token is replayed, the user's already issued access tokens are revoked too", func() {
			_, err := srv.Refresh(ctx, req)
			So(err, ShouldBeNil)

			_, err = srv.Refresh(ctx, authdomain.RefreshRequest{RefreshToken: rotatedRaw, UserAgent: "attacker", IPAddress: "203.0.113.9"})
			So(errors.Is(err, authdomain.ErrSessionRevoked), ShouldBeTrue)

			redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
			defer redisClient.Close()                                                         //nolint:errcheck // Close's error is never actionable in test cleanup
			t.Cleanup(func() { redisClient.Del(ctx, rediskeys.TokensRevokedBefore(userID)) }) //nolint:errcheck // best-effort cleanup
			So(redisClient.Exists(ctx, rediskeys.TokensRevokedBefore(userID)).Val(), ShouldEqual, 1)
		})
	})
}
