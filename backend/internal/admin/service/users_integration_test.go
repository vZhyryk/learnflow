//go:build integration

package adminservice

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	adminrepository "learnflow_backend/internal/admin/repository"
	auditdomain "learnflow_backend/internal/audit/domain"
	auditrepository "learnflow_backend/internal/audit/repository"
	authrepository "learnflow_backend/internal/auth/repository"
	"learnflow_backend/internal/events"
	"learnflow_backend/internal/infrastructure/db"
	redisinfra "learnflow_backend/internal/infrastructure/redis"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

type userOpFixture struct {
	srv      *Service
	pool     *pgxpool.Pool
	authRepo *authrepository.Repository
	redis    *redisinfra.Instance
	adminID  string
	targetID string
}

// newUserOpFixture wires the real service (Postgres + Redis) and inserts committed rows,
// removed on cleanup: the service opens its own transaction, so a rolled-back test tx can't be used.
func newUserOpFixture(t *testing.T, blocklist *redisinfra.Instance) *userOpFixture {
	t.Helper()

	pool := testutil.NewTestPool(t)
	realRedis := redisinfra.NewInstance(redis.NewClient(&redis.Options{Addr: "localhost:6379"}))
	t.Cleanup(func() { realRedis.Close() }) //nolint:errcheck // Close's error is never actionable in test cleanup

	if blocklist == nil {
		blocklist = realRedis
	}

	ctx := context.Background()
	adminID := testutil.InsertTestUser(t, pool, testutil.RandomTestEmail(t, "svc-admin"))
	targetID := testutil.InsertTestUser(t, pool, testutil.RandomTestEmail(t, "svc-target"))
	mustExec(t, pool, "UPDATE users SET role = 'admin' WHERE id = $1", adminID)
	mustExec(t, pool, "INSERT INTO user_profiles (user_id, first_name) VALUES ($1, 'Ada')", targetID)
	mustExec(t, pool, "INSERT INTO user_sessions (user_id, refresh_hash, expires_at) VALUES ($1, $2, now() + interval '1 day')",
		targetID, "svc-refresh-"+targetID)

	t.Cleanup(func() {
		realRedis.UnBlockUser(ctx, targetID)          //nolint:errcheck // best-effort cleanup
		realRedis.ClearUserRoleRevoked(ctx, targetID) //nolint:errcheck // best-effort cleanup
		testutil.DeleteAdminActionsByTarget(t, pool, targetID)
		for _, q := range []string{
			"DELETE FROM event_outbox WHERE aggregate_id = $1",
			"DELETE FROM user_sessions WHERE user_id = $1",
			"DELETE FROM user_profiles WHERE user_id = $1",
			"DELETE FROM users WHERE id = $1",
		} {
			pool.Exec(ctx, q, targetID) //nolint:errcheck // best-effort cleanup of throwaway test data
		}
		pool.Exec(ctx, "DELETE FROM users WHERE id = $1", adminID) //nolint:errcheck // best-effort cleanup
	})

	adminRepo := adminrepository.NewRepository(pool)
	authRepo := authrepository.NewRepository(pool)
	srv := New(
		Repos{AnnounRepo: adminRepo, UserRepo: adminRepo, ActionRepo: auditrepository.New(pool), SessionRepo: authRepo},
		Utils{Transactor: db.NewTransactor(pool), Outbox: events.NewOutboxWriter(pool), Blocklist: blocklist},
	)

	return &userOpFixture{srv: srv, pool: pool, authRepo: authRepo, redis: realRedis, adminID: adminID, targetID: targetID}
}

func mustExec(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func (f *userOpFixture) count(t *testing.T, query string, args ...any) int {
	t.Helper()

	var n int
	if err := f.pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}

	return n
}

func (f *userOpFixture) outboxEvents(t *testing.T) (eventTypes, eventIDs []string) {
	t.Helper()

	rows, err := f.pool.Query(context.Background(),
		"SELECT event_type, payload_json->>'event_id' FROM event_outbox WHERE aggregate_id = $1 ORDER BY created_at, id", f.targetID)
	if err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType, eventID string
		if err := rows.Scan(&eventType, &eventID); err != nil {
			t.Fatalf("scan outbox: %v", err)
		}
		eventTypes, eventIDs = append(eventTypes, eventType), append(eventIDs, eventID)
	}

	return eventTypes, eventIDs
}

