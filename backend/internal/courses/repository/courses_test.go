package courserepository

import (
	"context"
	"errors"
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

// testExecCourseMethod covers the shared shape of PublishCourse/ArchiveCourse/DeleteCourse:
// Exec, map 0 rows affected to ErrCourseNotFound, wrap any other error. Shared here instead
// of writing the same three Convey blocks three times over.
func testExecCourseMethod(t *testing.T, methodName string, call func(*Repository, context.Context, string) error) {
	Convey("Given a course repository", t, func() {
		var execTag pgconn.CommandTag
		var execErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
				return execTag, execErr
			},
		})

		Convey("When it succeeds", func() {
			execTag = pgconn.NewCommandTag("UPDATE 1")
			So(call(repo, context.Background(), "course-123"), ShouldBeNil)
		})

		Convey("When no row is matched (course not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := call(repo, context.Background(), "unknown")
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := call(repo, context.Background(), "course-123")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})
	})
}

func TestPublishCourse(t *testing.T) {
	testExecCourseMethod(t, "PublishCourse", (*Repository).PublishCourse)
}

func TestArchiveCourse(t *testing.T) {
	testExecCourseMethod(t, "ArchiveCourse", (*Repository).ArchiveCourse)
}

func TestDeleteCourse(t *testing.T) {
	testExecCourseMethod(t, "DeleteCourse", (*Repository).DeleteCourse)
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
			So(repo.UpdateCourse(context.Background(), course), ShouldBeNil)
		})

		Convey("When no row is matched (course not found)", func() {
			execTag = pgconn.NewCommandTag("UPDATE 0")
			err := repo.UpdateCourse(context.Background(), course)
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
		})

		Convey("When the new slug is already taken (pg 23505)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "courses_slug_unique"}
			err := repo.UpdateCourse(context.Background(), course)
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("When an unrelated unique violation occurs (pg 23505, different constraint)", func() {
			execErr = &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
			err := repo.UpdateCourse(context.Background(), course)
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeFalse)
			So(err, ShouldNotBeNil)
		})

		Convey("When the database returns an unexpected error", func() {
			execErr = testutil.ErrDBUnexpected
			err := repo.UpdateCourse(context.Background(), course)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository.UpdateCourse")
		})
	})
}

// testCourseListMethod covers the shared shape of every GetAll*Courses method
// (getAndParseCourses: query -> scan loop -> rows.Err), per go-testing.md's
// coverage-per-branch rule, without repeating all four branches per method.
func testCourseListMethod(t *testing.T, methodName string, call func(*Repository, context.Context, pagination.Params) ([]*coursedomain.Course, error)) {
	now := time.Now().UTC().Truncate(time.Second)

	Convey("Given a course repository", t, func() {
		var rows *testutil.MockRows
		var queryErr error
		repo := newTestRepo(&testutil.MockQueryRunner{
			QueryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
				return rows, queryErr
			},
		})

		Convey("When the query fails", func() {
			queryErr = testutil.ErrDBUnexpected
			_, err := call(repo, context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "repository."+methodName)
		})

		Convey("When a row fails to scan", func() {
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: func(_ ...any) error { return testutil.ErrDBUnexpected }},
			}}
			_, err := call(repo, context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan")
		})

		Convey("When rows.Err() reports a failure after iteration", func() {
			rows = &testutil.MockRows{RowsErr: testutil.ErrDBUnexpected}
			_, err := call(repo, context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "rows")
		})

		Convey("When rows return 2 courses", func() {
			course1, course2 := fakeCourse(now), fakeCourse(now)
			course1.ID, course2.ID = "course-1", "course-2"
			rows = &testutil.MockRows{Rows: []*testutil.MockRow{
				{ScanFn: fakeCourseScan(course1)},
				{ScanFn: fakeCourseScan(course2)},
			}}
			got, err := call(repo, context.Background(), pagination.NewParams(1, 20))
			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got[0], ShouldResemble, course1)
			So(got[1], ShouldResemble, course2)
		})
	})
}

func TestGetAllPublishedCourses(t *testing.T) {
	testCourseListMethod(t, "GetAllPublishedCourses", (*Repository).GetAllPublishedCourses)
}

func TestGetAllDraftCourses(t *testing.T) {
	testCourseListMethod(t, "GetAllDraftCourses", (*Repository).GetAllDraftCourses)
}

func TestGetAllArchivedCourses(t *testing.T) {
	testCourseListMethod(t, "GetAllArchivedCourses", (*Repository).GetAllArchivedCourses)
}

func TestGetAllCourses(t *testing.T) {
	testCourseListMethod(t, "GetAllCourses", (*Repository).GetAllCourses)
}
