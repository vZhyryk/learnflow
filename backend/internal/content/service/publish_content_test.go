package contentservice

import (
	"context"
	"errors"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func validGetContentItemByID(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
	description := "description"
	thumbnailURL := "https://example.com/thumbnail.jpg"
	videoURL := "https://example.com/video.mp4"
	seoTitle := "SeoTitle"
	seoDescription := "SeoDescription"
	return &contentdomain.ContentItem{
		Status:         contentdomain.DraftStatus,
		Title:          "title",
		ContentType:    contentdomain.VideoContent,
		Description:    &description,
		ThumbnailURL:   &thumbnailURL,
		VideoURL:       &videoURL,
		SeoTitle:       &seoTitle,
		SeoDescription: &seoDescription,
	}, nil
}

func TestPublishContentItem(t *testing.T) {
	Convey("PublishContentItem", t, func() {
		Convey("PublishContentItem - getContentItemByID error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: alwaysError,
			}

			srv := newTestService(cRepo)
			err := srv.PublishContentItem(context.Background(), "content_item_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("PublishContentItem - wrong status", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{Status: contentdomain.ArchivedStatus}, nil
				},
			}

			srv := newTestService(cRepo)
			err := srv.PublishContentItem(context.Background(), "content_item_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, contentdomain.ErrInvalidContentItemStatus), ShouldBeTrue)
		})

		Convey("PublishContentItem - not ready to publish", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{Status: contentdomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo)
			err := srv.PublishContentItem(context.Background(), "content_item_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.ReadyToPublish")
		})

		Convey("PublishContentItem - publish error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: validGetContentItemByID,
				publishContentItem: testutil.AlwaysFailsDB2,
			}

			srv := newTestService(cRepo)
			err := srv.PublishContentItem(context.Background(), "content_item_ID", "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "service.PublishContentItem")
		})

		Convey("Successful", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: validGetContentItemByID,
				publishContentItem: testutil.AlwaysNil2,
			}

			srv := newTestService(cRepo)
			err := srv.PublishContentItem(context.Background(), "content_item_ID", "user-1")
			So(err, ShouldBeNil)
		})
	})
}
