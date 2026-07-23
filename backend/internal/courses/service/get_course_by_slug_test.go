package courseservice

import (
	"context"
	"errors"
	coursedomain "learnflow_backend/internal/courses/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCourseBySlug(t *testing.T) {
	Convey("Given a course service", t, func() {
		Convey("When the course is published", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{Status: coursedomain.PublishedStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			course, err := srv.GetCourseBySlug(context.Background(), "course_slug")
			So(err, ShouldBeNil)
			So(course, ShouldNotBeNil)
			So(course.Status, ShouldEqual, coursedomain.PublishedStatus)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: AlwaysError,
			}

			srv := newTestService(cRepo, nil)
			course, err := srv.GetCourseBySlug(context.Background(), "course_slug")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(course, ShouldBeNil)
		})

		Convey("When the course is not published", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{Status: coursedomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			course, err := srv.GetCourseBySlug(context.Background(), "course_slug")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, coursedomain.ErrCourseNotFound), ShouldBeTrue)
			So(course, ShouldBeNil)
		})
	})
}
