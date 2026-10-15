package contentservice

import (
	"context"
	contentdomain "learnflow_backend/internal/content/domain"
	"learnflow_backend/internal/shared/pagination"
	"learnflow_backend/internal/shared/testutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getContentItemError(_ context.Context, _ pagination.Params) ([]*contentdomain.ContentItem, int, error) {
	return nil, 0, testutil.ErrDBUnexpected
}

func getValidList(_ context.Context, _ pagination.Params) ([]*contentdomain.ContentItem, int, error) {
	return []*contentdomain.ContentItem{{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}}, 4, nil
}

func testGetAllContentItemsByType(t *testing.T, scenario string, status contentdomain.ContentItemStatus, wireMock func(*mockContentItemRepo, func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, int, error))) {
	Convey("GetAllContentItems contentItem", t, func() {
		Convey(scenario+" - error", func() {
			cRepo := &mockContentItemRepo{}
			wireMock(cRepo, getContentItemError)

			srv := newTestService(cRepo)
			contentItem, gotTotal, err := srv.GetAllContentItems(context.Background(), status, pagination.Params{})
			So(err.Error(), ShouldContainSubstring, "db connection lost")
			So(contentItem, ShouldBeNil)
			So(gotTotal, ShouldEqual, 0)
		})

		Convey(scenario+" - success", func() {
			cRepo := &mockContentItemRepo{}
			wireMock(cRepo, getValidList)

			srv := newTestService(cRepo)
			ContentItem, total, err := srv.GetAllContentItems(context.Background(), status, pagination.Params{})
			So(err, ShouldBeNil)
			So(ContentItem, ShouldNotBeNil)
			So(ContentItem, ShouldNotBeEmpty)
			So(len(ContentItem), ShouldEqual, 4)
			So(total, ShouldEqual, 4)
		})
	})
}

func TestGetAllContentItemsArchived(t *testing.T) {
	testGetAllContentItemsByType(t, "getAllArchivedContentItems", contentdomain.ArchivedStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, int, error)) {
		r.getAllArchivedContentItems = fn
	})
}

func TestGetAllContentItemsPublished(t *testing.T) {
	testGetAllContentItemsByType(t, "GetAllPublishedContentItems", contentdomain.PublishedStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, int, error)) {
		r.getAllPublishedContentItems = fn
	})
}

func TestGetAllContentItemsDraft(t *testing.T) {
	testGetAllContentItemsByType(t, "GetAllDraftContentItems", contentdomain.DraftStatus, func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, int, error)) {
		r.getAllDraftContentItems = fn
	})
}

func TestGetAllContentItemsDefault(t *testing.T) {
	testGetAllContentItemsByType(t, "default", contentdomain.ContentItemStatus(""), func(r *mockContentItemRepo, fn func(context.Context, pagination.Params) ([]*contentdomain.ContentItem, int, error)) {
		r.getAllContentItems = fn
	})
}
