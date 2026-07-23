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
				deleteCourse: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteCourse(context.Background(), "course_ID")
			So(err, ShouldBeNil)
		})

		Convey("Error", func() {
			cRepo := &mockCourseRepoRepo{
				deleteCourse: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.DeleteCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
