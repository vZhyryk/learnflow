//go:build integration

package adminrepository

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/audit"
	auditdomain "learnflow_backend/internal/audit/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

const insertUserProfileSQL = `
	INSERT INTO user_profiles (user_id, first_name, last_name, date_of_birth)
	VALUES ($1, 'Ada', 'Lovelace', '1990-05-17')
`

func insertProfiledUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()

	userID := testutil.InsertRandomTestUser(t, tx)
	if _, err := tx.Exec(context.Background(), insertUserProfileSQL, userID); err != nil {
		t.Fatalf("insert user profile: %v", err)
	}

	return userID
}

func TestGetUserDataByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When the user has a profile with a date of birth", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				userID := insertProfiledUser(t, tx)

				got, err := repo.GetUserDataByID(ctx, userID)

				So(err, ShouldBeNil)
				So(got.UserID, ShouldEqual, userID)
				So(*got.FirstName, ShouldEqual, "Ada")
				So(*got.DateOfBirth, ShouldEqual, "1990-05-17")
				So(got.DeletedAt, ShouldBeNil)
				So(got.Role, ShouldEqual, admindomain.RoleUser)
			})
		})

		Convey("When the user has no profile row", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				userID := testutil.InsertRandomTestUser(t, tx)

				got, err := repo.GetUserDataByID(ctx, userID)

				So(err, ShouldBeNil)
				So(got.FirstName, ShouldBeNil)
				So(got.DateOfBirth, ShouldBeNil)
			})
		})

		Convey("When the user does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetUserDataByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, admindomain.ErrUserNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetUsersData_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When listing with LIMIT/OFFSET", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				insertProfiledUser(t, tx)
				insertProfiledUser(t, tx)
				insertProfiledUser(t, tx)

				page1, total, err := repo.GetUsersData(ctx, pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				So(page1, ShouldHaveLength, 2)
				So(total, ShouldBeGreaterThanOrEqualTo, 3)

				pageBeyond, _, err := repo.GetUsersData(ctx, pagination.NewParams(total, 1))
				So(err, ShouldBeNil)
				So(pageBeyond, ShouldHaveLength, 1)

				empty, _, err := repo.GetUsersData(ctx, pagination.NewParams(total+1, 1))
				So(err, ShouldBeNil)
				So(empty, ShouldBeEmpty)
			})
		})
	})
}

func TestChangeUserStatusAndRole_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("Block, unblock, delete and restore follow the status transitions", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				userID := testutil.InsertRandomTestUser(t, tx)

				So(repo.BlockUser(ctx, userID), ShouldBeNil)
				So(errors.Is(repo.BlockUser(ctx, userID), admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(repo.UnBlockUser(ctx, userID), ShouldBeNil)
				So(repo.DeleteUser(ctx, userID), ShouldBeNil)
				So(repo.RestoreUser(ctx, userID), ShouldBeNil)

				got, err := repo.GetUserDataByID(ctx, userID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, admindomain.StatusActive)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("Role assignment only moves user to subadmin and back", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				userID := testutil.InsertRandomTestUser(t, tx)

				So(repo.AssignUserRole(ctx, userID), ShouldBeNil)
				So(errors.Is(repo.AssignUserRole(ctx, userID), admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(repo.RevokeUserRole(ctx, userID), ShouldBeNil)
				So(errors.Is(repo.RevokeUserRole(ctx, userID), admindomain.ErrInvalidUserState), ShouldBeTrue)
			})
		})
	})
}

func TestAdminProtectionAndAudit_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("Admin accounts are never touched", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				adminID := testutil.InsertRandomTestUser(t, tx)
				if _, err := tx.Exec(ctx, "UPDATE users SET role = 'admin' WHERE id = $1", adminID); err != nil {
					t.Fatalf("promote to admin: %v", err)
				}

				So(errors.Is(repo.BlockUser(ctx, adminID), admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(errors.Is(repo.DeleteUser(ctx, adminID), admindomain.ErrInvalidUserState), ShouldBeTrue)
				So(errors.Is(repo.BlockUser(ctx, "00000000-0000-0000-0000-000000000000"), admindomain.ErrUserNotFound), ShouldBeTrue)
			})
		})

		Convey("Admin audit entries are persisted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &audit.Audit{BaseRepository: repository.BaseRepository{DB: tx}}
				adminID := testutil.InsertRandomTestUser(t, tx)
				targetID := testutil.InsertRandomTestUser(t, tx)

				err := repo.CreateAdminAction(ctx, &auditdomain.AdminAction{
					AdminUserID: adminID,
					ActionType:  auditdomain.ActionUnblockUser,
					TargetType:  auditdomain.TargetUser,
					TargetID:    targetID,
				})
				So(err, ShouldBeNil)

				var count int
				row := tx.QueryRow(ctx, "SELECT COUNT(*) FROM admin_actions WHERE admin_user_id = $1 AND target_id = $2 AND action_type = 'unblock_user'", adminID, targetID)
				So(row.Scan(&count), ShouldBeNil)
				So(count, ShouldEqual, 1)
			})
		})
	})
}
