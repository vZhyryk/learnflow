package contentservice

import (
	"context"
	"errors"
	contentdomain "learnflow_backend/internal/content/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetContentItemBySlug(t *testing.T) {
	Convey("Given a contentItem service", t, func() {
		Convey("When the contentItem is published", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{Status: contentdomain.PublishedStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			contentItem, err := srv.GetContentItemBySlug(context.Background(), "contentItem_slug")
			So(err, ShouldBeNil)
			So(contentItem, ShouldNotBeNil)
			So(contentItem.Status, ShouldEqual, contentdomain.PublishedStatus)
		})

		Convey("When the repository returns an error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			contentItem, err := srv.GetContentItemBySlug(context.Background(), "contentItem_slug")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(contentItem, ShouldBeNil)
		})

		Convey("When the contentItem is not published", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{Status: contentdomain.DraftStatus}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			contentItem, err := srv.GetContentItemBySlug(context.Background(), "contentItem_slug")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, contentdomain.ErrContentItemNotFound), ShouldBeTrue)
			So(contentItem, ShouldBeNil)
		})
	})
}
