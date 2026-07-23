//go:build integration

package courserepository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/smartystreets/goconvey/convey"
)

func randomTestSlug(t *testing.T) string {
	t.Helper()

	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("randomTestSlug: %v", err)
	}
	return fmt.Sprintf("courses-repo-integration-%s", hex.EncodeToString(buf))
}

func insertTestUser(t *testing.T, tx pgx.Tx) string {
	t.Helper()
	return testutil.InsertTestUser(t, tx, testutil.RandomTestEmail(t, "courses-repo-integration"))
}

// draftCourse returns a Course seed with every field populated, ready for CreateCourse —
// mirrors the shape a real CreateCourseRequest would produce after Apply/validation.
func draftCourse(t *testing.T, tx pgx.Tx) *coursedomain.Course {
	t.Helper()

	description := "A thorough introduction to the subject."
	thumbnailURL := "https://example.com/thumb.png"
	previewVideoURL := "https://example.com/preview.mp4"
	estimatedMinutes := 90
	seoTitle := "SEO Title"
	seoDescription := "SEO Description"
	ogImageURL := "https://example.com/og.png"
	canonicalURL := "https://example.com/courses/slug"

	return &coursedomain.Course{
		Slug:             randomTestSlug(t),
		Title:            "Integration Test Course",
		Description:      &description,
		ThumbnailURL:     &thumbnailURL,
		PreviewVideoURL:  &previewVideoURL,
		EstimatedMinutes: &estimatedMinutes,
		SeoTitle:         &seoTitle,
		SeoDescription:   &seoDescription,
		OgImageURL:       &ogImageURL,
		CanonicalURL:     &canonicalURL,
		IsIndexable:      true,
		CreatedByUserID:  insertTestUser(t, tx),
	}
}

func TestCreateCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When creating a course with all fields populated", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)

				got, err := repo.CreateCourse(ctx, seed)

				So(err, ShouldBeNil)
				So(got.ID, ShouldNotBeEmpty)
				So(got.Slug, ShouldEqual, seed.Slug)
				So(got.Status, ShouldEqual, coursedomain.DraftStatus)
				So(got.IsIndexable, ShouldBeTrue)
				So(got.CreatedByUserID, ShouldEqual, seed.CreatedByUserID)
				So(got.CreatedAt.IsZero(), ShouldBeFalse)
				So(got.PublishedAt, ShouldBeNil)
				So(got.DeletedAt, ShouldBeNil)
			})
		})

		Convey("When the slug is already taken by another course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				_, err := repo.CreateCourse(ctx, seed)
				So(err, ShouldBeNil)

				dup := draftCourse(t, tx)
				dup.Slug = seed.Slug

				_, err = repo.CreateCourse(ctx, dup)

				So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When created_by_user_id does not reference an existing user", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				seed.CreatedByUserID = "00000000-0000-0000-0000-000000000000"

				_, err := repo.CreateCourse(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503") // foreign_key_violation
			})
		})

		Convey("When title is blank (DB-level CHECK, defense-in-depth below domain validation)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				seed.Title = "   "

				_, err := repo.CreateCourse(ctx, seed)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23514") // check_violation
				So(pgErr.ConstraintName, ShouldEqual, "courses_title_nonempty")
			})
		})
	})
}