func TestBlockUserEndToEnd_Integration(t *testing.T) {
	Convey("Given the real service on Postgres and Redis", t, func() {
		f := newUserOpFixture(t, nil)
		ctx := context.Background()

		Convey("When an admin blocks a user, the status, audit, outbox event, sessions and Redis mark all change together", func() {
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "BlockUser"), ShouldBeNil)

			data, err := adminrepository.NewRepository(f.pool).GetUserDataByID(ctx, f.targetID)
			So(err, ShouldBeNil)
			So(data.Status, ShouldEqual, admindomain.StatusBlocked)

			So(f.count(t, "SELECT count(*) FROM admin_actions WHERE target_id = $1 AND action_type = $2",
				f.targetID, auditdomain.ActionBlockUser), ShouldEqual, 1)

			eventTypes, eventIDs := f.outboxEvents(t)
			So(eventTypes, ShouldResemble, []string{string(events.EventUserBlocked)})
			So(eventIDs[0], ShouldNotBeEmpty)

			var email, name string
			So(f.pool.QueryRow(ctx, "SELECT payload_json->>'email', payload_json->>'user_name' FROM event_outbox WHERE aggregate_id = $1",
				f.targetID).Scan(&email, &name), ShouldBeNil)
			So(email, ShouldEqual, *data.Email)
			So(name, ShouldEqual, "Ada")

			active, err := f.authRepo.GetActiveSessionsByUserID(ctx, f.targetID)
			So(err, ShouldBeNil)
			So(active, ShouldBeEmpty)

			blocked, err := f.redis.IsBlocked(ctx, f.targetID, "unused-jti", time.Now())
			So(err, ShouldBeNil)
			So(blocked, ShouldBeTrue)
		})

		Convey("When the same user is blocked, unblocked and blocked again, every action gets its own event id", func() {
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "BlockUser"), ShouldBeNil)
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "UnBlockUser"), ShouldBeNil)
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "BlockUser"), ShouldBeNil)

			eventTypes, eventIDs := f.outboxEvents(t)
			So(eventTypes, ShouldResemble, []string{"user.blocked", "user.unblocked", "user.blocked"})
			So(eventIDs[0], ShouldNotEqual, eventIDs[2])
		})

		Convey("When the user is deleted and restored, the events follow and the Redis mark is cleared", func() {
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "DeleteUser"), ShouldBeNil)

			blocked, err := f.redis.IsBlocked(ctx, f.targetID, "unused-jti", time.Now())
			So(err, ShouldBeNil)
			So(blocked, ShouldBeTrue)

			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "RestoreUser"), ShouldBeNil)

			blocked, err = f.redis.IsBlocked(ctx, f.targetID, "unused-jti", time.Now())
			So(err, ShouldBeNil)
			So(blocked, ShouldBeFalse)

			eventTypes, _ := f.outboxEvents(t)
			So(eventTypes, ShouldResemble, []string{"user.deleted", "user.restored"})
		})
	})
}

func TestBlockUserRollsBackWhenRedisFails_Integration(t *testing.T) {
	Convey("Given the real service whose Redis blocklist is unreachable", t, func() {
		f := newUserOpFixture(t, testutil.UnreachableRedisInstance())
		ctx := context.Background()

		Convey("When blocking fails after the DB writes, the transaction rolls back", func() {
			err := f.srv.ChangeUserField(ctx, f.targetID, f.adminID, "BlockUser")

			So(err, ShouldNotBeNil)

			data, getErr := adminrepository.NewRepository(f.pool).GetUserDataByID(ctx, f.targetID)
			So(getErr, ShouldBeNil)
			So(data.Status, ShouldEqual, admindomain.StatusActive)
			So(f.count(t, "SELECT count(*) FROM admin_actions WHERE target_id = $1", f.targetID), ShouldEqual, 0)
			So(f.count(t, "SELECT count(*) FROM event_outbox WHERE aggregate_id = $1", f.targetID), ShouldEqual, 0)

			active, sessErr := f.authRepo.GetActiveSessionsByUserID(ctx, f.targetID)
			So(sessErr, ShouldBeNil)
			So(active, ShouldHaveLength, 1)
		})
	})
}

