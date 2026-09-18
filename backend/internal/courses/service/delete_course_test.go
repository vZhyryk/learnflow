package courseservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeleteCourse(t *testing.T) {
	Convey("DeleteCourse Course", t, func() {
		Convey("Success", func() {
			cRepo := &mockCourseRepoRepo{
				deleteCourse: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo)
			err := srv.DeleteCourse(context.Background(), "course_ID", "user-1")
			So(err, ShouldBeNil)
		})

		Convey("Error", func() {
			cRepo := &mockCourseRepoRepo{
				deleteCourse: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo)
			err := srv.DeleteCourse(context.Background(), "course_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