func TestGetCourseByID_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When the course exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				created, err := repo.CreateCourse(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetCourseByID(ctx, created.ID)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
				So(got.Slug, ShouldEqual, created.Slug)
			})
		})

		Convey("When no course exists for the given ID", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				_, err := repo.GetCourseByID(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})

		Convey("When the course is soft-deleted", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				created, err := repo.CreateCourse(ctx, seed)
				So(err, ShouldBeNil)
				So(repo.DeleteCourse(ctx, created.ID), ShouldBeNil)

				_, err = repo.GetCourseByID(ctx, created.ID)

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetCourseBySlug_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When the course exists", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				seed := draftCourse(t, tx)
				created, err := repo.CreateCourse(ctx, seed)
				So(err, ShouldBeNil)

				got, err := repo.GetCourseBySlug(ctx, created.Slug)

				So(err, ShouldBeNil)
				So(got.ID, ShouldEqual, created.ID)
			})
		})

		Convey("When no course exists for the given slug", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				_, err := repo.GetCourseBySlug(ctx, "does-not-exist")

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestPublishCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When publishing an existing draft course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)

				So(repo.PublishCourse(ctx, created.ID), ShouldBeNil)

				got, err := repo.GetCourseByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, coursedomain.PublishedStatus)
				So(got.PublishedAt, ShouldNotBeNil)
			})
		})

		Convey("When the course does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				err := repo.PublishCourse(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestArchiveCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When archiving an existing draft course directly (no publish step)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)

				So(repo.ArchiveCourse(ctx, created.ID), ShouldBeNil)

				got, err := repo.GetCourseByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Status, ShouldEqual, coursedomain.ArchivedStatus)
			})
		})

		Convey("When the course does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}

				err := repo.ArchiveCourse(ctx, "00000000-0000-0000-0000-000000000000")

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDeleteCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When soft-deleting an existing course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)

				So(repo.DeleteCourse(ctx, created.ID), ShouldBeNil)

				_, err = repo.GetCourseByID(ctx, created.ID)
				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})

		Convey("When the course is already deleted (second delete affects 0 rows)", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)
				So(repo.DeleteCourse(ctx, created.ID), ShouldBeNil)

				err = repo.DeleteCourse(ctx, created.ID)

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpdateCourse_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a courses repository backed by real Postgres", t, func() {
		Convey("When updating every field of an existing course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)

				newDescription := "Updated description"
				created.Slug = randomTestSlug(t)
				created.Title = "Updated Title"
				created.Description = &newDescription

				err = repo.UpdateCourse(ctx, created)
				So(err, ShouldBeNil)

				got, err := repo.GetCourseByID(ctx, created.ID)
				So(err, ShouldBeNil)
				So(got.Slug, ShouldEqual, created.Slug)
				So(got.Title, ShouldEqual, "Updated Title")
				So(*got.Description, ShouldEqual, newDescription)
			})
		})

		Convey("When the new slug collides with another course", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				other, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)
				created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
				So(err, ShouldBeNil)

				created.Slug = other.Slug

				err = repo.UpdateCourse(ctx, created)

				So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
			})
		})

		Convey("When the course does not exist", func() {
			testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
				repo := &Repository{db: tx}
				ghost := draftCourse(t, tx)
				ghost.ID = "00000000-0000-0000-0000-000000000000"

				err := repo.UpdateCourse(ctx, ghost)

				So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestGetAllCoursesByStatus_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given draft, published, archived, and soft-deleted courses", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{db: tx}

			draft, err := repo.CreateCourse(ctx, draftCourse(t, tx))
			So(err, ShouldBeNil)

			published, err := repo.CreateCourse(ctx, draftCourse(t, tx))
			So(err, ShouldBeNil)
			So(repo.PublishCourse(ctx, published.ID), ShouldBeNil)

			archived, err := repo.CreateCourse(ctx, draftCourse(t, tx))
			So(err, ShouldBeNil)
			So(repo.ArchiveCourse(ctx, archived.ID), ShouldBeNil)

			deleted, err := repo.CreateCourse(ctx, draftCourse(t, tx))
			So(err, ShouldBeNil)
			So(repo.DeleteCourse(ctx, deleted.ID), ShouldBeNil)

			params := pagination.NewParams(1, 100)

			Convey("GetAllDraftCourses returns only the draft, excluding soft-deleted", func() {
				got, err := repo.GetAllDraftCourses(ctx, params)
				So(err, ShouldBeNil)
				ids := courseIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldNotContain, deleted.ID)
			})

			Convey("GetAllPublishedCourses returns only the published course", func() {
				got, err := repo.GetAllPublishedCourses(ctx, params)
				So(err, ShouldBeNil)
				ids := courseIDs(got)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldNotContain, draft.ID)
				So(ids, ShouldNotContain, archived.ID)
			})

			Convey("GetAllArchivedCourses includes the soft-deleted archived course too", func() {
				// GetAllArchivedCourses intentionally omits `deleted_at IS NULL` — it's an
				// admin "including soft-deleted ones" query per db-conventions.md.
				So(repo.DeleteCourse(ctx, archived.ID), ShouldBeNil)

				got, err := repo.GetAllArchivedCourses(ctx, params)
				So(err, ShouldBeNil)
				So(courseIDs(got), ShouldContain, archived.ID)
			})

			Convey("GetAllCourses returns every course regardless of status, including soft-deleted", func() {
				got, err := repo.GetAllCourses(ctx, params)
				So(err, ShouldBeNil)
				ids := courseIDs(got)
				So(ids, ShouldContain, draft.ID)
				So(ids, ShouldContain, published.ID)
				So(ids, ShouldContain, archived.ID)
				So(ids, ShouldContain, deleted.ID)
			})

			Convey("Pagination limits the returned page size", func() {
				got, err := repo.GetAllCourses(ctx, pagination.NewParams(1, 2))
				So(err, ShouldBeNil)
				So(len(got), ShouldEqual, 2)
			})
		})
	})
}

func courseIDs(courses []*coursedomain.Course) []string {
	ids := make([]string, 0, len(courses))
	for _, c := range courses {
		ids = append(ids, c.ID)
	}
	return ids
}

func TestCourseCreatedByUserForeignKeyRestrict_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)

	Convey("Given a course created by an existing user", t, func() {
		testutil.WithTestTx(t, pool, func(ctx context.Context, tx pgx.Tx) {
			repo := &Repository{db: tx}
			created, err := repo.CreateCourse(ctx, draftCourse(t, tx))
			So(err, ShouldBeNil)

			Convey("When hard-deleting the owning user row", func() {
				_, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, created.CreatedByUserID)

				So(err, ShouldNotBeNil)
				var pgErr *pgconn.PgError
				So(errors.As(err, &pgErr), ShouldBeTrue)
				So(pgErr.Code, ShouldEqual, "23503")
			})
		})
	})
}
