package courseservice

import (
	"context"
	"errors"
	coursedomain "learnflow_backend/internal/courses/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getValidCourse(_ context.Context, _ string) (*coursedomain.Course, error) {
	return &coursedomain.Course{
		Slug: "old slug",
		ID:   "course_ID",
	}, nil
}

// UpdateCourse patches an existing course, checking any new slug is not already in use.
func TestUpdateCourse(t *testing.T) {
	Convey("UpdateCourse Course", t, func() {
		Convey("UpdateCourse - GetCourseByID error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID"}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateCourse - getCourseBySlug error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateCourse - Same Slug different ID error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: getValidCourse,
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{Slug: "old slug", ID: "course_ID_2"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("UpdateCourse - updateCourse ID Match error ", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: getValidCourse,
				updateCourse:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateCourse - nil slug skips uniqueness check, update failure propagates", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: getValidCourse,
				updateCourse:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: nil}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("Success", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: getValidCourse,
				updateCourse:    alwaysSucceedsUpdate,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: nil}, "user-1")
			So(err, ShouldBeNil)
		})
	})
}
