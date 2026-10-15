package courseservice

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchiveCourse(t *testing.T) {
	Convey("Given a course service", t, func() {
		Convey("When ArchiveCourse succeeds", func() {
			cRepo := &mockCourseRepoRepo{
				archiveCourse: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo)
			err := srv.ArchiveCourse(context.Background(), "courseID", "user-1")
			So(err, ShouldBeNil)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockCourseRepoRepo{
				archiveCourse: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo)
			err := srv.ArchiveCourse(context.Background(), "courseID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
