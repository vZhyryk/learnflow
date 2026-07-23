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
				getCourseByID: AlwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateCourse - getCourseBySlug error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: AlwaysError,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug})
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
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("UpdateCourse - updateCourse ID Match error ", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: getValidCourse,
				updateCourse:    AlwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: &slug})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateCourse - updateCourse empty slug error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID:   getValidCourse,
				getCourseBySlug: getValidCourse,
				updateCourse:    AlwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: nil})
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
			err := srv.UpdateCourse(context.Background(), coursedomain.UpdateCourseRequest{ID: "course_ID", Slug: nil})
			So(err, ShouldBeNil)
		})
	})
}
