package contentservice

import (
	"context"
	"errors"
	contentdomain "learnflow_backend/internal/content/domain"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getValidContentItem(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
	return &contentdomain.ContentItem{
		Slug: "old slug",
		ID:   "content_item_id",
	}, nil
}

func TestUpdateContentItem(t *testing.T) {
	Convey("UpdateContentItem", t, func() {
		Convey("UpdateContentItem - GetContentItemByID error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id"}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateContentItem - getContentItemBySlug error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID:   getValidContentItem,
				getContentItemBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateContentItem - Same Slug different ID error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID: getValidContentItem,
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{Slug: "old slug", ID: "content_item_ID_2"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
		})

		Convey("UpdateContentItem - updateContentItem ID Match error ", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID:   getValidContentItem,
				getContentItemBySlug: getValidContentItem,
				updateContentItem:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			slug := "New Slug"
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id", Slug: &slug}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("UpdateContentItem - nil slug skips uniqueness check, update failure propagates", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID:   getValidContentItem,
				getContentItemBySlug: getValidContentItem,
				updateContentItem:    alwaysFailsErr,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id", Slug: nil}, "user-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
		})

		Convey("Success", func() {
			cRepo := &mockContentItemRepo{
				getContentItemByID:   getValidContentItem,
				getContentItemBySlug: getValidContentItem,
				updateContentItem:    alwaysSucceedsUpdate,
			}

			srv := newTestService(cRepo, nil)
			err := srv.UpdateContentItem(context.Background(), contentdomain.UpdateContentItemRequest{ID: "content_item_id", Slug: nil}, "user-1")
			So(err, ShouldBeNil)
		})
	})
}
