package courseservice

import (
	"context"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getCoursesError(_ context.Context, _ pagination.Params) ([]*coursedomain.Course, error) {
	return nil, testutil.ErrDBUnexpected
}

func getValidList(_ context.Context, _ pagination.Params) ([]*coursedomain.Course, error) {
	return []*coursedomain.Course{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}}, nil
}

// testGetAllCoursesByType covers the shared shape of GetAllCourses for a given status:
// wire the one mock field the service dispatches to for that status, then assert the
// error and success paths. Shared here instead of writing the same four Convey blocks
// four times over (mirrors testExecCourseMethod/testCourseListMethod in the repository
// layer's own test file).
func testGetAllCoursesByType(t *testing.T, scenario string, status coursedomain.CourseStatus, wireMock func(*mockCourseRepoRepo, func(context.Context, pagination.Params) ([]*coursedomain.Course, error))) {
	Convey("GetAllCourses Course", t, func() {
		Convey(scenario+" - error", func() {
			cRepo := &mockCourseRepoRepo{}
			wireMock(cRepo, getCoursesError)

			srv := newTestService(cRepo, nil)
			course, err := srv.GetAllCourses(context.Background(), status, pagination.Params{})
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(course, ShouldBeNil)
		})

		Convey(scenario+" - success", func() {
			cRepo := &mockCourseRepoRepo{}
			wireMock(cRepo, getValidList)

			srv := newTestService(cRepo, nil)
			courses, err := srv.GetAllCourses(context.Background(), status, pagination.Params{})
			So(err, ShouldBeNil)
			So(courses, ShouldNotBeNil)
			So(courses, ShouldNotBeEmpty)
			So(len(courses), ShouldEqual, 4)
		})
	})
}

func TestGetAllCoursesArchived(t *testing.T) {
	testGetAllCoursesByType(t, "GetAllArchivedCourses", coursedomain.ArchivedStatus, func(r *mockCourseRepoRepo, fn func(context.Context, pagination.Params) ([]*coursedomain.Course, error)) {
		r.getAllArchivedCourses = fn
	})
}

func TestGetAllCoursesPublished(t *testing.T) {
	testGetAllCoursesByType(t, "GetAllPublishedCourses", coursedomain.PublishedStatus, func(r *mockCourseRepoRepo, fn func(context.Context, pagination.Params) ([]*coursedomain.Course, error)) {
		r.getAllPublishedCourses = fn
	})
}

func TestGetAllCoursesDraft(t *testing.T) {
	testGetAllCoursesByType(t, "GetAllDraftCourses", coursedomain.DraftStatus, func(r *mockCourseRepoRepo, fn func(context.Context, pagination.Params) ([]*coursedomain.Course, error)) {
		r.getAllDraftCourses = fn
	})
}

func TestGetAllCoursesDefault(t *testing.T) {
	testGetAllCoursesByType(t, "default", coursedomain.CourseStatus(""), func(r *mockCourseRepoRepo, fn func(context.Context, pagination.Params) ([]*coursedomain.Course, error)) {
		r.getAllCourses = fn
	})
}