func TestSubAdminRoleEndToEnd_Integration(t *testing.T) {
	Convey("Given the real service on Postgres and Redis", t, func() {
		f := newUserOpFixture(t, nil)
		ctx := context.Background()
		role := func() string {
			var r string
			So(f.pool.QueryRow(ctx, "SELECT role FROM users WHERE id = $1", f.targetID).Scan(&r), ShouldBeNil)
			return r
		}

		Convey("When an admin assigns and then revokes the subadmin role, the role, audit, sessions and Redis mark follow", func() {
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, admindomain.AssignUserRole), ShouldBeNil)
			So(role(), ShouldEqual, "subadmin")
			So(f.count(t, "SELECT count(*) FROM admin_actions WHERE target_id = $1 AND action_type = $2",
				f.targetID, auditdomain.ActionAssignSubadmin), ShouldEqual, 1)

			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, admindomain.RevokeUserRole), ShouldBeNil)
			So(role(), ShouldEqual, "user")
			So(f.count(t, "SELECT count(*) FROM admin_actions WHERE target_id = $1 AND action_type = $2",
				f.targetID, auditdomain.ActionRevokeSubadmin), ShouldEqual, 1)

			active, err := f.authRepo.GetActiveSessionsByUserID(ctx, f.targetID)
			So(err, ShouldBeNil)
			So(active, ShouldBeEmpty)

			revoked, err := f.redis.IsUserRoleRevoked(ctx, f.targetID)
			So(err, ShouldBeNil)
			So(revoked, ShouldBeTrue)
		})

		Convey("When the role is revoked from a user who is not a subadmin, nothing is audited", func() {
			err := f.srv.ChangeUserField(ctx, f.targetID, f.adminID, admindomain.RevokeUserRole)

			So(errors.Is(err, admindomain.ErrInvalidUserState), ShouldBeTrue)
			So(f.count(t, "SELECT count(*) FROM admin_actions WHERE target_id = $1", f.targetID), ShouldEqual, 0)
		})
	})
}

func TestAdminReadsEndToEnd_Integration(t *testing.T) {
	Convey("Given the real service on Postgres and Redis", t, func() {
		f := newUserOpFixture(t, nil)
		ctx := context.Background()

		Convey("GetInstanceAdminActions returns the action written by ChangeUserField with the admin's name", func() {
			So(f.srv.ChangeUserField(ctx, f.targetID, f.adminID, admindomain.BlockUser), ShouldBeNil)

			actions, total, err := f.srv.GetInstanceAdminActions(ctx, f.adminID, auditdomain.TargetUser, f.targetID, pagination.NewParams(1, 10))

			So(err, ShouldBeNil)
			So(total, ShouldEqual, 1)
			So(actions, ShouldHaveLength, 1)
			So(actions[0].ActionType, ShouldEqual, auditdomain.ActionBlockUser)
			So(actions[0].AdminUserID, ShouldEqual, f.adminID)
			So(actions[0].AdminName, ShouldNotBeEmpty)
		})

		Convey("GetFailedJobs lists a dead-lettered job", func() {
			eventType := "svc-test-" + f.targetID
			mustExec(t, f.pool, `INSERT INTO failed_jobs (event_type, queue_name, payload_json, attempt_count, error_message)
				VALUES ($1, 'svc-test', '{}', 3, 'boom')`, eventType)
			t.Cleanup(func() { f.pool.Exec(ctx, "DELETE FROM failed_jobs WHERE event_type = $1", eventType) }) //nolint:errcheck // best-effort cleanup

			jobs, total, err := f.srv.GetFailedJobs(ctx, pagination.NewParams(1, 100))

			So(err, ShouldBeNil)
			So(total, ShouldBeGreaterThanOrEqualTo, 1)
			var found bool
			for _, job := range jobs {
				if job.EventType == eventType {
					found = true
					So(job.AttemptCount, ShouldEqual, 3)
					So(*job.ErrorMessage, ShouldEqual, "boom")
				}
			}
			So(found, ShouldBeTrue)
		})
	})
}
