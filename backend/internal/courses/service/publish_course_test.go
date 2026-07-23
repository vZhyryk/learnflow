package courseservice

import (
	"context"
	"errors"
	coursedomain "learnflow_backend/internal/courses/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func validGetCourseByID(_ context.Context, _ string) (*coursedomain.Course, error) {
	des := "description"
	ThumbnailURL := "ThumbnailURL"
	PreviewVideoURL := "PreviewVideoURL"
	SeoTitle := "SeoTitle"
	SeoDescription := "SeoDescription"
	return &coursedomain.Course{Status: coursedomain.DraftStatus, Title: "title", Description: &des, ThumbnailURL: &ThumbnailURL, PreviewVideoURL: &PreviewVideoURL, SeoTitle: &SeoTitle, SeoDescription: &SeoDescription}, nil
}

func TestPublishCourse(t *testing.T) {
	Convey("PublishCourse Course", t, func() {
		Convey("PublishCourse - GetCourseByID error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: AlwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("PublishCourse - wrong status", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{Status: coursedomain.ArchivedStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, coursedomain.ErrInvalidCourseStatus), ShouldBeTrue)
		})

		Convey("PublishCourse - not ready to publish", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: func(_ context.Context, _ string) (*coursedomain.Course, error) {
					return &coursedomain.Course{Status: coursedomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.ReadyToPublish")
		})

		Convey("PublishCourse - publish error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: validGetCourseByID,
				publishCourse: testutil.AlwaysFailsDB,
			}

			srv := newTestService(cRepo, nil)
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.PublishCourse")
		})

		Convey("PublishCourse - outbox emit error", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: validGetCourseByID,
				publishCourse: testutil.AlwaysNil,
			}

			srv := newTestService(cRepo, testutil.NewFailingOutbox(testutil.ErrDBUnexpected))
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, testutil.ErrDBUnexpected), ShouldBeTrue)
		})

		Convey("Successful", func() {
			cRepo := &mockCourseRepoRepo{
				getCourseByID: validGetCourseByID,
				publishCourse: testutil.AlwaysNil,
			}
			var captured []any

			srv := newTestService(cRepo, testutil.NewCapturingOutbox(&captured))
			err := srv.PublishCourse(context.Background(), "course_ID")
			So(err, ShouldBeNil)
		})
	})
}
