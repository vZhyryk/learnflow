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
				archiveCourse: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveCourse(context.Background(), "courseID")
			So(err, ShouldBeNil)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockCourseRepoRepo{
				archiveCourse: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.ArchiveCourse(context.Background(), "courseID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})
	})
}
