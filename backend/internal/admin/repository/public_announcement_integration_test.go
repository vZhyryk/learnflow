//go:build integration

package adminrepository

import (
	"context"
	"testing"
	"time"

	admindomain "learnflow_backend/internal/admin/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/repository"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	. "github.com/smartystreets/goconvey/convey"
)

type publicSeed struct {
	channels   []admindomain.Channel
	entityType *admindomain.EntityType
	entityID   *string
	unapproved bool
}

func seedPublicAnnouncement(t *testing.T, ctx context.Context, tx pgx.Tx, repo *Repository, s publicSeed) string {
	t.Helper()

	seed := draftAnnouncement(t, tx)
	seed.Channels = []admindomain.Channel{admindomain.BannerChannel}
	if s.channels != nil {
		seed.Channels = s.channels
	}
	seed.EntityType, seed.EntityID = s.entityType, s.entityID

	created, err := repo.CreateAnnouncement(ctx, seed)
	if err != nil {
		t.Fatalf("seedPublicAnnouncement create: %v", err)
	}
	if s.unapproved {
		return created.ID
	}
	if err := repo.ApproveAnnouncement(ctx, created.ID, testutil.InsertRandomTestUser(t, tx)); err != nil {
		t.Fatalf("seedPublicAnnouncement approve: %v", err)
	}
	return created.ID
}

func visiblePublicIDs(t *testing.T, ctx context.Context, repo *Repository, userID string) []string {
	t.Helper()

	list, err := repo.GetPublicAnnouncements(ctx, pagination.NewParams(1, 50), userID)
	if err != nil {
		t.Fatalf("visiblePublicIDs: %v", err)
	}
	ids := make([]string, 0, len(list))
	for _, a := range list {
		ids = append(ids, a.ID)
	}
	return ids
}

func entityRef(et admindomain.EntityType, id string) (*admindomain.EntityType, *string) {
	return &et, &id
}

func TestGetPublicAnnouncements_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given approved announcements backed by real Postgres", t, func() {
		Convey("When an announcement is platform-wide with the banner channel", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldContain, id)
			})
		})

		Convey("When an announcement has no banner channel", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{channels: []admindomain.Channel{admindomain.EmailChannel}})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldNotContain, id)
			})
		})

		Convey("When an announcement is not approved", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{unapproved: true})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldNotContain, id)
			})
		})

		Convey("When an announcement is scoped to an article", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				et, eid := entityRef(admindomain.ArticleEntityType, "00000000-0000-0000-0000-000000000001")
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldContain, id)
			})
		})
	})
}

func TestGetPublicAnnouncementsCourseAccess_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a course-scoped announcement backed by real Postgres", t, func() {
		Convey("When the user has active course access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})
				et, eid := entityRef(admindomain.CourseEntityType, courseID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, userID), ShouldContain, id)
			})
		})

		Convey("When the user has no access to the course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				courseID := testutil.InsertTestCourse(t, tx)
				et, eid := entityRef(admindomain.CourseEntityType, courseID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldNotContain, id)
			})
		})

		Convey("When the user's course access is revoked", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{Status: "revoked"})
				et, eid := entityRef(admindomain.CourseEntityType, courseID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, userID), ShouldNotContain, id)
			})
		})

		Convey("When the user's course access has expired", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				courseID := testutil.InsertTestCourse(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				created := time.Now().Add(-3 * 24 * time.Hour)
				granted := time.Now().Add(-2 * 24 * time.Hour)
				expires := time.Now().Add(-24 * time.Hour)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{
					CreatedAt: &created, GrantedAt: &granted, ExpiresAt: &expires,
				})
				et, eid := entityRef(admindomain.CourseEntityType, courseID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, userID), ShouldNotContain, id)
			})
		})
	})
}

func TestGetPublicAnnouncementsContentAccess_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a content-scoped announcement backed by real Postgres", t, func() {
		Convey("When the user has direct content access", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				contentID := testutil.InsertTestContentItem(t, tx)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantContentAccess(t, tx, userID, contentID, testutil.AccessGrant{})
				et, eid := entityRef(admindomain.ContentEntityType, contentID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, userID), ShouldContain, id)
			})
		})

		Convey("When the user has access only via a course containing the content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				courseID := testutil.InsertTestCourse(t, tx)
				contentID := testutil.InsertTestContentItem(t, tx)
				testutil.LinkContentItemToCourse(t, tx, courseID, contentID, 1, true)
				userID := testutil.InsertRandomTestUser(t, tx)
				testutil.GrantCourseAccess(t, tx, userID, courseID, testutil.AccessGrant{})
				et, eid := entityRef(admindomain.ContentEntityType, contentID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, userID), ShouldContain, id)
			})
		})

		Convey("When the user has no access to the content item", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{repository.BaseRepository{DB: tx}}
				contentID := testutil.InsertTestContentItem(t, tx)
				et, eid := entityRef(admindomain.ContentEntityType, contentID)
				id := seedPublicAnnouncement(t, ctx, tx, repo, publicSeed{entityType: et, entityID: eid})

				So(visiblePublicIDs(t, ctx, repo, testutil.InsertRandomTestUser(t, tx)), ShouldNotContain, id)
			})
		})
	})
}
