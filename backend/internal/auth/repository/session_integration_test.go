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

const nonExistentUUID = "00000000-0000-0000-0000-000000000000"

func fullTestSession(userId string) *authdomain.UserSession {
	refreshHash := fmt.Sprintf("refresh-hash-value-%s", uniqueSuffix())
	userAgent, ipAddress := "Mozilla/5.0 (Test)", "203.0.113.42"
	revokedByUserID := "admin-user-id"
	previousRefreshHash := fmt.Sprintf("prev-refresh-hash-%s", uniqueSuffix())
	lastSeenIP := "203.0.113.99"
	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(30 * 24 * time.Hour)
	revokedAt, lastSeenAt := now.Add(15*time.Minute), now

	return &authdomain.UserSession{
		UserID:              userId,
		RefreshHash:         refreshHash,
		UserAgent:           &userAgent,
		IPAddress:           &ipAddress,
		ExpiresAt:           expiresAt,
		RevokedAt:           &revokedAt,
		RevokedByUserID:     &revokedByUserID,
		PreviousRefreshHash: &previousRefreshHash,
		LastSeenAt:          &lastSeenAt,
		LastSeenIP:          &lastSeenIP,
	}
}

func newSessionFixture(t *testing.T, ctx context.Context, repo *Repository) (userId string, seed, session *authdomain.UserSession) {
	t.Helper()

	userId = newTestUser(t, ctx, repo)

	seed = fullTestSession(userId)
	session, err := repo.CreateUserSession(ctx, seed)
	So(err, ShouldBeNil)

	t.Cleanup(func() {
		repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
	})

	return userId, seed, session
}

func createSessions(t *testing.T, ctx context.Context, repo *Repository, userId string, n int) []*authdomain.UserSession {
	t.Helper()

	seeds := make([]*authdomain.UserSession, 0, n)
	for i := 0; i < n; i++ {
		seed := fullTestSession(userId)
		_, err := repo.CreateUserSession(ctx, seed)
		So(err, ShouldBeNil)
		seeds = append(seeds, seed)
	}
	return seeds
}

// assertFreshSession checks the fields of a session that has just been created
// and never rotated, revoked, or had a failed login attempt recorded against it.
func assertFreshSession(got, seed *authdomain.UserSession) {
	So(got.UserID, ShouldEqual, seed.UserID)
	So(got.RefreshHash, ShouldEqual, seed.RefreshHash)
	So(got.UserAgent, ShouldEqual, seed.UserAgent)
	So(got.IPAddress, ShouldEqual, seed.IPAddress)
	So(got.ExpiresAt, ShouldEqual, seed.ExpiresAt)
	So(got.RevokedAt, ShouldBeNil)
	So(got.RevokeReason, ShouldBeNil)
	So(got.RevokedByUserID, ShouldBeNil)
	So(got.CreatedAt.IsZero(), ShouldBeFalse)
	So(got.FailedAttemptCount, ShouldEqual, 0)
	So(got.LastAttemptAt, ShouldBeNil)
	So(got.LockedUntil, ShouldBeNil)
	So(got.TokenVersion, ShouldEqual, 1)
	So(got.PreviousRefreshHash, ShouldBeNil)
	So(got.LastSeenAt, ShouldBeNil)
	So(got.LastSeenIP, ShouldBeNil)
}

// assertRevokedSession checks the fields of a session after RevokeUserSession/
// RevokeAllUserSessions, otherwise untouched (no rotation, no failed attempts).
func assertRevokedSession(got *authdomain.UserSession, revokedByUserID string, reason authdomain.RevokeReason) {
	So(got.RevokedAt, ShouldNotBeNil)
	So(got.RevokeReason, ShouldEqual, &reason)
	So(got.RevokedByUserID, ShouldEqual, &revokedByUserID)
	So(got.CreatedAt.IsZero(), ShouldBeFalse)
	So(got.FailedAttemptCount, ShouldEqual, 0)
	So(got.LastAttemptAt, ShouldBeNil)
	So(got.LockedUntil, ShouldBeNil)
	So(got.TokenVersion, ShouldEqual, 1)
	So(got.PreviousRefreshHash, ShouldBeNil)
	So(got.LastSeenAt, ShouldBeNil)
	So(got.LastSeenIP, ShouldBeNil)
}

