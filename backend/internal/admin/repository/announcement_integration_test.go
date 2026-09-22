//go:build integration

package adminrepository

import (
	"context"
	"errors"
	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

// draftAnnouncement returns a platform-wide (no entity) Announcement seed, ready for
// CreateAnnouncement — mirrors the shape a real CreateAnnouncementRequest produces.
func draftAnnouncement(t *testing.T, tx pgx.Tx) *admindomain.Announcement {
	t.Helper()

	return &admindomain.Announcement{
		Title:           "Integration Test Announcement",
		Body:            "Integration test announcement body.",
		CreatedByUserID: testutil.InsertRandomTestUser(t, tx),
		ExpiresAt:       time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		Channels:        []admindomain.Channel{admindomain.EmailChannel},
	}
}

func TestCreateAnnouncement_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When creating a platform-wide announcement (no entity)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)

				got, err := repo.CreateAnnouncement(ctx, seed)

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.Title, ShouldEqual, seed.Title)
				So(got.CreatedByUserID, ShouldEqual, seed.CreatedByUserID)
				So(got.EntityID, ShouldBeNil)
				So(got.EntityType, ShouldBeNil)
				So(got.ApprovedAt, ShouldBeNil)
				So(got.UpdatedAt, ShouldBeNil)
				So(got.Channels, ShouldResemble, seed.Channels)
			})
		})

		Convey("When creating an entity-scoped announcement", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				entityID := "11111111-1111-1111-1111-111111111111"
				entityType := admindomain.CourseEntityType
				seed.EntityID = &entityID
				seed.EntityType = &entityType

				got, err := repo.CreateAnnouncement(ctx, seed)

				So(err, ShouldBeNil)
				So(*got.EntityID, ShouldEqual, entityID)
				So(*got.EntityType, ShouldEqual, entityType)
			})
		})

		Convey("When entity_type is set without entity_id (pairing CHECK)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				entityType := admindomain.CourseEntityType
				seed.EntityType = &entityType
				seed.EntityID = nil

				_, err := repo.CreateAnnouncement(ctx, seed)

				So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
			})
		})

		Convey("When entity_id is set without entity_type (pairing CHECK)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				entityID := "11111111-1111-1111-1111-111111111111"
				seed.EntityID = &entityID
				seed.EntityType = nil

				_, err := repo.CreateAnnouncement(ctx, seed)

				So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
			})
		})

		Convey("When a channel value is not in the allowed set (DB-level CHECK, defense-in-depth below Go validation)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				seed.Channels = []admindomain.Channel{"sms"}

				_, err := repo.CreateAnnouncement(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
				So(pgErr.ConstraintName, ShouldEqual, "announcements_channels_valid")
			})
		})

		Convey("When created_by_user_id does not reference an existing user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				seed.CreatedByUserID = "00000000-0000-0000-0000-000000000000"

				_, err := repo.CreateAnnouncement(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})
	})
}

func TestGetAnnouncementByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When the announcement exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
				So(err, ShouldBeNil)

				got, err := repo.GetAnnouncementByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no announcement exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				_, err := repo.GetAnnouncementByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateAnnouncement_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When updating an existing announcement", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
				So(err, ShouldBeNil)

				updatedByUserID := created.CreatedByUserID
				created.Title = "Updated Title"
				created.UpdatedByUserID = &updatedByUserID

				err = repo.UpdateAnnouncement(ctx, created)
				So(err, ShouldBeNil)

				got, err := repo.GetAnnouncementByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Title, ShouldEqual, "Updated Title")
				So(got.UpdatedAt, ShouldNotBeNil)
				So(*got.UpdatedByUserID, ShouldEqual, updatedByUserID)
			})
		})

		Convey("When the expiry is changed", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
				So(err, ShouldBeNil)

				updatedByUserID := created.CreatedByUserID
				created.UpdatedByUserID = &updatedByUserID
				created.ExpiresAt = time.Now().Add(72 * time.Hour).UTC().Truncate(time.Second)

				err = repo.UpdateAnnouncement(ctx, created)
				So(err, ShouldBeNil)

				got, err := repo.GetAnnouncementByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.ExpiresAt.Equal(created.ExpiresAt), ShouldBeTrue)
			})
		})

		Convey("When the announcement does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				ghost := draftAnnouncement(t, tx)
				ghost.ID = "00000000-0000-0000-0000-000000000000"

				err := repo.UpdateAnnouncement(ctx, ghost)

				So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
			})
		})

		Convey("When the update would violate the entity pairing CHECK", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
				So(err, ShouldBeNil)

				entityType := admindomain.CourseEntityType
				created.EntityType = &entityType
				created.EntityID = nil

				err = repo.UpdateAnnouncement(ctx, created)

				So(errors.Is(err, admindomain.ErrEntityDataMisMatch), ShouldBeTrue)
			})
		})
	})
}

