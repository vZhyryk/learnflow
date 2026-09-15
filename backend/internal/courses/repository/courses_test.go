package courserepository

import (
	"context"
	"errors"
	"fmt"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRepository(t *testing.T) {
	Convey("Given a nil connection pool", t, func() {
		Convey("NewRepository returns a non-nil Repository", func() {
			repo := NewRepository(nil)
			So(repo, ShouldNotBeNil)
		})
	})
}

func TestCreateCourse(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a course repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When creation succeeds", func() {
			expected := fakeCourse(now)
			row = &testutil.MockRow{ScanFn: fakeCourseScan(expected)}
			got, err := repo.CreateCourse(context.Background(), &coursedomain.Course{
				Slug: expected.Slug, Title: expected.Title, CreatedByUserID: expected.CreatedByUserID,
			})
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the slug is already taken (pg 23505)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "courses_slug_unique"}
			}}
			_, err := repo.CreateCourse(context.Background(), &coursedomain.Course{})
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error {
				return &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			}}
			_, err := repo.CreateCourse(context.Background(), &coursedomain.Course{})
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDB }}
			_, err := repo.CreateCourse(context.Background(), &coursedomain.Course{})
			testutil.AssertUnexpectedDBError(err, "db error")
		})
	})
}

func TestGetCourseByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a course repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the course exists", func() {
			expected := fakeCourse(now)
			row = &testutil.MockRow{ScanFn: fakeCourseScan(expected)}
			got, err := repo.GetCourseByID(context.Background(), "course-123")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the course does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetCourseByID(context.Background(), "unknown")
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetCourseByID(context.Background(), "course-123")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

func TestGetCourseBySlug(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a course repository", t, func() {
		var row *testutil.MockRow
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
				return row
			},
		})

		Convey("When the course exists", func() {
			expected := fakeCourse(now)
			row = &testutil.MockRow{ScanFn: fakeCourseScan(expected)}
			got, err := repo.GetCourseBySlug(context.Background(), "some-slug")
			So(err, ShouldBeNil)
			So(got, ShouldResemble, expected)
		})

		Convey("When the course does not exist", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return pgx.ErrNoRows }}
			_, err := repo.GetCourseBySlug(context.Background(), "unknown")
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			row = &testutil.MockRow{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }}
			_, err := repo.GetCourseBySlug(context.Background(), "some-slug")
			testutil.AssertUnexpectedDBError(err, "db connection lost")
		})
	})
}

// bindCourseExec adapts a repository exec method into the shape testutil.TestExecMethod
// expects: build a repo around the given runner, return the bound method.
func bindCourseExec(
	call func(*Repository, context.Context, string, string) error,
) func(*testutil.MockQueryRunner) func(context.Context, string, string) error {
	return func(runner *testutil.MockQueryRunner) func(context.Context, string, string) error {
		repo := newTestRepo(runner)
		return func(ctx context.Context, id, userID string) error {
			return call(repo, ctx, id, userID)
		}
	}
}

func TestPublishCourse(t *testing.T) {
	testutil.TestExecMethod(t, "PublishCourse", bindCourseExec((*Repository).PublishCourse), coursedomain.ErrCourseNotFound)
}

func TestArchiveCourse(t *testing.T) {
	testutil.TestExecMethod(t, "ArchiveCourse", bindCourseExec((*Repository).ArchiveCourse), coursedomain.ErrCourseNotFound)
}

func TestDeleteCourse(t *testing.T) {
	testutil.TestExecMethod(t, "DeleteCourse", bindCourseExec((*Repository).DeleteCourse), coursedomain.ErrCourseNotFound)
}

func TestUpdateCourse(t *testing.T) {
	Convey("Given a course repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})
		course := &coursedomain.Course{ID: "course-123"}

		Convey("When update succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(repo.UpdateCourse(context.Background(), course, "user-1"), ShouldBeNil)
		})

		Convey("When no row is matched (course not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateCourse(context.Background(), course, "user-1")
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
		})

		Convey("When the new slug is already taken (pg 23505)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "courses_slug_unique"}
			err := repo.UpdateCourse(context.Background(), course, "user-1")
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			err := repo.UpdateCourse(context.Background(), course, "user-1")
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateCourse(context.Background(), course, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateCourse")
		})
	})
}

// bindCourseList adapts a repository list method into the shape testutil.TestListMethod
// expects: build a repo around the given runner, return the bound method.
func bindCourseList(
	call func(*Repository, context.Context, pagination.Params) ([]*coursedomain.Course, error),
) func(*testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*coursedomain.Course, error) {
	return func(runner *testutil.MockQueryRunner) func(context.Context, pagination.Params) ([]*coursedomain.Course, error) {
		repo := newTestRepo(runner)
		return func(ctx context.Context, params pagination.Params) ([]*coursedomain.Course, error) {
			return call(repo, ctx, params)
		}
	}
}

// fakeCourseN builds the nth fake Course for TestListMethod's "2 items" case.
func fakeCourseN(n int) *coursedomain.Course {
	course := fakeCourse(time.Now().UTC().Truncate(time.Second))
	course.ID = fmt.Sprintf("course-%d", n)
	return course
}

func TestGetAllPublishedCourses(t *testing.T) {
	testutil.TestListMethod(t, "GetAllPublishedCourses",
		bindCourseList((*Repository).GetAllPublishedCourses), fakeCourseN, fakeCourseScan)
}

func TestGetAllDraftCourses(t *testing.T) {
	testutil.TestListMethod(t, "GetAllDraftCourses",
		bindCourseList((*Repository).GetAllDraftCourses), fakeCourseN, fakeCourseScan)
}

func TestGetAllArchivedCourses(t *testing.T) {
	testutil.TestListMethod(t, "GetAllArchivedCourses",
		bindCourseList((*Repository).GetAllArchivedCourses), fakeCourseN, fakeCourseScan)
}

func TestGetAllCourses(t *testing.T) {
	testutil.TestListMethod(t, "GetAllCourses",
		bindCourseList((*Repository).GetAllCourses), fakeCourseN, fakeCourseScan)
}