// assertFailedLoginState checks the fields of a session after one or more
// UpdateFailedLoginAttempts calls.
func assertFailedLoginState(t *testing.T, got, seed *authdomain.UserSession, expectFailedCount int, expectLocked bool) {
	t.Helper()
	So(got.UserID, ShouldEqual, seed.UserID)
	So(got.RefreshHash, ShouldEqual, seed.RefreshHash)
	So(got.UserAgent, ShouldEqual, seed.UserAgent)
	So(got.IPAddress, ShouldEqual, seed.IPAddress)
	So(got.ExpiresAt, ShouldEqual, seed.ExpiresAt)
	So(got.RevokedAt, ShouldBeNil)
	So(got.RevokeReason, ShouldBeNil)
	So(got.RevokedByUserID, ShouldBeNil)
	So(got.CreatedAt.IsZero(), ShouldBeFalse)
	So(got.FailedAttemptCount, ShouldEqual, expectFailedCount)
	So(got.LastAttemptAt, ShouldNotBeNil)
	if expectLocked {
		So(got.LockedUntil, ShouldNotBeNil)
	} else {
		So(got.LockedUntil, ShouldBeNil)
	}
	So(got.TokenVersion, ShouldEqual, 1)
	So(got.PreviousRefreshHash, ShouldBeNil)
	So(got.LastSeenAt, ShouldBeNil)
	So(got.LastSeenIP, ShouldBeNil)
}

// seedsByRefreshHash indexes seeds for matching returned rows in
// multi-session list assertions, where result order is not guaranteed.
func seedsByRefreshHash(seeds []*authdomain.UserSession) map[string]*authdomain.UserSession {
	m := make(map[string]*authdomain.UserSession, len(seeds))
	for _, s := range seeds {
		m[s.RefreshHash] = s
	}
	return m
}

func TestGetUserSessionByRefreshToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("GetUserSessionByRefreshToken", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, seed, _ := newSessionFixture(t, ctx, repo)

				got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				assertFreshSession(got, seed)
			})
		})

		Convey("Successful test - multiple sessions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				seeds := createSessions(t, ctx, repo, userId, 3)
				t.Cleanup(func() {
					repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
				})

				for _, seed := range seeds {
					got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
					So(err, ShouldBeNil)
					So(got, ShouldNotBeNil)
					assertFreshSession(got, seed)
				}
			})
		})

		Convey("user does not have a session", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				got, err := repo.GetUserSessionByRefreshToken(ctx, "non-existent-refresh-token")
				So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
				So(got, ShouldBeNil)
			})
		})
	})
}

func assertTokenRotation(t *testing.T, ctx context.Context, repo *Repository, seed, session *authdomain.UserSession, middle func(), end func(got *authdomain.UserSession)) {
	t.Helper()
	var ua string = "new-user-agent"
	var ip string = "new-ip-address"

	err := repo.UpdateSessionToken(ctx, session.ID, "new-refresh-token", ua, ip)
	So(err, ShouldBeNil)

	if middle != nil {
		middle()
	}

	got, err := repo.GetSessionByPrevHash(ctx, seed.RefreshHash)
	So(err, ShouldBeNil)
	So(got, ShouldNotBeNil)
	So(got.UserID, ShouldEqual, seed.UserID)
	So(got.RefreshHash, ShouldEqual, "new-refresh-token")
	So(got.UserAgent, ShouldEqual, &ua)
	So(got.IPAddress, ShouldEqual, session.IPAddress)
	So(got.ExpiresAt, ShouldEqual, seed.ExpiresAt)
	So(got.CreatedAt.IsZero(), ShouldBeFalse)
	So(got.FailedAttemptCount, ShouldEqual, 0)
	So(got.LastAttemptAt, ShouldNotBeNil)
	So(got.LockedUntil, ShouldBeNil)
	So(got.TokenVersion, ShouldEqual, 2)
	So(got.PreviousRefreshHash, ShouldEqual, &seed.RefreshHash)
	So(got.LastSeenAt, ShouldNotBeNil)
	So(got.LastSeenIP, ShouldEqual, &ip)

	if end != nil {
		end(got)
	}
}

func TestGetSessionByPrevHash_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("GetSessionByPrevHash", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, seed, session := newSessionFixture(t, ctx, repo)

				assertTokenRotation(t, ctx, repo, seed, session, nil, func(got *authdomain.UserSession) {
					So(got.RevokedAt, ShouldBeNil)
					So(got.RevokeReason, ShouldBeNil)
				})
			})
		})

		Convey("Successful test - multiple sessions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				createSessions(t, ctx, repo, userId, 2) // noise sessions for the same user

				seed := fullTestSession(userId)
				session, err := repo.CreateUserSession(ctx, seed)
				So(err, ShouldBeNil)
				t.Cleanup(func() {
					repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
				})

				assertTokenRotation(t, ctx, repo, seed, session, nil, func(got *authdomain.UserSession) {
					So(got.RevokedAt, ShouldBeNil)
					So(got.RevokeReason, ShouldBeNil)
				})
			})
		})

		Convey("user does not have a session", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				got, err := repo.GetSessionByPrevHash(ctx, "non-existent-prev-refresh-token")
				So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
				So(got, ShouldBeNil)
			})
		})
	})
}

