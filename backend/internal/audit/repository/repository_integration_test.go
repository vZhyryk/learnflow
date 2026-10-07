//go:build integration

package auditrepository

import (
	"context"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func newIntegrationRepo(tx pgx.Tx) *Audit {
	return &Audit{BaseRepository: repository.BaseRepository{DB: tx}}
}

func TestGetInstanceAdminActions_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an audit repository backed by real Postgres", t, func() {
		Convey("When the admin has no profile the name falls back to the email", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				email := testutil.RandomTestEmail(t, "audit-admin")
				adminID := testutil.InsertTestUser(t, tx, email)
				targetID := testutil.InsertRandomTestUser(t, tx)
				So(repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
					AdminUserID: adminID, ActionType: auditdomain.ActionBlockUser,
					TargetType: auditdomain.TargetUser, TargetID: targetID,
				}), ShouldBeNil)

				actions, total, err := repo.GetInstanceAdminActions(ctx, auditdomain.TargetUser, targetID, pagination.NewParams(1, 20))

				So(err, ShouldBeNil)
				So(total, ShouldEqual, 1)
				So(actions, ShouldHaveLength, 1)
				So(actions[0].AdminUserID, ShouldEqual, adminID)
				So(actions[0].AdminName, ShouldEqual, email)
				So(actions[0].ActionType, ShouldEqual, auditdomain.ActionBlockUser)
				So(actions[0].Details, ShouldBeNil)
				So(actions[0].CreatedAt.IsZero(), ShouldBeFalse)
			})
		})

		Convey("When the admin has a profile the first name is used and details round-trip", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				adminID := testutil.InsertRandomTestUser(t, tx)
				_, err := tx.Exec(ctx, `INSERT INTO user_profiles (user_id, first_name) VALUES ($1, 'Ada')`, adminID)
				So(err, ShouldBeNil)
				targetID := testutil.InsertRandomTestUser(t, tx)
				So(repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
					AdminUserID: adminID, ActionType: auditdomain.ActionGrantItemAccess,
					TargetType: auditdomain.TargetUser, TargetID: targetID,
					Details: map[string]any{"course_id": "course-1"},
				}), ShouldBeNil)

				actions, _, err := repo.GetInstanceAdminActions(ctx, auditdomain.TargetUser, targetID, pagination.NewParams(1, 20))

				So(err, ShouldBeNil)
				So(actions, ShouldHaveLength, 1)
				So(actions[0].AdminName, ShouldEqual, "Ada")
				So(actions[0].Details, ShouldResemble, map[string]any{"course_id": "course-1"})
			})
		})

		Convey("It returns only the target's entries and pages them with the full total", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				adminID := testutil.InsertRandomTestUser(t, tx)
				targetID := testutil.InsertRandomTestUser(t, tx)
				otherID := testutil.InsertRandomTestUser(t, tx)
				for _, target := range []string{targetID, targetID, targetID, otherID} {
					So(repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
						AdminUserID: adminID, ActionType: auditdomain.ActionBlockUser,
						TargetType: auditdomain.TargetUser, TargetID: target,
					}), ShouldBeNil)
				}

				firstPage, total, err := repo.GetInstanceAdminActions(ctx, auditdomain.TargetUser, targetID, pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				secondPage, _, err := repo.GetInstanceAdminActions(ctx, auditdomain.TargetUser, targetID, pagination.NewParams(2, 2))
				So(err, ShouldBeNil)

				So(total, ShouldEqual, 3)
				So(firstPage, ShouldHaveLength, 2)
				So(secondPage, ShouldHaveLength, 1)
				So(secondPage[0].ID, ShouldNotEqual, firstPage[0].ID)
				So(secondPage[0].ID, ShouldNotEqual, firstPage[1].ID)
			})
		})

		Convey("When the target has no entries it returns an empty, non-nil page", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)

				actions, total, err := repo.GetInstanceAdminActions(ctx, auditdomain.TargetUser, testutil.InsertRandomTestUser(t, tx), pagination.NewParams(1, 20))

				So(err, ShouldBeNil)
				So(total, ShouldEqual, 0)
				So(actions, ShouldNotBeNil)
				So(actions, ShouldBeEmpty)
			})
		})
	})
}

func TestGetFailedJobs_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an audit repository backed by real Postgres", t, func() {
		Convey("It returns the failed job without its payload and counts the table", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				var before int
				So(tx.QueryRow(ctx, `SELECT COUNT(*) FROM failed_jobs`).Scan(&before), ShouldBeNil)
				failedAt := time.Now().UTC().Truncate(time.Second)
				var id string
				So(tx.QueryRow(ctx, `
					INSERT INTO failed_jobs (event_type, queue_name, payload_json, attempt_count, error_message, failed_at)
					VALUES ('user.blocked', 'email', '{"raw_token":"secret"}', 3, 'smtp: connection refused', $1)
					RETURNING id`, failedAt).Scan(&id), ShouldBeNil)

				jobs, total, err := repo.GetFailedJobs(ctx, pagination.NewParams(1, 1000))

				So(err, ShouldBeNil)
				So(total, ShouldEqual, before+1)
				var got *auditdomain.FailedJob
				for _, job := range jobs {
					if job.ID == id {
						got = job
					}
				}
				So(got, ShouldNotBeNil)
				So(got.EventType, ShouldEqual, "user.blocked")
				So(got.QueueName, ShouldEqual, "email")
				So(got.AttemptCount, ShouldEqual, 3)
				So(*got.ErrorMessage, ShouldEqual, "smtp: connection refused")
				So(got.FailedAt.Equal(failedAt), ShouldBeTrue)
				So(got.ResolvedAt, ShouldBeNil)
			})
		})
	})
}

func TestCreateAdminActionAndWasDeletedByAdmin_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an audit repository backed by real Postgres", t, func() {
		Convey("A nil Details is stored as SQL NULL, not JSON null", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				adminID := testutil.InsertRandomTestUser(t, tx)
				targetID := testutil.InsertRandomTestUser(t, tx)
				So(repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
					AdminUserID: adminID, ActionType: auditdomain.ActionBlockUser,
					TargetType: auditdomain.TargetUser, TargetID: targetID,
				}), ShouldBeNil)

				var isNull bool
				So(tx.QueryRow(ctx, `SELECT details_json IS NULL FROM admin_actions WHERE target_id = $1`, targetID).Scan(&isNull), ShouldBeNil)

				So(isNull, ShouldBeTrue)
			})
		})

		Convey("WasDeletedByAdmin follows the latest delete or restore", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := newIntegrationRepo(tx)
				adminID := testutil.InsertRandomTestUser(t, tx)
				targetID := testutil.InsertRandomTestUser(t, tx)
				write := func(action auditdomain.AdminActionType) {
					So(repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
						AdminUserID: adminID, ActionType: action,
						TargetType: auditdomain.TargetUser, TargetID: targetID,
					}), ShouldBeNil)
				}

				deleted, err := repo.WasDeletedByAdmin(ctx, targetID)
				So(err, ShouldBeNil)
				So(deleted, ShouldBeFalse)

				write(auditdomain.ActionDeleteUser)
				deleted, err = repo.WasDeletedByAdmin(ctx, targetID)
				So(err, ShouldBeNil)
				So(deleted, ShouldBeTrue)
			})
		})
	})
}
