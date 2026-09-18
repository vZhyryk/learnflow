package contentservice

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getContentItemError(_ context.Context, _ pagination.Params) ([]*contentdomain.ContentItem, error) {
	return nil, testutil.ErrDBUnexpected
}

func getValidList(_ context.Context, _ pagination.Params) ([]*contentdomain.ContentItem, error) {
	return []*contentdomain.ContentItem{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}}, nil
}

func testGetAllContentItemsByType(t *testing.T, scenario string, status contentdomain.ContentItemStatus, wireMock func(*mockContentItemRepo, func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error))) {
	Convey("GetAllContentItems contentItem", t, func() {
		Convey(scenario+" - error", func() {
			cRepo := &mockContentItemRepo{}
			wireMock(cRepo, getContentItemError)

			srv := newTestService(cRepo)
			contentItem, err := srv.GetAllContentItems(context.Background(), status, pagination.Params{})
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(contentItem, ShouldBeNil)
		})

		Convey(scenario+" - success", func() {
			cRepo := &mockContentItemRepo{}
			wireMock(cRepo, getValidList)

			srv := newTestService(cRepo)
			ContentItem, err := srv.GetAllContentItems(context.Background(), status, pagination.Params{})
			So(err, ShouldBeNil)
			So(ContentItem, ShouldNotBeNil)
			So(ContentItem, ShouldNotBeEmpty)
			So(len(ContentItem), ShouldEqual, 4)
		})
	})
}

func TestGetAllContentItemsArchived(t *testing.T) {
	testGetAllContentItemsByType(t, "getAllArchivedContentItems", contentdomain.ArchivedStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error)) {
		r.getAllArchivedContentItems = fn
	})
}

func TestGetAllContentItemsPublished(t *testing.T) {
	testGetAllContentItemsByType(t, "GetAllPublishedContentItems", contentdomain.PublishedStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error)) {
		r.getAllPublishedContentItems = fn
	})
}

func TestGetAllContentItemsDraft(t *testing.T) {
	testGetAllContentItemsByType(t, "GetAllDraftContentItems", contentdomain.DraftStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error)) {
		r.getAllDraftContentItems = fn
	})
}

func TestGetAllContentItemsDefault(t *testing.T) {
	testGetAllContentItemsByType(t, "default", contentdomain.ContentItemStatus(""), func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, error)) {
		r.getAllContentItems = fn
	})
}