func TestRevokeUserSession_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("RevokeUserSession", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId, seed, session := newSessionFixture(t, ctx, repo)

				res := authdomain.RevokeReasonLogout

				assertTokenRotation(t, ctx, repo, seed, session, func() {
					err := repo.RevokeUserSession(ctx, session.ID, userId, res)
					So(err, ShouldBeNil)
					got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
					So(got, ShouldBeNil)
					So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
					So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
				},
					func(got *authdomain.UserSession) {
						So(got.RevokedAt, ShouldNotBeNil)
						So(got.RevokeReason, ShouldEqual, &res)
						So(got.RevokedByUserID, ShouldEqual, &userId)
					})
			})
		})

		Convey("session does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				err := repo.RevokeUserSession(ctx, nonExistentUUID, nonExistentUUID, authdomain.RevokeReasonLogout)
				So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestRevokeAllUserSessions_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("RevokeAllUserSessions", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId, _, _ := newSessionFixture(t, ctx, repo)

				res := authdomain.RevokeReasonLogout
				err := repo.RevokeAllUserSessions(ctx, userId, &userId, res)
				So(err, ShouldBeNil)

				result, err := repo.GetAllSessionsByUserID(ctx, userId)
				So(err, ShouldBeNil)
				So(len(result), ShouldEqual, 1)

				for _, session := range result {
					assertRevokedSession(session, userId, res)
				}
			})
		})

		Convey("user has no sessions to revoke", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				err := repo.RevokeAllUserSessions(ctx, nonExistentUUID, nil, authdomain.RevokeReasonLogout)
				So(err, ShouldBeNil)
			})
		})

		Convey("invalid revoke reason", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				err := repo.RevokeAllUserSessions(ctx, nonExistentUUID, nil, authdomain.RevokeReason("bogus"))
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateFailedLoginAttempts_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("UpdateFailedLoginAttempts", t, func() {
		Convey("Successful test - not locked", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, seed, session := newSessionFixture(t, ctx, repo)

				err := repo.UpdateFailedLoginAttempts(ctx, session.ID, "15 minutes", 3)
				So(err, ShouldBeNil)

				got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				assertFailedLoginState(t, got, seed, 1, false)
			})
		})

		Convey("Successful test - locked", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, seed, session := newSessionFixture(t, ctx, repo)

				err := repo.UpdateFailedLoginAttempts(ctx, session.ID, "15 minutes", 1)
				So(err, ShouldBeNil)

				got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				assertFailedLoginState(t, got, seed, 1, true)
			})
		})

		Convey("Successful test - couple in row", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				_, seed, session := newSessionFixture(t, ctx, repo)

				err := repo.UpdateFailedLoginAttempts(ctx, session.ID, "15 minutes", 5)
				So(err, ShouldBeNil)

				err = repo.UpdateFailedLoginAttempts(ctx, session.ID, "15 minutes", 5)
				So(err, ShouldBeNil)

				got, err := repo.GetUserSessionByRefreshToken(ctx, seed.RefreshHash)
				So(err, ShouldBeNil)
				So(got, ShouldNotBeNil)
				assertFailedLoginState(t, got, seed, 2, false)
			})
		})

		Convey("session does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				err := repo.UpdateFailedLoginAttempts(ctx, nonExistentUUID, "15 minutes", 3)
				So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetActiveSessionsByUserID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("GetActiveSessionsByUserID", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId, seed, _ := newSessionFixture(t, ctx, repo)

				sessionList, err := repo.GetActiveSessionsByUserID(ctx, userId)
				So(err, ShouldBeNil)
				So(sessionList, ShouldNotBeNil)
				So(len(sessionList), ShouldEqual, 1)

				for _, session := range sessionList {
					assertFreshSession(session, seed)
				}
			})
		})

		Convey("Successful test - multiple sessions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				seeds := createSessions(t, ctx, repo, userId, 3)
				t.Cleanup(func() {
					repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
				})

				sessionList, err := repo.GetActiveSessionsByUserID(ctx, userId)
				So(err, ShouldBeNil)
				So(sessionList, ShouldNotBeNil)
				So(len(sessionList), ShouldEqual, 3)

				byHash := seedsByRefreshHash(seeds)
				for _, session := range sessionList {
					seed, ok := byHash[session.RefreshHash]
					So(ok, ShouldBeTrue)
					assertFreshSession(session, seed)
				}
			})
		})

		Convey("user does not have a session", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				sessionList, err := repo.GetActiveSessionsByUserID(ctx, nonExistentUUID)
				So(err, ShouldBeNil)
				So(sessionList, ShouldBeNil)
				So(len(sessionList), ShouldEqual, 0)
			})
		})
	})
}

