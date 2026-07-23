package courseservice

import (
	"context"
	"errors"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCourse(t *testing.T) {
	Convey("Create Course", t, func() {
		Convey("GetCourseBySlug error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: AlwaysError,
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateCourse(context.Background(), coursedomain.CreateCourseRequest{})
			So(err, ShouldNotBeNil)
			So(id, ShouldBeEmpty)
		})

		Convey("GetCourseBySlug already exists", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateCourse(context.Background(), coursedomain.CreateCourseRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, coursedomain.ErrInvalidSlug), ShouldBeTrue)
			So(id, ShouldBeEmpty)
		})

		Convey("CreateCourse error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return nil, nil
				},
				createCourse: func(_ context.Context, _ *coursedomain.Course) (*coursedomain.Course, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateCourse(context.Background(), coursedomain.CreateCourseRequest{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(id, ShouldBeEmpty)
		})

		Convey("Success", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return nil, nil
				},
				createCourse: func(_ context.Context, _ *coursedomain.Course) (*coursedomain.Course, error) {
					return &coursedomain.Course{ID: "course_ID"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateCourse(context.Background(), coursedomain.CreateCourseRequest{})
			So(err, ShouldBeNil)
			So(id, ShouldEqual, "course_ID")
		})
	})
}

func TestCreateCourseIsIndexableOverride(t *testing.T) {
	Convey("Create Course with IsIndexable explicitly set to false", t, func() {
		var gotCourse *coursedomain.Course
		cRepo := &mockCourseRepoRepo{
			getCourseBySlug: func(_ context.Context, _ string) (*coursedomain.Course, error) {
				return nil, nil
			},
			createCourse: func(_ context.Context, course *coursedomain.Course) (*coursedomain.Course, error) {
				gotCourse = course
				return &coursedomain.Course{ID: "course_ID"}, nil
			},
		}

		srv := newTestService(cRepo, nil)
		isIndexable := false
		id, err := srv.CreateCourse(context.Background(), coursedomain.CreateCourseRequest{IsIndexable: &isIndexable})
		So(err, ShouldBeNil)
		So(id, ShouldEqual, "course_ID")
		So(gotCourse.IsIndexable, ShouldBeFalse)
	})
}