func TestApproveAnnouncement_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given an admin repository backed by real Postgres", t, func() {
		Convey("When approving an existing, non-expired announcement", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				created, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
				So(err, ShouldBeNil)
				approver := testutil.InsertRandomTestUser(t, tx)

				err = repo.ApproveAnnouncement(ctx, created.ID, approver)
				So(err, ShouldBeNil)

				got, err := repo.GetAnnouncementByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.ApprovedAt, ShouldNotBeNil)
				So(*got.ApprovedByUserID, ShouldEqual, approver)
			})
		})

		Convey("When the announcement does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}

				err := repo.ApproveAnnouncement(ctx, "00000000-0000-0000-0000-000000000000", testutil.InsertRandomTestUser(t, tx))

				So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
			})
		})

		Convey("When the announcement has already expired", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				seed := draftAnnouncement(t, tx)
				seed.ExpiresAt = time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
				created, err := repo.CreateAnnouncement(ctx, seed)
				So(err, ShouldBeNil)

				err = repo.ApproveAnnouncement(ctx, created.ID, testutil.InsertRandomTestUser(t, tx))

				So(errors.Is(err, admindomain.ErrAnnouncementNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetAnnouncementsByStatus_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given unapproved, approved, and expired announcements", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{repository.BaseRepository{DB: tx}}

			unapproved, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
			So(err, ShouldBeNil)

			approved, err := repo.CreateAnnouncement(ctx, draftAnnouncement(t, tx))
			So(err, ShouldBeNil)
			So(repo.ApproveAnnouncement(ctx, approved.ID, testutil.InsertRandomTestUser(t, tx)), ShouldBeNil)

			expiredSeed := draftAnnouncement(t, tx)
			expiredSeed.ExpiresAt = time.Now().Add(1 * time.Second).UTC()
			expired, err := repo.CreateAnnouncement(ctx, expiredSeed)
			So(err, ShouldBeNil)
			approver := testutil.InsertRandomTestUser(t, tx)
			So(repo.ApproveAnnouncement(ctx, expired.ID, approver), ShouldBeNil)
			// force it into the past now that it's approved, bypassing the
			// "expires_at > now()" WHERE clause on ApproveAnnouncement itself.
			_, err = tx.Exec(ctx, `UPDATE announcements SET expires_at = now() - interval '1 hour' WHERE id = $1`, expired.ID)
			So(err, ShouldBeNil)

			params := pagination.NewParams(1, 100)

			Convey("GetUnApprovedAnnouncements returns only the unapproved, non-expired one", func() {
				got, err := repo.GetUnApprovedAnnouncements(ctx, params)
				So(err, ShouldBeNil)
				ids := announcementIDs(got)
				So(ids, ShouldContain, unapproved.ID)
				So(ids, ShouldNotContain, approved.ID)
				So(ids, ShouldNotContain, expired.ID)
			})

			Convey("GetApprovedAnnouncements returns only the approved, non-expired one", func() {
				got, err := repo.GetApprovedAnnouncements(ctx, params)
				So(err, ShouldBeNil)
				ids := announcementIDs(got)
				So(ids, ShouldContain, approved.ID)
				So(ids, ShouldNotContain, unapproved.ID)
				So(ids, ShouldNotContain, expired.ID)
			})

			Convey("GetExpiredAnnouncements returns only the approved-but-expired one", func() {
				got, err := repo.GetExpiredAnnouncements(ctx, params)
				So(err, ShouldBeNil)
				ids := announcementIDs(got)
				So(ids, ShouldContain, expired.ID)
				So(ids, ShouldNotContain, unapproved.ID)
				So(ids, ShouldNotContain, approved.ID)
			})

			Convey("GetAnnouncements returns all three regardless of status", func() {
				got, err := repo.GetAnnouncements(ctx, params)
				So(err, ShouldBeNil)
				ids := announcementIDs(got)
				So(ids, ShouldContain, unapproved.ID)
				So(ids, ShouldContain, approved.ID)
				So(ids, ShouldContain, expired.ID)
			})
		})
	})
}

func announcementIDs(items []*admindomain.Announcement) []string {
	ids := make([]string, 0, len(items))
	for _, a := range items {
		ids = append(ids, a.ID)
	}
	return ids
}