func TestGetAllSessionsByUserID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("GetAllSessionsByUserID", t, func() {
		Convey("Successful test", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId, seed, session := newSessionFixture(t, ctx, repo)

				res := authdomain.RevokeReasonLogout
				err := repo.RevokeUserSession(ctx, session.ID, userId, res)
				So(err, ShouldBeNil)

				sessionList, err := repo.GetAllSessionsByUserID(ctx, userId)
				So(err, ShouldBeNil)
				So(sessionList, ShouldNotBeNil)
				So(len(sessionList), ShouldEqual, 1)

				for _, s := range sessionList {
					So(s.UserID, ShouldEqual, seed.UserID)
					So(s.RefreshHash, ShouldEqual, seed.RefreshHash)
					assertRevokedSession(s, userId, res)
				}
			})
		})

		Convey("Successful test - multiple sessions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)
				seeds := createSessions(t, ctx, repo, userId, 3)
				t.Cleanup(func() {
					repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
				})

				res := authdomain.RevokeReasonLogout
				err := repo.RevokeAllUserSessions(ctx, userId, &userId, res)
				So(err, ShouldBeNil)

				sessionList, err := repo.GetAllSessionsByUserID(ctx, userId)
				So(err, ShouldBeNil)
				So(sessionList, ShouldNotBeNil)
				So(len(sessionList), ShouldEqual, 3)

				byHash := seedsByRefreshHash(seeds)
				for _, session := range sessionList {
					_, ok := byHash[session.RefreshHash]
					So(ok, ShouldBeTrue)
					assertRevokedSession(session, userId, res)
				}
			})
		})

		Convey("user does not have a session", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				sessionList, err := repo.GetAllSessionsByUserID(ctx, nonExistentUUID)
				So(err, ShouldBeNil)
				So(sessionList, ShouldBeNil)
				So(len(sessionList), ShouldEqual, 0)
			})
		})
	})
}

func TestUpdateSessionToken_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("UpdateSessionToken", t, func() {
		Convey("Successful test - multiple sessions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				userId := newTestUser(t, ctx, repo)

				t.Cleanup(func() {
					repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", userId)
				})

				seed := fullTestSession(userId)

				for i := 0; i <= 3; i++ {
					session, err := repo.CreateUserSession(ctx, seed)
					So(err, ShouldBeNil)
					hash := fmt.Sprintf("session_%d_new_hash", i)
					ua := fmt.Sprintf("session_%d_user_agent", i)
					ip := fmt.Sprintf("session_%d_new_ip", i)
					err = repo.UpdateSessionToken(ctx, session.ID, hash, ua, ip)
					So(err, ShouldBeNil)

					gotSession, err := repo.GetUserSessionByRefreshToken(ctx, hash)
					So(err, ShouldBeNil)
					So(gotSession, ShouldNotBeNil)
					So(gotSession.RefreshHash, ShouldEqual, hash)
					So(gotSession.UserAgent, ShouldEqual, &ua)
					So(gotSession.LastSeenIP, ShouldEqual, &ip)
				}
			})
		})

		Convey("session does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				err := repo.UpdateSessionToken(ctx, nonExistentUUID, "new-hash", "new-ua", "new-ip")
				So(errors.Is(err, authdomain.ErrSessionNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestSessionIsolationBetweenUsers_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	Convey("Session queries do not leak across users", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{db: tx}

			userA := newTestUser(t, ctx, repo)
			userB := newTestUser(t, ctx, repo)

			seedA := fullTestSession(userA)
			_, err := repo.CreateUserSession(ctx, seedA)
			So(err, ShouldBeNil)

			seedB := fullTestSession(userB)
			sessionB, err := repo.CreateUserSession(ctx, seedB)
			So(err, ShouldBeNil)

			t.Cleanup(func() {
				repo.queryRunner(ctx).Exec(ctx, "DELETE FROM user_sessions WHERE user_id IN ($1,$2)", userA, userB)
			})

			err = repo.RevokeAllUserSessions(ctx, userA, &userA, authdomain.RevokeReasonLogout)
			So(err, ShouldBeNil)

			gotA, err := repo.GetAllSessionsByUserID(ctx, userA)
			So(err, ShouldBeNil)
			So(gotA[0].RevokedAt, ShouldNotBeNil)

			gotB, err := repo.GetAllSessionsByUserID(ctx, userB)
			So(err, ShouldBeNil)
			So(gotB[0].RevokedAt, ShouldBeNil)
			So(gotB[0].ID, ShouldEqual, sessionB.ID)

			activeA, err := repo.GetActiveSessionsByUserID(ctx, userA)
			So(err, ShouldBeNil)
			So(activeA, ShouldBeEmpty)

			activeB, err := repo.GetActiveSessionsByUserID(ctx, userB)
			So(err, ShouldBeNil)
			So(len(activeB), ShouldEqual, 1)
			So(activeB[0].UserID, ShouldEqual, userB)
		})
	})
}
