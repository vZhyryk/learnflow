package contentservice

import (
	"context"
	"errors"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateContentItem(t *testing.T) {
	Convey("Create ContentItem", t, func() {
		Convey("GetContentItemBySlug error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: alwaysError,
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateContentItem(context.Background(), contentdomain.CreateContentItemRequest{})
			So(err, ShouldNotBeNil)
			So(id, ShouldBeEmpty)
		})

		Convey("GetContentItemBySlug already exists", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateContentItem(context.Background(), contentdomain.CreateContentItemRequest{})
			So(err, ShouldNotBeNil)
			So(errors.Is(err, contentdomain.ErrInvalidSlug), ShouldBeTrue)
			So(id, ShouldBeEmpty)
		})

		Convey("CreateContentItem error", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return nil, nil
				},
				createContentItem: func(_ context.Context, _ *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
					return nil, testutil.ErrDBUnexpected
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateContentItem(context.Background(), contentdomain.CreateContentItemRequest{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(id, ShouldBeEmpty)
		})

		Convey("Success", func() {
			cRepo := &mockContentItemRepo{
				getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
					return nil, nil
				},
				createContentItem: func(_ context.Context, _ *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
					return &contentdomain.ContentItem{ID: "content_item_ID"}, nil
				},
			}

			srv := newTestService(cRepo, nil)
			id, err := srv.CreateContentItem(context.Background(), contentdomain.CreateContentItemRequest{})
			So(err, ShouldBeNil)
			So(id, ShouldEqual, "content_item_ID")
		})
	})
}

func TestCreateContentItemIsIndexableOverride(t *testing.T) {
	Convey("Create ContentItem with IsIndexable explicitly set to false", t, func() {
		var gotContentItem *contentdomain.ContentItem
		cRepo := &mockContentItemRepo{
			getContentItemBySlug: func(_ context.Context, _ string) (*contentdomain.ContentItem, error) {
				return nil, nil
			},
			createContentItem: func(_ context.Context, contentItem *contentdomain.ContentItem) (*contentdomain.ContentItem, error) {
				gotContentItem = contentItem
				return &contentdomain.ContentItem{ID: "content_item_ID"}, nil
			},
		}

		srv := newTestService(cRepo, nil)
		isIndexable := false
		id, err := srv.CreateContentItem(context.Background(), contentdomain.CreateContentItemRequest{IsIndexable: &isIndexable})
		So(err, ShouldBeNil)
		So(id, ShouldEqual, "content_item_ID")
		So(gotContentItem.IsIndexable, ShouldBeFalse)
	})
}
